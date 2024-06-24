package comm

import (
	"strconv"

	"github.com/stregato/mio/lib/security"
)

type MessageID uint64

type Message struct {
	Sender       security.ID
	EncryptionId int
	Recipient    string
	ID           MessageID
	Text         string
	Data         []byte
	File         string
}

func (id MessageID) String() string {
	return strconv.FormatUint(uint64(id), 16)
}

func (id MessageID) Uint64() uint64 {
	return uint64(id)
}
