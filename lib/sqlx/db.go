package sqlx

import (
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stregato/mio/lib/core"

	"github.com/sirupsen/logrus"
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

//go:embed ddl1_0.sql
var ddl1_0 string

func Open(dbPath string) (*DB, error) {
	if dbPath != MemoryDB {
		_, err := os.Stat(dbPath)
		if errors.Is(err, os.ErrNotExist) {
			err := os.WriteFile(dbPath, []byte{}, 0644)
			if core.IsErr(err, "cannot create SQLite db in %s: %v", dbPath, err) {
				return nil, err
			}

		} else if err != nil {
			logrus.Errorf("cannot access SQLite db file %s: %v", dbPath, err)
		}
	}

	db, err := sql.Open("sqlite3", dbPath)
	if core.IsErr(err, "cannot open SQLite db in %s: %v", dbPath, err) {
		return nil, err
	}

	d := &DB{DbPath: dbPath,
		Db:       db,
		versions: map[string]float32{},
		queries:  map[string]string{},
		stmts:    map[string]*sql.Stmt{},
	}

	err = d.Define(1.0, ddl1_0)
	if core.IsErr(err, "cannot define SQLite db in %s: %v", dbPath, err) {
		return nil, err
	}
	return d, nil
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

func NewTestDB(t *testing.T, persistent bool) *DB {
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
	core.TestErr(t, err, "invalid dll: %v")

	return db
}
