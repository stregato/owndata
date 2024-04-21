package db

import (
	"database/sql"
	_ "embed"

	"github.com/stregato/mio/lib/safe"
)

type PulseDB struct {
	*safe.Safe
	log       []Update
	tx        *sql.Tx
	groupName safe.GroupName
}

var (
	DBDir = "db"
)

type DDLs map[float32]string

func Open(s *safe.Safe, ddls DDLs, groupName safe.GroupName) (PulseDB, error) {
	for version, ddl := range ddls {
		err := s.DB.Define(version, ddl)
		if err != nil {
			return PulseDB{}, err
		}
	}

	return PulseDB{s, nil, nil, groupName}, nil
}
