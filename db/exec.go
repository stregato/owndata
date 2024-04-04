package db

import (
	s "database/sql"

	"github.com/stregato/mio/sql"
)

func (d *DB) Exec(key string, args sql.Args) (s.Result, error) {
	if d.tx == nil {
		tx, err := d.s.Db.GetConnection().Begin()
		if err != nil {
			return nil, err
		}
		d.tx = tx
	}

	res, err := d.s.Db.Exec(key, args)
	if err != nil {
		return nil, err
	}

	version := d.s.Db.GetVersion(key)

	d.log = append(d.log, Update{key, args, version})
	return res, nil
}
