package fs

import (
	"bytes"
	"io"
	"os"
	"path"

	"github.com/stregato/mio/lib/sqlx"
	"github.com/stregato/mio/lib/storage"
)

type GetOptions struct {
	Async    bool
	Range    *storage.Range
	Progress chan int64
}

func (f *FS) GetData(src string, options GetOptions) ([]byte, error) {
	var dest bytes.Buffer

	err := f.GetStream(src, &dest, options)
	if err != nil {
		return nil, err
	}

	return dest.Bytes(), nil
}

func (f *FS) GetFile(src, dest string, options GetOptions) ([]byte, error) {
	return nil, nil
}

func (f *FS) getSync(src string, localPath string, dest io.Writer, options GetOptions) error {
	dir, name := path.Split(src)
	if f.S.IsUpdated(HeadersDir, hashDir(dir)) {
		err := syncHeaders(f.S, dir)
		if err != nil {
			return err
		}
	}

	var (
		id            string
		size          int
		encryptionKey []byte
	)
	err := f.S.DB.QueryRow("GET_FILE_BY_NAME", sqlx.Args{"safeID": f.S.ID, "dir": dir, "name": name},
		&id, &size, &encryptionKey)
	if err == sqlx.ErrNoRows {
		return os.ErrNotExist
	}
	if err != nil {
		return err
	}

	dest, err = decryptWriter(dest, encryptionKey[0:32], encryptionKey[32:48])
	if err != nil {
		return err
	}

	err = f.S.Store.Read(path.Join(DataDir, id), options.Range, dest, options.Progress)
	if err != nil {
		return err
	}

	return nil
}
