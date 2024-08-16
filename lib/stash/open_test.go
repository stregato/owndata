package stash

import (
	"testing"

	"github.com/stregato/stash/lib/core"
	"github.com/stregato/stash/lib/security"
)

func TestOpen(t *testing.T) {
	alice := security.NewIdentityMust("alice")

	s := NewTestSafe(t, alice, "s3", alice.Id, false)
	core.Assert(t, s != nil, "Safe not created")

	s.Close()
}
