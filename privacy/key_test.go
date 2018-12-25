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

	//publicKey is compressed
	publicKey := GeneratePublicKey(spendingKey)
	publicKeyBytes := make([]byte, CompressedPointSize)
	copy(publicKeyBytes, publicKey[:])

	// decompress public key
	publicKeyPoint, err := DecompressKey(publicKey)
	if err != nil {
		fmt.Println(err)
	}

	assert.Equal(t, publicKeyBytes, publicKeyPoint.Compress())

	receivingKey := GenerateReceivingKey(spendingKey)

	transmissionKey := GenerateTransmissionKey(receivingKey)
	transmissionKeyBytes := make([]byte, CompressedPointSize)
	copy(transmissionKeyBytes, transmissionKey[:])

	transmissionKeyPoint, err := DecompressKey(transmissionKey)
	if err != nil {
		fmt.Println(err)
	}
	assert.Equal(t, transmissionKeyBytes, transmissionKeyPoint.Compress())

	paymentAddress := GeneratePaymentAddress(spendingKey)
	paymentAddrBytes := paymentAddress.Bytes()

	paymentAddress2 := new(PaymentAddress)
	paymentAddress2.SetBytes(paymentAddrBytes)

	assert.Equal(t, paymentAddress.Pk, paymentAddress2.Pk)
	assert.Equal(t, paymentAddress.Tk, paymentAddress2.Tk)

}


