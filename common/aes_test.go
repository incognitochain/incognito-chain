package common

import (
	"crypto/aes"
	"github.com/stretchr/testify/assert"
	"testing"
)

/*
	Unit test for Encrypt AES function
 */

func TestAESEncrypt(t *testing.T){
	data := [][]byte{
		{1},
		{1,2,3},
		{1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5},	// 25 bytes
		{1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5},	// 30 bytes
	}

	aesObj := AES{
		Key: []byte{1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32},
	}

	for _, item := range data{
		ciphertext, err := aesObj.Encrypt(item)
		assert.Equal(t, nil, err)
		assert.Equal(t, aes.BlockSize+len(item), len(ciphertext))
	}
}

func TestAESEncryptWithEmptyPlaintext(t *testing.T){
	aesObj := AES{
		Key: []byte{1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32},
	}

	ciphertext, err := aesObj.Encrypt([]byte{})
	assert.Equal(t, PlainTextIsEmptyErr, err)
	assert.Equal(t, 0, len(ciphertext))
}

func TestAESEncryptWithInvalidKey(t *testing.T){
	dataKey := [][]byte{
		{1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31},
		{1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33},
	}

	for _, item := range dataKey{
		aesObj := AES{
			Key: item,
		}

		ciphertext, err := aesObj.Encrypt([]byte{1,2,3})
		assert.Equal(t, InvalidAESKeyErr, err)
		assert.Equal(t, 0, len(ciphertext))
	}
}

/*
	Unit test for Decrypt AES function
 */

func TestAESDecrypt(t *testing.T){
	data := [][]byte{
		{1},
		{1,2,3},
		{1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5},	// 25 bytes
		{1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5},	// 30 bytes
	}

	aesObj := AES{
		Key: []byte{1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32},
	}

	for _, item := range data{
		// encrypt
		ciphertext, _ := aesObj.Encrypt(item)

		// decrypt
		plaintext, err := aesObj.Decrypt(ciphertext)

		assert.Equal(t, nil, err)
		assert.Equal(t, item, plaintext)
	}
}

func TestAESDecryptWithEmptyCiphertext(t *testing.T){
	aesObj := AES{
		Key: []byte{1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32},
	}

	ciphertext, err := aesObj.Decrypt([]byte{})
	assert.Equal(t, CipherTextIsEmptyErr, err)
	assert.Equal(t, 0, len(ciphertext))
}

func TestAESDecryptWithUnmatchedKey(t *testing.T){
	encryptionKey := []byte{1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32}
	decryptionKey := []byte{1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,33}
	plaintext := []byte{1,2,3}

	aesEncrypt := AES{
		Key: encryptionKey,
	}
	ciphertext, err := aesEncrypt.Encrypt(plaintext)

	aesDecrypt := AES{
		Key: decryptionKey,
	}
	plaintext2, err := aesDecrypt.Decrypt(ciphertext)

	assert.Equal(t, nil, err)
	assert.NotEqual(t, plaintext, plaintext2)
}