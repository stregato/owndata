package safe

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"fmt"
	"hash"
	"io/fs"
	"math/rand"
	"os"
	"path"
	"sort"
	"strconv"

	"github.com/stregato/mio/core"
	"github.com/stregato/mio/security"
	"github.com/stregato/mio/storage"
	"gopkg.in/yaml.v2"
)

type GroupName string
type Groups map[GroupName]core.Set[security.UserId]
type Endorsers core.Set[security.UserId]

const (
	UserGroup                   GroupName = "usr"
	AdminGroup                  GroupName = "adm"
	ErrGroupChangeSignature               = "errGroupChangeSignature: invalid signature for group change"
	ErrGroupChangeAuthorization           = "errGroupChangeAuthorization: user has no Admin rights"
	CompactThreshold                      = 32
)

var GroupChangeFileSize int64 = 1024 * 1024 * 128

type Change uint64

const (
	ChangeGrant   Change = iota // ChangeGrant grants access to a group
	ChangeRevoke                // ChangeRevoke revokes access to a group
	ChangeCurse                 // ChangeCurse revokes access to all groups and invalidate all changes done by the user
	ChangeEndorse               // ChangeEndorse endorses the validity of the group chain

	batchSize       = 1024
	ChangeCheckFreq = 8
)

type GroupChange struct {
	GroupName GroupName       `msgpack:"g"`
	UserId    security.UserId `msgpack:"u"`
	Change    Change          `msgpack:"c"`
	Timestamp int64           `msgpack:"t"`
	Signer    security.UserId `msgpack:"k"`
	Signature []byte          `msgpack:"s"`
}

type GroupChangeFile struct {
	Id   uint64
	Size int64
}

type GroupChain struct {
	Changes []GroupChange
	Groups  Groups
	Hash    []byte
}

func (s *Safe) GetGroups() (Groups, error) {
	g, err := syncGroupChain(s)
	if err != nil {
		return nil, err
	}
	return g.Groups, nil
}

func (s *Safe) UpdateGroup(groupName GroupName, change Change, users ...security.UserId) (Groups, error) {

	for {
		g, err := syncGroupChain(s)
		if err != nil {
			return nil, err
		}

		batchId := getLastBatchId(g)

		var gcs []GroupChange
		var isRevoked bool

		groups := core.CopyMap(g.Groups)
		h := security.NewHash(g.Hash)
		for _, user := range users {
			gc := GroupChange{
				GroupName: groupName,
				UserId:    user,
				Change:    change,
			}
			gc, err = signGroupChange(gc, h, s.CurrentUser) // sign the change
			if err != nil {
				return nil, err
			}
			err = applyChange(gc, groups) // apply the change to the local groups
			if err != nil {
				return nil, err
			}
			if gc.Change == ChangeRevoke {
				isRevoked = true
			}
			gcs = append(gcs, gc)
		}
		finalHash := h.Sum(nil)

		gcs = append(g.Changes, gcs...)
		err = writeGroupChanges(s.Store, gcs, batchId)
		if err != nil {
			return nil, err
		}
		storeHasChanged(s.Db, s.Store, GroupDir)

		rgcs, err := readGroupChanges(s.Store, batchId)
		if err != nil {
			return nil, err
		}

		var mismatch bool
		for i, gc := range gcs {
			if i >= len(rgcs) || !changeEqual(gc, rgcs[i]) {
				mismatch = true
				break
			}
		}
		if mismatch { // the local changes are not in the remote store yet, retry
			continue
		}

		g.Changes = gcs
		g.Hash = finalHash
		g.Groups = groups

		if isRevoked {
			_, err = addKey(s, groupName, groups)
			if err != nil {
				return nil, err
			}
			keysCache.Delete(fmt.Sprintf("%s/%s", s.Store.Url(), groupName))
		}

		err = SetConfigStruct(s.Db, GroupChainNode, s.Store.Url(), g)
		if err != nil {
			return nil, err
		}

		return groups, err
	}
}

func (g Groups) ToString() string {
	data, err := yaml.Marshal(g)
	if err != nil {
		return err.Error()
	}
	return string(data)
}

const (
	leadLocal  = -1
	leadRemote = 1
	leadEqual  = 0
)

