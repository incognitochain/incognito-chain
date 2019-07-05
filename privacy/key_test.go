package privacy

import (
	"encoding/hex"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestGeneratePrivateKey(t *testing.T) {
	privateKey := GeneratePrivateKey(new(big.Int).SetInt64(123).Bytes())
	expectedResult := []byte{2, 31, 181, 150, 219, 129, 230, 208, 43, 243, 210, 88, 110, 227, 152, 31, 229, 25, 242, 117, 192, 172, 156, 167, 107, 188, 242, 235, 180, 9, 125, 150}

	privateKeyArray := make([]uint8, PrivateKeySize)
	copy(privateKeyArray, privateKey)
	assert.Equal(t, expectedResult, privateKeyArray)
}

func TestPAdd1Div4(t *testing.T) {
	res := PAdd1Div4(new(big.Int).SetInt64(123))
	expectedResult := new(big.Int).SetInt64(31)
	assert.Equal(t, expectedResult, res)
}

func TestGenerateKey(t *testing.T) {
	privateKey := GeneratePrivateKey(new(big.Int).SetInt64(123).Bytes())

	//publicKey is compressed
	publicKey := GeneratePublicKey(privateKey)
	publicKeyBytes := make([]byte, CompressedPointSize)
	copy(publicKeyBytes, publicKey[:])

	// decompress public key
	publicKeyPoint := new(EllipticPoint)
	err := publicKeyPoint.Decompress(publicKey)
	if err != nil {
		return
	}

	assert.Equal(t, publicKeyBytes, publicKeyPoint.Compress())

	receivingKey := GenerateReceivingKey(privateKey)

	transmissionKey := GenerateTransmissionKey(receivingKey)
	transmissionKeyBytes := make([]byte, CompressedPointSize)
	copy(transmissionKeyBytes, transmissionKey[:])

	transmissionKeyPoint := new(EllipticPoint)
	err = transmissionKeyPoint.Decompress(transmissionKey)
	if err != nil {
		return
	}
	assert.Equal(t, transmissionKeyBytes, transmissionKeyPoint.Compress())

	paymentAddress := GeneratePaymentAddress(privateKey)
	paymentAddrBytes := paymentAddress.Bytes()

	paymentAddress2 := new(PaymentAddress)
	paymentAddress2.SetBytes(paymentAddrBytes)

	assert.Equal(t, paymentAddress.Pk, paymentAddress2.Pk)
	assert.Equal(t, paymentAddress.Tk, paymentAddress2.Tk)

	sk := GeneratePrivateKey([]byte{123})
	base58.Base58Check.Encode(base58.Base58Check{}, sk, 0x01)
}

func TestDecodePubKey(t *testing.T) {
	// shard 0
	hex.DecodeString("023db7a5efdc3c948d9882458e74568edf42ac0f7eaa1527beb457075d57028bfe")
}
