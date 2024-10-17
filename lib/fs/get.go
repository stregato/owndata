package fs

import (
	"bytes"
	"io"
	"os"
	"path"

	"github.com/stregato/stash/lib/core"
	"github.com/stregato/stash/lib/security"
	"github.com/stregato/stash/lib/sqlx"
)

type GetOptions struct {
	Async bool
}

func (f *FileSystem) GetData(src string, options GetOptions) ([]byte, error) {
	var dest bytes.Buffer

	if options.Async {
		return nil, core.Errorf("GetData does not support async mode. Use GetFile instead")
	}

	file, err := f.Stat(src)
	if err != nil {
		return nil, err
	}

	err = f.getSync(file, "", &dest)
	if err != nil {
		return nil, err
	}

	return dest.Bytes(), nil
}

func (f *FileSystem) GetFile(src, dest string, options GetOptions) (File, error) {
	file, err := f.Stat(src)
	if err != nil {
		return File{}, err
	}

	if options.Async {
		_, err = f.S.DB.Exec("STASH_INSERT_FILE_ASYNC", sqlx.Args{"id": file.ID, "safeID": f.S.ID,
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

func (f *FileSystem) getSync(file File, localPath string, dest io.Writer) error {
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

	dest, err := security.DecryptWriter(dest, encryptionKey[0:32], encryptionKey[32:48])
	if err != nil {
		return err
	}

	err = f.S.Store.Read(path.Join(DataDir, file.ID.String()), nil, dest, nil)
	if err != nil {
		return err
	}

	if localPath != "" {
		f.S.DB.Exec("STASH_UPDATE_LOCALPATH", sqlx.Args{"id": file.ID, "safeID": f.S.ID,
			"localCopy": localPath, "copyTime": core.Now()})
	}

	return nil
}
