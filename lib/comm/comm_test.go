package comm

import (
	"testing"

	"github.com/stregato/mio/lib/core"
	"github.com/stregato/mio/lib/safe"
	"github.com/stregato/mio/lib/security"
)

func TestBroadcast(t *testing.T) {
	alice := security.NewIdentityMust("alice")
	s := safe.NewTestSafe(t, alice, "local", alice.Id, true)

	c := Open(s)
	err := c.Broadcast(safe.UserGroup, Message{Text: "hello world"})
	core.TestErr(t, err, "cannot broadcast to user group: %v")

	ms, err := c.Receive("")
	core.TestErr(t, err, "cannot receive: %v")

	core.Assert(t, len(ms) == 1, "received messages: %v", ms)
	core.Assert(t, ms[0].Text == "hello world", "received message: %v", ms[0])
}
