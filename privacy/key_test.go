package privacy

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKey(t *testing.T) {
	for i:=0; i< 1000; i ++ {
		// random seed
		seed := ArrayToSlice(RandomScalar().ToBytes())

		// generate private key from seed
		privateKey := GeneratePrivateKey(seed)

		// generate public key from private key
		publicKey := GeneratePublicKey(privateKey)

		// check public key
		publicKeyPrime := new(Point).ScalarMultBase(new(Scalar).FromBytes(SliceToArray(privateKey)))
		assert.Equal(t, publicKeyPrime.ToBytes(), SliceToArray(publicKey))

		// generate receiving key from private key
		receivingKey := GenerateReceivingKey(privateKey)

		// generate transmission key from receiving key
		transmissionKey := GenerateTransmissionKey(receivingKey)

		// decompress transmission key to transmissionKeyPoint
		transmissionKeyPrime := new(Point).ScalarMultBase(new(Scalar).FromBytes(SliceToArray(receivingKey)))
		assert.Equal(t, transmissionKeyPrime.ToBytes(), SliceToArray(transmissionKey))

		// generate payment address from private key
		paymentAddress := GeneratePaymentAddress(privateKey)
		assert.Equal(t, publicKey, paymentAddress.Pk)
		assert.Equal(t, transmissionKey, paymentAddress.Tk)

		// convert payment address to bytes array
		paymentAddrBytes := paymentAddress.Bytes()

		// new payment address to set bytes
		paymentAddress2 := new(PaymentAddress)
		paymentAddress2.SetBytes(paymentAddrBytes)

		assert.Equal(t, paymentAddress.Pk, paymentAddress2.Pk)
		assert.Equal(t, paymentAddress.Tk, paymentAddress2.Tk)
	}
}
