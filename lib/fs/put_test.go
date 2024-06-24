package fs

import (
	"os"
	"testing"
	"time"

	"github.com/stregato/mio/lib/core"
	"github.com/stregato/mio/lib/safe"
	"github.com/stregato/mio/lib/security"
)

func TestPutData(t *testing.T) {
	alice := security.NewIdentityMust("alice")
	s := safe.NewTestSafe(t, alice, "local", alice.Id, true)

	f, err := Open(s)
	core.TestErr(t, err, "cannot open fs: %v")
	defer f.Close()

	_, err = f.PutData("test", []byte("hello world"), PutOptions{})
	core.TestErr(t, err, "cannot put data: %v")

	_, err = f.PutData("sub/test", []byte("hello world"), PutOptions{})
	core.TestErr(t, err, "cannot put data: %v")

	data, err := f.GetData("sub/test", GetOptions{})
	core.TestErr(t, err, "cannot get data: %v")
	core.Assert(t, string(data) == "hello world", "unexpected data: %s", data)
}
func TestPutFile(t *testing.T) {
	alice := security.NewIdentityMust("alice")
	s := safe.NewTestSafe(t, alice, "local", alice.Id, true)

	f, err := Open(s)
	core.TestErr(t, err, "cannot open fs: %v")
	defer f.Close()

	tf, err := os.CreateTemp("", "mio-test")
	core.TestErr(t, err, "cannot create temp file: %v")
	tf.WriteString("hello world")
	tf.Close()
	defer os.Remove(tf.Name())

	_, err = f.PutFile("test", tf.Name(), PutOptions{})
	core.TestErr(t, err, "cannot put file: %v")

	data, err := f.GetData("test", GetOptions{})
	core.TestErr(t, err, "cannot get data: %v")
	core.Assert(t, string(data) == "hello world", "unexpected data: %s", data)
}

func TestAsyncPutData(t *testing.T) {
	alice := security.NewIdentityMust("alice")
	s := safe.NewTestSafe(t, alice, "local", alice.Id, true)

	f, err := Open(s)
	core.TestErr(t, err, "cannot open fs: %v")
	defer f.Close()

	file, err := f.PutData("test", []byte("hello world"), PutOptions{Async: true})
	core.TestErr(t, err, "cannot put data: %v")

	for !f.HasPutCompleted(file.ID) {
		time.Sleep(100 * time.Millisecond)
		core.Info("waiting for async put to complete")
	}

	data, err := f.GetData("test", GetOptions{})
	core.TestErr(t, err, "cannot get data: %v")
	core.Assert(t, string(data) == "hello world", "unexpected data: %s", data)
}
