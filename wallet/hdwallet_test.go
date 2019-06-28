package wallet

import (
	"github.com/incognitochain/incognito-chain/common"
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

func TestHDWalletNewChildKey(t *testing.T) {
	seed := []byte{1,2,3}
	masterKey, _ := NewMasterKey(seed)

	data := []struct{
		childIdx uint32
	}{
		{uint32(0)},
		{uint32(1)},
		{uint32(2)},
		{uint32(3)},
		{uint32(4)},
	}

	for _, item := range data {
		childKey, err := masterKey.NewChildKey(item.childIdx)

		assert.Equal(t, nil, err)
		assert.Equal(t, common.Uint32ToBytes(item.childIdx), childKey.ChildNumber)
		assert.Equal(t, ChainCodeLen, len(childKey.ChainCode))
		assert.Equal(t, masterKey.Depth + 1, childKey.Depth)
		assert.Equal(t, privacy.PublicKeySize, len(childKey.KeySet.PaymentAddress.Pk))
		assert.Equal(t, privacy.TransmissionKeySize, len(childKey.KeySet.PaymentAddress.Tk))
		assert.Equal(t, privacy.PrivateKeySize, len(childKey.KeySet.PrivateKey))
		assert.Equal(t, privacy.ReceivingKeySize, len(childKey.KeySet.ReadonlyKey.Rk))
	}
}

func TestHDWalletNewChildKeyFromOtherChildKey(t *testing.T) {
	seed := []byte{1,2,3}
	masterKey, _ := NewMasterKey(seed)
	childKey1, _ := masterKey.NewChildKey(uint32(1))

	childIndex := uint32(10)
	childKey2, err := childKey1.NewChildKey(childIndex)

	assert.Equal(t, nil, err)
	assert.Equal(t, common.Uint32ToBytes(childIndex), childKey2.ChildNumber)
	assert.Equal(t, ChainCodeLen, len(childKey2.ChainCode))
	assert.Equal(t, childKey1.Depth + 1, childKey2.Depth)
	assert.Equal(t, privacy.PublicKeySize, len(childKey2.KeySet.PaymentAddress.Pk))
	assert.Equal(t, privacy.TransmissionKeySize, len(childKey2.KeySet.PaymentAddress.Tk))
	assert.Equal(t, privacy.PrivateKeySize, len(childKey2.KeySet.PrivateKey))
	assert.Equal(t, privacy.ReceivingKeySize, len(childKey2.KeySet.ReadonlyKey.Rk))
}

func TestHDWalletNewChildKeyWithSameChildIdx(t *testing.T) {
	seed := []byte{1,2,3}
	masterKey, _ := NewMasterKey(seed)

	childIndex := uint32(10)
	childKey1, err1 := masterKey.NewChildKey(childIndex)
	childKey2, err2 := masterKey.NewChildKey(childIndex)

	assert.Equal(t, nil, err1)
	assert.Equal(t, nil, err2)
	assert.Equal(t, childKey1.ChildNumber, childKey2.ChildNumber)
	assert.Equal(t, childKey1.ChainCode, childKey2.ChainCode)
	assert.Equal(t, childKey1.Depth, childKey2.Depth)
	assert.Equal(t, childKey1.KeySet.PaymentAddress.Pk, childKey2.KeySet.PaymentAddress.Pk)
	assert.Equal(t, ChainCodeLen, len(childKey2.ChainCode))
	assert.Equal(t, masterKey.Depth + 1, childKey2.Depth)
}

