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
	Async      bool             // put the file asynchronously
	DeleteSrc  bool             // delete the source file after a successful put
	GroupName  safe.GroupName   // the group name of the file. If empty, the group name is calculated from the directory
	Tags       core.Set[string] // the tags of the file
	Attributes map[string]any   // the attributes of the file
}

func (fs *FS) PutData(dest string, src []byte, options PutOptions) (File, error) {
	file, err := fs.createHeader(dest, len(src), options)
	if err != nil {
		return File{}, err
	}

	if options.Async {
		_, err = fs.S.DB.Exec("INSERT_FILE_ASYNC", sqlx.Args{"id": file.ID, "safeID": fs.S.ID,
			"localPath": "", "operation": "put", "file": file, "data": src, "deleteSrc": options.DeleteSrc})
		if err != nil {
			return File{}, err
		}
		triggerUpload <- file.ID
		return file, nil
	}
	err = fs.putSync(file, "", src, options.DeleteSrc)
	if err != nil {
		return File{}, err
	}

	return file, nil
}

func (fs *FS) PutFile(dest string, src string, options PutOptions) (File, error) {
	stat, err := os.Stat(src)
	if err != nil {
		return File{}, err
	}

	file, err := fs.createHeader(dest, int(stat.Size()), options)
	if err != nil {
		return File{}, err
	}
	file.LocalCopy = src

	if options.Async {
		_, err = fs.S.DB.Exec("INSERT_FILE_ASYNC", sqlx.Args{"id": file.ID,
			"localPath": src, "operation": "put", "file": file, "data": nil, "deleteSrc": options.DeleteSrc})
		if err != nil {
			return File{}, err
		}
		triggerUpload <- file.ID

		return file, nil
	}

	err = fs.putSync(file, src, nil, options.DeleteSrc)
	if err != nil {
		return File{}, err
	}
	return file, nil
}

func (fs *FS) createHeader(dest string, size int, options PutOptions) (File, error) {
	var err error
	dir, name := path.Split(dest)

	// get the group name and the corresponding key
	groupName := options.GroupName
	if groupName == "" {
		groupName, err = fs.calculateGroup(dir)
		if err != nil {
			return File{}, err
		}
	}

	return File{
		ID:            core.SnowIDString(),
		Dir:           dir,
		Name:          name,
		GroupName:     groupName,
		Creator:       fs.S.Identity.Id,
		Size:          size,
		ModTime:       core.Now(),
		Tags:          options.Tags,
		Attributes:    options.Attributes,
		EncryptionKey: core.GenerateRandomBytes(48),
	}, nil
}

func (fs *FS) putSync(file File, localPath string, data []byte, deleteSrc bool) error {
	var err error
	var src io.ReadSeeker

	switch {
	case data != nil:
		src = core.NewBytesReader(data)
	case localPath != "":
		f, err := os.Open(file.LocalCopy)
		if err != nil {
			return err
		}
		src = f
		defer f.Close()
	default:
		return core.Errorf("no data source provided for file %s", file.ID)
	}

	// write the body
	err = writeBody(fs.S, path.Join(DataDir, file.ID), src, file.EncryptionKey)
	if err != nil {
		return err
	}

	_, err = writeHeader(fs.S, file)
	if err != nil {
		fs.S.Store.Delete(path.Join(DataDir, file.ID))
		return err
	}

	if deleteSrc && file.LocalCopy != "" {
		os.Remove(file.LocalCopy)
	}

	err = syncHeaders(fs.S, file.Dir)
	if err != nil {
		core.Info("failed to sync headers: %v", err)
	}
	fs.S.Touch(HeadersDir, hashDir(file.Dir))

	return nil
}

func (fs *FS) calculateGroup(dir string) (safe.GroupName, error) {
	var groupName safe.GroupName
	for {
		err := fs.S.DB.QueryRow(GET_GROUP_NAME, sqlx.Args{"safeID": fs.S.ID, "dir": dir, "name": ""}, &groupName)
		if err != sqlx.ErrNoRows && err != nil {
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
