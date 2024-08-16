package comm

import (
	"fmt"

	"github.com/stregato/stash/lib/config"
	"github.com/stregato/stash/lib/core"
	"github.com/stregato/stash/lib/security"
	"github.com/stregato/stash/lib/stash"
)

type Comm struct {
	S *stash.Stash
}

var (
	CommDir = "comm"
)

func Open(s *stash.Stash) *Comm {
	return &Comm{S: s}
}

func (c *Comm) Rewind(dest string, messageID MessageID) error {
	err := config.SetConfigValue(c.S.DB, "comm", fmt.Sprintf("lastId-%s-%s", c.S.ID, dest), messageID.String(), 0, nil)
	if err != nil {
		return core.Errorf("cannot set lastId for %s: %s", dest, err)
	}
	core.Info("rewinded communication for %s to id %d", dest, messageID)
	return nil
}

func (c *Comm) getEncryptionKeys(sender security.ID, dest string) (keys []stash.Key, err error) {
	if len(dest) > 80 {
		var id security.ID
		if dest == c.S.Identity.Id.String() {
			id = sender
		} else {
			id, err = security.CastID(dest)
			if err != nil {
				return nil, err
			}
		}
		key, err := security.DiffieHellmanKey(c.S.Identity, id.String())
		return []stash.Key{stash.Key(key)}, err
	}

	keys, err = c.S.GetKeys(stash.GroupName(dest), 0)
	if err != nil {
		return nil, err
	}
	return keys, nil
}
