package fs

import (
	"io"
	"os"
	"path"

	"github.com/stregato/mio/lib/core"
	"github.com/stregato/mio/lib/safe"
	"github.com/stregato/mio/lib/sqlx"
)

type PutOptions struct {
	Async      chan map[string]error
	GroupName  safe.GroupName
	Tags       core.Set[string]
	Attributes map[string]any
}

func (fs *FS) PutData(dest string, src []byte, options PutOptions) error {
	return fs.PutStream(dest, core.NewBytesReader(src), options)
}

func (fs *FS) PutFile(dest string, src string, options PutOptions) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	if options.Async != nil {
		go func() {
			options.Async <- map[string]error{dest: fs.putSync(dest, f, options)}
			f.Close()
		}()
		return nil
	}
	defer f.Close()
	return fs.putSync(dest, f, options)
}

func (fs *FS) PutStream(dest string, src io.ReadSeekCloser, options PutOptions) error {
	if options.Async != nil {
		go func() {
			options.Async <- map[string]error{dest: fs.PutStream(dest, src, options)}
		}()
		return nil
	}
	return fs.putSync(dest, src, options)
}

func (fs *FS) putSync(dest string, src io.ReadSeekCloser, options PutOptions) error {
	var err error
	dir, name := path.Split(dest)

	// write the body
	id := core.SnowIDString()
	key := core.GenerateRandomBytes(64)
	err = writeBody(fs.S, path.Join(hashDir(dir), id), src, key)
	if err != nil {
		return err
	}

	// get the group name and the corresponding key
	groupName := options.GroupName
	if groupName == "" {
		groupName, err = fs.calculateGroup(dir)
		if err != nil {
			return err
		}
	}

	size, err := src.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	file := File{
		Id:            id,
		Dir:           dir,
		Name:          name,
		GroupName:     groupName,
		Creator:       fs.S.Identity.Id,
		Size:          int(size),
		ModTime:       core.Now(),
		Tags:          options.Tags,
		Attributes:    options.Attributes,
		EncryptionKey: key,
	}
	if src, ok := src.(*os.File); ok {
		file.LocalPath = src.Name()
	}

	err = writeHeader(fs.S, file)
	if err != nil {
		return err
	}

	return nil
}

func (fs *FS) calculateGroup(dir string) (safe.GroupName, error) {
	var groupName safe.GroupName
	for {
		err := fs.S.DB.QueryRow(GET_GROUP_NAME, sqlx.Args{"storeUrl": fs.StoreUrl, "dir": dir, "name": ""}, &groupName)
		if err != nil {
			return "", err
		}
		if groupName != "" {
			return groupName, nil
		}
		if dir == "" {
			return safe.UserGroup, nil
		}
		dir = path.Dir(dir)
	}
}

func writeBody(s *safe.Safe, dest string, src io.ReadSeeker, key []byte) error {
	aesKey := key[:32]
	aesIV := key[32:]
	r, err := encryptReader(src, aesKey, aesIV)
	if err != nil {
		return err
	}
	return s.Store.Write(dest, r, nil)
}
