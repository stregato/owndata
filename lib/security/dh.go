package security

import (
	"crypto/sha256"
	"fmt"

	eciesgo "github.com/ecies/go/v2"
	"github.com/stregato/mio/lib/core"
)

func DiffieHellmanKey(identity Identity, id string) ([]byte, error) {
	privateKey, _, err := DecodeKeys(identity.Private)
	if core.IsErr(err, "cannot decode keys: %v") {
		return nil, err
	}

	publicKey, _, err := DecodeKeys(id)
	if core.IsErr(err, "cannot decode keys: %v") {
		return nil, err
	}

	pr := eciesgo.NewPrivateKeyFromBytes(privateKey)
	if pr == nil {
		return nil, fmt.Errorf("cannot convert bytes to secp256k1 private key")
	}

	pu, err := eciesgo.NewPublicKeyFromBytes(publicKey)
	if core.IsErr(err, "cannot convert bytes to secp256k1 public key: %v") {
		return nil, err
	}

	data, err := pr.ECDH(pu)
	if core.IsErr(err, "cannot perform ECDH: %v") {
		return nil, err
	}

	h := sha256.Sum256(data)
	return h[:], nil
}
