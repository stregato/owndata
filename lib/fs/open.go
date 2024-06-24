package fs

import (
	"github.com/stregato/mio/lib/safe"
)

func Open(S *safe.Safe) (*FileSystem, error) {
	fs := &FileSystem{S: S}
	go fs.startUploadJob()
	return fs, nil
}
