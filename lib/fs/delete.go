package fs

import (
	"path"

	"github.com/stregato/stash/lib/core"
	"github.com/stregato/stash/lib/sqlx"
)

func (fs *FileSystem) Delete(name string) error {
	file, err := fs.Stat(name)
	if err != nil {
		return err
	}

	err = fs.S.Store.Delete(path.Join(HeadersDir, hashDir(file.Dir), file.ID.String()))
	if err != nil {
		return err
	}

	err = fs.S.Store.Delete(path.Join(DataDir, file.ID.String()))
	if err != nil {
		return err
	}

	_, err = fs.S.DB.Exec("MIO_DELETE_FILE", sqlx.Args{"safeID": fs.S.ID, "id": file.ID.Uint64()})
	if err != nil {
		return err
	}

	dir := core.Dir(file.Dir)
	_, err = fs.S.DB.Exec("MIO_DELETE_DIR", sqlx.Args{"safeID": fs.S.ID, "dir": dir})
	if err != nil {
		return err
	}

	return nil
}
