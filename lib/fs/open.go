package fs

import "github.com/stregato/stash/lib/stash"

func Open(S *stash.Stash) (*FileSystem, error) {
	fs := &FileSystem{S: S}
	go fs.startUploadJob()
	return fs, nil
}
