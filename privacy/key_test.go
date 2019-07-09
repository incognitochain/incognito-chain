package privacy

import (
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestKey(t *testing.T) {
	// random seed
	seed := RandBytes(10)

	// generate private key from seed
	privateKey := GeneratePrivateKey(seed)
	assert.Equal(t, PrivateKeySize, len(privateKey))

	// generate public key from private key
	publicKey := GeneratePublicKey(privateKey)
	assert.Equal(t, PublicKeySize, len(publicKey))

	// decompress public key to publicKeyPoint
	publicKeyPoint := new(EllipticPoint)
	err := publicKeyPoint.Decompress(publicKey)
	assert.Equal(t, nil, err)

	publicKeyExpected := PedCom.G[0].ScalarMult(new(big.Int).SetBytes(privateKey))
	assert.Equal(t, publicKeyExpected, publicKeyPoint)

	// generate receiving key from private key
	receivingKey := GenerateReceivingKey(privateKey)
	assert.Equal(t, ReceivingKeySize, len(receivingKey))

	// generate transmission key from receiving key
	transmissionKey := GenerateTransmissionKey(receivingKey)
	assert.Equal(t, TransmissionKeySize, len(transmissionKey))

	// decompress transmission key to transmissionKeyPoint
	transmissionKeyPoint := new(EllipticPoint)
	err = transmissionKeyPoint.Decompress(transmissionKey)
	assert.Equal(t, nil, err)

	transmissionKeyExpected := PedCom.G[0].ScalarMult(new(big.Int).SetBytes(receivingKey))
	assert.Equal(t, transmissionKeyExpected, transmissionKeyPoint)

	// generate payment address from private key
	paymentAddress := GeneratePaymentAddress(privateKey)
	assert.Equal(t, publicKey, paymentAddress.Pk)
	assert.Equal(t, transmissionKey, paymentAddress.Tk)

	// convert payment address to bytes array
	paymentAddrBytes := paymentAddress.Bytes()
	assert.Equal(t, PaymentAddressSize, len(paymentAddrBytes))

	// new payment address to set bytes
	paymentAddress2 := new(PaymentAddress)
	paymentAddress2.SetBytes(paymentAddrBytes)

	assert.Equal(t, paymentAddress.Pk, paymentAddress2.Pk)
	assert.Equal(t, paymentAddress.Tk, paymentAddress2.Tk)

	// convert payment address to hex encode string
	paymentAddrStr := paymentAddress2.String()
	assert.Equal(t, hex.EncodedLen(PaymentAddressSize), len(paymentAddrStr))

}
