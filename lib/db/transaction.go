package db

import (
	s "database/sql"
	"path"

	"github.com/stregato/stash/lib/core"
	"github.com/stregato/stash/lib/safe"
	"github.com/stregato/stash/lib/security"
	"github.com/stregato/stash/lib/sqlx"
	"github.com/stregato/stash/lib/storage"
	"github.com/vmihailenco/msgpack/v5"
)

type Update struct {
	Key     string
	Args    sqlx.Args
	Version float32
}

type Transaction struct {
	db        *DB
	tx        *s.Tx
	log       []Update
	Updates   []byte         // Updates is a list of Update encoded in msgpack and encrypted
	Version   float32        // Version is the highest version of the updates
	GroupName safe.GroupName // GroupName is the name of the group that the transaction is for
	KeyId     int            // KeyId is the id of the key used to encrypt the transaction
	Signer    security.ID    // Signer is the id of the user that signed the transaction
	Signature []byte         // Signature is the signature of the transaction
}

func (sq *DB) Transaction() (*Transaction, error) {
	tx, err := sq.Safe.DB.GetConnection().Begin()
	if err != nil {
		return nil, err
	}
	return &Transaction{
		db:        sq,
		tx:        tx,
		GroupName: sq.groupName,
	}, nil
}

func (t *Transaction) Exec(key string, args sqlx.Args) (s.Result, error) {
	res, err := t.db.Safe.DB.Exec(key, args)
	if err != nil {
		return nil, err
	}

	version := t.db.Safe.DB.GetVersion(key)

	t.log = append(t.log, Update{key, args, version})
	return res, nil
}

func (t *Transaction) Commit() error {
	var version float32
	for _, u := range t.log {
		if u.Version > version {
			version = u.Version
		}
	}

	data, err := msgpack.Marshal(t.log)
	if err != nil {
		return err
	}
	keys, err := t.db.Safe.GetKeys(t.db.groupName, 0)
	if err != nil {
		return err
	}
	lastKey := keys[len(keys)-1]

	encrypted, err := security.EncryptAES(data, lastKey)
	if err != nil {
		return err
	}
	signature, err := security.Sign(t.db.Safe.Identity, encrypted)
	if err != nil {
		return err
	}

	transaction := Transaction{
		Updates:   encrypted,
		Version:   version,
		GroupName: t.db.groupName,
		KeyId:     len(keys) - 1,
		Signer:    t.db.Safe.Identity.Id,
		Signature: signature,
	}

	id := core.SnowIDString()
	dest := path.Join(DBDir, t.db.groupName.String(), core.SnowIDString())
	err = storage.WriteMsgPack(t.db.Safe.Store, dest, transaction)
	if err != nil {
		return err
	}

	_, err = t.db.Safe.DB.Exec("STASH_STORE_TX", sqlx.Args{"safeID": t.db.Safe.ID, "groupName": t.db.groupName.String(), "kind": "skip", "id": id})
	if err != nil {
		t.db.Safe.Store.Delete(dest)
		return err
	}

	err = t.tx.Commit()
	if err != nil {
		t.db.Safe.Store.Delete(dest)
		return err
	}

	t.db.Safe.Touch(DBDir)
	return nil
}

func (t *Transaction) Rollback() error {
	return t.tx.Rollback()
}
