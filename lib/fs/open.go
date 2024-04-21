package fs

import (
	"github.com/stregato/mio/lib/safe"
)

func Open(S *safe.Safe) (*FS, error) {
	fs := &FS{S: S}
	go fs.startUploadJob()
	return fs, nil
}
