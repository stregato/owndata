package storage

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
	"time"

	"github.com/stregato/stash/lib/core"
	"gopkg.in/yaml.v2"
)

type Source struct {
	Name   string
	Data   []byte
	Reader io.Reader
	Size   int64
}

const SizeAll = -1

type ListOption uint32

const (
	// IncludeHiddenFiles includes hidden files in a list operation
	IncludeHiddenFiles ListOption = 1
)

type Range struct {
	From int64
	To   int64
}

type Filter struct {
	Prefix      string                 //Prefix filters on results starting with prefix
	Suffix      string                 //Suffix filters on results ending with suffix
	AfterName   string                 //After ignore all results before the provided one and the provided one
	After       time.Time              //After ignore all results before the provided one and the provided one
	MaxResults  int64                  //MaxResults limits the number of results returned
	OnlyFiles   bool                   //OnlyFiles returns only files
	OnlyFolders bool                   //OnlyFolders returns only folders
	Function    func(fs.FileInfo) bool //Function filters on a custom function
}

type Description struct {
	ReadCost  float64 //ReadCost is the cost of reading 1 byte in CHF as per 2023
	WriteCost float64 //WriteCost is the cost of writing 1 byte in CHF as per 2023
}

// Store is a low level interface to storage services such as S3 or SFTP
type Store interface {
	//ReadDir returns the entries of a folder content
	ReadDir(name string, filter Filter) ([]fs.FileInfo, error)

	// Read reads data from a file into a writer
	Read(name string, rang *Range, dest io.Writer, progress chan int64) error

	// Write writes data to a file name. An existing file is overwritten
	Write(name string, source io.ReadSeeker, progress chan int64) error

	// Stat provides statistics about a file
	Stat(name string) (os.FileInfo, error)

	// Delete deletes a file
	Delete(name string) error

	//ID returns an identifier for the store, typically the URL without credentials information and other parameters
	ID() string

	// Close releases resources
	Close() error

	// String returns a human-readable representation of the storer (e.g. sftp://user@host.cc/path)
	String() string

	Describe() Description
}

// Open creates a new exchanger giving a provided configuration
func Open(connectionUrl string) (Store, error) {
	switch {
	case strings.HasPrefix(connectionUrl, "sftp://"):
		return OpenSFTP(connectionUrl)
	case strings.HasPrefix(connectionUrl, "s3://"):
		return OpenS3(connectionUrl)
	case strings.HasPrefix(connectionUrl, "file:/"):
		return OpenLocal(connectionUrl)
	case strings.HasPrefix(connectionUrl, "dav://"):
		return OpenWebDAV(connectionUrl)
	case strings.HasPrefix(connectionUrl, "davs://"):
		return OpenWebDAV(connectionUrl)
	case strings.HasPrefix(connectionUrl, "mem://"):
		return OpenMemory(connectionUrl)
	}

	return nil, core.Errorf("unsupported store schema in %s", connectionUrl)
}

func LoadTestURLs() (urls map[string]string) {
	homeDir, err := os.UserHomeDir()
	if core.IsErr(err, "cannot get user home dir: %v", err) {
		panic(err)
	}
	filename := path.Join(homeDir, "stash_test_urls.yaml")
	_, err = os.Stat(filename)
	if err != nil {
		filename = "../test_urls.yaml"
	}

	data, err := os.ReadFile(filename)
	if core.IsErr(err, "cannot read file %s: %v", filename) {
		panic(err)
	}

	err = yaml.Unmarshal(data, &urls)
	if core.IsErr(err, "cannot parse file %s: %v", filename) {
		panic(err)
	}
	return urls
}

func NewTestStore(id string) Store {

	urls := LoadTestURLs()
	url := urls[id]
	if url == "" {
		panic(fmt.Errorf("store with id %s not found", id))
	}

	store, err := Open(url)
	if core.IsErr(err, "cannot open store %s: %v", url) {
		panic(err)
	}
	ls, _ := store.ReadDir("", Filter{})
	for _, l := range ls {
		store.Delete(l.Name())
	}

	return store
}
