package wallet

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/stretchr/testify/assert"
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

/*
		Unit test for Serialize function
 */

func TestHDWalletSerialize( t *testing.T){
	seed := []byte{1,2,3}
	masterKey, _ := NewMasterKey(seed)

	privKeyBytes, err := masterKey.Serialize(PriKeyType)
	paymentAddrBytes, err := masterKey.Serialize(PaymentAddressType)
	readonlyKeyBytes, err := masterKey.Serialize(ReadonlyKeyType)

	actualCheckSumPrivKey := privKeyBytes[PrivKeySerializedBytesLen - 4:]
	expectedCheckSumPrivKey := base58.ChecksumFirst4Bytes(privKeyBytes[:PrivKeySerializedBytesLen - 4])

	actualCheckSumPaymentAddr := paymentAddrBytes[PaymentAddrSerializedBytesLen - 4:]
	expectedCheckSumPaymentAddr := base58.ChecksumFirst4Bytes(paymentAddrBytes[:PaymentAddrSerializedBytesLen - 4])

	actualCheckSumReadOnlyKey := readonlyKeyBytes[ReadOnlyKeySerializedBytesLen - 4:]
	expectedCheckSumReadOnlyKey := base58.ChecksumFirst4Bytes(readonlyKeyBytes[:ReadOnlyKeySerializedBytesLen - 4])

	assert.Equal(t, err, nil)
	assert.Equal(t, PrivKeySerializedBytesLen, len(privKeyBytes))
	assert.Equal(t, PaymentAddrSerializedBytesLen, len(paymentAddrBytes))
	assert.Equal(t, ReadOnlyKeySerializedBytesLen, len(readonlyKeyBytes))

	assert.Equal(t, expectedCheckSumPrivKey, actualCheckSumPrivKey)
	assert.Equal(t, expectedCheckSumPaymentAddr, actualCheckSumPaymentAddr)
	assert.Equal(t, expectedCheckSumReadOnlyKey, actualCheckSumReadOnlyKey)
}

func TestHDWalletSerializeWithInvalidKeyType( t *testing.T){
	seed := []byte{1,2,3}
	masterKey, _ := NewMasterKey(seed)

	data := []struct{
		keyType byte
	}{
		{byte(3)},
		{byte(10)},
		{byte(123)},
		{byte(234)},
		{byte(255)},
	}

	for _, item := range data{
		serializedKey, err := masterKey.Serialize(item.keyType)

		assert.Equal(t, []byte{}, serializedKey)
		assert.Equal(t, NewWalletError(InvalidKeyTypeErr, nil), err)
	}
}

/*
		Unit test for Base58CheckSerialize function
 */

func TestHDWalletBase58CheckSerialize( t *testing.T){
	seed := []byte{1,2,3}
	masterKey, _ := NewMasterKey(seed)

	privKeyBytes := masterKey.Base58CheckSerialize(PriKeyType)
	paymentAddrBytes := masterKey.Base58CheckSerialize(PaymentAddressType)
	readonlyKeyBytes := masterKey.Base58CheckSerialize(ReadonlyKeyType)

	assert.Equal(t, PrivKeyBase58CheckSerializedBytesLen, len(privKeyBytes))
	assert.Equal(t, PaymentAddrBase58CheckSerializedBytesLen, len(paymentAddrBytes))
	assert.Equal(t, ReadOnlyKeyBase58CheckSerializedBytesLen, len(readonlyKeyBytes))
}

func TestHDWalletBase58CheckSerializeWithInvalidKeyType( t *testing.T){
	seed := []byte{1,2,3}
	masterKey, _ := NewMasterKey(seed)

	data := []struct{
		keyType byte
	}{
		{byte(3)},
		{byte(10)},
		{byte(123)},
		{byte(234)},
		{byte(255)},
	}

	for _, item := range data{
		serializedKey := masterKey.Base58CheckSerialize(item.keyType)

		assert.Equal(t, "", serializedKey)
	}
}
