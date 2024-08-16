package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/stregato/stash/lib/core"
	"github.com/vmihailenco/msgpack/v5"
)

func ReadFile(s Store, name string) ([]byte, error) {
	var b bytes.Buffer
	err := s.Read(name, nil, &b, nil)
	return b.Bytes(), err
}

func WriteFile(s Store, name string, data []byte) error {
	b := core.NewBytesReader(data)
	defer b.Close()
	return s.Write(name, b, nil)
}

func ReadJSON(s Store, name string, v any, hash hash.Hash) error {
	data, err := ReadFile(s, name)
	if err == nil {
		if hash != nil {
			hash.Write(data)
		}

		err = json.Unmarshal(data, v)
	}
	return err
}

func WriteJSON(s Store, name string, v any, hash hash.Hash) error {
	b, err := json.Marshal(v)
	if err == nil {
		if hash != nil {
			hash.Write(b)
		}
		err = s.Write(name, core.NewBytesReader(b), nil)
	}
	return err
}

func ReadMsgPack(s Store, name string, v any) error {
	data, err := ReadFile(s, name)
	if os.IsNotExist(err) || core.IsErr(err, "msgpackErr: cannot read file %s from store %s: %v", name, s) {
		return err
	}
	err = msgpack.Unmarshal(data, v)
	if core.IsErr(err, "msgpackErr: cannot unmarshal msgpack file %s from store %s : %v", name, s) {
		return err
	}

	return err
}

func WriteMsgPack(s Store, name string, v any) error {
	b, err := msgpack.Marshal(v)
	if core.IsErr(err, "msgpackErr: cannot marshal in store %s msgpack file %s: %v", s, name) {
		return err
	}
	err = s.Write(name, core.NewBytesReader(b), nil)
	if core.IsErr(err, "msgpackErr: cannot write file %s into store %s: %v", name, s) {
		return err
	}
	return nil
}

func ReadYAML(s Store, name string, v interface{}, hash hash.Hash) error {
	data, err := ReadFile(s, name)
	if err == nil {
		if hash != nil {
			hash.Write(data)
		}

		err = yaml.Unmarshal(data, v)
	}
	return err
}

func WriteYAML(s Store, name string, v interface{}, hash hash.Hash) error {
	b, err := yaml.Marshal(v)
	if err == nil {
		if hash != nil {
			hash.Write(b)
		}
		err = s.Write(name, core.NewBytesReader(b), nil)
	}
	return err
}

const maxSizeForMemoryCopy = 1024 * 1024

func CopyFile(dest Store, destName string, source Store, sourceName string) error {
	stat, err := source.Stat(sourceName)
	if core.IsErr(err, "cannot stat %s/%s: %v", source, sourceName) {
		return err
	}

	var r io.ReadSeeker
	if stat.Size() <= maxSizeForMemoryCopy {
		buf := bytes.Buffer{}
		err = source.Read(sourceName, nil, &buf, nil)
		if core.IsErr(err, "cannot read %s/%s: %v", source, sourceName) {
			return err
		}
		r = core.NewBytesReader(buf.Bytes())
	} else {
		file, err := os.CreateTemp("", "woland")
		if core.IsErr(err, "cannot create temporary file for CopyFile: %v") {
			return err
		}

		err = source.Read(sourceName, nil, file, nil)
		if core.IsErr(err, "cannot read %s/%s: %v", source, sourceName) {
			return err
		}
		file.Seek(0, 0)
		r = file
		defer func() {
			file.Close()
			os.Remove(file.Name())
		}()
	}

	err = dest.Write(destName, r, nil)
	if core.IsErr(err, "cannot write %s/%s: %v", dest, destName) {
		dest.Delete(destName)
		return err
	}

	return nil
}

func Dump(store Store, dir string, content bool) string {
	var builder strings.Builder
	files, err := store.ReadDir(dir, Filter{})
	if err != nil {
		return ""
	}
	var subdirs []string
	for _, file := range files {
		if file.IsDir() {
			subdir := filepath.Join(dir, file.Name())
			subdirs = append(subdirs, subdir)
		}
	}
	sort.Strings(subdirs)
	for _, subdir := range subdirs {
		subdirOutput := Dump(store, subdir, content)
		builder.WriteString(subdirOutput)
	}
	for _, file := range files {
		if !file.IsDir() {
			fmt.Fprintf(&builder, "%s\n", filepath.Join(dir, file.Name()))
			if content {
				data, _ := ReadFile(store, filepath.Join(dir, file.Name()))
				fmt.Fprintf(&builder, "%s\n", string(data))
			}
		}
	}
	return builder.String()
}
