package stash

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/stregato/stash/lib/config"
	"github.com/stregato/stash/lib/core"
	"github.com/stregato/stash/lib/security"
	"github.com/stregato/stash/lib/storage"
	"gopkg.in/yaml.v2"
)

type GroupName string
type Groups map[GroupName]core.Set[security.ID]
type Endorsers core.Set[security.ID]

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
	Grant   Change = iota // Grant grants access to a group
	Revoke                // Revoke revokes access to a group
	Curse                 // Curse revokes access to all groups and invalidate all changes done by the user
	Endorse               // Endorse endorses the validity of the group chain

	batchSize       = 1024
	ChangeCheckFreq = 8
)

type GroupChange struct {
	GroupName GroupName   `msgpack:"g"`
	Change    Change      `msgpack:"c"`
	UserId    security.ID `msgpack:"u"`
	Signer    security.ID `msgpack:"k"`
	Signature []byte      `msgpack:"s"`
	Timestamp int64       `msgpack:"t"`
}

type GroupChangeFile struct {
	Id   uint64
	Size int64
}

type GroupChain struct {
	Changes []GroupChange
	Groups  Groups
}

func (s *Stash) GetGroups() (Groups, error) {
	g, err := SyncGroupChain(s)
	if err != nil {
		return nil, err
	}
	return g.Groups, nil
}

