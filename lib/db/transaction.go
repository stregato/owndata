package db

import (
	"path"

	"github.com/stregato/mio/lib/core"
	"github.com/stregato/mio/lib/safe"
	"github.com/stregato/mio/lib/security"
	"github.com/stregato/mio/lib/sqlx"
	"github.com/stregato/mio/lib/storage"
	"github.com/vmihailenco/msgpack/v5"
)

type Update struct {
	Key     string
	Args    sqlx.Args
	Version float32
}

type Transaction struct {
	Updates   []byte         // Updates is a list of Update encoded in msgpack and encrypted
	Version   float32        // Version is the highest version of the updates
	GroupName safe.GroupName // GroupName is the name of the group that the transaction is for
	KeyId     int            // KeyId is the id of the key used to encrypt the transaction
	Signer    security.ID    // Signer is the id of the user that signed the transaction
	Signature []byte         // Signature is the signature of the transaction
}

func (d *PulseDB) Commit() error {
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
	keys, err := d.Safe.GetKeys(d.groupName, 0)
	if err != nil {
		return err
	}
	lastKey := keys[len(keys)-1]

	encrypted, err := security.EncryptAES(data, lastKey)
	if err != nil {
		return err
	}
	signature, err := security.Sign(d.Safe.Identity, encrypted)
	if err != nil {
		return err
	}

	transaction := Transaction{
		Updates:   encrypted,
		Version:   version,
		GroupName: d.groupName,
		KeyId:     len(keys) - 1,
		Signer:    d.Safe.Identity.Id,
		Signature: signature,
	}

	id := core.SnowIDString()
	dest := path.Join(DBDir, d.groupName.String(), core.SnowIDString())
	err = storage.WriteMsgPack(d.Safe.Store, dest, transaction)
	if err != nil {
		return err
	}

	_, err = d.Safe.DB.Exec("MIO_STORE_TX", sqlx.Args{"safeID": d.Safe.ID, "groupName": d.groupName.String(), "kind": "skip", "id": id})
	if err != nil {
		d.Safe.Store.Delete(dest)
		return err
	}

	err = d.tx.Commit()
	if err != nil {
		d.Safe.Store.Delete(dest)
		return err
	}

	d.Safe.Touch(DBDir)
	return nil
}

func (d *PulseDB) Rollback() error {
	d.tx = nil
	d.log = nil
	return d.tx.Rollback()
}
