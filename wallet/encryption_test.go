package wallet

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

/*
		Unit test for deriveKey function
 */
func TestAESDeriveKey(t *testing.T){
	passPhrase := "123"
	salt := []byte{1,2,3}

	newKey, newSalt := deriveKey(passPhrase, salt)

	assert.Equal(t, common.AESKeySize, len(newKey))
	assert.Equal(t, salt, newSalt)
}

func TestAESDeriveKeyWithEmptyPassPhrase(t *testing.T) {
	passPhrase := ""
	salt := []byte{1, 2, 3}

	newKey, newSalt := deriveKey(passPhrase, salt)

	assert.Equal(t, common.AESKeySize, len(newKey))
	assert.Equal(t, salt, newSalt)
}

func TestAESDeriveKeyWithEmptySalt(t *testing.T) {
	passPhrase := "123"
	salt := []byte{}

	newKey, newSalt := deriveKey(passPhrase, salt)

	assert.Equal(t, common.AESKeySize, len(newKey))
	assert.Equal(t, 8, len(newSalt))
}

/*
		Unit test for EncryptByPassPhrase function
 */

func TestEncryptionEncryptByPassPhrase(t *testing.T){
	passPhrase := "123"
	plaintext := []byte{1,2,3,4}

	ciphertextStr, err := EncryptByPassPhrase(passPhrase, plaintext)
	fmt.Println("ciphertextStr : ", ciphertextStr )

	assert.Equal(t, nil, err)
	assert.Greater(t, len(ciphertextStr), 0)

	plaintext2, err := DecryptByPassPhrase(passPhrase, ciphertextStr)
	assert.Equal(t, plaintext, plaintext2)
}

func TestEncryptionEncryptByPassPhraseWithEmptyPassPhrase(t *testing.T){
	passPhrase := ""
	plaintext := []byte{1,2,3,4}

	ciphertextStr, err := EncryptByPassPhrase(passPhrase, plaintext)
	fmt.Println("ciphertextStr : ", ciphertextStr)

	assert.Equal(t, nil, err)
	assert.Greater(t, len(ciphertextStr), 0)

	plaintext2, err := DecryptByPassPhrase(passPhrase, ciphertextStr)
	assert.Equal(t, plaintext, plaintext2)
}

func TestEncryptionEncryptByPassPhraseWithEmptyPlaintext(t *testing.T){
	passPhrase := "123"
	plaintext := []byte{}

	ciphertextStr, err := EncryptByPassPhrase(passPhrase, plaintext)

	assert.Equal(t, NewWalletError(InvalidPlaintextErr, nil), err)
	assert.Equal(t, "", ciphertextStr)
}


/*
		Unit test for DecryptByPassPhrase function
 */

func TestEncryptionDecryptByPassPhraseWithUnmatchedPass(t *testing.T){
	passPhrase := "123"
	plaintext := []byte{1,2,3,4}
	ciphertextStr, _ := EncryptByPassPhrase(passPhrase, plaintext)

	passPhrase2 := "1234"
	plaintext2, err := DecryptByPassPhrase(passPhrase2, ciphertextStr)

	assert.NotEqual(t, plaintext, plaintext2)
	assert.Equal(t, nil, err)
}

func TestEncryptionDecryptByPassPhraseWithInvalidCipher(t *testing.T) {
	passPhrase := "123"
	ciphertextStr := "ciphertextabc"

	_, err := DecryptByPassPhrase(passPhrase, ciphertextStr)

	assert.NotEqual(t, nil, err)
}


