package db

import "github.com/stregato/stash/lib/sqlx"

func (d *DB) Query(query string, args sqlx.Args) (sqlx.Rows, error) {
	return d.Safe.DB.Query(query, args)
}

func (d *DB) QueryRow(query string, args sqlx.Args, dest ...any) error {
	return d.Safe.DB.QueryRow(query, args, dest...)
}
