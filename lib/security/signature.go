package security

import (
	"bytes"
	"crypto/ed25519"

	"github.com/stregato/mio/lib/core"
)

type PublicKey ed25519.PublicKey
type PrivateKey ed25519.PrivateKey

const (
	PublicKeySize  = ed25519.PublicKeySize
	PrivateKeySize = ed25519.PrivateKeySize
	SignatureSize  = ed25519.SignatureSize
)

type SignedData struct {
	Signature [SignatureSize]byte
	Signer    PublicKey
}

type Public struct {
	Id    PublicKey
	Nick  string
	Email string
}

func Sign(identity *Identity, data []byte) ([]byte, error) {
	_, signKey, err := DecodeKeys(identity.Private)
	if core.IsErr(err, "cannot decode keys: %v") {
		return nil, err
	}

	return ed25519.Sign(ed25519.PrivateKey(signKey), data), nil
}

func Verify(id ID, data []byte, sig []byte) bool {
	_, signKey, err := DecodeKeys(string(id))
	if core.IsErr(err, "cannot decode keys: %v") {
		return false
	}

	for off := 0; off < len(sig); off += SignatureSize {
		if func() bool {
			defer func() { recover() }()
			return ed25519.Verify(ed25519.PublicKey(signKey), data, sig[off:off+SignatureSize])
		}() {
			return true
		}
	}
	return false
}

type SignedHashEvidence struct {
	Key       []byte `json:"k"`
	Signature []byte `json:"s"`
}

type SignedHash struct {
	Hash       []byte
	Signatures map[ID][]byte
}

func NewSignedHash(hash []byte, i *Identity) (SignedHash, error) {
	signature, err := Sign(i, hash)
	if core.IsErr(err, "cannot sign with identity %s: %v", i.Id) {
		return SignedHash{}, err
	}

	return SignedHash{
		Hash:       hash,
		Signatures: map[ID][]byte{i.Id: signature},
	}, nil
}

func AppendToSignedHash(s SignedHash, i *Identity) error {
	signature, err := Sign(i, s.Hash)
	if core.IsErr(err, "cannot sign with identity %s: %v", i.Id) {
		return err
	}
	s.Signatures[i.Id] = signature
	return nil
}

func VerifySignedHash(s SignedHash, trusts []Identity, hash []byte) bool {
	if !bytes.Equal(s.Hash, hash) {
		return false
	}

	for _, trust := range trusts {
		id := trust.Id
		if signature, ok := s.Signatures[id]; ok {
			if Verify(id, hash, signature) {
				return true
			}
		}
	}
	return false
}
