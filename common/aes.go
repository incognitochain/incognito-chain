package common

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"errors"
)

var PlainTextIsEmptyErr = errors.New("plaintext is empty")
var CipherTextIsEmptyErr = errors.New("ciphertext is empty")
var InvalidAESKeyErr = errors.New("aes key is invalid")

type AES struct {
	Key []byte
}

func (aesObj *AES) Encrypt(plaintext []byte) ([]byte, error) {
	if len(plaintext) == 0{
		return []byte{}, PlainTextIsEmptyErr
	}

	block, err := aes.NewCipher(aesObj.Key)
	if err != nil {
		return nil, InvalidAESKeyErr
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))

	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)
	return ciphertext, nil
}

func (aesObj *AES) Decrypt(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) == 0 {
		return []byte{}, CipherTextIsEmptyErr
	}

	plaintext := make([]byte, len(ciphertext[aes.BlockSize:]))

	block, err := aes.NewCipher(aesObj.Key)
	if err != nil {
		return nil, err
	}

	iv := ciphertext[:aes.BlockSize]
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(plaintext, ciphertext[aes.BlockSize:])

	return plaintext, nil
}
