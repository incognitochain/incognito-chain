package client

import (
	"crypto/rand"
	// "golang.org/x/crypto/curve25519"
	"math/big"
	"golang.org/x/crypto/openpgp/elgamal"
)

const (
	SpendingKeyLength     = 32 // bytes
	ReceivingKeyLength    = 32
	SpendingAddressLength = 32

	// This is the 1024-bit MODP group from RFC 5114, section 2.1:
	// Parameter for Elgamal encryption
	primeHex = "B10B8F96A080E01DDE92DE5EAE5D54EC52C99FBCFB06A3C69A6A9DCA52D23B616073E28675A23D189838EF1E2EE652C013ECB4AEA906112324975C3CD49B83BFACCBDD7D90C4BD7098488E9C219A73724EFFD6FAE5644738FAA31A4FF55BCCC0A151AF5F0DC8B4BD45BF37DF365C1A65E68CFDA76D4DA708DF1FB2BC2E4A4371"
	generatorHex = "A4D1CBD5C3FD34126765A442EFB99905F8104DD258AC507FD6406CFF14266D31266FEA1E5C41564B777E690F5504F213160217B4B01B886A5E91547F9E2749F4D7FBD7D3B9A92EE1909D0D2263F80A76A6A24C087A091F531DBF0A0169B6A28AD662A4D18E73AFA32D779D5918D08BC8858F4DCEF97C2A24855E6EEB22B3B2E5"
	subgroupSizeHex = "F518AA8781A8DF278ABA4E7D64B7CB9D49462353"
)

type SpendingKey [SpendingKeyLength]byte

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

// RandSpendingKey generates a random SpendingKey
func RandSpendingKey() SpendingKey {
	b := RandBits(SpendingKeyLength*8 - 4)
	// b := make([]byte, SpendingKeyLength)
	// rand.Read(b)
	// b[SpendingKeyLength-1] &= 0x0F // First 4 bits are 0

	ask := *new(SpendingKey)
	copy(ask[:], b)
	return ask
}

type ReceivingKey elgamal.PrivateKey
type TransmissionKey elgamal.PublicKey

func GenReceivingKey() ReceivingKey {
	var skenc ReceivingKey
	
	skenc.X = RandBigInt(FromHexToBigInt(subgroupSizeHex))
	skenc.PublicKey.G = generator
	skenc.PublicKey.P = prime
	skenc.PublicKey.Y = new(big.Int).Exp(generator, skenc.X, prime)
	
	return skenc
}

func GenTransmissionKey(skenc ReceivingKey) TransmissionKey {
	return (TransmissionKey)(skenc.PublicKey)
}

/*
func GenReceivingKey(ask SpendingKey) ReceivingKey {
	data := PRF_addr_x(ask[:], 1)
	clamped := clampCurve25519(data)
	var skenc ReceivingKey
	copy(skenc[:], clamped)
	return skenc
}
*/

func clampCurve25519(x []byte) []byte {
	x[0] &= 0xF8  // Clear bit 0, 1, 2 of first byte
	x[31] &= 0x7F // Clear bit 7 of last byte
	x[31] |= 0x40 // Set bit 6 of last byte
	return x
}

type SpendingAddress [SpendingAddressLength]byte

func GenSpendingAddress(ask SpendingKey) SpendingAddress {
	data := PRF_addr_x(ask[:], 0)
	var apk SpendingAddress
	copy(apk[:], data)
	return apk
}

type ViewingKey struct {
	Apk   SpendingAddress
	Skenc ReceivingKey
}

func GenViewingKey(ask SpendingKey) ViewingKey {
	var ivk ViewingKey

	ivk.Apk = GenSpendingAddress(ask)
	ivk.Skenc = GenReceivingKey()

	return ivk
}

type PaymentAddress struct {
	Apk   SpendingAddress
	Pkenc TransmissionKey
}
//
/*
func GenTransmissionKey(skenc ReceivingKey) TransmissionKey {
	// TODO: reduce copy
	var x, y [32]byte
	copy(y[:], skenc[:])
	curve25519.ScalarBaseMult(&x, &y)

	var pkenc TransmissionKey
	copy(pkenc[:], x[:])
	return pkenc
}
*/


func GenPaymentAddress(ask SpendingKey) PaymentAddress {
	var addr PaymentAddress
	
	addr.apk = GenSpendingAddress(ask)
	addr.pkenc = GenTransmissionKey(GenReceivingKey())

	return addr
}

// FullKey convenient struct storing all keys and addresses
type FullKey struct {
	Ask  SpendingKey
	Ivk  ViewingKey
	Addr PaymentAddress
}

// GenFullKey generates all needed keys from a single SpendingKey
func (ask SpendingKey) GenFullKey() FullKey {
	return FullKey{Ask: ask, Ivk: GenViewingKey(ask), Addr: GenPaymentAddress(ask)}
}