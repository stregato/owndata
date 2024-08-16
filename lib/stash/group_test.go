package stash

import (
	"testing"

	"github.com/stregato/stash/lib/core"
	"github.com/stregato/stash/lib/security"
)

func TestGroupChain(t *testing.T) {
	alice, err := security.NewIdentity("alice")
	core.TestErr(t, err, "cannot create identity")

	bob, err := security.NewIdentity("bob")
	core.TestErr(t, err, "cannot create identity")

	carl, err := security.NewIdentity("carl")
	core.TestErr(t, err, "cannot create identity")

	groups := Groups{}
	gc0 := GroupChange{AdminGroup, Grant, alice.Id, "", nil, 0}
	gc0, err = signGroupChange(gc0, nil, alice)
	core.TestErr(t, err, "cannot create group change")
	err = applyChange(gc0, groups, alice.Id)
	core.TestErr(t, err, "cannot resolve group chain: %v")
	core.Assert(t, len(groups) == 1, "wrong number of groups")

	gc1 := GroupChange{AdminGroup, Grant, bob.Id, "", nil, 0}
	gc1, err = signGroupChange(gc1, nil, alice)
	core.TestErr(t, err, "cannot create group change")
	err = applyChange(gc1, groups, alice.Id)
	t.Log(groups)

	core.TestErr(t, err, "cannot resolve group chain: %v")
	core.Assert(t, len(groups) == 1, "wrong number of groups")
	core.Assert(t, len(groups[AdminGroup]) == 2, "wrong number of users in group")

	gc2 := GroupChange{UserGroup, Grant, carl.Id, "", nil, 0}
	gc2, err = signGroupChange(gc2, gc1.Signature, bob)
	core.TestErr(t, err, "cannot create group change")
	err = applyChange(gc2, groups, alice.Id)
	core.TestErr(t, err, "cannot resolve group chain: %v")
	core.Assert(t, len(groups) == 2, "wrong number of groups")
	core.Assert(t, len(groups[UserGroup]) == 1, "wrong number of users in group")

	gc3 := GroupChange{AdminGroup, Revoke, bob.Id, "", nil, 0}
	gc3, err = signGroupChange(gc3, gc2.Signature, bob)
	core.TestErr(t, err, "cannot create group change")
	err = applyChange(gc3, groups, alice.Id)
	core.TestErr(t, err, "cannot resolve group chain: %v")
	core.Assert(t, len(groups[AdminGroup]) == 1, "wrong number of users in group")

	gc4 := GroupChange{UserGroup, Revoke, carl.Id, "", nil, 0}
	gc4, err = signGroupChange(gc4, gc3.Signature, bob)
	core.TestErr(t, err, "cannot create group change")
	err = applyChange(gc4, groups, alice.Id)
	core.Assert(t, err != nil, "bob should not be able to revoke carl when he is not an admin")
	core.Assert(t, len(groups[AdminGroup]) == 1, "wrong number of users in group")

	err = validateGroupChain(gc1, nil)
	core.TestErr(t, err, "cannot validate group chain: %v")

	err = validateGroupChain(gc2, gc1.Signature)
	core.TestErr(t, err, "cannot validate group chain: %v")

	err = validateGroupChain(gc3, gc2.Signature)
	core.TestErr(t, err, "cannot validate group chain: %v")

	err = validateGroupChain(gc4, gc3.Signature)
	core.TestErr(t, err, "cannot validate group chain: %v")
}

func TestGroupSync(t *testing.T) {
	alice := security.NewIdentityMust("alice")
	s := NewTestSafe(t, alice, "local", alice.Id, true)

	groups, err := s.UpdateGroup(UserGroup, Grant, alice.Id)
	core.TestErr(t, err, "cannot update group: %v")
	core.Assert(t, len(groups) == 2, "wrong number of groups: %d", len(groups))
	core.Assert(t, len(groups[AdminGroup]) == 1, "wrong number of users in group")
	core.Assert(t, len(groups[UserGroup]) == 1, "wrong number of users in group")

	bob := security.NewIdentityMust("bob")
	s2 := NewTestSafe(t, bob, "local", alice.Id, false)

	groups, err = s2.GetGroups()
	core.TestErr(t, err, "cannot sync groups: %v")

	core.Assert(t, len(groups) == 2, "wrong number of groups: %d", len(groups))

	_, err = s2.UpdateGroup(UserGroup, Grant, s2.Identity.Id)
	core.Assert(t, err != nil, "cannot update group: %v")
}