const GroupChainNode = "GroupChain"

func syncGroupChain(s *Safe) (GroupChain, error) {
	var g GroupChain
	err := GetConfigStruct(s.Db, GroupChainNode, s.Store.Url(), &g)
	if err == sql.ErrNoRows {
		g.Groups = Groups{AdminGroup: core.NewSet(s.CreatorId)}
	} else if err != nil {
		return GroupChain{}, err
	}
	if !hasStoreChanged(s.Db, s.Store, GroupDir) {
		return g, nil
	}

	var batchId int
	if rand.Intn(ChangeCheckFreq) > 0 { // occasionally check the full history on the remote store
		batchId = getLastBatchId(g)
	}

	rgcs, err := readGroupChanges(s.Store, batchId)
	if err != nil {
		return GroupChain{}, err
	}

	var lead int
	lead, g = addChanges(g, rgcs, batchId)
	switch lead {
	case leadLocal:
		err = writeGroupChanges(s.Store, g.Changes, batchId)
		storeHasChanged(s.Db, s.Store, GroupDir)
	case leadRemote:
		err = SetConfigStruct(s.Db, GroupChainNode, s.Store.Url(), g)
	}
	if err != nil {
		return GroupChain{}, err
	}

	return g, nil
}

func getLastBatchId(g GroupChain) int {
	var n int

	ln := len(g.Changes)
	n = ln / batchSize
	if ln%batchSize > 0 {
		n++
	}
	return n
}

func changeEqual(local, remote GroupChange) bool {
	return local.GroupName == remote.GroupName && local.UserId == remote.UserId && local.Change == remote.Change &&
		local.Signer == remote.Signer && bytes.Equal(local.Signature, remote.Signature)
}

func addChanges(g GroupChain, remoteGcs []GroupChange, batchId int) (int, GroupChain) {
	firstMismatch := -1

	localGcs := g.Changes
	h := security.NewHash(g.Hash)

	i, offset := 0, batchId*batchSize
	for i := 0; i+offset < len(localGcs) && i < len(remoteGcs); i++ {
		if firstMismatch == -1 && !changeEqual(localGcs[i+offset], remoteGcs[i]) {
			firstMismatch = i
		}
	}

	hasFork := firstMismatch != -1
	localEnd := i == offset+len(localGcs)
	remoteEnd := i == len(remoteGcs)

	var err error
	groups := core.CopyMap(g.Groups)
	for j := i; j < len(remoteGcs); j++ {
		err = validateGroupChain(remoteGcs[j], h)
		if err != nil {
			return -1, g
		}

		err = applyChange(remoteGcs[j], groups)
		if err != nil {
			core.Info("failed to apply group change: %v", err)
		}
	}

	if !hasFork && localEnd && remoteEnd { // the chains are identical
		return 0, g
	}

	if !hasFork && localEnd && !remoteEnd { // the local chain is a prefix of the remote chain
		changes := append(localGcs, remoteGcs[offset:]...)
		return 1, GroupChain{Changes: changes, Groups: groups, Hash: h.Sum(nil)}
	}

	if !hasFork && !localEnd && remoteEnd { // the remote chain is a prefix of the local chain
		return -1, g
	}

	if hasFork {
		localEndorsement := calculateEndorsement(localGcs, firstMismatch, len(localGcs))
		remoteEndorsement := calculateEndorsement(remoteGcs, firstMismatch, len(remoteGcs))
		if localEndorsement < remoteEndorsement {
			changes := append(localGcs, remoteGcs[offset:]...)
			return 1, GroupChain{Changes: changes, Groups: groups, Hash: h.Sum(nil)}
		} else if localEndorsement > remoteEndorsement {
			return -1, g
		} else {
			return 0, g
		}
	}

	// the below code should be unreachable
	panic("unexpected state")
}

func validateGroupChain(gc GroupChange, h hash.Hash) error {
	err := updateGroupChangeHash(gc, h)
	if err != nil {
		return err
	}

	data := h.Sum(nil)
	if !security.Verify(gc.Signer, data, gc.Signature) {
		return fmt.Errorf(ErrGroupChangeSignature)
	}
	return nil
}

