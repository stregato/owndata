package db

import (
	"database/sql"
	_ "embed"

	"github.com/stregato/stash/lib/core"
	"github.com/stregato/stash/lib/safe"
)

type Database struct {
	*safe.Safe
	log       []Update
	tx        *sql.Tx
	groupName safe.GroupName
}

var (
	DBDir = "db"
)

type DDLs map[float32]string

func Open(s *safe.Safe, groupName safe.GroupName, ddls DDLs) (Database, error) {
	for version, ddl := range ddls {
		err := s.DB.Define(version, ddl)
		if err != nil {
			return Database{}, err
		}
	}

	core.Info("Opening database on safe %s with group %s", s.URL, groupName)

	return Database{s, nil, nil, groupName}, nil
}
