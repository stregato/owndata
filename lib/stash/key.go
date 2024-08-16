package stash

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/stregato/stash/lib/config"
	"github.com/stregato/stash/lib/core"
	"github.com/stregato/stash/lib/security"
	"github.com/stregato/stash/lib/storage"
	"github.com/vmihailenco/msgpack/v5"
)

const (
	KeysDir = "keys"
)

type Key []byte
type Keystore struct {
	EnvelopeKey map[security.ID]Key
	DataKeys    []byte
	Signature   []byte
	Signer      security.ID
}

type KeyData struct {
	Keys map[uint64]Key
}

var keysCache = cache.New(time.Minute, time.Hour)

// GetKeys returns the encryption keys for the given group. If the user is not authorized to access the keys, it returns a AuthErr.
// The parameter expectedMinLength is used to check if the number of keys is at least the expected value. If it is 0, the check is skipped.
func (s *Stash) GetKeys(groupName GroupName, expectedMinLength int) ([]Key, error) {
	k, found := keysCache.Get(fmt.Sprintf("%s/%s", s.Store.ID(), groupName))
	if found {
		keys, ok := k.([]Key)
		if ok && (expectedMinLength == 0 || len(keys) >= expectedMinLength) {
			return keys, nil
		}
	}
	groups, err := s.GetGroups()
	if err != nil {
		return nil, err
	}

	g := groups[groupName]
	if !g.Contains(s.Identity.Id) {
		return nil, core.Errorf("AuthErr: user %s is not in the group %s", s.Identity.Id, groupName)
	}

	return syncKeys(s, groupName, groups)
}

func syncKeys(s *Stash, groupName GroupName, groups Groups) ([]Key, error) {
	var keys []Key
	var err error
	_keys, ok := keysCache.Get(path.Join(s.ID, groupName.String()))
	if ok {
		keys = _keys.([]Key)
	} else {
		keys, err = readKeysFromDb(s, groupName)
		if err != nil {
			keys = nil
		}
	}
	if keys != nil && !s.IsUpdated(KeysDir) {
		return keys, nil
	}

	keys, err = readKeystore(s, groupName, groups)
	if os.IsNotExist(err) {
		keys, err = readKeysFromDb(s, groupName)
		if err == sql.ErrNoRows {
			keys = []Key{core.GenerateRandomBytes(32)}
		} else if err != nil {
			return nil, err
		}

		err = writeKeystore(s, groupName, groups, keys)
		if err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}
	err = writeKeysToDb(s, groupName, keys)
	if err != nil {
		return nil, err
	}
	keysCache.Set(fmt.Sprintf("%s/%s", s.ID, groupName), keys, cache.DefaultExpiration)
	s.Touch(KeysDir)
	return keys, nil
}

func updateKeys(c *Stash, groupName GroupName, groups Groups, createNewDataKey bool) ([]Key, error) {

	lock, err := storage.Lock(c.Store, KeysDir, "keys", time.Minute)
	if err != nil {
		return nil, err
	}
	defer storage.Unlock(lock)

	keys, err := syncKeys(c, groupName, groups)
	if err != nil {
		return nil, err
	}

	if createNewDataKey {
		keys = append(keys, core.GenerateRandomBytes(32))
	}
	err = writeKeystore(c, groupName, groups, keys)
	if err != nil {
		return nil, err
	}

	err = writeKeysToDb(c, groupName, keys)
	if err != nil {
		return nil, err
	}
	keysCache.Set(fmt.Sprintf("%s/%s", c.Store.ID(), groupName), keys, cache.DefaultExpiration)

	return keys, nil
}

func writeKeysToDb(c *Stash, groupName GroupName, keys []Key) error {
	k := path.Join(KeysDir, string(groupName))
	return config.SetConfigStruct(c.DB, config.KeystoreDomain, k, keys)
}

// readKeysFromDb reads keys from the sql. It returns sql.ErrNoRows if the keys are not found.
func readKeysFromDb(c *Stash, groupName GroupName) ([]Key, error) {
	var keys []Key
	k := path.Join(KeysDir, string(groupName))
	err := config.GetConfigStruct(c.DB, config.KeystoreDomain, k, &keys)
	return keys, err
}

