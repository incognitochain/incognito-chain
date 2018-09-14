package cashec

import (
	"bytes"
	"encoding/base64"
	"errors"

	"golang.org/x/crypto/ed25519"
)

type KeyPair struct {
	PrivateKey []byte
	PublicKey  []byte
	Curve      string // current support ed25519 and default is ed25519
}

func (self *KeyPair) GenerateKey(seed []byte) (*KeyPair, error) {
	if self.Curve == "" {
		self.Curve = "ed25519"
	}
	var err error
	switch self.Curve {
	case "ed25519":
		self.PublicKey, self.PrivateKey, err = ed25519.GenerateKey(bytes.NewBuffer(seed))
		if err != nil {
			return self, err
		}
	default:
		return self, errors.New("this curve isn't supported")
	}
	return self, nil
}

func (self *KeyPair) Import(privateKey string) (*KeyPair, error) {
	if self.Curve == "" {
		self.Curve = "ed25519"
	}
	switch self.Curve {
	case "ed25519":
		key := ed25519.PrivateKey{}
		key, err := base64.StdEncoding.DecodeString(privateKey)
		if err != nil {
			return self, err
		}
		self.PublicKey = key.Public().(ed25519.PublicKey)
		self.PrivateKey = key
	default:
		return self, errors.New("this curve isn't supported")
	}
	return self, nil
}

func (self *KeyPair) Verify(data, signature []byte) (bool, error) {
	if self.Curve == "" {
		self.Curve = "ed25519"
	}
	isValid := false
	switch self.Curve {
	case "ed25519":
		isValid = ed25519.Verify(self.PublicKey, data, signature)
	default:
		return isValid, errors.New("this curve isn't supported")
	}
	return isValid, nil
}

func (self *KeyPair) Sign(data []byte) ([]byte, error) {
	var result []byte
	switch self.Curve {
	case "ed25519":
		result = ed25519.Sign(self.PrivateKey, data)
	default:
		return result, errors.New("this curve isn't supported")
	}
	return result, nil
}
