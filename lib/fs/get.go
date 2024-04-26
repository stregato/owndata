package fs

import (
	"bytes"
	"io"
	"os"
	"path"

	"github.com/stregato/mio/lib/core"
	"github.com/stregato/mio/lib/sqlx"
)

type GetOptions struct {
	Async bool
}

func (f *FS) GetData(src string, options GetOptions) ([]byte, error) {
	var dest bytes.Buffer

	if options.Async {
		return nil, core.Errorf("GetData does not support async mode. Use GetFile instead")
	}

	file, err := f.getFileRecord(src)
	if err != nil {
		return nil, err
	}

	err = f.getSync(file, "", &dest)
	if err != nil {
		return nil, err
	}

	return dest.Bytes(), nil
}

func (f *FS) GetFile(src, dest string, options GetOptions) (File, error) {
	file, err := f.getFileRecord(src)
	if err != nil {
		return File{}, err
	}

	if options.Async {
		_, err = f.S.DB.Exec("MIO_INSERT_FILE_ASYNC", sqlx.Args{"id": file.ID, "safeID": f.S.ID,
			"operation": "get", "file": file, "data": nil, "localCopy": dest, "deleteSrc": false})
		if err != nil {
			return File{}, err
		}
		triggerAsync <- file.ID
		return file, nil
	}

	err = f.getSync(file, dest, nil)
	if err != nil {
		return File{}, err
	}
	return file, nil
}

func (f *FS) getFileRecord(src string) (File, error) {
	dir, name := path.Split(src)

	var file File
	err := f.S.DB.QueryRow("MIO_GET_FILE_BY_NAME", sqlx.Args{"safeID": f.S.ID, "dir": dir, "name": name},
		&file.ID, &file.Dir, &file.GroupName, &file.Tags, &file.ModTime, &file.Size, &file.Creator, &file.Attributes, &file.LocalCopy, &file.EncryptionKey)
	if err == sqlx.ErrNoRows {
		return File{}, os.ErrNotExist
	}
	if err != nil {
		return File{}, err
	}
	return file, nil
}

func (f *FS) getSync(file File, localPath string, dest io.Writer) error {
	encryptionKey := file.EncryptionKey

	if dest == nil {
		if localPath == "" {
			return core.Errorf("no destination specified")
		}

		destFile, err := os.Create(localPath)
		if err != nil {
			return err
		}
		defer destFile.Close()
		dest = destFile
	}

	dest, err := decryptWriter(dest, encryptionKey[0:32], encryptionKey[32:48])
	if err != nil {
		return err
	}

	err = f.S.Store.Read(path.Join(DataDir, file.ID), nil, dest, nil)
	if err != nil {
		return err
	}

	return nil
}
