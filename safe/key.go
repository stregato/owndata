package safe

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/stregato/mio/core"
	"github.com/stregato/mio/security"
	"github.com/stregato/mio/storage"
	"github.com/vmihailenco/msgpack/v5"
)

const (
	KeysDir = "keys"
	KeyNode = "keys"
)

type Key []byte
type Keystore struct {
	MasterKey map[security.UserId]Key
	DataKeys  []byte
	Signature []byte
	Signer    security.UserId
}

type KeyData struct {
	Keys map[uint64]Key
}

var keysCache = cache.New(time.Minute, time.Hour)

// GetKeys returns the encryption keys for the given group. If the user is not authorized to access the keys, it returns a AuthErr.
// The parameter expectedMinLength is used to check if the number of keys is at least the expected value. If it is 0, the check is skipped.
func (s *Safe) GetKeys(groupName GroupName, expectedMinLength int) ([]Key, error) {
	k, found := keysCache.Get(fmt.Sprintf("%s/%s", s.Store.Url(), groupName))
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
	if !g.Contains(s.CurrentUser.Id) {
		return nil, core.Errorf("AuthErr: user %s is not in the group %s", s.CurrentUser.Id, groupName)
	}

	return syncKeys(s, groupName, groups)
}

func syncKeys(c *Safe, groupName GroupName, groups Groups) ([]Key, error) {
	keys, err := readKeystore(c, groupName, groups)
	if os.IsNotExist(err) {
		keys, err = readKeysFromDb(c, groupName)
		if err == sql.ErrNoRows {
			keys = []Key{core.GenerateRandomBytes(32)}
		} else if err != nil {
			return nil, err
		}

		err = writeKeystore(c, groupName, groups, keys)
		if err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}
	err = writeKeysToDb(c, groupName, keys)
	if err != nil {
		return nil, err
	}
	keysCache.Set(fmt.Sprintf("%s/%s", c.Store.Url(), groupName), keys, cache.DefaultExpiration)
	return keys, nil
}

func addKey(c *Safe, groupName GroupName, groups Groups) ([]Key, error) {
	var retries int
retry:
	if retries++; retries > 10 {
		return nil, core.Errorf("keyAddErr: cannot add new excryption key after %d retries", retries)
	}

	keys, err := syncKeys(c, groupName, groups)
	if err != nil {
		return nil, err
	}

	keys = append(keys, core.GenerateRandomBytes(32))
	err = writeKeystore(c, groupName, groups, keys)
	if err != nil {
		goto retry
	}

	err = writeKeysToDb(c, groupName, keys)
	if err != nil {
		return nil, err
	}
	keysCache.Set(fmt.Sprintf("%s/%s", c.Store.Url(), groupName), keys, cache.DefaultExpiration)

	return keys, nil

}

func writeKeysToDb(c *Safe, groupName GroupName, keys []Key) error {
	k := path.Join(KeysDir, string(groupName))
	return SetConfigStruct(c.Db, KeyNode, k, keys)
}

// readKeysFromDb reads keys from the sql. It returns sql.ErrNoRows if the keys are not found.
func readKeysFromDb(c *Safe, groupName GroupName) ([]Key, error) {
	var keys []Key
	k := path.Join(KeysDir, string(groupName))
	err := GetConfigStruct(c.Db, KeyNode, k, &keys)
	return keys, err
}

// readKeystore reads the keystore from the store and verifies the signature. It returns ErrNotExist if the keystore is not found.
func readKeystore(c *Safe, groupName GroupName, groups Groups) ([]Key, error) {
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

	masterKey, ok := keystore.MasterKey[c.CurrentUser.Id]
	if !ok {
		return nil, core.Errorf("AuthErr: user %s is not authorized to access the keys for group %s", c.CurrentUser.Id, groupName)
	}

	data, err := security.DecryptAES(keystore.DataKeys, masterKey)
	if err != nil {
		return nil, err
	}

	var keys []Key
	err = msgpack.Unmarshal(data, &keys)
	if err != nil {
		return nil, err
	}

	return keys, nil
}

func writeKeystore(c *Safe, groupName GroupName, groups Groups, keys []Key) error {
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
		MasterKey: make(map[security.UserId]Key),
		DataKeys:  data,
		Signer:    c.CurrentUser.Id,
	}
	group := groups[groupName]
	for userId := range group {
		encryptedMasterKey, err := security.EcEncrypt(string(userId), masterKey)
		if core.IsErr(err, "cannot encrypt master key for user id %s: %v", userId) {
			continue
		}

		keystore.MasterKey[userId] = encryptedMasterKey
	}
	keystore.Signature, err = security.Sign(c.CurrentUser, data)
	if err != nil {
		return err
	}

	filename := path.Join(KeysDir, fmt.Sprintf("%s.ks", groupName))
	err = storage.WriteMsgPack(c.Store, filename, keystore)
	if err != nil {
		return err
	}

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
