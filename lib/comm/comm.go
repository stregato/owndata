package comm

import (
	"github.com/stregato/mio/lib/safe"
	"github.com/stregato/mio/lib/security"
)

type Comm struct {
	S *safe.Safe
}

var (
	CommDir = "comm"
)

func Open(s *safe.Safe) *Comm {
	return &Comm{S: s}
}

func (c *Comm) getEncryptionKeys(sender security.ID, dest string) (keys []safe.Key, err error) {
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
