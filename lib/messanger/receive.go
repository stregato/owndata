package messanger

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/stregato/stash/lib/config"
	"github.com/stregato/stash/lib/core"
	"github.com/stregato/stash/lib/security"
	"github.com/stregato/stash/lib/storage"
	"golang.org/x/crypto/blake2b"
)

func (c *Messenger) Receive(filter string) ([]Message, error) {
	var dests []string
	if filter != "" {
		dests = []string{filter}
	} else {
		groups, err := c.S.GetGroups()
		if err != nil {
			return nil, err
		}
		dests = append(dests, c.S.Identity.Id.String())
		for name, users := range groups {
			if users.Contains(c.S.Identity.Id) {
				dests = append(dests, name.String())
			}
		}
	}

	var messages []Message
	for _, dest := range dests {
		dir := path.Join(MessangerDir, dest)
		lastId, _, _, _ := config.GetConfigValue(c.S.DB, "messanger", fmt.Sprintf("lastId-%s-%s", c.S.ID, dest))

		files, err := c.S.Store.ReadDir(dir, storage.Filter{AfterName: lastId})
		if err != nil {
			continue
		}

		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".data") {
				continue
			}
			m, err := c.receiveMessage(dest, file)
			if err != nil {
				core.Errorf("error reading %s: %s", file.Name(), err)
				continue
			}
			messages = append(messages, m)
			if m.ID.String() > lastId {
				lastId = m.ID.String()
			}
		}
		err = config.SetConfigValue(c.S.DB, "messanger", fmt.Sprintf("lastId-%s-%s", c.S.ID, dest), lastId, 0, nil)
		if err != nil {
			core.Errorf("cannot set lastId for %s: %s", dest, err)
		}
	}

	core.Info("received %d messages from safe %s", len(messages), c.S.URL)
	return messages, nil
}

func (c *Messenger) receiveMessage(dest string, file fs.FileInfo) (Message, error) {
	var m Message
	err := storage.ReadJSON(c.S.Store, path.Join(MessangerDir, dest, file.Name()), &m, nil)
	if err != nil {
		return Message{}, err
	}

	keys, err := c.getEncryptionKeys(m.Sender, dest)
	if err != nil {
		return Message{}, err
	}
	if len(keys) == 0 {
		return Message{}, nil
	}
	key := keys[m.EncryptionId]

	if m.File != "" {
		data, err := core.DecodeBinary(m.File)
		if err != nil {
			return Message{}, err
		}
		data, err = security.DecryptAES(data, key)
		if err != nil {
			return Message{}, err
		}
		m.File = string(data)
	}
	if m.Text != "" {
		data, err := core.DecodeBinary(m.Text)
		if err != nil {
			return Message{}, err
		}
		data, err = security.DecryptAES(data, key)
		if err != nil {
			return Message{}, err
		}
		m.Text = string(data)
	}
	if m.Data != nil {
		data, err := security.DecryptAES(m.Data, key)
		if err != nil {
			return Message{}, err
		}
		m.Data = data
	}
	return m, nil
}

func (c *Messenger) DownloadFile(m Message, dest string) error {
	if m.File == "" {
		return core.Errorf("no file to download")
	}

	keys, err := c.getEncryptionKeys(m.Sender, m.Recipient)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	key := keys[m.EncryptionId]
	iv := blake2b.Sum256([]byte(m.File))

	file, err := os.Create(dest)
	if err != nil {
		return err
	}
	w, err := security.DecryptWriter(file, key, iv[:16])
	if err != nil {
		return err
	}

	source := path.Join(MessangerDir, m.Recipient, m.ID.String()+".data")
	return c.S.Store.Read(source, nil, w, nil)
}
