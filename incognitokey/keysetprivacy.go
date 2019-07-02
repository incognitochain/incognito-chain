package incognitokey

import (
	"encoding/json"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/privacy"
)

// KeySet is real raw data of wallet account, which user can use to
// - spend and check double spend coin with private key
// - receive coin with payment address
// - read tx data with readonly key
type KeySet struct {
	PrivateKey     privacy.PrivateKey
	PaymentAddress privacy.PaymentAddress
	ReadonlyKey    privacy.ViewingKey
}

// GenerateKey generates key set from seed in byte array
func (keySet *KeySet) GenerateKey(seed []byte) *KeySet {
	keySet.PrivateKey = privacy.GeneratePrivateKey(seed)
	keySet.PaymentAddress = privacy.GeneratePaymentAddress(keySet.PrivateKey[:])
	keySet.ReadonlyKey = privacy.GenerateViewingKey(keySet.PrivateKey[:])
	return keySet
}

// ImportFromPrivateKeyByte receives private key in bytes array,
// and regenerates payment address and readonly key
// returns error if private key is invalid
func (keySet *KeySet) ImportFromPrivateKeyByte(privateKey []byte) error {
	if len(privateKey) != privacy.PrivateKeySize {
		return NewCashecError(InvalidPrivateKeyErr, nil)
	}

	keySet.PrivateKey = privateKey
	keySet.PaymentAddress = privacy.GeneratePaymentAddress(keySet.PrivateKey[:])
	keySet.ReadonlyKey = privacy.GenerateViewingKey(keySet.PrivateKey[:])
	return nil
}

// ImportFromPrivateKey receives private key in PrivateKey type,
// and regenerates payment address and readonly key
// returns error if private key is invalid
func (keySet *KeySet) ImportFromPrivateKey(privateKey *privacy.PrivateKey) error {
	if privateKey == nil || len(*privateKey) != privacy.PrivateKeySize {
		return NewCashecError(InvalidPrivateKeyErr, nil)
	}

	keySet.PrivateKey = *privateKey
	keySet.PaymentAddress = privacy.GeneratePaymentAddress(keySet.PrivateKey[:])
	keySet.ReadonlyKey = privacy.GenerateViewingKey(keySet.PrivateKey[:])

	return nil
}

// Sign receives data in bytes array and
// returns the signature of that data using Schnorr Signature Scheme with signing key is private key in ketSet
func (keySet *KeySet) Sign(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return []byte{}, NewCashecError(InvalidDataSignErr, nil)
	}

	hash := common.HashB(data)
	privKeySig := new(privacy.SchnPrivKey)
	privKeySig.Set(new(big.Int).SetBytes(keySet.PrivateKey), big.NewInt(0))

	signature, err := privKeySig.Sign(hash)
	return signature.Bytes(), err
}

// Verify receives data and signature
// It checks whether the given signature is the signature of data
// and was signed by private key corresponding to public key in keySet or not
func (keySet *KeySet) Verify(data, signature []byte) (bool, error) {
	hash := common.HashB(data)
	isValid := false

	pubKeySig := new(privacy.SchnPubKey)
	PK := new(privacy.EllipticPoint)
	err := PK.Decompress(keySet.PaymentAddress.Pk)
	if err != nil {
		return false, NewCashecError(InvalidVerificationKeyErr, nil)
	}
	pubKeySig.Set(PK)

	signatureSetBytes := new(privacy.SchnSignature)
	signatureSetBytes.SetBytes(signature)

	isValid = pubKeySig.Verify(signatureSetBytes, hash)
	return isValid, nil
}

// EncodeToString encodes keySet to base58 check encode string
func (keySet *KeySet) EncodeToString() string {
	val, _ := json.Marshal(keySet)
	result := base58.Base58Check{}.Encode(val, common.ZeroByte)
	return result
}

// DecodeToKeySet decodes keySet from base58 check encode string
func (keySet *KeySet) DecodeToKeySet(keystring string) (*KeySet, error) {
	keyBytes, _, err := base58.Base58Check{}.Decode(keystring)
	if err != nil {
		return nil, NewCashecError(DecodeFromStringErr, nil)
	}
	json.Unmarshal(keyBytes, keySet)
	return keySet, nil
}

// GetPublicKeyB58 returns the public key which is base58 check encoded
func (keySet *KeySet) GetPublicKeyB58() string {
	return base58.Base58Check{}.Encode(keySet.PaymentAddress.Pk, common.ZeroByte)
}

// SignDataB58 receives data and
// returns the signature that is base58 check encoded and is signed by private key in keySet
func (keySet *KeySet) SignDataB58(data []byte) (string, error) {
	signatureByte, err := keySet.Sign(data)
	if err != nil {
		return "", NewCashecError(SignDataB58Err, err)
	}
	return base58.Base58Check{}.Encode(signatureByte, common.ZeroByte), nil
}

// ValidateDataB58 receives a data, a base58 check encoded signature (sigB58)
// and a base58 check encoded public key (pbkB58)
// It decodes pbkB58 and sigB58
// after that, it verifies the given signature is corresponding to data using verification key is pbkB58
func ValidateDataB58(pbkB58 string, sigB58 string, data []byte) error {
	// decode public key (verification key)
	decodedPubKey, _, err := base58.Base58Check{}.Decode(pbkB58)
	if err != nil {
		return NewCashecError(B58DecodePubKeyErr, nil)
	}
	validatorKeySet := KeySet{}
	validatorKeySet.PaymentAddress.Pk = make([]byte, len(decodedPubKey))
	copy(validatorKeySet.PaymentAddress.Pk[:], decodedPubKey)

	// decode the signature
	decodedSig, _, err := base58.Base58Check{}.Decode(sigB58)
	if err != nil {
		return NewCashecError(B58DecodeSigErr, nil)
	}

	// validate the data and signature
	isValid, err := validatorKeySet.Verify(data, decodedSig)
	if err != nil {
		return NewCashecError(B58ValidateErr, nil)
	}
	if !isValid {
		return NewCashecError(InvalidDataValidateErr, nil)
	}
	return nil
}
