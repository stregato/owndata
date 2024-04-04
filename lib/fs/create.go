package fs

import (
	"fmt"

	"github.com/stregato/mio/lib/safe"
	"github.com/stregato/mio/lib/storage"
)

type CreateOptions struct {
	Quota       int64  // Quota is the maximum size of the filesystem in bytes
	Overwrite   bool   // Overwrite is whether to overwrite the filesystem if it already exists
	Description string // Description is a human-readable description of the filesystem
}

func CreateFS(c *safe.Safe, options CreateOptions) (*FS, error) {
	_, err := c.Store.Stat(ConfigPath)
	if err == nil {
		if options.Overwrite {
			err = c.Store.Delete(FSDir)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf(ErrExists, c.Store.Url())
		}
	}

	fsc := Config{
		Quota:       options.Quota,
		Description: options.Description,
	}
	err = storage.WriteYAML(c.Store, ConfigPath, fsc, nil)
	if err != nil {
		return nil, err
	}

	return &FS{
		S:        c,
		StoreUrl: c.Store.Url(),
	}, nil
}
