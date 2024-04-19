package cmd

import (
	"context"
	"os"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/stregato/mio/lib/safe"
)

type Mount struct {
	fs.Inode
}

func (m *Mount) OnAdd(ctx context.Context) {
	ch := m.NewPersistentInode(
		ctx, &fs.MemRegularFile{
			Data: []byte("file.txt"),
			Attr: fuse.Attr{
				Mode: 0644,
			},
		}, fs.StableAttr{Ino: 2})
	m.AddChild("file.txt", ch, false)
}

func mountFS(s *safe.Safe, path string) error {
	os.MkdirAll(path, 0755)

	s.Close()
	server, err := fs.Mount(path, &Mount{}, &fs.Options{})
	if err != nil {
		return err
	}
	server.Wait()
	return nil
}
