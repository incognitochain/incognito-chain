package cashec

import (
	// "github.com/ninjadotorg/constant/privacy/client"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/privacy"
)

type KeySet struct {
	// ProducerKeyPair KeyPair
	PrivateKey     privacy.SpendingKey
	PaymentAddress privacy.PaymentAddress
	ReadonlyKey    privacy.ViewingKey
}

/*
GenerateKey - generate key set from seed byte[]
*/
func (self *KeySet) GenerateKey(seed []byte) *KeySet {
	self.PrivateKey = privacy.GenerateSpendingKey(seed)
	self.PaymentAddress = privacy.GeneratePaymentAddress(self.PrivateKey[:])
	self.ReadonlyKey = privacy.GenerateViewingKey(self.PrivateKey[:])
	return self
}

/*
ImportFromPrivateKeyByte - from private-key byte[], regenerate pub-key and readonly-key
*/
func (self *KeySet) ImportFromPrivateKeyByte(privateKey []byte) {
	copy(self.PrivateKey[:], privateKey)
	self.PaymentAddress = privacy.GeneratePaymentAddress(self.PrivateKey[:])
	self.ReadonlyKey = privacy.GenerateViewingKey(self.PrivateKey[:])
}

/*
ImportFromPrivateKeyByte - from private-key data, regenerate pub-key and readonly-key
*/
func (self *KeySet) ImportFromPrivateKey(privateKey *privacy.SpendingKey) {
	self.PrivateKey = *privateKey
	self.PaymentAddress = privacy.GeneratePaymentAddress(self.PrivateKey[:])
	self.ReadonlyKey = privacy.GenerateViewingKey(self.PrivateKey[:])
}

func (self *KeySet) Verify(data, signature []byte) (bool, error) {
	hash := common.HashB(data)
	isValid := false

	pubKeySig := new(privacy.SchnPubKey)
	PK, err := privacy.DecompressKey(self.PaymentAddress.Pk)
	if err != nil{
		return false, err
	}
	pubKeySig.Set(PK)

	signatureSetBytes := new(privacy.SchnSignature)
	signatureSetBytes.SetBytes(signature)

	isValid = pubKeySig.Verify(signatureSetBytes, hash)
	return isValid, nil
}

func (self *KeySet) Sign(data []byte) ([]byte, error) {
	hash := common.HashB(data)
	privKeySig := new(privacy.SchnPrivKey)
	privKeySig.Set(new(big.Int).SetBytes(self.PrivateKey), big.NewInt(0))

	signature, err := privKeySig.Sign(hash)
	return signature.Bytes(), err
}

func (self *KeySet) SignDataB58(data []byte) (string, error) {
	signatureByte, err := self.Sign(data)
	if err != nil {
		return common.EmptyString, errors.New("Can't sign data. " + err.Error())
	}
	return base58.Base58Check{}.Encode(signatureByte, common.ZeroByte), nil
}

func (self *KeySet) Encrypt(data []byte) ([]byte, error) {
	// encryptText := client.Encrypt(self.PaymentAddress.Pk[:], data)
	encryptText := []byte{0}
	return encryptText, nil
}

func (self *KeySet) Decrypt(data []byte) ([]byte, error) {
	// data, err := client.Decrypt(self.PrivateKey[:], data)
	data = []byte{0}
	return data, nil
}

func (self *KeySet) EncodeToString() string {
	val, _ := json.Marshal(self)
	result := base58.Base58Check{}.Encode(val, common.ZeroByte)
	return result
}

func (self *KeySet) DecodeToKeySet(keystring string) (*KeySet, error) {
	base58C := base58.Base58Check{}
	keyBytes, _, _ := base58C.Decode(keystring)
	json.Unmarshal(keyBytes, self)
	return self, nil
}

func (self *KeySet) GetViewingKey() (privacy.ViewingKey, error) {
	return self.ReadonlyKey, nil
}

func (self *KeySet) GetPublicKeyB58() string {
	return base58.Base58Check{}.Encode(self.PaymentAddress.Pk, common.ZeroByte)
}

func ValidateDataB58(pbkB58 string, sigB58 string, data []byte) error {
	decPubkey, _, err := base58.Base58Check{}.Decode(pbkB58)
	if err != nil {
		return errors.New("can't decode public key:" + err.Error())
	}

	validatorKp := KeySet{}
	validatorKp.PaymentAddress.Pk = make([]byte, len(decPubkey))
	copy(validatorKp.PaymentAddress.Pk[:], decPubkey)

	decSig, _, err := base58.Base58Check{}.Decode(sigB58)
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
