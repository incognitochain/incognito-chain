package cashec

import (
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

/*
		Unit test for GenerateKey function
 */

func TestKeySetGenerateKey(t *testing.T){
	data := [][]byte {
		{},
		{1},
		{1,2,3},
		{1,2,3,4,1,2,3,4,1,2,3,4,1,2,3,4,1,2,3,4,1,2,3,4,1,2,3,4,1,2,3,4},	// 32 bytes
	}

	keySet := new(KeySet)

	for _, item := range data {
		keySet = keySet.GenerateKey(item)
		assert.Equal(t, privacy.PrivateKeySize, len(keySet.PrivateKey))
		assert.Equal(t, privacy.PublicKeySize, len(keySet.PaymentAddress.Pk))
		assert.Equal(t, privacy.TransmissionKeySize, len(keySet.PaymentAddress.Tk))
		assert.Equal(t, privacy.ReceivingKeySize, len(keySet.ReadonlyKey.Rk))
	}
}

/*
		Unit test for ImportFromPrivateKeyByte function
 */

func TestKeySetImportFromPrivateKeyByte(t *testing.T){
	privateKey := privacy.RandBytes(privacy.PrivateKeySize)

	keySet := new(KeySet)
	err := keySet.ImportFromPrivateKeyByte(privateKey)

	assert.Equal(t, nil, err)
	assert.Equal(t, privacy.PrivateKeySize, len(keySet.PrivateKey[:]))
	assert.Equal(t, privacy.PublicKeySize, len(keySet.PaymentAddress.Pk))
	assert.Equal(t, privacy.TransmissionKeySize, len(keySet.PaymentAddress.Tk))
	assert.Equal(t, privacy.ReceivingKeySize, len(keySet.ReadonlyKey.Rk))
}

func TestKeySetImportFromPrivateKeyByteWithInvalidPrivKey(t *testing.T){
	keySet := new(KeySet)
	err := keySet.ImportFromPrivateKeyByte(nil)
	assert.Equal(t, errors.Wrap(nil, "Priv key is invalid"), err)

	err2 := keySet.ImportFromPrivateKeyByte([]byte{1,2,3})
	assert.Equal(t, errors.Wrap(nil, "Priv key is invalid"), err2)
}
