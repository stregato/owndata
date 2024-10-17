package fs

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/stregato/stash/lib/core"
	"github.com/stregato/stash/lib/safe"
	"github.com/stregato/stash/lib/security"
	"github.com/stregato/stash/lib/sqlx"
)

type PutOptions struct {
	ID         FileID         `json:"id"`         // the ID of the file, used to overwrite an existing file
	Async      bool           `json:"async"`      // put the file asynchronously
	DeleteSrc  bool           `json:"deleteSrc"`  // delete the source file after putting it
	GroupName  safe.GroupName `json:"groupName"`  // the group name of the file. If empty, the group name is calculated from the directory
	Tags       []string       `json:"tags"`       // the tags of the file
	Attributes map[string]any `json:"attributes"` // the attributes of the file
}

func (fs *FileSystem) PutData(dest string, src []byte, options PutOptions) (File, error) {
	file, err := fs.createHeader(dest, len(src), options)
	if err != nil {
		return File{}, err
	}

	if options.Async {
		core.Info("putting file %s asynchronously", dest)
		_, err = fs.S.DB.Exec("STASH_INSERT_FILE_ASYNC", sqlx.Args{"id": file.ID, "safeID": fs.S.ID,
			"operation": "put", "file": file, "data": src, "localCopy": "", "deleteSrc": options.DeleteSrc})
		if err != nil {
			return File{}, err
		}
		triggerAsync <- file.ID
		return file, nil
	}
	err = fs.putSync(file, "", src, options.DeleteSrc)
	if err != nil {
		return File{}, err
	}

	return file, nil
}

func (fs *FileSystem) PutFile(dest string, src string, options PutOptions) (File, error) {
	stat, err := os.Stat(src)
	if err != nil {
		return File{}, err
	}

	file, err := fs.createHeader(dest, int(stat.Size()), options)
	if err != nil {
		return File{}, err
	}
	localCopy, err := filepath.Abs(src)
	if err != nil {
		return File{}, err
	}
	file.LocalCopy = localCopy

	if options.Async {
		core.Info("putting file %s asynchronously", dest)
		_, err = fs.S.DB.Exec("STASH_INSERT_FILE_ASYNC", sqlx.Args{"id": file.ID,
			"operation": "put", "file": file, "data": nil, "localCopy": src, "deleteSrc": options.DeleteSrc})
		if err != nil {
			return File{}, err
		}
		triggerAsync <- file.ID

		return file, nil
	}

	err = fs.putSync(file, src, nil, options.DeleteSrc)
	if err != nil {
		return File{}, err
	}
	return file, nil
}

func (fs *FileSystem) createHeader(dest string, size int, options PutOptions) (File, error) {
	var err error
	dir, name := core.SplitPath(dest)

	// get the group name and the corresponding key
	groupName := options.GroupName
	if groupName == "" {
		groupName, err = fs.calculateGroup(dir)
		if err != nil {
			return File{}, err
		}
	}

	id := options.ID
	if id == 0 {
		id = FileID(core.SnowID())
	}

	return File{
		ID:            id,
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

func (fs *FileSystem) putSync(file File, localPath string, data []byte, deleteSrc bool) error {
	var err error
	var src io.ReadSeeker

	switch {
	case data != nil:
		core.Info("putting file %s from data", file.ID)
		src = core.NewBytesReader(data)
	case localPath != "":
		core.Info("putting file %s from local file %s", file.ID, localPath)
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
	err = writeBody(fs.S, path.Join(DataDir, file.ID.String()), src, file.EncryptionKey)
	if err != nil {
		return err
	}

	_, err = writeHeader(fs.S, file)
	if err != nil {
		fs.S.Store.Delete(path.Join(DataDir, file.ID.String()))
		return err
	}

	if deleteSrc && file.LocalCopy != "" {
		os.Remove(file.LocalCopy)
	}

	err = syncHeaders(fs.S, file.Dir)
	if err != nil {
		core.Info("failed to sync headers: %v", err)
	}

	dir := file.Dir
	for dir != "" {
		fs.S.Touch(HeadersDir, hashDir(dir))
		dir = core.Dir(dir)
	}

	return nil
}

func (fs *FileSystem) calculateGroup(dir string) (safe.GroupName, error) {
	var groupName safe.GroupName
	for {
		err := fs.S.DB.QueryRow(STASH_GET_GROUP_NAME, sqlx.Args{"safeID": fs.S.ID, "dir": dir, "name": ""}, &groupName)
		if err != sqlx.ErrNoRows && err != nil {
			return "", err
		}
		if groupName != "" {
			return groupName, nil
		}
		if dir == "" {
			return safe.UserGroup, nil
		}
		index := strings.LastIndex(dir, "/")
		if index == -1 {
			dir = ""
		} else {
			dir = dir[:index]
		}
	}
}

func writeBody(s *safe.Safe, dest string, src io.ReadSeeker, key []byte) error {
	aesKey := key[:32]
	aesIV := key[32:]
	r, err := security.EncryptReader(src, aesKey, aesIV)
	if err != nil {
		return err
	}
	core.Info("writing body to %s", dest)
	return s.Store.Write(dest, r, nil)
}
