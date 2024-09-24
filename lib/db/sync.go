package db

import (
	"path"

	"github.com/stregato/stash/lib/core"
	"github.com/stregato/stash/lib/safe"
	"github.com/stregato/stash/lib/security"
	"github.com/stregato/stash/lib/sqlx"
	"github.com/stregato/stash/lib/storage"
	"github.com/vmihailenco/msgpack/v5"
)

func (d *Database) processTransaction(dir, id string, keys []safe.Key) ([]Update, error) {
	var tx Transaction
	var updates []Update

	err := storage.ReadMsgPack(d.Safe.Store, path.Join(dir, id), &tx)
	if err != nil {
		return nil, err
	}

	if tx.KeyId >= len(keys) {
		keys, err = d.Safe.GetKeys(tx.GroupName, tx.KeyId)
		if err != nil {
			return nil, err
		}
	}

	if tx.GroupName != d.groupName {
		return nil, core.Errorf("wrong group name %s", tx.GroupName)
	}

	if !security.Verify(tx.Signer, tx.Updates, tx.Signature) {
		return nil, core.Errorf("cannot verify transaction %s", id)
	}

	decrypted, err := security.DecryptAES(tx.Updates, keys[tx.KeyId])
	if err != nil {
		return nil, err
	}

	err = msgpack.Unmarshal(decrypted, &updates)
	if err != nil {
		return nil, err
	}

	sqlTx, err := d.Safe.DB.GetConnection().Begin()
	if err != nil {
		return nil, err
	}
	for _, u := range updates {
		_, err = d.Safe.DB.Exec(u.Key, u.Args)
		if err != nil {
			break
		}
	}
	if err != nil {
		sqlTx.Rollback()
		return nil, err
	}

	err = sqlTx.Commit()
	if err != nil {
		sqlTx.Rollback()
		return nil, err
	}
	return updates, nil
}

func (d *Database) Sync() ([]Update, error) {
	err := d.commit()
	if err != nil {
		return nil, err
	}
	return d.sync(false)
}

func (d *Database) sync(force bool) ([]Update, error) {
	if !force && !d.Safe.IsUpdated(DBDir) {
		return nil, nil
	}

	keys, err := d.Safe.GetKeys(d.groupName, 0)
	if err != nil {
		return nil, err
	}

	var updates []Update
	var lastId string

	var ids []string
	ignores := core.Set[string]{}
	rows, err := d.Safe.DB.Query("MIO_GET_TX", sqlx.Args{"groupName": d.groupName.String(), "safeID": d.Safe.ID})
	if err == nil {
		for rows.Next() {
			var kind, id string
			if rows.Scan(&kind, &id) == nil {
				switch kind {
				case "failed":
					ids = append(ids, id)
				case "skip":
					ignores.Add(id)
				case "last":
					lastId = id
				}
			}
		}
		rows.Close()
	}

	groupName := d.groupName.String()
	dir := path.Join(DBDir, groupName)
	ls, err := d.Safe.Store.ReadDir(dir, storage.Filter{})
	if err != nil {
		return nil, err
	}
	for _, l := range ls {
		if l.Name() > lastId {
			ids = append(ids, l.Name())
		}
	}

	if len(ids) == 0 {
		return nil, nil
	}

	for _, id := range ids {
		if ignores.Contains(id) {
			continue
		}

		u, err := d.processTransaction(dir, id, keys)
		if err != nil {
			d.Safe.DB.Exec("MIO_STORE_TX", sqlx.Args{"groupName": groupName, "safeID": d.Safe.ID, "kind": "failed", "id": id})
		}
		updates = append(updates, u...)
		lastId = id
	}

	_, err = d.Safe.DB.Exec("MIO_STORE_TX", sqlx.Args{"groupName": groupName, "safeID": d.Safe.ID, "kind": "last", "lastId": lastId})
	if err != nil {
		return nil, err
	}
	_, err = d.Safe.DB.Exec("MIO_DEL_TX_KIND", sqlx.Args{"groupName": groupName, "safeID": d.Safe.ID, "kind": "skip"})
	if err != nil {
		return nil, err
	}

	d.Safe.Touch(DBDir)

	return updates, nil
}
