package wallet

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

type AES struct {
}

// create a salt
func (self AES) deriveKey(passPhrase string, salt []byte) ([]byte, []byte) {
	if salt == nil {
		salt = make([]byte, 8)
		rand.Read(salt)
	}
	return pbkdf2.Key([]byte(passPhrase), salt, 1000, 32, sha256.New), salt
}

// encrypt with AES
func (self AES) Encrypt(passphrase string, plaintext []byte) (string, error) {
	key, salt := self.deriveKey(passphrase, nil)
	iv := make([]byte, 12)
	// http://nvlpubs.nist.gov/nistpubs/Legacy/SP/nistspecialpublication800-38d.pdf
	// Section 8.2
	rand.Read(iv)
	b, err := aes.NewCipher(key)
	aesgcm, err := cipher.NewGCM(b)
	data := aesgcm.Seal(nil, iv, plaintext, nil)
	return hex.EncodeToString(salt) + "-" + hex.EncodeToString(iv) + "-" + hex.EncodeToString(data), err
}

// decrypt with AES
func (self AES) Decrypt(passPhrase, cipherText string) ([]byte, error) {
	arr := strings.Split(cipherText, "-")
	salt, err := hex.DecodeString(arr[0])
	iv, err := hex.DecodeString(arr[1])
	data, err := hex.DecodeString(arr[2])
	key, _ := self.deriveKey(passPhrase, salt)
	b, err := aes.NewCipher(key)
	aesgcm, err := cipher.NewGCM(b)
	data, err = aesgcm.Open(nil, iv, data, nil)
	return data, err
}
