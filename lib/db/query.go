package db

import "github.com/stregato/mio/lib/sqlx"

func (d *DB) Query(query string, args sqlx.Args) (sqlx.Rows, error) {
	return d.s.Db.Query(query, args)
}
