package db

import "github.com/stregato/mio/lib/sqlx"

func (d *PulseDB) Query(query string, args sqlx.Args) (sqlx.Rows, error) {
	args["safe"] = d.Safe.ID
	args["groupName"] = string(d.groupName)
	return d.Safe.DB.Query(query, args)
}

func (d *PulseDB) QueryRow(query string, args sqlx.Args, dest ...any) error {
	args["safe"] = d.Safe.ID
	args["groupName"] = string(d.groupName)
	return d.Safe.DB.QueryRow(query, args, dest...)
}
