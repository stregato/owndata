package fs

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/stregato/mio/lib/core"
	"github.com/stregato/mio/lib/safe"
	"github.com/stregato/mio/lib/security"
	"github.com/stregato/mio/lib/sqlx"

	"github.com/stregato/mio/lib/storage"
	"github.com/vmihailenco/msgpack/v5"
)

type FileID uint64

type File struct {
	ID            FileID
	Dir           string
	Name          string
	IsDir         bool
	GroupName     safe.GroupName
	Creator       security.ID
	Size          int
	ModTime       time.Time
	Tags          core.Set[string]
	Attributes    map[string]any
	LocalCopy     string
	CopyTime      time.Time
	EncryptionKey []byte
}

func (fileID FileID) String() string {
	return strconv.FormatUint(uint64(fileID), 16)
}

func (fileID FileID) Uint64() uint64 {
	return uint64(fileID)
}

func (f File) Path() string {
	return path.Join(f.Dir, f.Name)
}

type FileWrap struct {
	Group        safe.GroupName
	EncryptionId int
	Data         []byte
}

func hashDir(dir string) string {
	hasher := fnv.New64a() // FNV-1a variant provides slightly better dispersion for tiny differences in strings
	_, err := hasher.Write([]byte(dir))
	if err != nil {
		// Handle error in a real application
		panic(err)
	}
	return strconv.FormatUint(hasher.Sum64(), 16)
}

func writeHeader(s *safe.Safe, f File) (string, error) {
	dest := path.Join(HeadersDir, hashDir(f.Dir), f.ID.String())

	keys, err := s.GetKeys(f.GroupName, 0)
	if err != nil {
		return "", err
	}
	lastKey := keys[len(keys)-1]

	data, err := msgpack.Marshal(f)
	if err != nil {
		return "", core.Errorf("failed to marshal file header of %s/%s: %w", f.Dir, f.Name, err)
	}

	core.Info("encrypting header %s/%s", f.Dir, f.Name)
	data, err = security.EncryptAES(data, lastKey)
	if err != nil {
		return "", err
	}

	core.Info("writing header %s/%s to %s", f.Dir, f.Name, dest)
	fw := FileWrap{
		Group:        f.GroupName,
		EncryptionId: len(keys) - 1,
		Data:         data,
	}
	err = storage.WriteMsgPack(s.Store, dest, fw)
	if err != nil {
		return "", err
	}

	return dest, err
}

func readHeader(s *safe.Safe, dir, name string) (File, error) {
	src := path.Join(HeadersDir, hashDir(dir), name)

	var fw FileWrap
	err := storage.ReadMsgPack(s.Store, src, &fw)
	if err != nil {
		return File{}, err
	}

	keys, err := s.GetKeys(fw.Group, fw.EncryptionId+1)
	if err != nil {
		return File{}, err
	}
	if fw.EncryptionId >= len(keys) {
		return File{}, core.Errorf("invalid encryption id %d for group %s", fw.EncryptionId, fw.Group)
	}
	key := keys[fw.EncryptionId]

	data, err := security.DecryptAES(fw.Data, key)
	if err != nil {
		return File{}, err
	}

	f := File{}
	err = msgpack.Unmarshal(data, &f)
	if err != nil {
		return File{}, core.Errorf("failed to unmarshal file header of %s/%s: %w", dir, name, err)
	}
	core.Info("read header %s/%s with id %d", f.Dir, f.Name, f.ID)
	return f, nil
}

func syncHeaders(s *safe.Safe, dir string) error {
	ls, err := s.Store.ReadDir(path.Join(HeadersDir, hashDir(dir)), storage.Filter{})
	if os.IsNotExist(err) {
		return nil
	}
	core.Info("found %d headers in %s", len(ls), dir)

	if err != nil {
		return err
	}

	var lastID string
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if r.Intn(10) == 0 {
		err = s.DB.QueryRow("MIO_GET_LAST_ID", sqlx.Args{"safeID": s.ID}, &lastID)
		if err != nil && err != sqlx.ErrNoRows {
			return err
		}
	}

	for _, l := range ls {
		name := l.Name()
		if name <= lastID || strings.HasPrefix(name, ".") {
			core.Info("skipping header %s/%s", dir, name)
			continue
		}
		f, err := readHeader(s, dir, name)
		if err != nil {
			log.Error("failed to read header %s/%s: %w", dir, l.Name(), err)
			continue
		}
		err = writeFileToDB(s, f)
		if err != nil {
			log.Error("failed to write header %s/%s to DB: %w", dir, l.Name(), err)
			continue
		}
	}
	return nil
}

const MIO_STORE_FILE = "MIO_STORE_FILE"

func writeFileToDB(s *safe.Safe, f File) error {
	tags := fmt.Sprintf(" %s ", strings.Join(f.Tags.Slice(), " "))
	if len(tags) > 4096 {
		return core.Errorf("ErrTags: tags too long: %d", len(tags))
	}

	args := sqlx.Args{"safeID": s.ID, "name": f.Name, "dir": f.Dir, "id": f.ID.Uint64(),
		"creator": f.Creator, "groupName": f.GroupName, "tags": tags,
		"encryptionKey": f.EncryptionKey, "modTime": f.ModTime, "size": f.Size,
		"localCopy": f.LocalCopy, "copyTime": core.Now(), "attributes": f.Attributes}
	_, err := s.DB.Exec(MIO_STORE_FILE, args)
	if err != nil {
		return err
	}

	dir := f.Dir
	for dir != "" {
		var name string
		dir, name = core.SplitPath(dir)
		if name != "" {
			res, err := s.DB.Exec("MIO_STORE_DIR", sqlx.Args{"safeID": s.ID, "dir": dir, "name": name})
			if err != nil {
				core.Errorf("failed to store dir %s/%s: %w", dir, name, err)
			}
			count, _ := res.RowsAffected()
			if count > 0 {
				core.Info("stored dir %s", path.Join(dir, name))
			}
		}
	}
	core.Info("stored file %s/%s with id %d, args %+v", f.Dir, f.Name, f.ID, args)
	return err
}

const MIO_GET_FILES_BY_DIR = "MIO_GET_FILES_BY_DIR"

func searchFiles(s *safe.Safe, dir string, after, before time.Time, prefix, suffix, tag string, orderBy string,
	limit, offset int) ([]File, error) {
	args := sqlx.Args{"dir": dir, "safeID": s.ID, "name": "", "groupName": "", "tag": tag, "creator": "",
		"before": before.UnixNano(), "after": after.UnixNano(), "prefix": prefix, "suffix": suffix,
		"limit": limit, "offset": offset}

	if orderBy != "" {
		args["#orderBy"] = " ORDER BY "
	} else {
		args["#orderBy"] = ""
	}

	rows, err := s.DB.Query(MIO_GET_FILES_BY_DIR, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []File
	for rows.Next() {
		var f File
		var tags string
		err := rows.Scan(&f.ID, &f.Name, &f.Dir, &f.GroupName, &tags, &f.ModTime, &f.Size, &f.Creator,
			&f.Attributes, &f.LocalCopy, &f.CopyTime, &f.EncryptionKey)
		if err != nil {
			return nil, err
		}
		f.Tags = core.NewSet(strings.Split(strings.TrimSpace(tags), " ")...)
		f.IsDir = f.ID == 0
		files = append(files, f)
		core.Info("found file %s/%s", f.Dir, f.Name)
	}
	core.Info("found %d files in %s with search options %+v", len(files), dir, args)
	return files, nil
}
