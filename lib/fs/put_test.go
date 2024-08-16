package fs

import (
	"os"
	"testing"
	"time"

	"github.com/stregato/stash/lib/core"
	"github.com/stregato/stash/lib/security"
	"github.com/stregato/stash/lib/stash"
)

func TestPutData(t *testing.T) {
	alice := security.NewIdentityMust("alice")
	s := stash.NewTestSafe(t, alice, "local", alice.Id, true)

	f, err := Open(s)
	core.TestErr(t, err, "cannot open fs: %v")
	defer f.Close()

	file, err := f.PutData("test", []byte("hello world"), PutOptions{})
	core.TestErr(t, err, "cannot put data: %v")

	_, err = f.PutData("sub/test", []byte("hello world"), PutOptions{})
	core.TestErr(t, err, "cannot put data: %v")

	files, err := f.List("", ListOptions{})
	core.TestErr(t, err, "cannot list files: %v")
	core.Assert(t, len(files) == 2, "unexpected number of files: %d", len(files))
	core.Assert(t, files[0].Name == "sub", "unexpected file: %s", files[0].Name)
	core.Assert(t, files[1].Name == "test", "unexpected file: %s", files[1].Name)
	core.Assert(t, files[1].ID == file.ID, "unexpected file id: %s", files[1].ID)

	data, err := f.GetData("sub/test", GetOptions{})
	core.TestErr(t, err, "cannot get data: %v")
	core.Assert(t, string(data) == "hello world", "unexpected data: %s", data)
}
func TestPutFile(t *testing.T) {
	alice := security.NewIdentityMust("alice")
	s := stash.NewTestSafe(t, alice, "local", alice.Id, true)

	f, err := Open(s)
	core.TestErr(t, err, "cannot open fs: %v")
	defer f.Close()

	tf, err := os.CreateTemp("", "stash-test")
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
	s := stash.NewTestSafe(t, alice, "local", alice.Id, true)

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
