package db

import (
	"database/sql"
	_ "embed"

	"github.com/stregato/mio/safe"
)

type DB struct {
	s         *safe.Safe
	log       []Update
	tx        *sql.Tx
	groupName safe.GroupName
}

var (
	DBDir = "db"
)

//go:embed tx.sql
var txDdl string

type DDLs map[float32]string

func Open(s *safe.Safe, ddls DDLs, groupName safe.GroupName) (DB, error) {

	err := s.Db.Define(1.0, txDdl)
	if err != nil {
		return DB{}, err
	}
	for version, ddl := range ddls {
		err := s.Db.Define(version, ddl)
		if err != nil {
			return DB{}, err
		}
	}

	return DB{s, nil, nil, groupName}, nil
}
