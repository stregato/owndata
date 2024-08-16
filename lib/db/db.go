package db

import (
	"database/sql"
	_ "embed"

	"github.com/stregato/stash/lib/core"
	"github.com/stregato/stash/lib/stash"
)

type Database struct {
	*stash.Stash
	log       []Update
	tx        *sql.Tx
	groupName stash.GroupName
}

var (
	DBDir = "db"
)

type DDLs map[float32]string

func Open(s *stash.Stash, groupName stash.GroupName, ddls DDLs) (Database, error) {
	for version, ddl := range ddls {
		err := s.DB.Define(version, ddl)
		if err != nil {
			return Database{}, err
		}
	}

	core.Info("Opening database on safe %s with group %s", s.URL, groupName)

	return Database{s, nil, nil, groupName}, nil
}