func applyChange(gc GroupChange, groups Groups) error {
	switch gc.Change {
	case ChangeGrant:
		if !groups[AdminGroup].Contains(gc.Signer) {
			return fmt.Errorf(ErrGroupChangeAuthorization)
		}
		if groups[gc.GroupName] == nil {
			groups[gc.GroupName] = core.NewSet(gc.UserId)
		} else {
			groups[gc.GroupName].Add(gc.UserId)
		}

	case ChangeRevoke:
		if !groups[AdminGroup].Contains(gc.Signer) {
			return fmt.Errorf(ErrGroupChangeAuthorization)
		}
		if groups[gc.GroupName] != nil {
			groups[gc.GroupName].Remove(gc.UserId)
		}
	}

	return nil
}

func calculateEndorsement(gcs []GroupChange, start, end int) int {
	var endorsers Endorsers
	var endorsement int
	for i := start; i < end; i++ {
		switch gcs[i].Change {
		case ChangeEndorse:
			endorsers[gcs[i].UserId] = true
		case ChangeRevoke, ChangeGrant:
			endorsers = Endorsers{}
			endorsement++
		}
	}
	return endorsement
}

// }

func updateGroupChangeHash(gc GroupChange, h hash.Hash) error {
	var buf []byte
	var err error

	buf = append(buf, gc.GroupName...)
	buf = append(buf, gc.UserId...)
	buf = binary.AppendUvarint(buf, uint64(gc.Change))
	buf = binary.AppendUvarint(buf, uint64(gc.Timestamp))
	buf = append(buf, gc.Signer...)

	_, err = h.Write(buf)
	if err != nil {
		return core.Errorw(err, "failed to write to blake2b hash: %v")
	}
	return nil
}

func newGroupChange(group GroupName, user security.UserId, change Change, h hash.Hash, signer security.Identity) (GroupChange, error) {
	gc := GroupChange{
		GroupName: group,
		UserId:    user,
		Change:    change,
		Signer:    signer.Id,
	}
	return signGroupChange(gc, h, signer)
}

func signGroupChange(gc GroupChange, h hash.Hash, signer security.Identity) (GroupChange, error) {
	gc.Timestamp = core.Now().UnixMicro()
	gc.Signer = signer.Id

	err := updateGroupChangeHash(gc, h)
	if err != nil {
		return GroupChange{}, core.Errorw(err, "failed to calculate group change hash: %v")
	}

	data := h.Sum(nil)
	sig, err := security.Sign(signer, data)
	if core.IsErr(err, "failed to sign group change: %v") {
		return GroupChange{}, err
	}
	gc.Signature = sig
	return gc, nil
}

const GroupDir = "groups"

func readGroupChanges(store storage.Store, firstBatchId int) ([]GroupChange, error) {
	var gcs []GroupChange
	ls, err := store.ReadDir(GroupDir, storage.Filter{})
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// extract the ids of the files containing the group changes
	ids := core.Apply(ls, func(l fs.FileInfo) (int, bool) {
		id, err := strconv.Atoi(l.Name())
		return id, err == nil && id >= firstBatchId
	})
	sort.Ints(ids)

	// read the group changes from the files in increasing order of their ids
	for _, id := range ids {
		var changes []GroupChange
		err = storage.ReadMsgPack(store, path.Join(GroupDir, strconv.Itoa(id)), &changes)
		if err != nil {
			return nil, err
		}
		gcs = append(gcs, changes...)
	}
	return gcs, nil
}

func writeGroupChanges(store storage.Store, gcs []GroupChange, fromBatchId int) error {
	i := fromBatchId
	// loop over the changes in batches, each batch is batchSize long and is stored in a separate file
	for offset := fromBatchId * batchSize; offset < len(gcs); offset += batchSize {
		end := offset + batchSize
		if end > len(gcs) { // the last batch may be shorter
			end = len(gcs)
		}
		// write the batch to a file whose name is the batch sequence number
		err := storage.WriteMsgPack(store, path.Join(GroupDir, strconv.Itoa(i)), gcs[offset:end])
		if err != nil {
			return core.Errorw(err, "failed to write group changes: %v")
		}
		i++
	}

	return nil
}
