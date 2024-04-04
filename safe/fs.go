package safe

import (
	"github.com/stregato/mio/storage"
)

type CDN struct {
	Store storage.Store
	Quota int64
}

type FS struct {
	S        *Safe
	StoreUrl string
	Quota    int64
}
