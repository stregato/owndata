package fs

import (
	"os"
	"strings"

	"github.com/stregato/stash/lib/core"
	"github.com/stregato/stash/lib/sqlx"
)

func (f *FileSystem) Stat(name string) (File, error) {
	dir, name := core.SplitPath(name)

	var file File
	var tags string
	err := f.S.DB.QueryRow("STASH_GET_FILE_BY_NAME", sqlx.Args{"safeID": f.S.ID, "dir": dir, "name": name},
		&file.ID, &file.GroupName, &tags, &file.ModTime, &file.Size, &file.Creator, &file.Attributes,
		&file.LocalCopy, &file.CopyTime, &file.EncryptionKey)
	if err == sqlx.ErrNoRows {
		return File{}, os.ErrNotExist
	}
	if err != nil {
		return File{}, err
	}
	file.Name = name
	file.Dir = dir
	file.Tags = strings.Split(strings.TrimSpace(tags), " ")
	file.IsDir = file.ID == 0
	return file, nil
}
