package db

import (
	"path"

	"github.com/stregato/stash/lib/core"
	"github.com/stregato/stash/lib/security"
	"github.com/stregato/stash/lib/sqlx"
	"github.com/stregato/stash/lib/stash"
	"github.com/stregato/stash/lib/storage"
	"github.com/vmihailenco/msgpack/v5"
)

type Update struct {
	Key     string
	Args    sqlx.Args
	Version float32
}

type Transaction struct {
	Updates   []byte          // Updates is a list of Update encoded in msgpack and encrypted
	Version   float32         // Version is the highest version of the updates
	GroupName stash.GroupName // GroupName is the name of the group that the transaction is for
	KeyId     int             // KeyId is the id of the key used to encrypt the transaction
	Signer    security.ID     // Signer is the id of the user that signed the transaction
	Signature []byte          // Signature is the signature of the transaction
}

func (d *Database) commit() error {
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
	keys, err := d.Stash.GetKeys(d.groupName, 0)
	if err != nil {
		return err
	}
	lastKey := keys[len(keys)-1]

	encrypted, err := security.EncryptAES(data, lastKey)
	if err != nil {
		return err
	}
	signature, err := security.Sign(d.Stash.Identity, encrypted)
	if err != nil {
		return err
	}

	transaction := Transaction{
		Updates:   encrypted,
		Version:   version,
		GroupName: d.groupName,
		KeyId:     len(keys) - 1,
		Signer:    d.Stash.Identity.Id,
		Signature: signature,
	}

	id := core.SnowIDString()
	dest := path.Join(DBDir, d.groupName.String(), core.SnowIDString())
	err = storage.WriteMsgPack(d.Stash.Store, dest, transaction)
	if err != nil {
		return err
	}

	_, err = d.Stash.DB.Exec("MIO_STORE_TX", sqlx.Args{"safeID": d.Stash.ID, "groupName": d.groupName.String(), "kind": "skip", "id": id})
	if err != nil {
		d.Stash.Store.Delete(dest)
		return err
	}

	err = d.tx.Commit()
	if err != nil {
		d.Stash.Store.Delete(dest)
		return err
	}

	d.tx = nil
	d.Stash.Touch(DBDir)
	return nil
}

func (d *Database) Cancel() error {
	d.tx = nil
	d.log = nil
	return d.tx.Rollback()
}
