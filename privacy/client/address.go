package client

import (
	"crypto/rand"

	"golang.org/x/crypto/curve25519"
)

const (
	SpendingKeyLength     = 32 // bytes
	ReceivingKeyLength    = 32
	SpendingAddressLength = 32
	TransmissionKeyLength = 32
	EphemeralKeyLength    = 32
)

type SpendingKey [SpendingKeyLength]byte
type SpendingAddress [SpendingAddressLength]byte
type TransmissionKey [TransmissionKeyLength]byte
type ReceivingKey [ReceivingKeyLength]byte
type EphemeralPubKey [EphemeralKeyLength]byte
type EphemeralPrivKey [EphemeralKeyLength]byte

type ViewingKey struct {
	Apk   SpendingAddress // use to receive coin
	Skenc ReceivingKey    // use to decrypt data
}

type PaymentAddress struct {
	Apk   SpendingAddress // use to receive coin
	Pkenc TransmissionKey // use to encrypt data
}

type PaymentInfo struct {
	PaymentAddress PaymentAddress
	Amount         uint64
}

// RandBits generates random bits and return as bytes; zero out redundant bits
func RandBits(n int) []byte {
	m := 1 + (n-1)/8
	b := make([]byte, m)
	rand.Read(b)

	if n%8 > 0 {
		b[m-1] &= ((1 << uint(n%8)) - 1)
	}
	return b
}

func clampCurve25519(x []byte) []byte {
	x[0] &= 0xF8  // Clear bit 0, 1, 2 of first byte
	x[31] &= 0x7F // Clear bit 7 of last byte
	x[31] |= 0x40 // Set bit 6 of last byte
	return x
}

// RandSpendingKey generates a random SpendingKey
func RandSpendingKey() SpendingKey {
	b := RandBits(SpendingKeyLength*8 - 4)

	ask := *new(SpendingKey)
	copy(ask[:], b)
	return ask
}

func GenSpendingAddress(ask SpendingKey) SpendingAddress {
	data := PRF_addr_x(ask[:], 0)
	var apk SpendingAddress
	copy(apk[:], data)
	return apk
}

func GenViewingKey(ask SpendingKey) ViewingKey {
	var ivk ViewingKey
	ivk.Apk = GenSpendingAddress(ask)
	ivk.Skenc = GenReceivingKey(ask)
	return ivk
}

// FromBytes converts a byte stream to PaymentAddress
func (addr *PaymentAddress) FromBytes(data []byte) *PaymentAddress {
	copy(addr.Apk[:], data[:32])   // First 32 bytes are Apk's
	copy(addr.Pkenc[:], data[32:]) // Last 32 bytes are Pkenc's
	return addr
}

// ToBytes converts a PaymentAddress to a byte slice
func (addr *PaymentAddress) ToBytes() []byte {
	result := make([]byte, 32)
	pkenc := make([]byte, 32)
	copy(result, addr.Apk[:32])
	copy(pkenc, addr.Pkenc[:32])
	result = append(result, pkenc...)
	return result
}

func GenReceivingKey(ask SpendingKey) ReceivingKey {
	data := PRF_addr_x(ask[:], 1)
	clamped := clampCurve25519(data)
	var skenc ReceivingKey
	copy(skenc[:], clamped)
	return skenc
}

func GenTransmissionKey(skenc ReceivingKey) TransmissionKey {
	// TODO: reduce copy
	var x, y [32]byte
	copy(y[:], skenc[:])
	curve25519.ScalarBaseMult(&x, &y)

	var pkenc TransmissionKey
	copy(pkenc[:], x[:])
	return pkenc
}

func GenPaymentAddress(ask SpendingKey) PaymentAddress {
	var addr PaymentAddress
	addr.Apk = GenSpendingAddress(ask)
	addr.Pkenc = GenTransmissionKey(GenReceivingKey(ask))
	return addr
}

func (esk EphemeralPrivKey) GenPubKey() EphemeralPubKey {
	var x, y [32]byte
	var epk EphemeralPubKey

	copy(y[:], esk[:])
	curve25519.ScalarBaseMult(&x, &y)
	copy(epk[:], x[:])
	return epk
}

func GenEphemeralKey() (EphemeralPubKey, EphemeralPrivKey) {
	var esk EphemeralPrivKey

	esk_tmp := RandBits(256)
	copy(esk[:], esk_tmp[:])

	return esk.GenPubKey(), esk
}
