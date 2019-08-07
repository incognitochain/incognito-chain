package privacy

import (
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"math/big"
)

// HybridCipherText represents to HybridCipherText for Hybrid encryption
// Hybrid encryption uses AES scheme to encrypt message with arbitrary size
// and uses Elgamal encryption to encrypt AES key
type HybridCipherText struct {
	MsgEncrypted    []byte
	SymKeyEncrypted []byte
}

func (ciphertext HybridCipherText) GetMsgEncrypted() []byte {
	return ciphertext.MsgEncrypted
}

func (ciphertext HybridCipherText) GetSymKeyEncrypted() []byte {
	return ciphertext.SymKeyEncrypted
}

// isNil check whether ciphertext is nil or not
func (ciphertext HybridCipherText) isNil() bool {
	if len(ciphertext.MsgEncrypted) == 0 {
		return true
	}

	return len(ciphertext.SymKeyEncrypted) == 0
}

// Bytes converts ciphertext to bytes array
// if ciphertext is nil, return empty byte array
func (ciphertext HybridCipherText) Bytes() []byte {
	if ciphertext.isNil() {
		return []byte{}
	}

	res := make([]byte, 0)
	res = append(res, ciphertext.SymKeyEncrypted...)
	res = append(res, ciphertext.MsgEncrypted...)

	return res
}

// SetBytes reverts bytes array to HybridCipherText
func (ciphertext *HybridCipherText) SetBytes(bytes []byte) error {
	if len(bytes) == 0 {
		return NewPrivacyErr(InvalidInputToSetBytesErr, nil)
	}

	ciphertext.SymKeyEncrypted = bytes[0:elGamalCiphertextSize]
	ciphertext.MsgEncrypted = bytes[elGamalCiphertextSize:]
	return nil
}

// hybridEncrypt encrypts message with any size, using Publickey to encrypt
// hybridEncrypt generates AES key by randomize an elliptic point aesKeyPoint and get X-coordinate
// using AES key to encrypt message
// After that, using ElGamal encryption encrypt aesKeyPoint using PublicKey
func hybridEncrypt(msg []byte, publicKey *EllipticPoint) (ciphertext *HybridCipherText, err error) {
	ciphertext = new(HybridCipherText)
	// Generate a AES key as the abscissa of a random elliptic point
	aesKeyPoint := new(EllipticPoint)
	aesKeyPoint.randomize()
	aesKeyByte := common.AddPaddingBigInt(aesKeyPoint.x, common.BigIntSize)

	// Encrypt msg using aesKeyByte
	aesScheme := &common.AES{
		Key: aesKeyByte,
	}

	ciphertext.MsgEncrypted, err = aesScheme.Encrypt(msg)
	if err != nil {
		return nil, err
	}

	// Using ElGamal cryptosystem for encrypting AES sym key
	pubKey := new(elGamalPublicKey)
	pubKey.h = new(EllipticPoint)
	pubKey.h.Set(publicKey.x, publicKey.y)

	ciphertext.SymKeyEncrypted = pubKey.encrypt(aesKeyPoint).Bytes()

	return ciphertext, nil
}

// hybridDecrypt receives a ciphertext and privateKey
// it decrypts aesKeyPoint, using ElGamal encryption with privateKey
// Using X-coordinate of aesKeyPoint to decrypts message
func hybridDecrypt(ciphertext *HybridCipherText, privateKey *big.Int) (msg []byte, err error) {
	// Validate ciphertext
	if ciphertext.isNil() {
		return []byte{}, errors.New("ciphertext must not be nil")
	}

	// Get receiving key, which is a private key of ElGamal cryptosystem
	privKey := new(elGamalPrivateKey)
	privKey.set(privateKey)

	// Parse encrypted AES key encoded as an elliptic point from EncryptedSymKey
	encryptedAESKey := new(elGamalCipherText)
	err = encryptedAESKey.SetBytes(ciphertext.SymKeyEncrypted)
	if err != nil {
		return []byte{}, err
	}

	// Decrypt encryptedAESKey using recipient's receiving key
	aesKeyPoint, err := privKey.decrypt(encryptedAESKey)
	if err != nil {
		return []byte{}, err
	}

	// Get AES key
	aesScheme := &common.AES{
		Key: common.AddPaddingBigInt(aesKeyPoint.x, common.BigIntSize),
	}

	// Decrypt encrypted coin Randomness using AES keysatt
	msg, err = aesScheme.Decrypt(ciphertext.MsgEncrypted)
	if err != nil {
		return []byte{}, err
	}
	return msg, nil
}
