package safe

import (
	"testing"

	"github.com/stregato/mio/core"
	"github.com/stregato/mio/security"
)

func TestKeys(t *testing.T) {
	alice := security.NewIdentityMust("alice")
	s := NewTestSafe(t, alice.Id, alice, "local", false)

	groups, err := s.UpdateGroup(AdminGroup, ChangeGrant, alice.Id)
	core.TestErr(t, err, "cannot update group: %v")
	core.Assert(t, len(groups) == 1, "wrong number of groups: %d", len(groups))
	groups, err = s.UpdateGroup(UserGroup, ChangeGrant, alice.Id)
	core.TestErr(t, err, "cannot update group: %v")
	core.Assert(t, len(groups) == 2, "wrong number of groups: %d", len(groups))

	keys, err := s.GetKeys(AdminGroup, 0)
	core.TestErr(t, err, "cannot get keys: %v")
	core.Assert(t, len(keys) != 0, "wrong number of keys: %d", len(keys))

	groups, err = s.UpdateGroup(UserGroup, ChangeRevoke, alice.Id)
	core.TestErr(t, err, "cannot update group: %v")
	core.Assert(t, len(groups) == 1, "wrong number of groups: %d", len(groups))

	_, err = s.GetKeys(UserGroup, 0)
	core.Assert(t, err != nil, "expected error")
}
