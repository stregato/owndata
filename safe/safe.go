package safe

import (
	"sync"
	"testing"

	"github.com/stregato/mio/core"
	"github.com/stregato/mio/security"
	"github.com/stregato/mio/sql"
	"github.com/stregato/mio/storage"
)

type Safe struct {
	Id          int
	Db          sql.DB
	Store       storage.Store
	CreatorId   security.UserId
	CurrentUser security.Identity
	Lock        sync.RWMutex
}

var nextId int
var nextIdLock sync.Mutex

func NewSafe(dbUrl, storeUrl string, creatorId security.UserId, currentUser security.Identity) (*Safe, error) {
	db, err := sql.Open(dbUrl)
	if err != nil {
		return nil, err
	}
	store, err := storage.Open(storeUrl)

	nextIdLock.Lock()
	defer nextIdLock.Unlock()
	nextId++

	if err != nil {
		return nil, err
	}
	return &Safe{
		Id:          nextId,
		Db:          db,
		Store:       store,
		CreatorId:   creatorId,
		CurrentUser: currentUser,
	}, nil
}

func NewTestSafe(t *testing.T, creatorId security.UserId, currentUser security.Identity, storeId string, persistent bool) *Safe {
	core.T = t
	db := sql.NewTestDB(t, persistent)

	store := storage.NewTestStore(storeId)
	return &Safe{
		Db:          db,
		Store:       store,
		CreatorId:   creatorId,
		CurrentUser: currentUser,
	}
}

func CopySafe(c *Safe) *Safe {
	return &Safe{
		Db:          c.Db,
		Store:       c.Store,
		CreatorId:   c.CreatorId,
		CurrentUser: c.CurrentUser,
	}
}
