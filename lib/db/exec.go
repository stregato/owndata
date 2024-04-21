package db

import (
	s "database/sql"

	"github.com/stregato/mio/lib/sqlx"
)

func (d *PulseDB) Exec(key string, args sqlx.Args) (s.Result, error) {
	args["safe"] = d.Safe.ID
	args["groupName"] = string(d.groupName)
	if d.tx == nil {
		tx, err := d.Safe.DB.GetConnection().Begin()
		if err != nil {
			return nil, err
		}
		d.tx = tx
	}

	res, err := d.Safe.DB.Exec(key, args)
	if err != nil {
		return nil, err
	}

	version := d.Safe.DB.GetVersion(key)

	d.log = append(d.log, Update{key, args, version})
	return res, nil
}
