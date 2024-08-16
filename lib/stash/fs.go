package stash

import "github.com/stregato/stash/lib/storage"

type CDN struct {
	Store storage.Store
	Quota int64
}

type FS struct {
	S        *Stash
	StoreUrl string
	Quota    int64
}
