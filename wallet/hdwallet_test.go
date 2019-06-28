package wallet

import (
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/magiconair/properties/assert"
	"testing"
)

/*
		Unit test for NewMasterKey function
 */

func TestHDWalletNewMasterKey(t *testing.T){
	data := []struct{
		seed []byte
	}{
		{[]byte{1,2,3}},
		{[]byte{}},		// empty array
		{[]byte{1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1}},			// 32 bytes
		{[]byte{1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1}},  // 64 bytes
	}

	for _, item := range data {
		masterKey, err := NewMasterKey(item.seed)

		assert.Equal(t, nil, err)
		assert.Equal(t, ChildNumberLen, len(masterKey.ChildNumber))
		assert.Equal(t, ChainCodeLen, len(masterKey.ChainCode))
		assert.Equal(t, privacy.PublicKeySize, len(masterKey.KeySet.PaymentAddress.Pk))
		assert.Equal(t, privacy.TransmissionKeySize, len(masterKey.KeySet.PaymentAddress.Tk))
		assert.Equal(t, privacy.PrivateKeySize, len(masterKey.KeySet.PrivateKey))
		assert.Equal(t, privacy.ReceivingKeySize, len(masterKey.KeySet.ReadonlyKey.Rk))
	}
}

/*
		Unit test for NewChildKey function
 */
