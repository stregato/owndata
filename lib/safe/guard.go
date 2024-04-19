package safe

import (
	"path"

	"github.com/stregato/mio/lib/config"
	"github.com/stregato/mio/lib/storage"
)

func (s *Safe) IsUpdated(dir string) bool {
	var (
		name = path.Join(dir, ".touch")
	)

	_, lastChange, _, ok := config.GetConfigValue(s.DB, config.GuardDomain, path.Join(s.ID, dir))
	if !ok {
		return true
	}
	st, err := s.Store.Stat(name)
	if err != nil {
		return true
	}
	fileChange := st.ModTime().UnixMilli()

	return fileChange > lastChange
}

func (s *Safe) Touch(dir string) error {
	var (
		name = path.Join(dir, ".touch")
	)
	err := storage.WriteFile(s.Store, name, []byte{})
	if err != nil {
		return err
	}

	st, err := s.Store.Stat(name)
	if err != nil {
		return err
	}

	err = config.SetConfigValue(s.DB, config.GuardDomain, path.Join(s.ID, dir), "", st.ModTime().UnixMilli(), nil)
	return err
}
