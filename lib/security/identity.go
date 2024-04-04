package security

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	eciesgo "github.com/ecies/go/v2"
	"github.com/stregato/mio/lib/core"
	"github.com/stregato/mio/lib/sqlx"
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

type UserId string
type Identity struct {
	Id      UserId    `json:"i"`           // public key
	Nick    string    `json:"n,omitempty"` // nickname
	Email   string    `json:"e,omitempty"` // email
	ModTime time.Time `json:"m"`           // last modification time

	Private string `json:"p,omitempty"` // private key
	Avatar  []byte `json:"a,omitempty"` // avatar
}

func NewIdentity(nick string) (*Identity, error) {
	var identity Identity

	if strings.Contains(nick, ":") {
		return nil, core.Errorf("invalid nickname '%s': %w", nick, ErrInvalidID)
	}

	identity.ModTime = core.Now()
	identity.Nick = nick
	privateCrypt, err := eciesgo.GenerateKey()
	if core.IsErr(err, "cannot generate secp256k1 key: %v") {
		return nil, err
	}
	publicCrypt := privateCrypt.PublicKey.Bytes(true)

	publicSign, privateSign, err := ed25519.GenerateKey(rand.Reader)
	if core.IsErr(err, "cannot generate ed25519 key: %v") {
		return nil, err
	}

	identity.Id = UserId(fmt.Sprintf("%s:%s", nick, core.EncodeBinary(append(publicCrypt, publicSign[:]...))))
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

func NewIdentityFromId(nick, privateId string) (Identity, error) {
	var identity Identity

	if len(privateId) == secp256k1PublicKeySize+ed25519.PublicKeySize {
		return identity, ErrInvalidID
	}
	privateCrypt, privateSign, err := DecodeKeys(privateId)
	if core.IsErr(err, "invalid private id: %v") {
		return identity, err
	}
	publicCrypt, err := eciesgo.NewPublicKeyFromBytes(privateCrypt)
	if core.IsErr(err, "cannot convert bytes to secp256k1 public key: %v") {
		return identity, err
	}
	publicSign := ed25519.PrivateKey(privateSign).Public().(ed25519.PublicKey)

	identity.ModTime = core.Now()
	identity.Nick = nick
	identity.Id = UserId(core.EncodeBinary(append(publicCrypt.Bytes(true), publicSign[:]...)))
	identity.Private = core.EncodeBinary(append(privateCrypt, privateSign[:]...))

	return identity, nil

}

func (i Identity) Public() Identity {
	return Identity{
		Id:      i.Id,
		Nick:    i.Nick,
		Email:   i.Email,
		ModTime: i.ModTime,
		Avatar:  i.Avatar,
	}
}

func SetIdentity(i Identity) error {
	data, err := json.Marshal(i)
	if core.IsErr(err, "cannot marshal identity: %v") {
		return err
	}

	_, err = sqlx.Default.Exec("SET_IDENTITY", sqlx.Args{
		"id":   i.Id,
		"data": data,
	})
	return err
}

func DelIdentity(id string) error {
	_, err := sqlx.Default.Exec("DEL_IDENTITY", sqlx.Args{
		"id": id,
	})
	return err
}

func GetIdentity(id string) (Identity, error) {
	var data []byte
	var identity Identity
	err := sqlx.Default.QueryRow("GET_IDENTITY", sqlx.Args{"id": id}, &data)
	if err == nil {
		err = json.Unmarshal(data, &identity)
		if core.IsErr(err, "corrupted identity on db: %v") {
			return identity, err
		}
	}
	return identity, err
}

func GetIdentities() ([]Identity, error) {
	rows, err := sqlx.Default.Query("GET_IDENTITIES", sqlx.Args{})
	if core.IsErr(err, "cannot get trusted identities from db: %v") {
		return nil, err
	}
	defer rows.Close()

	var identities []Identity
	for rows.Next() {
		var i64 []byte
		err = rows.Scan(&i64)
		if core.IsErr(err, "cannot read pool feeds from db: %v") {
			continue
		}

		var identity Identity
		err := json.Unmarshal(i64, &identity)
		if core.IsErr(err, "invalid identity record '%s': %v", i64) {
			continue
		}

		identities = append(identities, identity)
	}
	return identities, nil
}

func DecodeKeys(id string) (cryptKey []byte, signKey []byte, err error) {
	idx := strings.Index(id, ":")
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
