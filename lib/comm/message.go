package comm

import (
	"strconv"

	"github.com/stregato/stash/lib/security"
)

type MessageID uint64

type Message struct {
	Sender       security.ID `json:"sender"`
	EncryptionId int         `json:"encryptionId"`
	Recipient    string      `json:"recipient"`
	ID           MessageID   `json:"id"`
	Text         string      `json:"text"`
	Data         []byte      `json:"data"`
	File         string      `json:"file"`
}

func (id MessageID) String() string {
	return strconv.FormatUint(uint64(id), 16)
}

func (id MessageID) Uint64() uint64 {
	return uint64(id)
}
