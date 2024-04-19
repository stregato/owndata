package safe

import (
	urllib "net/url"
	"path"

	"github.com/stregato/mio/lib/core"
	"github.com/stregato/mio/lib/security"
	"github.com/stregato/mio/lib/sqlx"
)

func Create(db *sqlx.DB, identity *security.Identity, storeUrl, name string) (*Safe, error) {
	u, err := urllib.Parse(storeUrl)
	if err != nil {
		return nil, core.Errorw(err, "invalid url %s : %v", storeUrl)
	}

	u.Path = path.Join(u.Path, string(identity.Id), name)

	s, err := Open(db, identity, u.String())
	if err != nil {
		return nil, err
	}

	return s, nil
}
