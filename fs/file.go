package fs

import (
	"encoding/hex"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/stregato/mio/core"
	"github.com/stregato/mio/safe"
	"github.com/stregato/mio/security"
	"github.com/stregato/mio/sql"
	"github.com/stregato/mio/storage"
	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/crypto/blake2b"
)

type File struct {
	Id            string
	Dir           string
	Name          string
	GroupName     safe.GroupName
	Creator       security.UserId
	Size          int
	ModTime       time.Time
	Tags          core.Set[string]
	Attributes    map[string]any
	LocalPath     string
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
	h, err := blake2b.New256(nil)
	if err != nil {
		panic(err)
	}

	paddingLen := 128 - len(dir)%128
	data := []byte(dir)
	if paddingLen > 0 {
		data = append(data, 0)
		data = append(data, core.GenerateRandomBytes(paddingLen-1)...)
	}

	_, err = h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func writeHeader(s *safe.Safe, f File) error {
	dest := path.Join(HeadersDir, hashDir(f.Dir), core.SnowIDString())

	keys, err := s.GetKeys(f.GroupName, 0)
	if err != nil {
		return err
	}
	lastKey := keys[len(keys)-1]

	data, err := msgpack.Marshal(f)
	if err != nil {
		return core.Errorf("failed to marshal file header of %s/%s: %w", f.Dir, f.Name, err)
	}

	data, err = security.EncryptAES(data, lastKey)
	if err != nil {
		return err
	}

	fw := FileWrap{
		Group:        f.GroupName,
		EncryptionId: len(keys) - 1,
		Data:         data,
	}
	err = storage.WriteMsgPack(s.Store, dest, fw)
	return err
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
	ls, err := s.Store.ReadDir(path.Join(HeadersDir, dir), storage.Filter{})
	if err != nil {
		return err
	}

	for _, l := range ls {
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

	_, err := s.Db.Exec(INSERT_FILE, sql.Args{"id": f.Id, "dir": f.Dir, "name": f.Name, "group": f.GroupName,
		"creator": f.Creator, "size": f.Size, "mod_time": f.ModTime, "tags": tags, "attributes": f.Attributes,
		"local_path": f.LocalPath, "encryption_key": f.EncryptionKey})
	return err
}

const GET_FILES_BY_DIR = "GET_FILES_BY_DIR"

func searchFiles(s *safe.Safe, dir string, after, before time.Time, prefix, suffix, tag string, orderBy string,
	limit, offset int) ([]File, error) {
	var query string
	args := sql.Args{"dir": dir, "after": after, "before": before, "limit": limit, "offset": offset,
		"prefix": prefix, "suffix": suffix, "tag": tag}

	if orderBy != "" {
		query += " ORDER BY " + orderBy
	}

	rows, err := s.Db.QueryExt(GET_FILES_BY_DIR, query, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var files []File
	for rows.Next() {
		var f File
		var tags string
		err := rows.Scan(&f.Id, &f.Name, &f.Dir, &f.GroupName, &tags, &f.ModTime, &f.Size, &f.Creator,
			&f.Attributes, &f.LocalPath, &f.EncryptionKey)
		if err != nil {
			return nil, err
		}
		f.Tags = core.NewSet(strings.Split(strings.TrimSpace(tags), " ")...)
		files = append(files, f)
	}
	return files, nil
}
