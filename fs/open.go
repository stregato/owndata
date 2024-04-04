package fs

import (
	"github.com/stregato/mio/safe"
	"github.com/stregato/mio/storage"
)

func Open(S *safe.Safe) (*FS, error) {
	var config Config
	err := storage.ReadYAML(S.Store, ConfigPath, &config, nil)
	if err != nil {
		return nil, err
	}

	return &FS{S: S,
		StoreUrl: S.Store.Url(),
		Config:   config}, nil
}
