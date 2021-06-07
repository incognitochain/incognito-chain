package incognitokey

import (
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/stretchr/testify/assert"
)

/*
	Unit test for GenerateKey function
*/

func TestKeySetGenerateKey(t *testing.T) {
	data := [][]byte{
		{},
		{1},
		{1, 2, 3},
		{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4}, // 32 bytes
	}

	keySet := new(KeySet)

	for _, item := range data {
		keySet = keySet.GenerateKey(item)
		assert.Equal(t, common.PrivateKeySize, len(keySet.PrivateKey))
		assert.Equal(t, common.PublicKeySize, len(keySet.PaymentAddress.Pk))
		assert.Equal(t, common.TransmissionKeySize, len(keySet.PaymentAddress.Tk))
		assert.Equal(t, common.ReceivingKeySize, len(keySet.ReadonlyKey.Rk))
	}
}

/*
	Unit test for InitFromPrivateKeyByte function
*/

func TestKeySetImportFromPrivateKeyByte(t *testing.T) {
	privateKey := common.RandBytes(common.PrivateKeySize)

	keySet := new(KeySet)
	err := keySet.InitFromPrivateKeyByte(privateKey)

	assert.Equal(t, nil, err)
	assert.Equal(t, common.PrivateKeySize, len(keySet.PrivateKey[:]))
	assert.Equal(t, common.PublicKeySize, len(keySet.PaymentAddress.Pk))
	assert.Equal(t, common.TransmissionKeySize, len(keySet.PaymentAddress.Tk))
	assert.Equal(t, common.ReceivingKeySize, len(keySet.ReadonlyKey.Rk))
}

func TestKeySetImportFromPrivateKeyByteWithInvalidPrivKey(t *testing.T) {
	keySet := new(KeySet)
	err := keySet.InitFromPrivateKeyByte(nil)
	assert.Equal(t, NewCashecError(InvalidPrivateKeyErr, nil), err)

	err2 := keySet.InitFromPrivateKeyByte([]byte{1, 2, 3})
	assert.Equal(t, NewCashecError(InvalidPrivateKeyErr, nil), err2)
}

/*
	Unit test for InitFromPrivateKey function
*/

func TestKeySetImportFromPrivateKey(t *testing.T) {
	privateKey := key.GeneratePrivateKey([]byte{1, 2, 3})

	keySet := new(KeySet)
	err := keySet.InitFromPrivateKey(&privateKey)

	assert.Equal(t, nil, err)
	assert.Equal(t, privateKey, keySet.PrivateKey[:])
	assert.Equal(t, common.PublicKeySize, len(keySet.PaymentAddress.Pk))
	assert.Equal(t, common.TransmissionKeySize, len(keySet.PaymentAddress.Tk))
	assert.Equal(t, common.ReceivingKeySize, len(keySet.ReadonlyKey.Rk))
}

func TestKeySetImportFromPrivateKeyWithInvalidPrivKey(t *testing.T) {
	keySet := new(KeySet)

	// private is not enough length
	privateKey := key.PrivateKey(common.RandBytes(10))
	err := keySet.InitFromPrivateKey(&privateKey)
	assert.Equal(t, NewCashecError(InvalidPrivateKeyErr, nil), err)

	// nil private key
	err = keySet.InitFromPrivateKey(nil)
	assert.Equal(t, NewCashecError(InvalidPrivateKeyErr, nil), err)
}

/*
	Unit test for Sign function
*/

func TestKeySetSign(t *testing.T) {
	data := [][]byte{
		{1},
		{1, 2, 3},
		{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4}, // 32 bytes
	}

	keySet := new(KeySet)
	keySet.GenerateKey([]byte{1, 2, 2})

	for _, item := range data {
		sig, err := keySet.Sign(item)
		assert.Equal(t, nil, err)
		assert.Equal(t, common.SigNoPrivacySize, len(sig))
	}
}

func TestKeySetSignWithEmptyData(t *testing.T) {
	keySet := new(KeySet)
	keySet.GenerateKey([]byte{1, 2, 2})

	sig, err := keySet.Sign([]byte{})
	assert.Equal(t, ErrCodeMessage[InvalidDataSignErr].Code, err.(*CashecError).Code)
	assert.Equal(t, 0, len(sig))
}

/*
	Unit test for Verify function
*/

func TestKeySetVerify(t *testing.T) {
	data := [][]byte{
		{1},
		{1, 2, 3},
		{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4}, // 32 bytes
	}

	keySet := new(KeySet)
	keySet.GenerateKey([]byte{1, 2, 2})

	for _, item := range data {
		sig, _ := keySet.Sign(item)
		isValid, err := keySet.Verify(item, sig)

		assert.Equal(t, nil, err)
		assert.Equal(t, true, isValid)
	}
}

