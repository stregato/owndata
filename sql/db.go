package sql

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/sirupsen/logrus"
	"github.com/stregato/mio/core"
)

type DB struct {
	DbPath string
	Db     *sql.DB

	queries  map[string]string
	stmts    map[string]*sql.Stmt
	versions map[string]float32
}

var Default *DB

var MemoryDB = ":memory:"

func Open(dbPath string) (DB, error) {
	if dbPath != MemoryDB {
		_, err := os.Stat(dbPath)
		if errors.Is(err, os.ErrNotExist) {
			err := os.WriteFile(dbPath, []byte{}, 0644)
			if core.IsErr(err, "cannot create SQLite db in %s: %v", dbPath, err) {
				return DB{}, err
			}

		} else if err != nil {
			logrus.Errorf("cannot access SQLite db file %s: %v", dbPath, err)
		}
	}

	db, err := sql.Open("sqlite3", dbPath)
	if core.IsErr(err, "cannot open SQLite db in %s: %v", dbPath, err) {
		return DB{}, err
	}
	return DB{DbPath: dbPath,
		Db:       db,
		versions: map[string]float32{},
		queries:  map[string]string{},
		stmts:    map[string]*sql.Stmt{},
	}, nil
}

func (db *DB) Close() error {
	return db.Db.Close()
}

func (db *DB) Delete() error {
	if db.DbPath != MemoryDB {
		return os.Remove(db.DbPath)
	} else {
		return nil
	}
}

func (db *DB) GetConnection() *sql.DB {
	return db.Db
}

func NewTestDB(t *testing.T, persistent bool) DB {
	var name string
	if persistent {
		name = path.Join(os.TempDir(), fmt.Sprintf("test-%d.db", os.Getpid()))
	} else {
		name = MemoryDB
	}
	db, err := Open(name)
	if err != nil {
		panic(err)
	}
	data, err := os.ReadFile("../db.sql")
	core.TestErr(t, err, "cannot read dll: %v")
	dll := string(data)

	err = db.Define(1.0, dll)
	core.TestErr(t, err, "invalid dll: %v")

	return db
}
