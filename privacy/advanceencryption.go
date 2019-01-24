package privacy

import (
	"errors"
	"math/big"
)

type Ciphertext struct {
	msgEncrypted    []byte
	symKeyEncrypted []byte
}

func (ciphertext *Ciphertext) IsNil() bool {
	if len(ciphertext.msgEncrypted) == 0 {
		return true
	}
	if len(ciphertext.symKeyEncrypted) == 0 {
		return true
	}
	return false
}

func (ciphertext *Ciphertext) Bytes() []byte {
	if ciphertext.IsNil() {
		return []byte{}
	}

	res := make([]byte, 0)
	res = append(res, ciphertext.symKeyEncrypted...)
	res = append(res, ciphertext.msgEncrypted...)

	return res
}

func (ciphertext *Ciphertext) SetBytes(bytes []byte) error {
	if len(bytes) == 0 {
		return errors.New("SetBytes ciphertext encryption: invalid input")
	}
	ciphertext.symKeyEncrypted = bytes[0:66]
	ciphertext.msgEncrypted = bytes[66:]
	return nil
}

// AdvanceEncrypt encrypts message with any size, using Publickey to encrypt
func AdvanceEncrypt(msg []byte, publicKey *EllipticPoint) (ciphertext *Ciphertext, err error) {
	ciphertext = new(Ciphertext)
	// Generate a AES key as the abscissa of a random elliptic point
	aesKeyPoint := new(EllipticPoint)
	aesKeyPoint.Randomize()
	aesKeyByte := aesKeyPoint.X.Bytes()

	// Encrypt msg using aesKeyByte
	aesScheme := &AES{
		Key: aesKeyByte,
	}

	ciphertext.msgEncrypted, err = aesScheme.encrypt(msg)
	if err != nil {
		return nil, err
	}

	// Using ElGamal cryptosystem for encrypting AES sym key
	pubKey := new(ElGamalPubKey)
	pubKey.H = new(EllipticPoint)
	pubKey.H.Set(publicKey.X, publicKey.Y)

	ciphertext.symKeyEncrypted = pubKey.Encrypt(aesKeyPoint).Bytes()

	return ciphertext, nil
}

func AdvanceDecrypt(ciphertext *Ciphertext, privateKey *big.Int) (msg []byte, err error) {
	// Validate ciphertext
	if ciphertext.IsNil() {
		return []byte{}, errors.New("ciphertext must not be nil")
	}

	// Get receiving key, which is a private key of ElGamal cryptosystem
	privKey := new(ElGamalPrivKey)
	privKey.Set(privateKey)

	// Parse encrypted AES key encoded as an elliptic point from EncryptedSymKey
	encryptedAESKey := new(ElGamalCiphertext)
	err = encryptedAESKey.SetBytes(ciphertext.symKeyEncrypted)
	if err != nil {
		return []byte{}, err
	}

	// Decrypt encryptedAESKey using recipient's receiving key
	aesKeyPoint, _ := privKey.Decrypt(encryptedAESKey)

	// Get AES key
	aesScheme := &AES{
		Key: aesKeyPoint.X.Bytes(),
	}

	// Decrypt encrypted coin randomness using AES key
	msg, err = aesScheme.decrypt(ciphertext.msgEncrypted)
	if err != nil {
		return []byte{}, err
	}
	return msg, nil
}
