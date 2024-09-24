package safe

import (
	"path"

	"github.com/stregato/stash/lib/config"
	"github.com/stregato/stash/lib/storage"
)

func (s *Safe) IsUpdated(dirs ...string) bool {
	name := path.Join(dirs...)
	name = path.Join(name, ".touch")

	_, lastChange, _, ok := config.GetConfigValue(s.DB, config.GuardDomain, path.Join(s.ID, name))
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

func (s *Safe) Touch(dirs ...string) error {
	name := path.Join(dirs...)
	name = path.Join(name, ".touch")

	err := storage.WriteFile(s.Store, name, []byte{})
	if err != nil {
		return err
	}

	st, err := s.Store.Stat(name)
	if err != nil {
		return err
	}

	err = config.SetConfigValue(s.DB, config.GuardDomain, path.Join(s.ID, name), "", st.ModTime().UnixMilli(), nil)
	return err
}
