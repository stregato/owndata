package security

import (
	eciesgo "github.com/ecies/go/v2"
	"github.com/stregato/stash/lib/core"
)

func EcEncrypt(id ID, data []byte) ([]byte, error) {
	cryptKey, _, err := DecodeKeys(id.String())
	if core.IsErr(err, "cannot decode keys: %v") {
		return nil, err
	}

	pk, err := eciesgo.NewPublicKeyFromBytes(cryptKey)
	if core.IsErr(err, "cannot convert bytes to secp256k1 public key: %v") {
		return nil, err
	}
	data, err = eciesgo.Encrypt(pk, data)
	if core.IsErr(err, "cannot encrypt with secp256k1: %v") {
		return nil, err
	}
	return data, err
}

func EcDecrypt(identity *Identity, data []byte) ([]byte, error) {
	cryptKey, _, err := DecodeKeys(identity.Private)
	if core.IsWarn(err, "cannot decode keys: %v") {
		return nil, err
	}

	data, err = eciesgo.Decrypt(eciesgo.NewPrivateKeyFromBytes(cryptKey), data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
