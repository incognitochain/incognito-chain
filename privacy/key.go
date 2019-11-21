package privacy

import (
	"encoding/hex"
)

// 32-byte spending key
type PrivateKey []byte

// 32-byte public key
type PublicKey []byte

// 32-byte receiving key
type ReceivingKey []byte

// 32-byte transmission key
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
	Message        []byte // 512 bytes
}

// GeneratePrivateKey generates a random 32-byte spending key
func GeneratePrivateKey(seed []byte) PrivateKey {
	bip32PrivKey := HashToScalar(seed)
	privateKey := bip32PrivKey.ToBytesS()
	return privateKey
}

// GeneratePublicKey computes a 32-byte public-key corresponding to a spending key
func GeneratePublicKey(privateKey []byte) PublicKey {
	privScalar := new(Scalar).FromBytesS(privateKey)
	publicKey := new(Point).ScalarMultBase(privScalar)
	return publicKey.ToBytesS()
}

// GenerateReceivingKey generates a 32-byte receiving key
func GenerateReceivingKey(privateKey []byte) ReceivingKey {
	receivingKey := HashToScalar(privateKey[:])
	return receivingKey.ToBytesS()
}

// GenerateTransmissionKey computes a 33-byte transmission key corresponding to a receiving key
func GenerateTransmissionKey(receivingKey []byte) TransmissionKey {
	receiScalar := new(Scalar).FromBytesS(receivingKey)
	transmissionKey := new(Point).ScalarMultBase(receiScalar)
	return transmissionKey.ToBytesS()
}

// GenerateViewingKey generates a viewingKey corresponding to a spending key
func GenerateViewingKey(privateKey []byte) ViewingKey {
	var viewingKey ViewingKey
	viewingKey.Pk = GeneratePublicKey(privateKey)
	viewingKey.Rk = GenerateReceivingKey(privateKey)
	return viewingKey
}

// GeneratePaymentAddress generates a payment address corresponding to a spending key
func GeneratePaymentAddress(privateKey []byte) PaymentAddress {
	var paymentAddress PaymentAddress
	paymentAddress.Pk = GeneratePublicKey(privateKey)
	paymentAddress.Tk = GenerateTransmissionKey(GenerateReceivingKey(privateKey))
	return paymentAddress
}

// Bytes converts payment address to bytes array
func (addr *PaymentAddress) Bytes() []byte {
	return append(addr.Pk[:], addr.Tk[:]...)
}

// SetBytes reverts bytes array to payment address
func (addr *PaymentAddress) SetBytes(bytes []byte) *PaymentAddress {
	// the first 33 bytes are public key
	addr.Pk = bytes[:Ed25519KeySize]
	// the last 33 bytes are transmission key
	addr.Tk = bytes[Ed25519KeySize:]
	return addr
}

// String encodes a payment address as a hex string
func (addr PaymentAddress) String() string {
	byteArrays := addr.Bytes()
	return hex.EncodeToString(byteArrays[:])
}

func SliceToArray(slice []byte) [Ed25519KeySize]byte {
	var array [Ed25519KeySize]byte
	copy(array[:], slice)
	return array
}

func ArrayToSlice(array [Ed25519KeySize]byte) []byte {
	var slice []byte
	slice = array[:]
	return slice
}
