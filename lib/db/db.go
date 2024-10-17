package db

import (
	"database/sql"
	_ "embed"

	"github.com/stregato/stash/lib/core"
	"github.com/stregato/stash/lib/safe"
	"github.com/stregato/stash/lib/sqlx"
)

type DB struct {
	*safe.Safe
	log       []Update
	tx        *sql.Tx
	groupName safe.GroupName
	counters  map[string]bool
}

var (
	DBDir = "db"
)

type DDLs map[float32]string

func Open(s *safe.Safe, groupName safe.GroupName, ddls DDLs) (DB, error) {
	for version, ddl := range ddls {
		err := s.DB.Define(version, ddl)
		if err != nil {
			return DB{}, err
		}
	}

	counters, err := findCounterTables(s.DB)
	if err != nil {
		return DB{}, err
	}

	core.Info("Opening database on safe %s with group %s", s.URL, groupName)

	return DB{s, nil, nil, groupName, counters}, nil
}

func findCounterTables(db *sqlx.DB) (map[string]bool, error) {
	rows, err := db.Db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name LIKE 'counter_%'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var counters = map[string]bool{}
	for rows.Next() {
		var table string
		err = rows.Scan(&table)
		if err != nil {
			return nil, err
		}
		counters[table] = true
	}
	return counters, nil
}
