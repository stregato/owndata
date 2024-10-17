package messanger

import (
	"fmt"

	"github.com/stregato/stash/lib/config"
	"github.com/stregato/stash/lib/core"
	"github.com/stregato/stash/lib/safe"
	"github.com/stregato/stash/lib/security"
)

type Messenger struct {
	S *safe.Safe
}

var (
	MessangerDir = "messanger"
)

func Open(s *safe.Safe) *Messenger {
	return &Messenger{S: s}
}

func (c *Messenger) Rewind(dest string, messageID MessageID) error {
	err := config.SetConfigValue(c.S.DB, "messanger", fmt.Sprintf("lastId-%s-%s", c.S.ID, dest), messageID.String(), 0, nil)
	if err != nil {
		return core.Errorf("cannot set lastId for %s: %s", dest, err)
	}
	core.Info("rewinded communication for %s to id %d", dest, messageID)
	return nil
}

func (c *Messenger) getEncryptionKeys(sender security.ID, dest string) (keys []safe.Key, err error) {
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
		return []safe.Key{safe.Key(key)}, err
	}

	keys, err = c.S.GetKeys(safe.GroupName(dest), 0)
	if err != nil {
		return nil, err
	}
	return keys, nil
}
