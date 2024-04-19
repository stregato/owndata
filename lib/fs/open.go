package fs

import (
	"github.com/stregato/mio/lib/safe"
	"github.com/stregato/mio/lib/storage"
)

func Open(S *safe.Safe) (*FS, error) {
	var config Config
	err := storage.ReadYAML(S.Store, ConfigPath, &config, nil)
	if err != nil {
		return nil, err
	}

	return &FS{S: S,
		StoreUrl: S.Store.ID(),
		Config:   config}, nil
}