func TestKeySetVerifyWithWrongSig(t *testing.T) {
	data := [][]byte{
		{1},
		{1, 2, 3},
		{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4}, // 32 bytes
	}

	keySet := new(KeySet)
	keySet.GenerateKey([]byte{1, 2, 2})

	for _, item := range data {
		sig, _ := keySet.Sign(item)

		// edit signature
		sig[len(sig)-1] = 0
		isValid, err := keySet.Verify(item, sig)

		assert.Equal(t, nil, err)
		assert.Equal(t, false, isValid)
	}
}

func TestKeySetVerifyWithUnmatchedPubKey(t *testing.T) {
	data := [][]byte{
		{1},
		{1, 2, 3},
		{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4}, // 32 bytes
	}

	keySet := new(KeySet)
	keySet.GenerateKey([]byte{1, 2, 2})
	keySet.PaymentAddress.Pk[0] = 0

	for _, item := range data {
		sig, _ := keySet.Sign(item)
		isValid, err := keySet.Verify(item, sig)

		assert.Equal(t, NewCashecError(InvalidVerificationKeyErr, nil), err)
		assert.Equal(t, false, isValid)
	}
}

/*
	Unit test for SignDataB58 function
*/
func TestKeySetSignDataB58(t *testing.T) {
	data := [][]byte{
		{1},
		{1, 2, 3},
		{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4}, // 32 bytes
	}

	// generate key set
	keySet := new(KeySet)
	keySet.GenerateKey([]byte{1})

	for _, item := range data {
		sig, err := keySet.SignDataInBase58CheckEncode(item)
		assert.Equal(t, nil, err)
		assert.Greater(t, len(sig), 0)
	}
}

func TestKeySetSignDataB58WithEmptyData(t *testing.T) {
	// generate key set
	keySet := new(KeySet)
	keySet.GenerateKey([]byte{1})

	_, err := keySet.SignDataInBase58CheckEncode([]byte{})
	assert.Equal(t, ErrCodeMessage[SignDataB58Err].Code, err.(*CashecError).GetCode())
}

/*
	Unit test for ValidateDataB58 function
*/
func TestKeySetValidateDataB58(t *testing.T) {
	// generate key set
	keySet := new(KeySet)
	keySet.GenerateKey([]byte{1})

	// sign data
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	sigB58, _ := keySet.SignDataInBase58CheckEncode(data)

	// get pubB58
	pubB58 := keySet.GetPublicKeyInBase58CheckEncode()

	// validate
	err := ValidateDataB58(pubB58, sigB58, data)
	assert.Equal(t, nil, err)
}

func TestKeySetValidateDataB58WithUnmatchedData(t *testing.T) {
	// generate key set
	keySet := new(KeySet)
	keySet.GenerateKey([]byte{1})

	// sign data
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	data2 := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}
	sigB58, _ := keySet.SignDataInBase58CheckEncode(data)

	// get pubB58
	pubB58 := keySet.GetPublicKeyInBase58CheckEncode()

	// validate
	err := ValidateDataB58(pubB58, sigB58, data2)
	assert.Equal(t, NewCashecError(InvalidDataValidateErr, nil), err)
}

func TestKeySetValidateDataB58WithWrongSig(t *testing.T) {
	// generate key set
	keySet := new(KeySet)
	keySet.GenerateKey([]byte{1})

	// sign data
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	sigB58, _ := keySet.SignDataInBase58CheckEncode(data)
	// edit the signature
	sigB58 += "abc"

	// get pubB58
	pubB58 := keySet.GetPublicKeyInBase58CheckEncode()

	// validate
	err := ValidateDataB58(pubB58, sigB58, data)
	assert.Equal(t, NewCashecError(B58DecodeSigErr, nil), err)
}

func TestKeySetValidateDataB58WithWrongPub(t *testing.T) {
	// generate key set
	keySet := new(KeySet)
	keySet.GenerateKey([]byte{1})

	// sign data
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	sigB58, _ := keySet.SignDataInBase58CheckEncode(data)

	// get pubB58
	pubB58 := keySet.GetPublicKeyInBase58CheckEncode()
	// edit the public key
	pubB58 += "abc"

	// validate
	err := ValidateDataB58(pubB58, sigB58, data)
	assert.Equal(t, NewCashecError(B58DecodePubKeyErr, nil), err)
}
