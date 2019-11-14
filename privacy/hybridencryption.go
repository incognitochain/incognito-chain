package privacy

import (
	"encoding/json"
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
)

// hybridCipherText_Old represents to hybridCipherText_Old for Hybrid encryption
// Hybrid encryption uses AES scheme to encrypt message with arbitrary size
// and uses Elgamal encryption to encrypt AES key
type HybridCipherText struct {
	msgEncrypted    []byte
	symKeyEncrypted []byte
}

func (ciphertext HybridCipherText) GetMsgEncrypted() []byte {
	return ciphertext.msgEncrypted
}

func (ciphertext HybridCipherText) GetSymKeyEncrypted() []byte {
	return ciphertext.symKeyEncrypted
}

// isNil check whether ciphertext is nil or not
func (ciphertext HybridCipherText) IsNil() bool {
	if len(ciphertext.msgEncrypted) == 0 {
		return true
	}

	return len(ciphertext.symKeyEncrypted) == 0
}

func (hybridCipherText HybridCipherText) MarshalJSON() ([]byte, error) {
	data := hybridCipherText.Bytes()
	temp := base58.Base58Check{}.Encode(data, common.ZeroByte)
	return json.Marshal(temp)
}

func (hybridCipherText *HybridCipherText) UnmarshalJSON(data []byte) error {
	dataStr := ""
	_ = json.Unmarshal(data, &dataStr)
	temp, _, err := base58.Base58Check{}.Decode(dataStr)
	if err != nil {
		return err
	}
	hybridCipherText.SetBytes(temp)
	return nil
}

// Bytes converts ciphertext to bytes array
// if ciphertext is nil, return empty byte array
func (ciphertext HybridCipherText) Bytes() []byte {
	if ciphertext.IsNil() {
		return []byte{}
	}

	res := make([]byte, 0)
	res = append(res, ciphertext.symKeyEncrypted...)
	res = append(res, ciphertext.msgEncrypted...)

	return res
}

// SetBytes reverts bytes array to hybridCipherText_Old
func (ciphertext *HybridCipherText) SetBytes(bytes []byte) error {
	if len(bytes) == 0 {
		return NewPrivacyErr(InvalidInputToSetBytesErr, nil)
	}

	if len(bytes) < elGamalCiphertextSize {
		// out of range
		return errors.New("out of range Parse ciphertext")
	}
	ciphertext.symKeyEncrypted = bytes[0:elGamalCiphertextSize]
	ciphertext.msgEncrypted = bytes[elGamalCiphertextSize:]
	return nil
}

// hybridEncrypt_Old encrypts message with any size, using Publickey to encrypt
// hybridEncrypt_Old generates AES key by randomize an elliptic point aesKeyPoint and get X-coordinate
// using AES key to encrypt message
// After that, using ElGamal encryption encrypt aesKeyPoint using publicKey
func HybridEncrypt(msg []byte, publicKey *Point) (ciphertext *HybridCipherText, err error) {
	ciphertext = new(HybridCipherText)

	// Generate a AES key bytes
	sKeyPoint := RandomPoint()
	sKeyByte := sKeyPoint.ToBytes()
	// Encrypt msg using aesKeyByte

	aesKey := sKeyByte[:]
	aesScheme := &common.AES{
		Key: aesKey,
	}
	ciphertext.msgEncrypted, err = aesScheme.Encrypt(msg)
	if err != nil {
		return nil, err
	}

	// Using ElGamal cryptosystem for encrypting AES sym key
	pubKey := new(elGamalPublicKey)
	pubKey.h = publicKey
	ciphertext.symKeyEncrypted = pubKey.encrypt(sKeyPoint).Bytes()

	return ciphertext, nil
}

// hybridDecrypt_Old receives a ciphertext and privateKey
// it decrypts aesKeyPoint, using ElGamal encryption with privateKey
// Using X-coordinate of aesKeyPoint to decrypts message
func HybridDecrypt(ciphertext *HybridCipherText, privateKey *Scalar) (msg []byte, err error) {
	// Validate ciphertext
	if ciphertext.IsNil() {
		return []byte{}, errors.New("ciphertext must not be nil")
	}

	// Get receiving key, which is a private key of ElGamal cryptosystem
	privKey := new(elGamalPrivateKey)
	privKey.set(privateKey)

	// Parse encrypted AES key encoded as an elliptic point from EncryptedSymKey
	encryptedAESKey := new(elGamalCipherText)
	err = encryptedAESKey.SetBytes(ciphertext.symKeyEncrypted)
	if err != nil {
		return []byte{}, err
	}

	// Decrypt encryptedAESKey using recipient's receiving key
	aesKeyPoint, err := privKey.decrypt(encryptedAESKey)
	if err != nil {
		return []byte{}, err
	}

	// Get AES key
	aesKeyByte := aesKeyPoint.ToBytes()
	aesKey := aesKeyByte[:]
	aesScheme := &common.AES{
		Key: aesKey,
	}

	// Decrypt encrypted coin randomness using AES keysatt
	msg, err = aesScheme.Decrypt(ciphertext.msgEncrypted)
	if err != nil {
		return []byte{}, err
	}
	return msg, nil
}
