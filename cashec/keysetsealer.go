package cashec

import (
	"bytes"
	"encoding/base64"
	"encoding/json"

	"github.com/ninjadotorg/cash-prototype/privacy/client"
	"golang.org/x/crypto/ed25519"
)

type KeySetSealer struct {
	SprivateKey     []byte
	SpublicKey      []byte
	SpendingAddress [client.SpendingAddressLength]byte
	TransmissionKey [client.TransmissionKeyLength]byte
	ReceivingKey    [client.ReceivingKeyLength]byte
}

func (self *KeySetSealer) GenerateKey(seed []byte) (*KeySetSealer, error) {
	var err error
	self.SpublicKey, self.SprivateKey, err = ed25519.GenerateKey(bytes.NewBuffer(seed))
	if err != nil {
		return self, err
	}
	return self, nil
}

func (self *KeySetSealer) Import(privateKey string) (*KeySetSealer, error) {
	key := ed25519.PrivateKey{}
	key, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		return self, err
	}
	self.SpublicKey = key.Public().(ed25519.PublicKey)
	self.SprivateKey = key
	return self, nil
}

func (self *KeySetSealer) Verify(data, signature []byte) (bool, error) {
	isValid := false
	isValid = ed25519.Verify(self.SpublicKey, data, signature)
	return isValid, nil
}

func (self *KeySetSealer) Sign(data []byte) ([]byte, error) {
	result := ed25519.Sign(self.SprivateKey, data)
	return result, nil
}

func (self *KeySetSealer) EncodeToString() string {
	val, _ := json.Marshal(self)
	result := base64.StdEncoding.EncodeToString(val)
	return result
}

func (self *KeySetSealer) DecodeToKeySet(keystring string) (*KeySetSealer, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(keystring)
	if err != nil {
		return self, err
	}
	json.Unmarshal(keyBytes, self)
	return self, nil
}

func (self *KeySetSealer) GetPaymentAddress() (client.PaymentAddress, error) {
	var paymentAddr client.PaymentAddress
	paymentAddr.Apk = self.SpendingAddress
	paymentAddr.Pkenc = self.TransmissionKey
	return paymentAddr, nil
}

func (self *KeySetSealer) GetViewingKey() (client.ViewingKey, error) {
	var viewingKey client.ViewingKey
	viewingKey.Apk = self.SpendingAddress
	viewingKey.Skenc = self.ReceivingKey
	return viewingKey, nil
}
