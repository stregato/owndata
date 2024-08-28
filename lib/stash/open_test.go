package stash

import (
	"testing"

	"github.com/stregato/stash/lib/core"
	"github.com/stregato/stash/lib/security"
	"github.com/stregato/stash/lib/sqlx"
)

func TestOpen(t *testing.T) {
	alice := security.NewIdentityMust("alice")

	s := NewTestSafe(t, alice, "s3", alice.Id, false)
	core.Assert(t, s != nil, "Safe not created")

	s.Close()
}

func TestCreate(t *testing.T) {
	alice := security.NewIdentityMust("alice")
	db := sqlx.NewTestDB(t, false)

	url := "file:///tmp/stash/" + alice.Id.String() + "/test"
	s, err := Create(db, alice, url, Config{})
	core.Assert(t, err == nil, "Create failed: %v", err)

	s.Close()

	s, err = Open(db, alice, url)
	core.Assert(t, err == nil, "Open failed: %v", err)
	s.Close()
}
