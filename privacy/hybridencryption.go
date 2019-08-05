package privacy

import (
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"math/big"
)

// Ciphertext represents to Ciphertext for Hybrid encryption
// Hybrid encryption uses AES scheme to encrypt message with arbitrary size
// and uses Elgamal encryption to encrypt AES key
type Ciphertext struct {
	MsgEncrypted    []byte
	SymKeyEncrypted []byte
}

// IsNil check whether ciphertext is nil or not
func (ciphertext *Ciphertext) IsNil() bool {
	if len(ciphertext.MsgEncrypted) == 0 {
		return true
	}

	return len(ciphertext.SymKeyEncrypted) == 0
}

// Bytes converts ciphertext to bytes array
// if ciphertext is nil, return empty byte array
func (ciphertext *Ciphertext) Bytes() []byte {
	if ciphertext.IsNil() {
		return []byte{}
	}

	res := make([]byte, 0)
	res = append(res, ciphertext.SymKeyEncrypted...)
	res = append(res, ciphertext.MsgEncrypted...)

	return res
}

// SetBytes reverts bytes array to Ciphertext
func (ciphertext *Ciphertext) SetBytes(bytes []byte) error {
	if len(bytes) == 0 {
		return NewPrivacyErr(InvalidInputToSetBytesErr, nil)
	}
	ciphertext.SymKeyEncrypted = bytes[0:elGamalCiphertextSize]
	ciphertext.MsgEncrypted = bytes[elGamalCiphertextSize:]
	return nil
}

// HybridEncrypt encrypts message with any size, using Publickey to encrypt
// HybridEncrypt generates AES key by randomize an elliptic point aesKeyPoint and get X-coordinate
// using AES key to encrypt message
// After that, using ElGamal encryption encrypt aesKeyPoint using publicKey
func HybridEncrypt(msg []byte, publicKey *EllipticPoint) (ciphertext *Ciphertext, err error) {
	ciphertext = new(Ciphertext)
	// Generate a AES key as the abscissa of a random elliptic point
	aesKeyPoint := new(EllipticPoint)
	aesKeyPoint.Randomize()
	aesKeyByte := common.AddPaddingBigInt(aesKeyPoint.X, common.BigIntSize)

	// Encrypt msg using aesKeyByte
	aesScheme := &common.AES{
		Key: aesKeyByte,
	}

	ciphertext.MsgEncrypted, err = aesScheme.Encrypt(msg)
	if err != nil {
		return nil, err
	}

	// Using ElGamal cryptosystem for encrypting AES sym key
	pubKey := new(ElGamalPubKey)
	pubKey.H = new(EllipticPoint)
	pubKey.H.Set(publicKey.X, publicKey.Y)

	ciphertext.SymKeyEncrypted = pubKey.Encrypt(aesKeyPoint).Bytes()

	return ciphertext, nil
}

// HybridDecrypt receives a ciphertext and privateKey
// it decrypts aesKeyPoint, using ElGamal encryption with privateKey
// Using X-coordinate of aesKeyPoint to decrypts message
func HybridDecrypt(ciphertext *Ciphertext, privateKey *big.Int) (msg []byte, err error) {
	// Validate ciphertext
	if ciphertext.IsNil() {
		return []byte{}, errors.New("ciphertext must not be nil")
	}

	// Get receiving key, which is a private key of ElGamal cryptosystem
	privKey := new(ElGamalPrivKey)
	privKey.Set(privateKey)

	// Parse encrypted AES key encoded as an elliptic point from EncryptedSymKey
	encryptedAESKey := new(ElGamalCiphertext)
	err = encryptedAESKey.SetBytes(ciphertext.SymKeyEncrypted)
	if err != nil {
		return []byte{}, err
	}

	// Decrypt encryptedAESKey using recipient's receiving key
	aesKeyPoint, err := privKey.Decrypt(encryptedAESKey)
	if err != nil {
		return []byte{}, err
	}

	// Get AES key
	aesScheme := &common.AES{
		Key: common.AddPaddingBigInt(aesKeyPoint.X, common.BigIntSize),
	}

	// Decrypt encrypted coin randomness using AES key
	msg, err = aesScheme.Decrypt(ciphertext.MsgEncrypted)
	if err != nil {
		return []byte{}, err
	}
	return msg, nil
}
