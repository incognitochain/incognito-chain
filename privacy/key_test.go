package privacy

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSpendingKey(t *testing.T) {
	spendingKey := GenerateSpendingKey(new(big.Int).SetInt64(123).Bytes())
	fmt.Printf("\nSpending key: %v\n", spendingKey)
	fmt.Println(len(spendingKey))
	expectedResult := []byte{2, 31, 181, 150, 219, 129, 230, 208, 43, 243, 210, 88, 110, 227, 152, 31, 229, 25, 242, 117, 192, 172, 156, 167, 107, 188, 242, 235, 180, 9, 125, 150}

	assert.Equal(t, expectedResult, spendingKey)
}

func TestPAdd1Div4(t *testing.T) {
	res := new(big.Int)
	res = PAdd1Div4(new(big.Int).SetInt64(123))
	expectedResult := new(big.Int).SetInt64(31)
	assert.Equal(t, expectedResult, res)
}

func TestGenerateKey(t *testing.T){
	spendingKey := GenerateSpendingKey(new(big.Int).SetInt64(123).Bytes())
	fmt.Printf("\nSpending key: %v\n", spendingKey)
	lenSK := len(spendingKey)

	//publicKey is compressed
	publicKey := GeneratePublicKey(spendingKey)
	fmt.Printf("\nPublic key: %v\n", publicKey)
	lenPK := len(publicKey)

	point, err := DecompressKey(publicKey)
	if err != nil {
	fmt.Println(err)
	}
	fmt.Printf("Public key decompress: %v\n", point)

	receivingKey := GenerateReceivingKey(spendingKey)
	fmt.Printf("\nReceiving key: %v\n", receivingKey)
	lenRK := len(receivingKey)
	fmt.Println(len(receivingKey))

	transmissionKey := GenerateTransmissionKey(receivingKey)
	fmt.Printf("\nTransmission key: %v\n", transmissionKey)
	lenTK := len(transmissionKey)

	point, err = DecompressKey(transmissionKey)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Transmission key point decompress: %+v\n ", point)

	paymentAddress := GeneratePaymentAddress(spendingKey)
	lenPaymentAddr := len(paymentAddress.ToBytes())

	assert.Equal(t, lenSK, 32)
	assert.Equal(t, lenPK, CompressedPointSize)
	assert.Equal(t, lenRK, 32)
	assert.Equal(t, lenTK, CompressedPointSize)
	assert.Equal(t, lenPaymentAddr, 66)
}


