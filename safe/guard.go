package safe

import (
	"path"

	"github.com/stregato/mio/sql"
	"github.com/stregato/mio/storage"
)

const GuardNode = "guard"

func hasStoreChanged(db sql.DB, store storage.Store, dir string) bool {
	var (
		name = path.Join(dir, ".touch")
	)

	_, lastChange, _, ok := GetConfig(db, GuardNode, path.Join(store.Url(), dir))
	if !ok {
		return true
	}
	st, err := store.Stat(name)
	if err != nil {
		return true
	}
	fileChange := st.ModTime().UnixMilli()

	return fileChange > lastChange
}

func storeHasChanged(db sql.DB, store storage.Store, dir string) error {
	var (
		name = path.Join(dir, ".touch")
	)
	err := storage.WriteFile(store, name, []byte{})
	if err != nil {
		return err
	}

	st, err := store.Stat(name)
	if err != nil {
		return err
	}

	err = SetConfig(db, GuardNode, path.Join(store.Url(), dir), "", st.ModTime().UnixMilli(), nil)
	return err
}
