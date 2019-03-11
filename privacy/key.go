package privacy

import (
	"encoding/hex"
	"math/big"

	"github.com/big0t/constant-chain/common"
)

// 32-byte spending key
type SpendingKey []byte

// 33-byte public key
type PublicKey []byte

// 32-byte receiving key
type ReceivingKey []byte

// 33-byte transmission key
type TransmissionKey []byte

// ViewingKey is a public/private key pair to encrypt coins in an outgoing transaction
// and decrypt coins in an incoming transaction
type ViewingKey struct {
	Pk PublicKey    // 33 bytes, use to receive coin
	Rk ReceivingKey // 32 bytes, use to decrypt pointByte
}

// PaymentAddress is an address of a payee
type PaymentAddress struct {
	Pk PublicKey       // 33 bytes, use to receive coin
	Tk TransmissionKey // 33 bytes, use to encrypt pointByte
}

// PaymentInfo contains an address of a payee and a value of coins he/she will receive
type PaymentInfo struct {
	PaymentAddress PaymentAddress
	Amount         uint64
}

// GenerateSpendingKey generates a random 32-byte spending key
func GenerateSpendingKey(seed []byte) SpendingKey {
	spendingKey := common.HashB(seed)

	temp := new(big.Int)
	for temp.SetBytes(spendingKey).Cmp(Curve.Params().N) == 1 {
		spendingKey = common.HashB(spendingKey)
	}
	return spendingKey
}

// GeneratePublicKey computes a 33-byte public-key corresponding to a spending key
func GeneratePublicKey(spendingKey []byte) PublicKey {
	var publicKey EllipticPoint
	publicKey.X, publicKey.Y = Curve.ScalarBaseMult(spendingKey)
	return publicKey.Compress()
}

// GenerateReceivingKey generates a 32-byte receiving key
func GenerateReceivingKey(spendingKey []byte) ReceivingKey {
	receivingKey := common.HashB(spendingKey)

	temp := new(big.Int)
	for temp.SetBytes(receivingKey).Cmp(Curve.Params().N) == 1 {
		receivingKey = common.HashB(receivingKey)
	}
	return receivingKey
}

// GenerateTransmissionKey computes a 33-byte transmission key corresponding to a receiving key
func GenerateTransmissionKey(receivingKey []byte) TransmissionKey {
	var transmissionKey EllipticPoint
	transmissionKey.X, transmissionKey.Y = Curve.ScalarBaseMult(receivingKey)
	return transmissionKey.Compress()
}

// GenerateViewingKey generates a viewingKey corresponding to a spending key
func GenerateViewingKey(spendingKey []byte) ViewingKey {
	var viewingKey ViewingKey
	viewingKey.Pk = GeneratePublicKey(spendingKey)
	viewingKey.Rk = GenerateReceivingKey(spendingKey)
	return viewingKey
}

// GeneratePaymentAddress generates a payment address corresponding to a spending key
func GeneratePaymentAddress(spendingKey []byte) PaymentAddress {
	var paymentAddress PaymentAddress
	paymentAddress.Pk = GeneratePublicKey(spendingKey)
	paymentAddress.Tk = GenerateTransmissionKey(GenerateReceivingKey(spendingKey))
	return paymentAddress
}

// Bytes converts payment address to bytes array
func (addr *PaymentAddress) Bytes() []byte {
	return append(addr.Pk, addr.Tk...)
}

// SetBytes reverts bytes array to payment address
func (addr *PaymentAddress) SetBytes(bytes []byte) *PaymentAddress {
	// the first 33 bytes are public key
	addr.Pk = bytes[:CompressedPointSize]
	// the last 33 bytes are transmission key
	addr.Tk = bytes[CompressedPointSize:]
	return addr
}

// NewPaymentAddressFromByte reverts a byte array to a payment address
// TODO: remove this function 0xjacklope
func NewPaymentAddressFromByte(b []byte) *PaymentAddress {
	tmp := make([]byte, len(b))
	copy(tmp, b)
	paymentAddress := PaymentAddress{}
	paymentAddress.SetBytes(tmp)
	return &paymentAddress
}

// Size returns the size of a payment address
func (addr *PaymentAddress) Size() int {
	return len(addr.Pk) + len(addr.Tk)
}

// String encodes a payment address as a hex string
func (addr PaymentAddress) String() string {
	byteArrays := addr.Bytes()
	return hex.EncodeToString(byteArrays[:])
}
