package cashec

import (
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/privacy/client"
)

type KeySet struct {
	PrivateKey  client.SpendingKey
	PublicKey   client.PaymentAddress
	ReadonlyKey client.ViewingKey
}

/**
GenerateKey - generate key set from seed byte[]
 */
func (self *KeySet) GenerateKey(seed []byte) (*KeySet) {
	copy(self.PrivateKey[:], common.HashB(seed))
	self.PublicKey = client.GenPaymentAddress(self.PrivateKey)
	self.ReadonlyKey = client.GenViewingKey(self.PrivateKey)
	return self
}

/**
ImportFromPrivateKeyByte - from private-key byte[], regenerate pub-key and readonly-key
 */
func (self *KeySet) ImportFromPrivateKeyByte(privateKey []byte) {
	copy(self.PrivateKey[:], privateKey)
	self.PublicKey = client.GenPaymentAddress(self.PrivateKey)
	self.ReadonlyKey = client.GenViewingKey(self.PrivateKey)
}

/**
ImportFromPrivateKeyByte - from private-key data, regenerate pub-key and readonly-key
 */
func (self *KeySet) ImportFromPrivateKey(privateKey *client.SpendingKey) {
	self.PrivateKey = *privateKey
	self.PublicKey = client.GenPaymentAddress(self.PrivateKey)
	self.ReadonlyKey = client.GenViewingKey(self.PrivateKey)
}

func (self *KeySet) Verify(data, signature []byte) (bool, error) {
	isValid := true
	return isValid, nil
}

func (self *KeySet) Sign(data []byte) ([]byte, error) {
	// TODO(@0xkraken): implement signing using keypair
	result := [32]byte{1, 2, 3}
	return result[:], nil
}
