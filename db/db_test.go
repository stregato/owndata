package db

import (
	_ "embed"
	"testing"

	"github.com/stregato/mio/core"
	"github.com/stregato/mio/safe"
	"github.com/stregato/mio/security"
	"github.com/stregato/mio/sql"
)

//go:embed test.sql
var testDdl string

func TestExec(t *testing.T) {
	alice := security.NewIdentityMust("alice")
	s := safe.NewTestSafe(t, alice.Id, alice, "local", true)

	groups, err := s.UpdateGroup(safe.UserGroup, safe.ChangeGrant, alice.Id)
	core.TestErr(t, err, "cannot update group: %v")
	core.Assert(t, len(groups) == 2, "wrong number of groups: %d", len(groups))

	db, err := Open(s, DDLs{1.0: testDdl}, safe.UserGroup)
	core.TestErr(t, err, "cannot open db: %v")

	_, err = db.Exec("INSERT_TEST_DATA", sql.Args{"msg": "hello world"})
	core.TestErr(t, err, "cannot insert test data: %v")

	err = db.Commit()
	core.TestErr(t, err, "cannot commit: %v")

	db.s.Db.GetConnection().Exec("DELETE FROM db_test")
	db.s.Db.GetConnection().Exec("DELETE FROM MIO_STORE_TX")

	rows, err := db.Query("SELECT_TEST_DATA", sql.Args{})
	core.TestErr(t, err, "cannot select test data: %v")
	core.Assert(t, !rows.Next(), "unexpected rows")

	db.Sync()

	rows, err = db.Query("SELECT_TEST_DATA", sql.Args{})
	core.TestErr(t, err, "cannot select test data: %v")
	core.Assert(t, rows.Next(), "no rows")

	var msg string
	err = rows.Scan(&msg)
	core.TestErr(t, err, "cannot scan: %v")
	core.Assert(t, msg == "hello world", "unexpected msg: %s", msg)

}
