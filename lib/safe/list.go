package safe

import (
	"github.com/stregato/mio/lib/sqlx"
)

// List returns the list of safes in the provided DB
func List(db sqlx.DB) ([]Safe, error) {
	db.Exec("", sqlx.Args{})
	return nil, nil
}
