package cashec

import (
	// "github.com/ninjadotorg/constant/privacy/client"
	"encoding/json"
	"errors"
	errors2 "github.com/pkg/errors"
	"math/big"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/privacy"
)

// This is real raw data of wallet account, which user can use to
// - spend and check double spend coin with private key
// - receive coin with payment address
// - read tx data with readonly key
type KeySet struct {
	PrivateKey     privacy.SpendingKey
	PaymentAddress privacy.PaymentAddress
	ReadonlyKey    privacy.ViewingKey
}

/*
GenerateKey - generate key set from seed byte[]
*/
func (keysetObj *KeySet) GenerateKey(seed []byte) *KeySet {
	keysetObj.PrivateKey = privacy.GenerateSpendingKey(seed)
	keysetObj.PaymentAddress = privacy.GeneratePaymentAddress(keysetObj.PrivateKey[:])
	keysetObj.ReadonlyKey = privacy.GenerateViewingKey(keysetObj.PrivateKey[:])
	return keysetObj
}

/*
ImportFromPrivateKeyByte - from private-key byte[], regenerate pub-key and readonly-key
*/
func (keysetObj *KeySet) ImportFromPrivateKeyByte(privateKey []byte) (error) {
	if privateKey == nil {
		return errors2.Wrap(nil, "Priv key is invalid")
	}
	if len(privateKey) != 32 {
		return errors2.Wrap(nil, "Priv key is invalid")
	}
	copy(keysetObj.PrivateKey[:], privateKey)
	keysetObj.PaymentAddress = privacy.GeneratePaymentAddress(keysetObj.PrivateKey[:])
	keysetObj.ReadonlyKey = privacy.GenerateViewingKey(keysetObj.PrivateKey[:])
	return nil
}

/*
ImportFromPrivateKeyByte - from private-key data, regenerate pub-key and readonly-key
*/
func (keysetObj *KeySet) ImportFromPrivateKey(privateKey *privacy.SpendingKey) {
	keysetObj.PrivateKey = *privateKey
	keysetObj.PaymentAddress = privacy.GeneratePaymentAddress(keysetObj.PrivateKey[:])
	keysetObj.ReadonlyKey = privacy.GenerateViewingKey(keysetObj.PrivateKey[:])
}

func (keysetObj *KeySet) Verify(data, signature []byte) (bool, error) {
	hash := common.HashB(data)
	isValid := false

	pubKeySig := new(privacy.SchnPubKey)
	PK := new(privacy.EllipticPoint)
	err := PK.Decompress(keysetObj.PaymentAddress.Pk)
	if err != nil {
		return false, err
	}
	pubKeySig.Set(PK)

	signatureSetBytes := new(privacy.SchnSignature)
	signatureSetBytes.SetBytes(signature)

	isValid = pubKeySig.Verify(signatureSetBytes, hash)
	return isValid, nil
}

func (keysetObj *KeySet) Sign(data []byte) ([]byte, error) {
	hash := common.HashB(data)
	privKeySig := new(privacy.SchnPrivKey)
	privKeySig.Set(new(big.Int).SetBytes(keysetObj.PrivateKey), big.NewInt(0))

	signature, err := privKeySig.Sign(hash)
	return signature.Bytes(), err
}

func (keysetObj *KeySet) SignBase58(data []byte) (string, error) {
	signatureByte, err := keysetObj.Sign(data)
	if err != nil {
		return "", errors.New("Can't sign data. " + err.Error())
	}
	return base58.Base58Check{}.Encode(signatureByte, common.ZeroByte), nil
}

func (keysetObj *KeySet) EncodeToString() string {
	val, _ := json.Marshal(keysetObj)
	result := base58.Base58Check{}.Encode(val, common.ZeroByte)
	return result
}

func (keysetObj *KeySet) DecodeToKeySet(keystring string) (*KeySet, error) {
	base58C := base58.Base58Check{}
	keyBytes, _, _ := base58C.Decode(keystring)
	json.Unmarshal(keyBytes, keysetObj)
	return keysetObj, nil
}

func (keysetObj *KeySet) GetViewingKey() (privacy.ViewingKey, error) {
	return keysetObj.ReadonlyKey, nil
}

func (keysetObj *KeySet) GetPublicKeyB58() string {
	return base58.Base58Check{}.Encode(keysetObj.PaymentAddress.Pk, common.ZeroByte)
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
		return errors.New("invalid signature")
	}
	return nil
}

func (keysetObj *KeySet) SignDataB58(data []byte) (string, error) {
	signatureByte, err := keysetObj.Sign(data)
	if err != nil {
		return "", errors.New("can't sign data. " + err.Error())
	}
	return base58.Base58Check{}.Encode(signatureByte, common.ZeroByte), nil
}
