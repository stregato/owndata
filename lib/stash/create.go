package stash

import (
	urllib "net/url"
	"strings"

	"github.com/stregato/stash/lib/core"
	"github.com/stregato/stash/lib/security"
	"github.com/stregato/stash/lib/sqlx"
)

func Create(db *sqlx.DB, identity *security.Identity, url string, config Config) (*Stash, error) {
	u, err := urllib.Parse(url)
	if err != nil {
		return nil, core.Errorw(err, "invalid url %s : %v", url)
	}
	parts := strings.Split(strings.TrimLeft(u.Path, "/"), "/")
	if len(parts) < 2 {
		return nil, core.Errorf("missing creator id and safe name in %s", url)
	}
	if parts[len(parts)-2] != identity.Id.String() {
		return nil, core.Errorf("creator id %s does not match identity %s", parts[len(parts)-2], identity.Id.String())
	}

	s, err := connect(db, identity, u.String())
	if err != nil {
		return nil, err
	}

	err = s.WriteConfig(config)
	if err != nil {
		return nil, err
	}
	s.Config = config

	_, err = s.UpdateGroup(AdminGroup, Grant, identity.Id)
	if err != nil {
		return nil, err
	}

	_, err = s.UpdateGroup(UserGroup, Grant, s.CreatorID)
	if err != nil {
		return nil, err
	}

	return s, nil
}
