package safe

import (
	"testing"

	"github.com/stregato/mio/core"
	"github.com/stregato/mio/security"
)

func TestGroupChain(t *testing.T) {
	alice, err := security.NewIdentity("alice")
	core.TestErr(t, err, "cannot create identity")

	bob, err := security.NewIdentity("bob")
	core.TestErr(t, err, "cannot create identity")

	carl, err := security.NewIdentity("carl")
	core.TestErr(t, err, "cannot create identity")

	h := security.NewHash([]byte(nil))

	gc1, err := newGroupChange(AdminGroup, bob.Id, ChangeGrant, h, alice)
	core.TestErr(t, err, "cannot create group change")
	groups := Groups{AdminGroup: core.NewSet(alice.Id)}
	err = applyChange(gc1, groups)
	t.Log(groups)

	core.TestErr(t, err, "cannot resolve group chain: %v")
	core.Assert(t, len(groups) == 1, "wrong number of groups")
	core.Assert(t, len(groups[AdminGroup]) == 2, "wrong number of users in group")

	gc2, err := newGroupChange(UserGroup, carl.Id, ChangeGrant, h, bob)
	core.TestErr(t, err, "cannot create group change")
	err = applyChange(gc2, groups)
	core.TestErr(t, err, "cannot resolve group chain: %v")
	core.Assert(t, len(groups) == 2, "wrong number of groups")
	core.Assert(t, len(groups[UserGroup]) == 1, "wrong number of users in group")

	gc3, err := newGroupChange(AdminGroup, alice.Id, ChangeRevoke, h, alice)
	core.TestErr(t, err, "cannot create group change")
	err = applyChange(gc3, groups)
	core.TestErr(t, err, "cannot resolve group chain: %v")
	core.Assert(t, len(groups[AdminGroup]) == 1, "wrong number of users in group")

	gc4, err := newGroupChange(AdminGroup, bob.Id, ChangeRevoke, h, alice)
	core.TestErr(t, err, "cannot create group change")
	err = applyChange(gc4, groups)
	core.Assert(t, err != nil, "cannot resolve group chain: %v")
	core.Assert(t, len(groups[AdminGroup]) == 1, "wrong number of users in group")

}

func TestGroupSync(t *testing.T) {
	alice := security.NewIdentityMust("alice")
	s := NewTestSafe(t, alice.Id, alice, "local", true)

	groups, err := s.UpdateGroup(UserGroup, ChangeGrant, alice.Id)
	core.TestErr(t, err, "cannot update group: %v")
	core.Assert(t, len(groups) == 1, "wrong number of groups: %d", len(groups))
	groups, err = s.UpdateGroup(UserGroup, ChangeGrant, alice.Id)
	core.TestErr(t, err, "cannot update group: %v")
	core.Assert(t, len(groups) == 2, "wrong number of groups: %d", len(groups))
	core.Assert(t, len(groups[AdminGroup]) == 1, "wrong number of users in group")
	core.Assert(t, len(groups[UserGroup]) == 1, "wrong number of users in group")

	bob := security.NewIdentityMust("bob")
	s2 := NewTestSafe(t, alice.Id, bob, "local", false)

	groups, err = s2.GetGroups()
	core.TestErr(t, err, "cannot sync groups: %v")

	core.Assert(t, len(groups) == 2, "wrong number of groups: %d", len(groups))

	_, err = s2.UpdateGroup(UserGroup, ChangeGrant, s2.CurrentUser.Id)
	core.Assert(t, err != nil, "cannot update group: %v")
}