func (s *Stash) UpdateGroup(groupName GroupName, change Change, users ...security.ID) (Groups, error) {

	var lastSignature []byte

	lock, err := storage.Lock(s.Store, GroupDir, "chain", time.Minute)
	defer storage.Unlock(lock)
	if err != nil {
		return nil, err
	}

	g, err := SyncGroupChain(s)
	if err != nil {
		return nil, err
	}

	batchId := len(g.Changes) / batchSize

	var gcs []GroupChange
	var createNewKey bool

	if len(g.Changes) > 0 {
		lastSignature = g.Changes[len(g.Changes)-1].Signature
	}
	groups := core.CopyMap(g.Groups)
	for _, user := range users {
		// check if the user is already in the group and skip the change in case of Grant
		if change == Grant && groups[groupName].Contains(user) {
			core.Info("user %s is already in the group %s", user.Nick(), groupName)
			continue
		}
		// check if the user is not in the group and skip the change in case of Revoke
		if change == Revoke && !groups[groupName].Contains(user) {
			core.Info("user %s is not in the group %s", user.Nick(), groupName)
			continue
		}

		// create a new group change
		gc := GroupChange{
			GroupName: groupName,
			UserId:    user,
			Change:    change,
		}
		gc, err = signGroupChange(gc, lastSignature, s.Identity) // sign the change
		if err != nil {
			return nil, err
		}
		err = applyChange(gc, groups, s.CreatorID) // apply the change to the local groups
		if err != nil {
			return nil, err
		}
		// if the change is a Revoke, a new key must be added to the store
		if gc.Change == Revoke {
			createNewKey = true
		}
		gcs = append(gcs, gc)
		lastSignature = gc.Signature
		core.Info("group change created and added to the chain: %s", gc)
	}
	if len(gcs) == 0 {
		return groups, nil
	}

	gcs = append(g.Changes, gcs...)
	err = writeGroupChanges(s.Store, gcs, batchId)
	if err != nil {
		return nil, err
	}
	s.Touch(GroupDir)

	g.Changes = gcs
	g.Groups = groups

	_, err = updateKeys(s, groupName, groups, createNewKey)
	if err != nil {
		return nil, err
	}

	err = config.SetConfigStruct(s.DB, config.GroupChainDomain, s.Store.ID(), g)
	if err != nil {
		return nil, err
	}

	return groups, err
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

func SyncGroupChain(s *Stash) (GroupChain, error) {
	var g GroupChain
	err := config.GetConfigStruct(s.DB, config.GroupChainDomain, s.Store.ID(), &g)
	if err != sql.ErrNoRows && err != nil {
		return GroupChain{}, err
	}

	noChainInDB := err != sql.ErrNoRows
	if noChainInDB && !s.IsUpdated(GroupDir) {
		core.Info("group chain is up to date, using the local copy")
		return g, nil
	}

	var batchId int
	if rand.Intn(ChangeCheckFreq) > 0 { // occasionally check the full history on the remote store
		batchId = len(g.Changes) / batchSize
	}

	rgcs, err := readGroupChanges(s.Store, batchId)
	if err != nil {
		return GroupChain{}, err
	}

	var lead int
	lead, g = addChanges(g, rgcs, batchId, s.CreatorID)
	switch lead {
	case leadLocal:
		core.Info("local group chain is lead, writing the changes to the store")
		err = writeGroupChanges(s.Store, g.Changes, batchId)
		s.Touch(GroupDir)
	case leadRemote:
		core.Info("remote group chain is lead, updating the local copy")
		err = config.SetConfigStruct(s.DB, config.GroupChainDomain, s.Store.ID(), g)
	default:
		core.Info("neither local nor remote group chain is lead, doing nothing hoping another peer will resolve the conflict")
	}
	if err != nil {
		return GroupChain{}, err
	}

	return g, nil
}

func changeEqual(local, remote GroupChange) bool {
	return local.GroupName == remote.GroupName && local.UserId == remote.UserId && local.Change == remote.Change &&
		local.Signer == remote.Signer && bytes.Equal(local.Signature, remote.Signature)
}

func addChanges(g GroupChain, remoteGcs []GroupChange, batchId int, creatorId security.ID) (int, GroupChain) {
	firstMismatch := -1

	localGcs := g.Changes

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
	var lastSignature []byte

	if len(localGcs) > 0 {
		lastSignature = localGcs[len(localGcs)-1].Signature
	}
	groups := core.CopyMap(g.Groups)
	for j := i; j < len(remoteGcs); j++ {
		err = validateGroupChain(remoteGcs[j], lastSignature)
		if err != nil {
			return -1, g
		}

		err = applyChange(remoteGcs[j], groups, creatorId)
		if err != nil {
			core.Info("failed to apply group change: %v", err)
		}
		lastSignature = remoteGcs[j].Signature
	}

	if !hasFork && localEnd && remoteEnd { // the chains are identical
		core.Info("local and remote group chains are identical")
		return 0, g
	}

	if !hasFork && localEnd && !remoteEnd { // the local chain is a prefix of the remote chain
		core.Info("local group chain is a prefix of the remote group chain")
		changes := append(localGcs, remoteGcs[offset:]...)
		return 1, GroupChain{Changes: changes, Groups: groups}
	}

	if !hasFork && !localEnd && remoteEnd { // the remote chain is a prefix of the local chain
		core.Info("remote group chain is a prefix of the local group chain")
		return -1, g
	}

	if hasFork {
		localEndorsement := calculateEndorsement(localGcs, firstMismatch, len(localGcs))
		remoteEndorsement := calculateEndorsement(remoteGcs, firstMismatch, len(remoteGcs))
		if localEndorsement < remoteEndorsement {
			core.Info("remote group chain has more endorsements, using it")
			changes := append(localGcs, remoteGcs[offset:]...)
			return 1, GroupChain{Changes: changes, Groups: groups}
		} else if localEndorsement > remoteEndorsement {
			core.Info("local group chain has more endorsements, ignoring the remote changes")
			return -1, g
		} else {
			core.Info("local and remote group chains have the same number of endorsements. Do nothing")
			return 0, g
		}
	}

	// the below code should be unreachable
	panic("unexpected state")
}

func validateGroupChain(gc GroupChange, lastSignature []byte) error {
	h, err := getGroupChangeHash(gc, lastSignature)
	if err != nil {
		return err
	}

	if !security.Verify(gc.Signer, h, gc.Signature) {
		return fmt.Errorf(ErrGroupChangeSignature)
	}

	core.Info("group change validated: %s", gc)
	return nil
}

func applyChange(gc GroupChange, groups Groups, creatorId security.ID) error {
	var skipCheck bool
	if len(groups) == 0 {
		skipCheck = gc.UserId == creatorId
	}

	switch gc.Change {
	case Grant:
		if !skipCheck && !groups[AdminGroup].Contains(gc.Signer) {
			return fmt.Errorf(ErrGroupChangeAuthorization)
		}
		if groups[gc.GroupName] == nil {
			groups[gc.GroupName] = core.NewSet(gc.UserId)
		} else {
			groups[gc.GroupName].Add(gc.UserId)
		}

	case Revoke:
		if !skipCheck && !groups[AdminGroup].Contains(gc.Signer) {
			return fmt.Errorf(ErrGroupChangeAuthorization)
		}
		if groups[gc.GroupName] != nil {
			groups[gc.GroupName].Remove(gc.UserId)
		}
	}

	core.Info("group change applied: %s", gc)

	return nil
}

func calculateEndorsement(gcs []GroupChange, start, end int) int {
	var endorsers Endorsers
	var endorsement int
	for i := start; i < end; i++ {
		switch gcs[i].Change {
		case Endorse:
			endorsers[gcs[i].UserId] = true
		case Revoke, Grant:
			endorsers = Endorsers{}
			endorsement++
		}
	}
	return endorsement
}

func getGroupChangeHash(gc GroupChange, lastSig []byte) ([]byte, error) {
	var buf []byte
	var err error

	buf = append(buf, lastSig...)
	buf = append(buf, gc.GroupName...)
	buf = append(buf, gc.UserId...)
	buf = binary.AppendUvarint(buf, uint64(gc.Change))
	buf = append(buf, gc.Signer...)

	h := security.NewHash(nil)
	_, err = h.Write(buf)
	if err != nil {
		return nil, core.Errorw(err, "failed to write to blake2b hash: %v")
	}
	return h.Sum(nil), nil
}

func signGroupChange(gc GroupChange, lastSignature []byte, signer *security.Identity) (GroupChange, error) {
	gc.Timestamp = core.Now().UnixMicro()
	gc.Signer = signer.Id

	h, err := getGroupChangeHash(gc, lastSignature)
	if err != nil {
		return GroupChange{}, core.Errorw(err, "failed to calculate group change hash: %v")
	}

	signature, err := security.Sign(signer, h)
	if core.IsErr(err, "failed to sign group change: %v") {
		return GroupChange{}, err
	}
	gc.Signature = signature
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

	var batches []string
	// read the group changes from the files in increasing order of their ids
	for _, id := range ids {
		var changes []GroupChange
		err = storage.ReadMsgPack(store, path.Join(GroupDir, strconv.Itoa(id)), &changes)
		if err != nil {
			return nil, err
		}
		gcs = append(gcs, changes...)
		batches = append(batches, strconv.Itoa(id))
	}

	core.Info("group changes read from the store from batches [%s]", strings.Join(batches, " "))
	return gcs, nil
}

func writeGroupChanges(store storage.Store, gcs []GroupChange, fromBatchId int) error {
	i := fromBatchId
	var batches []string

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
		batches = append(batches, strconv.Itoa(i))
		i++
	}

	core.Info("group changes written to the store on batches [%s]", strings.Join(batches, " "))
	return nil
}

func (g GroupName) String() string {
	return string(g)
}

func (groups Groups) String() string {
	var buf bytes.Buffer
	for n, g := range groups {
		buf.WriteString(fmt.Sprintf("%s: ", n))
		for _, u := range g.Slice() {
			buf.WriteString(fmt.Sprintf("%s ", u.Nick()))
		}
		buf.WriteString("\n")
	}
	return buf.String()
}

func (gc GroupChange) String() string {
	var change string
	switch gc.Change {
	case Grant:
		change = "granted to"
	case Revoke:
		change = "revoked from"
	case Curse:
		change = "cursed from"
	}
	return fmt.Sprintf("%s %s %s by %s", gc.UserId.Nick(), change, gc.GroupName, gc.Signer.Nick())
}

func (gc GroupChain) String() string {
	var buf bytes.Buffer
	for i, c := range gc.Changes {
		buf.WriteString(fmt.Sprintf("%d: %s\n", i, c))
	}

	buf.WriteString("Groups\n")
	buf.WriteString(gc.Groups.String())

	return buf.String()
}
