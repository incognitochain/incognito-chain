package cashec

import (
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/privacy/client"
)

type KeySet struct {
	// SealerKeyPair KeyPair
	PrivateKey  client.SpendingKey
	PublicKey   client.PaymentAddress
	ReadonlyKey client.ViewingKey
}

/**
GenerateKey - generate key set from seed byte[]
*/
func (self *KeySet) GenerateKey(seed []byte) *KeySet {
	hash := common.HashB(seed)
	hash[len(hash)-1] &= 0x0F // Private key only has 252 bits
	copy(self.PrivateKey[:], hash)
	self.PublicKey = client.GenPaymentAddress(self.PrivateKey)
	self.ReadonlyKey = client.GenViewingKey(self.PrivateKey)
	// self.SealerKeyPair.GenerateKey(self.PrivateKey[:])
	return self
}

/**
ImportFromPrivateKeyByte - from private-key byte[], regenerate pub-key and readonly-key
*/
func (self *KeySet) ImportFromPrivateKeyByte(privateKey []byte) {
	copy(self.PrivateKey[:], privateKey)
	self.PublicKey = client.GenPaymentAddress(self.PrivateKey)
	self.ReadonlyKey = client.GenViewingKey(self.PrivateKey)
	// self.SealerKeyPair.GenerateKey(self.PrivateKey[:])
}

/**
ImportFromPrivateKeyByte - from private-key data, regenerate pub-key and readonly-key
*/
func (self *KeySet) ImportFromPrivateKey(privateKey *client.SpendingKey) {
	self.PrivateKey = *privateKey
	self.PublicKey = client.GenPaymentAddress(self.PrivateKey)
	self.ReadonlyKey = client.GenViewingKey(self.PrivateKey)
	// self.SealerKeyPair.GenerateKey(self.PrivateKey[:])
}

func (self *KeySet) CreateSealerKeySet() (*KeySetSealer, error) {
	var sealerKeySet KeySetSealer
	sealerKeySet.GenerateKey(self.PrivateKey[:])
	sealerKeySet.SpendingAddress = self.PublicKey.Apk
	sealerKeySet.TransmissionKey = self.PublicKey.Pkenc
	sealerKeySet.ReceivingKey = self.ReadonlyKey.Skenc
	return &sealerKeySet, nil
}

// func (self *KeySet) GenerateSignKey() (client.PrivateKey, error){
// 	// Generate signing key
// 	privKey, err := client.GenerateKey(rand.Reader)
// 	return *privKey, err
// }

// func (self *KeySet) Verify(data, signature []byte, pubKey client.PublicKey) (bool, error) {
// 	jsSig := new(JSSig)
// 	err := json.Unmarshal(signature, jsSig)
// 	if err != nil {
// 		return false, err
// 	}
// 	valid := client.VerifySign(&pubKey, data[:], jsSig.R, jsSig.S)
// 	return valid, nil
// }

// func (self *KeySet) Sign(data []byte, privKey client.PrivateKey) ([]byte, error) {
// 	// TODO(@0xkraken): implement signing using keypair
// 	jsSig := *new(JSSig)
// 	jsSig.R, jsSig.S, _= client.Sign(rand.Reader, &privKey, data[:])

// 	signed_data, err := json.Marshal(jsSig)
// 	if err != nil {
// 		return nil, err
// 	}

// 	//Calculate hi
// 	return signed_data, nil
// }
