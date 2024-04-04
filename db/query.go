package db

import "github.com/stregato/mio/sql"

func (d *DB) Query(query string, args sql.Args) (sql.Rows, error) {
	return d.s.Db.Query(query, args)
}
