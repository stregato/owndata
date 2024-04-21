package fs

import (
	"time"
)

type ListOptions struct {
	After   time.Time // After is the minimum modification time of files to list
	Before  time.Time // Before is the maximum modification time of files to list
	OrderBy string    // OrderBy is the list of fields to order by in SQL style
	Reverse bool      // Reverse is whether to reverse the order of files
	Limit   int       // Limit is the maximum number of files to list
	Offset  int       // Offset is the number of files to skip
	Prefix  string    // Prefix is a filter on the prefix of the name
	Suffix  string    // Suffix is a filter on the suffix of the name
	Tag     string    // Tag is a filter on the tag of the file
}

func (f *FS) List(dir string, options ListOptions) ([]File, error) {
	if f.S.IsUpdated(HeadersDir, hashDir(dir)) {
		err := syncHeaders(f.S, dir)
		if err != nil {
			return nil, err
		}
	}

	return searchFiles(f.S, dir, options.After, options.Before, options.Prefix, options.Suffix, options.Tag,
		options.OrderBy, options.Limit, options.Offset)
}
