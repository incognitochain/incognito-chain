package cashec

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy/client"
	"github.com/ninjadotorg/constant/common/base58"
	"encoding/json"
	"errors"
)

type KeySet struct {
	// ProducerKeyPair KeyPair
	PrivateKey     client.SpendingKey
	PaymentAddress client.PaymentAddress
	ReadonlyKey    client.ViewingKey
}

/*
GenerateKey - generate key set from seed byte[]
*/
func (self *KeySet) GenerateKey(seed []byte) *KeySet {
	hash := common.HashB(seed)
	hash[len(hash)-1] &= 0x0F // Private key only has 252 bits
	copy(self.PrivateKey[:], hash)
	self.PaymentAddress = client.GenPaymentAddress(self.PrivateKey)
	self.ReadonlyKey = client.GenViewingKey(self.PrivateKey)
	return self
}

/*
ImportFromPrivateKeyByte - from private-key byte[], regenerate pub-key and readonly-key
*/
func (self *KeySet) ImportFromPrivateKeyByte(privateKey []byte) {
	copy(self.PrivateKey[:], privateKey)
	self.PaymentAddress = client.GenPaymentAddress(self.PrivateKey)
	self.ReadonlyKey = client.GenViewingKey(self.PrivateKey)
}

/*
ImportFromPrivateKeyByte - from private-key data, regenerate pub-key and readonly-key
*/
func (self *KeySet) ImportFromPrivateKey(privateKey *client.SpendingKey) {
	self.PrivateKey = *privateKey
	self.PaymentAddress = client.GenPaymentAddress(self.PrivateKey)
	self.ReadonlyKey = client.GenViewingKey(self.PrivateKey)
}

/*
Generate Producer keyset from privacy key set
*/
func (self *KeySet) CreateProducerKeySet() (*KeySetProducer, error) {
	var producerKeySet KeySetProducer
	producerKeySet.GenerateKey(self.PrivateKey[:])
	producerKeySet.SpendingAddress = self.PaymentAddress.Apk
	producerKeySet.TransmissionKey = self.PaymentAddress.Pkenc
	producerKeySet.ReceivingKey = self.ReadonlyKey.Skenc
	return &producerKeySet, nil
}

func (self *KeySet) Verify(data, signature []byte) (bool, error) {
	/*isValid := false
	hash := common.HashB(data)
	isValid = privacy.Verify(signature, hash[:], self.PaymentAddress.Pk)
	return isValid, nil*/
	return true, nil
}

func (self *KeySet) Sign(data []byte) ([]byte, error) {
	/*hash := common.HashB(data)
	signature, err := privacy.Sign(hash[:], self.PrivateKey)
	return signature, err*/
	return []byte{}, nil
}

func (self *KeySet) Encrypt(data []byte) ([]byte, error) {
	encryptText := client.Encrypt(self.PaymentAddress.Pkenc[:], data)
	return encryptText, nil
}

func (self *KeySet) Decrypt(data []byte) ([]byte, error) {
	data, err := client.Decrypt(self.ReadonlyKey.Skenc[:], data)
	return data, err
}

func (self *KeySet) EncodeToString() string {
	val, _ := json.Marshal(self)
	result := base58.Base58Check{}.Encode(val, byte(0x00))
	return result
}

func (self *KeySet) DecodeToKeySet(keystring string) (*KeySet, error) {
	base58C := base58.Base58Check{}
	keyBytes, _, _ := base58C.Decode(keystring)
	json.Unmarshal(keyBytes, self)
	return self, nil
}

func (self *KeySet) GetPaymentAddress() (client.PaymentAddress, error) {
	return self.PaymentAddress, nil
}

func (self *KeySet) GetViewingKey() (client.ViewingKey, error) {
	return self.ReadonlyKey, nil
}

func ValidateDataB58_(pubkey string, sig string, data []byte) error {
	decPubkey, _, err := base58.Base58Check{}.Decode(pubkey)
	if err != nil {
		return errors.New("can't decode public key:" + err.Error())
	}

	validatorKp := KeySet{}
	copy(validatorKp.PaymentAddress.Apk[:], decPubkey)
	decSig, _, err := base58.Base58Check{}.Decode(sig)
	if err != nil {
		return errors.New("can't decode signature: " + err.Error())
	}

	isValid, err := validatorKp.Verify(data, decSig)
	if err != nil {
		return errors.New("error when verify data: " + err.Error())
	}
	if !isValid {
		return errors.New("Invalid signature")
	}
	return nil
}
