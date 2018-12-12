package privacy

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ninjadotorg/constant/common"
)

// const (
// const (
// 	P = 0xFFFFFFFF00000001000000000000000000000000FFFFFFFFFFFFFFFFFFFFFFFF
// 	N = 0xFFFFFFFF00000000FFFFFFFFFFFFFFFFBCE6FAADA7179E84F3B9CAC2FC632551
// 	B = 0x5AC635D8AA3A93E7B3EBBD55769886BC651D06B0CC53B0F63BCE3C3E27D2604B
// )

// These constants define the lengths of serialized public keys.
//const (
//	PubKeyBytesLenCompressed      = 33
//	pubkeyCompressed         byte = 0x2 // y_bit + x coord
//)

// fmt.Printf("N: %v\n", curve.N)
// fmt.Printf("P: %v\n", curve.P)
// fmt.Printf("B: %v\n", curve.B)
// fmt.Printf("Gx: %v\n", curve.Gx)
// fmt.Printf("Gy: %v\n", curve.Gy)
// fmt.Printf("BitSize: %v\n", curve.BitSize)

// SpendingKey 32 bytes
type SpendingKey []byte

// Pk 33 bytes
type PublicKey []byte

// Rk 32 bytes
type ReceivingKey []byte

// Tk 33 bytes
type TransmissionKey []byte

// ViewingKey represents an key that be used to view transactions
type ViewingKey struct {
	Pk PublicKey    // 33 bytes, use to receive coin
	Rk ReceivingKey // 32 bytes, use to decrypt pointByte
}

// PaymentAddress represents an payment address of receiver
type PaymentAddress struct {
	Pk PublicKey       // 33 bytes, use to receive coin
	Tk TransmissionKey // 33 bytes, use to encrypt pointByte
}

type PaymentInfo struct {
	PaymentAddress PaymentAddress
	Amount         uint64
}

// GenerateSpendingKey generates a random SpendingKey
// SpendingKey: 32 bytes
func GenerateSpendingKey(seed []byte) SpendingKey {
	temp := new(big.Int)
	spendingKey := make([]byte, 32)
	spendingKey = common.HashB(seed)
	for temp.SetBytes(spendingKey).Cmp(Curve.Params().N) == 1 {
		spendingKey = common.HashB(spendingKey)
	}

	return spendingKey
}

// GeneratePublicKey computes an address corresponding with spendingKey
// Pk : 33 bytes
func GeneratePublicKey(spendingKey []byte) PublicKey {
	var p EllipticPoint
	p.X, p.Y = Curve.ScalarBaseMult(spendingKey)
	//fmt.Printf("Public key is not compressed: %+v\n\n", p)
	publicKey := p.Compress()

	return publicKey
}

// GenerateReceivingKey computes a receiving key corresponding with spendingKey
// Rk : 32 bytes
func GenerateReceivingKey(spendingKey []byte) ReceivingKey {
	hash := sha256.Sum256(spendingKey)
	receivingKey := make([]byte, 32)
	copy(receivingKey, hash[:])
	return receivingKey
}

// GenerateTransmissionKey computes a transmission key corresponding with receivingKey
// Tk : 33 bytes
func GenerateTransmissionKey(receivingKey []byte) TransmissionKey {
	var p EllipticPoint

	p.X, p.Y = Curve.ScalarBaseMult(receivingKey)
	//fmt.Printf("Transmission key is not compressed: %+v\n\n", p)
	transmissionKey := p.Compress()
	return transmissionKey
}

// GenerateViewingKey generates a viewingKey corresponding with spendingKey
func GenerateViewingKey(spendingKey []byte) ViewingKey {
	var viewingKey ViewingKey
	viewingKey.Pk = GeneratePublicKey(spendingKey)
	viewingKey.Rk = GenerateReceivingKey(spendingKey)
	return viewingKey
}

// GeneratePaymentAddress generates a payment address corresponding with spendingKey
func GeneratePaymentAddress(spendingKey []byte) PaymentAddress {
	var paymentAddress PaymentAddress
	paymentAddress.Pk = GeneratePublicKey(spendingKey)
	paymentAddress.Tk = GenerateTransmissionKey(GenerateReceivingKey(spendingKey))
	return paymentAddress
}

//// FromPointToByteArray converts an elliptic point to byte array
//func FromPointToByteArray(p EllipticPoint) []byte {
//	var pointByte []byte
//	x := p.X.Bytes()
//	y := p.Y.Bytes()
//	pointByte = append(pointByte, x...)
//	pointByte = append(pointByte, y...)
//	return pointByte
//}
//
//// FromByteArrayToPoint converts a byte array to elliptic point
//func FromByteArrayToPoint(pointByte []byte) EllipticPoint {
//	point := new(EllipticPoint)
//	point.X = new(big.Int).SetBytes(pointByte[0:32])
//	point.Y = new(big.Int).SetBytes(pointByte[32:64])
//	return *point
//}

func isOdd(a *big.Int) bool {
	return a.Bit(0) == 1
}

// DecompressKey decompress public key to elliptic point
func DecompressKey(pubKeyStr []byte) (pubkey *EllipticPoint, err error) {
	if len(pubKeyStr) == 0 || len(pubKeyStr) != 33 {
		return nil, fmt.Errorf("pubkey string len is wrong")
	}

	pubkey = new(EllipticPoint)

	err = pubkey.Decompress(pubKeyStr)
	if err != nil {
		return nil, err
	}

	if pubkey.X.Cmp(Curve.Params().P) >= 0 {
		return nil, fmt.Errorf("pubkey X parameter is >= to P")
	}
	if pubkey.Y.Cmp(Curve.Params().P) >= 0 {
		return nil, fmt.Errorf("pubkey Y parameter is >= to P")
	}
	if !Curve.Params().IsOnCurve(pubkey.X, pubkey.Y) {
		return nil, fmt.Errorf("pubkey isn't on P256 curve")
	}
	return pubkey, nil
}

// PAdd1Div4 computes (p + 1) / 4
func PAdd1Div4(p *big.Int) (res *big.Int) {
	res = new(big.Int)
	res.Add(p, new(big.Int).SetInt64(1))
	res.Div(res, new(big.Int).SetInt64(4))
	return
}

// paddedAppend appends the src byte slice to dst, returning the new slice.
// If the length of the source is smaller than the passed size, leading zero
// bytes are appended to the dst slice before appending src.
func paddedAppend(size uint, dst, src []byte) []byte {
	for i := 0; i < int(size)-len(src); i++ {
		dst = append(dst, 0)
	}
	return append(dst, src...)
}

func (addr *PaymentAddress) ToBytes() []byte {
	result := make([]byte, 33)
	pkenc := make([]byte, 33)
	copy(result, addr.Pk[:33])
	copy(pkenc, addr.Tk[:33])
	result = append(result, pkenc...)
	return result
}

func (addr *PaymentAddress) Size() int {
	return len(addr.Pk) + len(addr.Tk)
}

func (addr *PaymentAddress) FromBytes(data []byte) *PaymentAddress {
	addr.Pk = make([]byte, 33)
	addr.Tk = make([]byte, 33)
	copy(addr.Pk[:], data[:33]) // First 33 bytes are PaymentAddress's
	copy(addr.Tk[:], data[33:]) // Last 33 bytes are Pkenc's
	return addr
}

func (spendingKey SpendingKey) String() string {
	for i := 0; i < 32/2; i++ {
		spendingKey[i], spendingKey[32-1-i] = spendingKey[32-1-i], spendingKey[i]
	}
	return hex.EncodeToString(spendingKey[:])
}
