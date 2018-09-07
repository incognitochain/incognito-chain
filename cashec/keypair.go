package cashec

import (
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/privacy/client"
)

type KeyPair struct {
	PrivateKey client.SpendingKey
	PublicKey  client.PaymentAddress
}

func (self *KeyPair) GenerateKey(seed []byte) (*KeyPair, error) {
	copy(self.PrivateKey[:], common.HashB(seed))
	self.PublicKey = client.GenPaymentAddress(self.PrivateKey)
	return self, nil
}

func (self *KeyPair) Import(privateKey []byte) (*KeyPair, error) {
	copy(self.PrivateKey[:], privateKey)
	self.PublicKey = client.GenPaymentAddress(self.PrivateKey)
	return self, nil
}

func (self *KeyPair) Verify(data, signature []byte) (bool, error) {
	isValid := true
	return isValid, nil
}

func (self *KeyPair) Sign(data []byte) ([]byte, error) {
	// TODO(@0xbunyip): implement signing using keypair
	result := [32]byte{1, 2, 3}
	return result[:], nil
}
