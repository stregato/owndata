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

type File struct {
	ID            string
	Dir           string
	Name          string
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
	dest := path.Join(HeadersDir, hashDir(f.Dir), f.ID)

	keys, err := s.GetKeys(f.GroupName, 0)
	if err != nil {
		return "", err
	}
	lastKey := keys[len(keys)-1]

	data, err := msgpack.Marshal(f)
	if err != nil {
		return "", core.Errorf("failed to marshal file header of %s/%s: %w", f.Dir, f.Name, err)
	}

	data, err = security.EncryptAES(data, lastKey)
	if err != nil {
		return "", err
	}

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
	return f, nil
}

func syncHeaders(s *safe.Safe, dir string) error {
	ls, err := s.Store.ReadDir(path.Join(HeadersDir, hashDir(dir)), storage.Filter{})
	if os.IsNotExist(err) {
		return nil
	}

	if err != nil {
		return err
	}

	var lastID string
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if r.Intn(10) == 0 {
		err = s.DB.QueryRow("GET_LAST_ID", sqlx.Args{"safeID": s.ID}, &lastID)
		if err != nil && err != sqlx.ErrNoRows {
			return err
		}
	}

	for _, l := range ls {
		if l.Name() <= lastID {
			continue
		}
		f, err := readHeader(s, dir, l.Name())
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

const INSERT_FILE = "INSERT_FILE"

func writeFileToDB(s *safe.Safe, f File) error {
	tags := fmt.Sprintf(" %s ", strings.Join(f.Tags.Slice(), " "))
	if len(tags) > 4096 {
		return core.Errorf("ErrTags: tags too long: %d", len(tags))
	}

	_, err := s.DB.Exec(INSERT_FILE, sqlx.Args{"id": f.ID, "safeID": s.ID, "name": f.Name, "dir": f.Dir,
		"creator": f.Creator, "groupName": f.GroupName, "tags": tags, "localPath": f.LocalCopy,
		"encryptionKey": f.EncryptionKey, "modTime": f.ModTime, "size": f.Size, "attributes": f.Attributes})
	return err
}

const GET_FILES_BY_DIR = "GET_FILES_BY_DIR"

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

	rows, err := s.DB.Query(GET_FILES_BY_DIR, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var files []File
	for rows.Next() {
		var f File
		var tags string
		err := rows.Scan(&f.ID, &f.Name, &f.Dir, &f.GroupName, &tags, &f.ModTime, &f.Size, &f.Creator,
			&f.Attributes, &f.LocalCopy, &f.EncryptionKey)
		if err != nil {
			return nil, err
		}
		f.Tags = core.NewSet(strings.Split(strings.TrimSpace(tags), " ")...)
		files = append(files, f)
	}
	return files, nil
}
