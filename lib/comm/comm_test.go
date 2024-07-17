package comm

import (
	"bytes"
	"os"
	"testing"

	"github.com/stregato/mio/lib/core"
	"github.com/stregato/mio/lib/safe"
	"github.com/stregato/mio/lib/security"
	"github.com/stregato/mio/lib/sqlx"
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

	file, err := os.CreateTemp("", "mio-test-send.txt")
	core.TestErr(t, err, "cannot create temp file: %v")

	_, err = file.WriteString("hello world")
	core.TestErr(t, err, "cannot write to temp file: %v")
	file.Close()

	err = c.Broadcast(safe.UserGroup, Message{File: file.Name()})
	core.TestErr(t, err, "cannot broadcast file to user group: %v")

	file, err = os.CreateTemp("", "mio-test-recv.txt")
	core.TestErr(t, err, "cannot create temp file: %v")
	file.Close()

	ms, err = c.Receive("")
	core.TestErr(t, err, "cannot receive: %v")

	core.Assert(t, len(ms) == 1, "received messages: %v", ms)
	core.Assert(t, ms[0].File != "", "received message: %v", ms[0])

	err = c.DownloadFile(ms[0], file.Name())
	core.TestErr(t, err, "cannot download file: %v")

	data, err := os.ReadFile(file.Name())
	core.TestErr(t, err, "cannot read downloaded file: %v")
	core.Assert(t, string(data) == "hello world", "downloaded file: %v", string(data))
	s.Close()
}

func TestSend(t *testing.T) {
	alice := security.NewIdentityMust("alice")
	bob := security.NewIdentityMust("bob")
	s := safe.NewTestSafe(t, alice, "local", alice.Id, true)
	_, err := s.UpdateGroup(safe.UserGroup, safe.Grant, bob.Id)
	core.TestErr(t, err, "cannot update group: %v")

	c := Open(s)
	data := core.GenerateRandomBytes(100)
	err = c.Send(bob.Id, Message{Data: data})
	core.TestErr(t, err, "cannot send to bob: %v")

	s.Close()

	db := sqlx.NewTestDB(t, true)

	s, err = safe.Open(db, bob, s.URL)
	core.TestErr(t, err, "cannot open safe %s", s.URL)

	c = Open(s)
	ms, err := c.Receive("")
	core.TestErr(t, err, "cannot receive: %v")
	core.Assert(t, len(ms) == 1, "received messages: %v", ms)
	core.Assert(t, bytes.Equal(ms[0].Data, data), "received message: %v", ms[0])

	c.Rewind(bob.Id.String(), 0)
	ms, err = c.Receive("")
	core.TestErr(t, err, "cannot receive: %v")
	core.Assert(t, len(ms) == 1, "received messages: %v", ms)

	s.Close()
}
