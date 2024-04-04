package safe

import (
	"sync"
	"testing"

	"github.com/stregato/mio/lib/core"
	"github.com/stregato/mio/lib/security"
	"github.com/stregato/mio/lib/settings"
	"github.com/stregato/mio/lib/sqlx"
	"github.com/stregato/mio/lib/storage"
)

var nextId int
var nextIdLock sync.Mutex

func Open(url string, creator security.UserId) (*Safe, error) {
	if defaultDB == nil {
		db, err := sqlx.Open(defaultDBPath)
		if err != nil {
			return nil, core.Errorw(err, "cannot open default database '%s': %v", defaultDBPath)
		}
		defaultDB = db
	}
	if currentUser == nil {
		err := GetConfigStruct(defaultDB, settings.KeySettingNode, string(settings.CurrentUserSettings), &currentUser)
		if err == sqlx.ErrNoRows {
			currentUser, err = security.NewIdentity(settings.DefaultNick)
			if err != nil {
				return nil, core.Errorw(err, "cannot create current user: %v")
			}
			err = SetConfigStruct(defaultDB, settings.KeySettingNode, string(settings.CurrentUserSettings), currentUser)
			if err != nil {
				return nil, core.Errorw(err, "cannot set current user: %v")
			}
		}
	}
	return OpenEx(defaultDB, url, creator, currentUser)
}

func OpenEx(db *sqlx.DB, url string, creatorId security.UserId, currentUser *security.Identity) (*Safe, error) {
	store, err := storage.Open(url)

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

func NewTestSafe(t *testing.T, creatorId security.UserId, currentUser *security.Identity, storeId string, persistent bool) *Safe {
	core.T = t
	db := sqlx.NewTestDB(t, persistent)

	store := storage.NewTestStore(storeId)
	return &Safe{
		Db:          db,
		Store:       store,
		CreatorId:   creatorId,
		CurrentUser: currentUser,
	}
}
