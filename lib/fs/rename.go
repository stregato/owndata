package fs

import (
	"os"
	"path"

	"github.com/stregato/mio/lib/sqlx"
)

func (fs *FileSystem) Rename(old, new string) (File, error) {
	file, err := fs.Stat(old)
	if err != nil {
		return File{}, err
	}

	if file.ID == 0 {
		return File{}, os.ErrPermission
	}

	err = fs.S.Store.Delete(path.Join(HeadersDir, hashDir(file.Dir), file.ID.String()))
	if err != nil {
		return File{}, err
	}

	oldName := file.Name
	oldDir := file.Dir
	file.Name = path.Base(new)
	file.Dir = path.Dir(new)
	if file.Dir == "." {
		file.Dir = ""
	}
	_, err = writeHeader(fs.S, file)
	if err != nil {
		return File{}, err
	}

	_, err = fs.S.DB.Exec("MIO_RENAME_FILE", sqlx.Args{"safeID": fs.S.ID, "oldDir": oldDir, "oldName": oldName,
		"newDir": file.Dir, "newName": file.Name, "id": file.ID.Uint64()})
	if err != nil {
		return File{}, err
	}

	return file, nil
}
