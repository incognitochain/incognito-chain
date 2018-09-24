package cashec

import (
	"bytes"
	"encoding/base64"

	"golang.org/x/crypto/ed25519"
)

type KeyPair struct {
	PrivateKey []byte
	PublicKey  []byte
}

func (self *KeyPair) GenerateKey(seed []byte) (*KeyPair, error) {
	var err error
	self.PublicKey, self.PrivateKey, err = ed25519.GenerateKey(bytes.NewBuffer(seed))
	if err != nil {
		return self, err
	}
	return self, nil
}

func (self *KeyPair) Import(privateKey string) (*KeyPair, error) {
	key := ed25519.PrivateKey{}
	key, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		return self, err
	}
	self.PublicKey = key.Public().(ed25519.PublicKey)
	self.PrivateKey = key
	return self, nil
}

func (self *KeyPair) Verify(data, signature []byte) (bool, error) {
	isValid := false
	isValid = ed25519.Verify(self.PublicKey, data, signature)
	return isValid, nil
}

func (self *KeyPair) Sign(data []byte) ([]byte, error) {
	result := ed25519.Sign(self.PrivateKey, data)
	return result, nil
}
