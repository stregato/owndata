package stash

import (
	"github.com/stregato/stash/lib/core"
	"github.com/stregato/stash/lib/security"
	"github.com/stregato/stash/lib/storage"
	"golang.org/x/crypto/blake2b"
)

func (s *Stash) WriteConfig(config Config) error {
	if s.Identity.Id != s.CreatorID {
		return core.Errorf("only the creator can write the config")
	}

	h := hashOfConfig(config)
	signature, err := security.Sign(s.Identity, h)
	if err != nil {
		return err
	}
	config.Signature = signature
	err = storage.WriteYAML(s.Store, ".config.yaml", &config, nil)
	if err != nil {
		return err
	}
	return nil
}

func (s *Stash) ReadConfig() (Config, error) {
	var config Config
	err := storage.ReadYAML(s.Store, ".config.yaml", &config, nil)
	if err != nil {
		return Config{}, err
	}
	if !security.Verify(s.CreatorID, hashOfConfig(config), config.Signature) {
		return Config{}, core.Errorf("config signature is invalid")
	}
	return config, nil
}

func hashOfConfig(config Config) []byte {
	h, err := blake2b.New256(nil)
	if err != nil {
		panic(err)
	}
	h.Write([]byte(config.Description))
	h.Write([]byte(config.Signature))
	return h.Sum(nil)
}
