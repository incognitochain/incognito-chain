package cashec

import (
	"github.com/ninjadotorg/cash/common"
	"github.com/ninjadotorg/cash/privacy/client"
)

type KeySet struct {
	// ProducerKeyPair KeyPair
	PrivateKey  client.SpendingKey
	PublicKey   client.PaymentAddress
	ReadonlyKey client.ViewingKey
}

/*
GenerateKey - generate key set from seed byte[]
*/
func (self *KeySet) GenerateKey(seed []byte) *KeySet {
	hash := common.HashB(seed)
	hash[len(hash)-1] &= 0x0F // Private key only has 252 bits
	copy(self.PrivateKey[:], hash)
	self.PublicKey = client.GenPaymentAddress(self.PrivateKey)
	self.ReadonlyKey = client.GenViewingKey(self.PrivateKey)
	return self
}

/*
ImportFromPrivateKeyByte - from private-key byte[], regenerate pub-key and readonly-key
*/
func (self *KeySet) ImportFromPrivateKeyByte(privateKey []byte) {
	copy(self.PrivateKey[:], privateKey)
	self.PublicKey = client.GenPaymentAddress(self.PrivateKey)
	self.ReadonlyKey = client.GenViewingKey(self.PrivateKey)
}

/*
ImportFromPrivateKeyByte - from private-key data, regenerate pub-key and readonly-key
*/
func (self *KeySet) ImportFromPrivateKey(privateKey *client.SpendingKey) {
	self.PrivateKey = *privateKey
	self.PublicKey = client.GenPaymentAddress(self.PrivateKey)
	self.ReadonlyKey = client.GenViewingKey(self.PrivateKey)
}

/*
Generate Producer keyset from privacy key set
*/
func (self *KeySet) CreateProducerKeySet() (*KeySetProducer, error) {
	var producerKeySet KeySetProducer
	producerKeySet.GenerateKey(self.PrivateKey[:])
	producerKeySet.SpendingAddress = self.PublicKey.Apk
	producerKeySet.TransmissionKey = self.PublicKey.Pkenc
	producerKeySet.ReceivingKey = self.ReadonlyKey.Skenc
	return &producerKeySet, nil
}
