package safe

import (
	_ "embed"
	"path"
	"strings"
	"sync"
	"testing"

	url_ "net/url"

	"github.com/stregato/stash/lib/core"
	"github.com/stregato/stash/lib/security"
	"github.com/stregato/stash/lib/sqlx"
	"github.com/stregato/stash/lib/storage"
)

var nextHnd int
var nextHndLock sync.Mutex

func connect(db *sqlx.DB, identity *security.Identity, url string) (*Safe, error) {
	u, err := url_.Parse(url)
	if err != nil {
		return nil, core.Errorw(err, "invalid url %s : %v", url)
	}

	parts := strings.Split(strings.TrimLeft(u.Path, "/"), "/")
	if len(parts) < 2 {
		return nil, core.Errorf("missing creator hash and safe name in %s", url)
	}

	creatorId, err := security.CastID(parts[len(parts)-2])
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
		CreatorID: creatorId,
		Identity:  identity,
	}
	return s, nil
}

func Open(db *sqlx.DB, identity *security.Identity, url string) (*Safe, error) {
	s, err := connect(db, identity, url)
	if err != nil {
		return nil, err
	}

	config, err := s.ReadConfig()
	if err != nil {
		return nil, err
	}
	s.Config = config
	return s, nil
}

func NewTestSafe(t *testing.T, identity *security.Identity, storeId string, creatorId security.ID, persistent bool) *Safe {
	core.T = t
	db := sqlx.NewTestDB(t, persistent)

	urls := storage.LoadTestURLs()

	url := urls[storeId]
	if url == "" {
		t.Fatalf("unknown store id %s", storeId)
	}
	u, err := url_.Parse(url)
	core.TestErr(t, err, "cannot parse url %s", url)
	u.Path = path.Join(u.Path, string(creatorId)+"/test")
	s, err := Create(db, identity, u.String(), Config{})
	core.TestErr(t, err, "cannot open safe %s", u.String())

	_, err = s.UpdateGroup(AdminGroup, Grant, identity.Id)
	core.TestErr(t, err, "cannot update group: %v")

	_, err = s.UpdateGroup(UserGroup, Grant, identity.Id)
	core.TestErr(t, err, "cannot update group: %v")

	return s
}