// readKeystore reads the keystore from the store and verifies the signature. It returns ErrNotExist if the keystore is not found.
func readKeystore(c *Stash, groupName GroupName, groups Groups) ([]Key, error) {
	var keystore Keystore
	filename := path.Join(KeysDir, fmt.Sprintf("%s.ks", groupName))
	err := storage.ReadMsgPack(c.Store, filename, &keystore) // read and parse the keystore
	if err != nil {
		return nil, err
	}

	adminGroup := groups[AdminGroup]
	if !adminGroup.Contains(keystore.Signer) { // check if the signer is in the group
		return nil, core.Errorf("InvalidSignerErr: signer %s is not in the group %s", keystore.Signer, AdminGroup)
	}

	if !security.Verify(keystore.Signer, keystore.DataKeys, keystore.Signature) {
		return nil, core.Errorf("InvalidSignatureErr: invalid signature for group %s", groupName)
	}

	core.Info("keystore %s.ks read successfully: users %s", groupName, strings.Join(
		core.Apply(core.Keys(keystore.EnvelopeKey), func(id security.ID) (string, bool) {
			return id.Nick(), true
		}), ", "))

	envelopeKey, ok := keystore.EnvelopeKey[c.Identity.Id]
	if !ok {
		return nil, core.Errorf("AuthErr: user %s is not authorized to access the keys for group %s", c.Identity.Id, groupName)
	}
	envelopeKey, err = security.EcDecrypt(c.Identity, envelopeKey)
	if err != nil {
		return nil, core.Errorw(err, "cannot decrypt master key for %s", c.Identity.Id.Nick())
	}
	core.Info("envelope key decrypted successfully for user %s", c.Identity.Id.Nick())

	data, err := security.DecryptAES(keystore.DataKeys, envelopeKey)
	if err != nil {
		return nil, err
	}

	var keys []Key
	err = msgpack.Unmarshal(data, &keys)
	if err != nil {
		return nil, err
	}
	core.Info("data keys decrypted successfully for user %s", c.Identity.Id.Nick())

	return keys, nil
}

func writeKeystore(c *Stash, groupName GroupName, groups Groups, keys []Key) error {
	adminGroup := groups[AdminGroup]
	if !adminGroup.Contains(c.Identity.Id) {
		return core.Errorf("AuthErr: user %s is not in the group %s", c.Identity.Id, AdminGroup)
	}

	data, err := msgpack.Marshal(keys)
	if err != nil {
		return err
	}

	masterKey := core.GenerateRandomBytes(32)
	data, err = security.EncryptAES(data, masterKey)
	if err != nil {
		return err
	}

	keystore := Keystore{
		EnvelopeKey: make(map[security.ID]Key),
		DataKeys:    data,
		Signer:      c.Identity.Id,
	}
	group := groups[groupName]
	var users []string
	for userId := range group {
		encryptedMasterKey, err := security.EcEncrypt(userId, masterKey)
		if core.IsErr(err, "cannot encrypt master key for user id %s: %v", userId) {
			continue
		}

		keystore.EnvelopeKey[userId] = encryptedMasterKey
		users = append(users, userId.Nick())
	}
	keystore.Signature, err = security.Sign(c.Identity, data)
	if err != nil {
		return err
	}

	filename := path.Join(KeysDir, fmt.Sprintf("%s.ks", groupName))
	err = storage.WriteMsgPack(c.Store, filename, keystore)
	if err != nil {
		return err
	}
	core.Info("keystore %s.ks written successfully by %s: users %s", groupName, c.Identity.Id.Nick(), strings.Join(users, ", "))

	var keystore2 Keystore
	err = storage.ReadMsgPack(c.Store, filename, &keystore2) // read and parse the keystore
	if err != nil {
		return err
	}

	if !bytes.Equal(keystore.Signature, keystore2.Signature) || !bytes.Equal(keystore.DataKeys, keystore2.DataKeys) {
		return core.Errorf("mismatched keystore")
	}
	return nil
}
