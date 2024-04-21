package safe

import (
	_ "embed"
	"path"
	"strings"
	"sync"
	"testing"

	url_ "net/url"

	"github.com/stregato/mio/lib/core"
	"github.com/stregato/mio/lib/security"
	"github.com/stregato/mio/lib/sqlx"
	"github.com/stregato/mio/lib/storage"
)

var nextHnd int
var nextHndLock sync.Mutex

func Open(db *sqlx.DB, identity *security.Identity, url string) (*Safe, error) {
	u, err := url_.Parse(url)
	if err != nil {
		return nil, core.Errorw(err, "invalid url %s : %v", url)
	}

	parts := strings.Split(strings.TrimLeft(u.Path, "/"), "/")
	if len(parts) < 2 {
		return nil, core.Errorf("missing creator hash and safe name in %s", url)
	}

	creatorId, err := security.NewUserId(parts[len(parts)-2])
	if err != nil {
		return nil, core.Errorf("invalid creator id in %s", url)
	}

	store, err := storage.Open(url)
	if err != nil {
		return nil, err
	}

	nextHndLock.Lock()
	defer nextHndLock.Unlock()
	nextHnd++

	s := &Safe{
		Hnd:       nextHnd,
		URL:       url,
		ID:        store.ID(),
		DB:        db,
		Store:     store,
		CreatorId: creatorId,
		Identity:  identity,
	}

	config, err := s.ReadConfig()
	if err == nil {
		s.Config = config
	}

	return s, nil
}

func NewTestSafe(t *testing.T, identity *security.Identity, storeId string, creatorId security.ID, persistent bool) *Safe {
	core.T = t
	db := sqlx.NewTestDB(t, persistent)

	urls := storage.LoadTestURLs()

	url := urls[storeId]
	u, err := url_.Parse(url)
	core.TestErr(t, err, "cannot parse url %s", url)
	u.Path = path.Join(u.Path, string(creatorId)+"/test")
	s, err := Open(db, identity, u.String())
	core.TestErr(t, err, "cannot open safe %s", u.String())

	_, err = s.UpdateGroup(AdminGroup, Grant, identity.Id)
	core.TestErr(t, err, "cannot update group: %v")

	_, err = s.UpdateGroup(UserGroup, Grant, identity.Id)
	core.TestErr(t, err, "cannot update group: %v")

	return s
}
