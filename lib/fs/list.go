package fs

import (
	"time"

	"github.com/stregato/stash/lib/core"
)

type ListOptions struct {
	After   time.Time `json:"after"`   // After is the minimum modification time of files to list
	Before  time.Time `json:"before"`  // Before is the maximum modification time of files to list
	OrderBy string    `json:"orderBy"` // OrderBy is the list of fields to order by in SQL style
	Reverse bool      `json:"reverse"` // Reverse is true if the order should be reversed
	Limit   int       `json:"limit"`   // Limit is the maximum number of files to return
	Offset  int       `json:"offset"`  // Offset is the number of files to skip
	Prefix  string    `json:"prefix"`  // Prefix is a filter on the prefix of the name
	Suffix  string    `json:"suffix"`  // Suffix is a filter on the suffix of the name
	Tag     string    `json:"tag"`     // Tag is a filter on the tag of the file
}

func (f *FileSystem) List(dir string, options ListOptions) ([]File, error) {
	if f.S.IsUpdated(HeadersDir, hashDir(dir)) {
		core.Info("syncing headers of %s", dir)
		err := syncHeaders(f.S, dir)
		if err != nil {
			return nil, err
		}
	}

	return searchFiles(f.S, dir, options.After, options.Before, options.Prefix, options.Suffix, options.Tag,
		options.OrderBy, options.Limit, options.Offset)
}
