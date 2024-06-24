package comm

import (
	"os"
	"path"

	"github.com/stregato/mio/lib/core"
	"github.com/stregato/mio/lib/safe"
	"github.com/stregato/mio/lib/security"
	"github.com/stregato/mio/lib/storage"
)

func (c *Comm) Send(userId security.ID, m Message) error {
	m.Dest = userId.String()
	return c.send(m)
}

func (c *Comm) Broadcast(groupName safe.GroupName, m Message) error {
	m.Dest = groupName.String()
	return c.send(m)
}

func (c *Comm) send(m Message) error {
	m.Sender = c.S.Identity.Id
	id := MessageID(core.SnowID())
	keys, err := c.getEncryptionKeys(m.Dest)
	if err != nil {
		return err
	}
	m.EncryptionId = len(keys) - 1
	key := keys[m.EncryptionId]

	messageFile := path.Join(CommDir, m.Dest, id.String())

	if m.File != "" {
		source, err := os.Open(m.File)
		if err != nil {
			return err
		}
		defer source.Close()
		aesKey := key[:32]
		aesIV := key[32:]
		r, err := security.EncryptReader(source, aesKey, aesIV)
		if err != nil {
			return err
		}
		err = c.S.Store.Write(messageFile+".data", r, nil)
		if err != nil {
			return err
		}

		name := path.Base(m.File)
		data, err := security.EncryptAES([]byte(name), key)
		if err != nil {
			return err
		}
		m.File = core.EncodeBinary(data)
	}
	if m.Text != "" {
		data, err := security.EncryptAES([]byte(m.Text), key)
		if err != nil {
			return err
		}
		m.Text = core.EncodeBinary(data)
	}
	if m.Data != nil {
		data, err := security.EncryptAES(m.Data, key)
		if err != nil {
			return err
		}
		m.Data = data
	}

	err = storage.WriteJSON(c.S.Store, messageFile, m, nil)
	if err != nil {
		return err
	}

	return nil
}
