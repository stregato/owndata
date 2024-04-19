package safe

import (
	"testing"

	"github.com/stregato/mio/lib/core"
	"github.com/stregato/mio/lib/security"
)

func TestOpen(t *testing.T) {
	alice := security.NewIdentityMust("alice")

	s := NewTestSafe(t, alice, "s3", alice.Id, false)
	core.Assert(t, s != nil, "Safe not created")

	s.Close()
}
