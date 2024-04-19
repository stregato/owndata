package safe

import (
	"testing"

	"github.com/stregato/mio/lib/core"
	"github.com/stregato/mio/lib/security"
)

func TestKeys(t *testing.T) {
	alice := security.NewIdentityMust("alice")
	s := NewTestSafe(t, alice, "local", alice.Id, false)

	groups, err := s.UpdateGroup(AdminGroup, Grant, alice.Id)
	core.TestErr(t, err, "cannot update group: %v")
	core.Assert(t, len(groups) == 1, "wrong number of groups: %d", len(groups))
	groups, err = s.UpdateGroup(UserGroup, Grant, alice.Id)
	core.TestErr(t, err, "cannot update group: %v")
	core.Assert(t, len(groups) == 2, "wrong number of groups: %d", len(groups))

	keys, err := s.GetKeys(AdminGroup, 0)
	core.TestErr(t, err, "cannot get keys: %v")
	core.Assert(t, len(keys) != 0, "wrong number of keys: %d", len(keys))

	groups, err = s.UpdateGroup(UserGroup, Revoke, alice.Id)
	core.TestErr(t, err, "cannot update group: %v")
	core.Assert(t, len(groups) == 1, "wrong number of groups: %d", len(groups))

	_, err = s.GetKeys(UserGroup, 0)
	core.Assert(t, err != nil, "expected error")
}
