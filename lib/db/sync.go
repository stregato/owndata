package db

import (
	"path"

	"github.com/stregato/mio/lib/core"
	"github.com/stregato/mio/lib/security"
	"github.com/stregato/mio/lib/sqlx"
	"github.com/stregato/mio/lib/storage"
	"github.com/vmihailenco/msgpack/v5"
)

func (d *DB) Sync() error {

	var lastId string
	err := d.s.Db.QueryRow("MIO_GET_LAST_TX", sqlx.Args{"group": d.groupName, "storeUrl": d.s.Store.Url()}, &lastId)
	if err != nil && err != sqlx.ErrNoRows {
		return err
	}

	dir := path.Join(DBDir, string(d.groupName))
	ls, err := d.s.Store.ReadDir(dir, storage.Filter{})
	if err != nil {
		return err
	}
	for _, l := range ls {
		if l.Name() > lastId {
			var tx Transaction

			err = storage.ReadMsgPack(d.s.Store, path.Join(dir, l.Name()), &tx)
			if err != nil {
				return err
			}

			if tx.GroupName != d.groupName {
				return core.Errorf("wrong group name %s", tx.GroupName)
			}

			if !security.Verify(tx.Signer, tx.Updates, tx.Signature) {
				return core.Errorf("cannot verify transaction %s", l.Name())
			}

			_, err = d.s.Db.Exec("MIO_STORE_TX", sqlx.Args{"group": tx.GroupName, "storeUrl": d.s.Store.Url(), "id": l.Name(), "version": 0, "consumed": false})
			if err != nil {
				return err
			}

			keys, err := d.s.GetKeys(tx.GroupName, tx.KeyId)
			if err != nil {
				continue
			}

			decrypted, err := security.DecryptAES(tx.Updates, keys[tx.KeyId])
			if err != nil {
				continue
			}

			var updates []Update
			err = msgpack.Unmarshal(decrypted, &updates)
			if err != nil {
				continue
			}

			sqlTx, err := d.s.Db.GetConnection().Begin()
			if err != nil {
				continue
			}
			for _, u := range updates {
				_, err = d.s.Db.Exec(u.Key, u.Args)
				if err != nil {
					break
				}
			}
			if err != nil {
				sqlTx.Rollback()
				continue
			}

			_, err = d.s.Db.Exec("MIO_STORE_TX", sqlx.Args{"group": tx.GroupName, "storeUrl": d.s.Store.Url(), "id": l.Name(), "version": tx.Version, "consumed": true})
			if err != nil {
				sqlTx.Rollback()
				continue
			}
			sqlTx.Commit()
		}
	}
	return nil
}
