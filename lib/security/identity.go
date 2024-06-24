package security

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"

	eciesgo "github.com/ecies/go/v2"
	"github.com/stregato/mio/lib/core"
)

var ErrInvalidSignature = errors.New("signature is invalid")
var ErrInvalidID = errors.New("ID is neither a public or private key")

const (
	Secp256k1               = "secp256k1"
	secp256k1PublicKeySize  = 33
	secp256k1PrivateKeySize = 32

	Ed25519 = "ed25519"
)

type Key struct {
	Public  []byte `json:"pu"`
	Private []byte `json:"pr,omitempty"`
}

type ID string
type Identity struct {
	Id      ID     `json:"i"`           // public key
	Private string `json:"p,omitempty"` // private key
}

func NewIdentity(nick string) (*Identity, error) {
	var identity Identity

	privateCrypt, err := eciesgo.GenerateKey()
	if core.IsErr(err, "cannot generate secp256k1 key: %v") {
		return nil, err
	}
	publicCrypt := privateCrypt.PublicKey.Bytes(true)

	publicSign, privateSign, err := ed25519.GenerateKey(rand.Reader)
	if core.IsErr(err, "cannot generate ed25519 key: %v") {
		return nil, err
	}

	identity.Id = ID(fmt.Sprintf("%s.%s", nick, core.EncodeBinary(append(publicCrypt, publicSign[:]...))))
	identity.Private = core.EncodeBinary(append(privateCrypt.Bytes(), privateSign[:]...))

	return &identity, nil
}

func NewIdentityMust(nick string) *Identity {
	identity, err := NewIdentity(nick)
	if err != nil {
		panic(err)
	}
	return identity
}

func CastID(id string) (ID, error) {
	id = strings.TrimSpace(id)
	_, _, err := DecodeKeys(id)
	if core.IsErr(err, "invalid ID '%s': %v", id) {
		return "", err
	}

	return ID(id), nil
}

func (u ID) String() string {
	return string(u)
}

func (userId ID) Nick() string {
	idx := strings.LastIndex(string(userId), ".")
	if idx > 0 {
		return string(userId[:idx])
	}
	return ""
}

func DecodeKeys(id string) (cryptKey []byte, signKey []byte, err error) {
	idx := strings.LastIndex(id, ".")
	if idx > 0 {
		id = id[idx+1:]
	}

	data, err := core.DecodeBinary(id)
	if core.IsErr(err, "cannot decode base64: %v") {
		return nil, nil, err
	}

	var split int
	if len(data) == secp256k1PrivateKeySize+ed25519.PrivateKeySize {
		split = secp256k1PrivateKeySize
	} else if len(data) == secp256k1PublicKeySize+ed25519.PublicKeySize {
		split = secp256k1PublicKeySize
	} else {
		core.IsErr(ErrInvalidID, "invalid ID %s with length %d: %v", id, len(data))
		return nil, nil, ErrInvalidID
	}

	return data[:split], data[split:], nil
}
