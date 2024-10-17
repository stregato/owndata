package db

import (
	_ "embed"
	"testing"

	"github.com/stregato/stash/lib/core"
	"github.com/stregato/stash/lib/safe"
	"github.com/stregato/stash/lib/security"

	"github.com/stregato/stash/lib/sqlx"
)

//go:embed test.sql
var testDdl string

func TestExec(t *testing.T) {
	alice := security.NewIdentityMust("alice")
	s := safe.NewTestSafe(t, alice, "local", alice.Id, true)

	groups, err := s.UpdateGroup(safe.UserGroup, safe.Grant, alice.Id)
	core.TestErr(t, err, "cannot update group: %v")
	core.Assert(t, len(groups) == 2, "wrong number of groups: %d", len(groups))

	db, err := Open(s, safe.UserGroup, DDLs{1.0: testDdl})
	core.TestErr(t, err, "cannot open db: %v")

	tx, err := db.Transaction()
	core.TestErr(t, err, "cannot start transaction: %v")

	_, err = tx.Exec("INSERT_TEST_DATA", sqlx.Args{"msg": "hello world", "cnt": 1, "ratio": 0.5, "bin": []byte{1, 2, 3}})
	core.TestErr(t, err, "cannot insert test data: %v")

	err = tx.Commit()
	core.TestErr(t, err, "cannot commit: %v")

	_, err = db.Sync()
	core.TestErr(t, err, "cannot sync: %v")

	db.Safe.DB.GetConnection().Exec("DELETE FROM db_test")
	db.Safe.DB.GetConnection().Exec("DELETE FROM STASH_STORE_TX")

	rows, err := db.Query("SELECT_TEST_DATA", sqlx.Args{})
	core.TestErr(t, err, "cannot select test data: %v")
	core.Assert(t, !rows.Next(), "unexpected rows")

	db.Safe.ResetTouch(DBDir)
	_, err = db.Sync()
	core.TestErr(t, err, "cannot sync: %v")

	rows, err = db.Query("SELECT_TEST_DATA", sqlx.Args{})
	core.TestErr(t, err, "cannot select test data: %v")
	core.Assert(t, rows.Next(), "no rows")

	var msg string
	var ratio float64
	var cnt int
	var bin string
	err = rows.Scan(&msg, &cnt, &ratio, &bin)
	core.TestErr(t, err, "cannot scan: %v")
	core.Assert(t, msg == "hello world", "unexpected msg: %s", msg)
	rows.Close()

	rows, err = db.Query("SELECT_TEST_DATA", sqlx.Args{})
	core.TestErr(t, err, "cannot select test data: %v")

	values, err := rows.NextRow()
	core.TestErr(t, err, "cannot get row: %v")
	core.Assert(t, len(values) == 4, "unexpected number of values: %d", len(values))
	core.Assert(t, values[0] == "hello world", "unexpected value: %s", values[0])

	rows.Close()
	db.Close()
	s.Close()
}

func TestCounter(t *testing.T) {
	alice := security.NewIdentityMust("alice")
	s := safe.NewTestSafe(t, alice, "local", alice.Id, true)

	groups, err := s.UpdateGroup(safe.UserGroup, safe.Grant, alice.Id)
	core.TestErr(t, err, "cannot update group: %v")
	core.Assert(t, len(groups) == 2, "wrong number of groups: %d", len(groups))

	db, err := Open(s, safe.UserGroup, DDLs{1.0: testDdl})
	core.TestErr(t, err, "cannot open db: %v")

	tx, err := db.Transaction()
	core.TestErr(t, err, "cannot start transaction: %v")

	err = tx.IncCounter("counter_test", "key1", 1)
	core.TestErr(t, err, "cannot increment counter: %v")

	err = tx.Commit()
	core.TestErr(t, err, "cannot commit: %v")

	_, err = db.Sync()
	core.TestErr(t, err, "cannot sync: %v")

	cnt, err := db.GetCounter("counter_test", "key1")
	core.TestErr(t, err, "cannot get counter: %v")
	core.Assert(t, cnt == 1, "unexpected counter: %d", cnt)
	db.Close()
	s.Close()
}
