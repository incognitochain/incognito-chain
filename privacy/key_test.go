package privacy

import (
	"fmt"
	"github.com/constant-money/constant-chain/common/base58"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestGenerateSpendingKey(t *testing.T) {
	spendingKey := GenerateSpendingKey(new(big.Int).SetInt64(123).Bytes())
	expectedResult := []byte{2, 31, 181, 150, 219, 129, 230, 208, 43, 243, 210, 88, 110, 227, 152, 31, 229, 25, 242, 117, 192, 172, 156, 167, 107, 188, 242, 235, 180, 9, 125, 150}

	spendingKeyArray := make([]uint8, SpendingKeySize)
	copy(spendingKeyArray, spendingKey)
	assert.Equal(t, expectedResult, spendingKeyArray)
}

func TestPAdd1Div4(t *testing.T) {
	res := PAdd1Div4(new(big.Int).SetInt64(123))
	expectedResult := new(big.Int).SetInt64(31)
	assert.Equal(t, expectedResult, res)
}

func TestGenerateKey(t *testing.T) {
	spendingKey := GenerateSpendingKey(new(big.Int).SetInt64(123).Bytes())

	//publicKey is compressed
	publicKey := GeneratePublicKey(spendingKey)
	publicKeyBytes := make([]byte, CompressedPointSize)
	copy(publicKeyBytes, publicKey[:])

	// decompress public key
	publicKeyPoint := new(EllipticPoint)
	err := publicKeyPoint.Decompress(publicKey)
	if err != nil {
		return
	}

	assert.Equal(t, publicKeyBytes, publicKeyPoint.Compress())

	receivingKey := GenerateReceivingKey(spendingKey)

	transmissionKey := GenerateTransmissionKey(receivingKey)
	transmissionKeyBytes := make([]byte, CompressedPointSize)
	copy(transmissionKeyBytes, transmissionKey[:])

	transmissionKeyPoint := new(EllipticPoint)
	err = transmissionKeyPoint.Decompress(transmissionKey)
	if err != nil {
		return
	}
	assert.Equal(t, transmissionKeyBytes, transmissionKeyPoint.Compress())

	paymentAddress := GeneratePaymentAddress(spendingKey)
	paymentAddrBytes := paymentAddress.Bytes()

	paymentAddress2 := new(PaymentAddress)
	paymentAddress2.SetBytes(paymentAddrBytes)

	assert.Equal(t, paymentAddress.Pk, paymentAddress2.Pk)
	assert.Equal(t, paymentAddress.Tk, paymentAddress2.Tk)

	sk := GenerateSpendingKey([]byte{123})
	fmt.Printf("Spending key byte : %v\n", sk)
	skStr := base58.Base58Check.Encode(base58.Base58Check{}, sk, 0x01)
	fmt.Printf("Spending key string after encode : %v\n", skStr)
}
