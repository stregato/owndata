package db

import "github.com/stregato/stash/lib/sqlx"

func (d *Database) Query(query string, args sqlx.Args) (sqlx.Rows, error) {
	return d.Stash.DB.Query(query, args)
}

func (d *Database) QueryRow(query string, args sqlx.Args, dest ...any) error {
	return d.Stash.DB.QueryRow(query, args, dest...)
}
