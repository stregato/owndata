package db

import (
	"path"

	"github.com/stregato/mio/core"
	"github.com/stregato/mio/safe"
	"github.com/stregato/mio/security"
	"github.com/stregato/mio/sql"
	"github.com/stregato/mio/storage"
	"github.com/vmihailenco/msgpack/v5"
)

type Update struct {
	Key     string
	Args    sql.Args
	Version float32
}

type Transaction struct {
	Updates   []byte          // Updates is a list of Update encoded in msgpack and encrypted
	Version   float32         // Version is the highest version of the updates
	GroupName safe.GroupName  // GroupName is the name of the group that the transaction is for
	KeyId     int             // KeyId is the id of the key used to encrypt the transaction
	Signer    security.UserId // Signer is the id of the user that signed the transaction
	Signature []byte          // Signature is the signature of the transaction
}

func (d *DB) Commit() error {
	if d.tx == nil {
		return nil
	}

	var version float32
	for _, u := range d.log {
		if u.Version > version {
			version = u.Version
		}
	}

	data, err := msgpack.Marshal(d.log)
	if err != nil {
		return err
	}
	keys, err := d.s.GetKeys(d.groupName, 0)
	if err != nil {
		return err
	}
	lastKey := keys[len(keys)-1]

	encrypted, err := security.EncryptAES(data, lastKey)
	if err != nil {
		return err
	}
	signature, err := security.Sign(d.s.CurrentUser, encrypted)
	if err != nil {
		return err
	}

	transaction := Transaction{
		Updates:   encrypted,
		Version:   version,
		GroupName: d.groupName,
		KeyId:     len(keys) - 1,
		Signer:    d.s.CurrentUser.Id,
		Signature: signature,
	}

	dest := path.Join(DBDir, string(d.groupName), core.SnowIDString())
	err = storage.WriteMsgPack(d.s.Store, dest, transaction)
	if err != nil {
		return err
	}

	_, err = d.s.Db.Exec("MIO_STORE_TX", sql.Args{"dest": dest, "version": version,
		"group": d.groupName, "storeUrl": d.s.Store.Url(), "consumed": true})
	if err != nil {
		d.s.Store.Delete(dest)
		return err
	}

	err = d.tx.Commit()
	if err != nil {
		d.s.Store.Delete(dest)
		return err
	}
	return nil
}

func (d *DB) Rollback() error {
	d.tx = nil
	d.log = nil
	return d.tx.Rollback()
}
