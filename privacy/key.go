package privacy

import (
	"encoding/hex"
	"github.com/pkg/errors"
	"math/big"

	"github.com/ninjadotorg/constant/common"
)

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
	spendingKey := common.HashB(seed)

	temp := new(big.Int)
	for temp.SetBytes(spendingKey).Cmp(Curve.Params().N) == 1 {
		spendingKey = common.HashB(spendingKey)
	}

	return spendingKey
}

// GeneratePublicKey computes an address corresponding with spendingKey
// Pk : 33 bytes
func GeneratePublicKey(spendingKey []byte) PublicKey {
	var publicKey EllipticPoint
	publicKey.X, publicKey.Y = Curve.ScalarBaseMult(spendingKey)
	return publicKey.Compress()
}

// GenerateReceivingKey computes a receiving key corresponding with spendingKey
// Rk : 32 bytes
func GenerateReceivingKey(spendingKey []byte) ReceivingKey {
	return common.HashB(spendingKey)
}

// GenerateTransmissionKey computes a transmission key corresponding with receivingKey
// Tk : 33 bytes
func GenerateTransmissionKey(receivingKey []byte) TransmissionKey {
	var transmissionKey EllipticPoint
	transmissionKey.X, transmissionKey.Y = Curve.ScalarBaseMult(receivingKey)
	return transmissionKey.Compress()
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

// DecompressKey decompress public key to elliptic point
func DecompressKey(pubKeyStr []byte) (pubkey *EllipticPoint, err error) {
	if len(pubKeyStr) == 0 || len(pubKeyStr) != CompressedPointSize {
		return nil, NewPrivacyErr(UnexpectedErr, errors.New("pubkey string len is wrong"))
	}

	pubkey = new(EllipticPoint)

	err = pubkey.Decompress(pubKeyStr)
	if err != nil {
		return nil, err
	}

	if pubkey.X.Cmp(Curve.Params().P) >= 0 {
		return nil, NewPrivacyErr(UnexpectedErr, errors.New("pubkey X parameter is >= to P"))
	}
	if pubkey.Y.Cmp(Curve.Params().P) >= 0 {
		return nil, NewPrivacyErr(UnexpectedErr, errors.New("pubkey Y parameter is >= to P"))
	}
	if !Curve.Params().IsOnCurve(pubkey.X, pubkey.Y) {
		return nil, NewPrivacyErr(UnexpectedErr, errors.New("pubkey isn't on P256 curve"))
	}
	return pubkey, nil
}

// Bytes converts payment address to bytes array
func (addr *PaymentAddress) Bytes() []byte {
	return append(addr.Pk, addr.Tk...)
}

// SetBytes reverts bytes array to payment address
func (addr *PaymentAddress) SetBytes(bytes []byte) *PaymentAddress {
	// First 33 bytes are public key
	addr.Pk = bytes[:33]
	// Last 33 bytes are transmission key
	addr.Tk = bytes[33:]
	return addr
}

// Size returns size of payment address
func (addr *PaymentAddress) Size() int {
	return len(addr.Pk) + len(addr.Tk)
}

// String converts spending key to string
func (spendingKey SpendingKey) String() string {
	for i := 0; i < SpendingKeySize/2; i++ {
		spendingKey[i], spendingKey[SpendingKeySize-1-i] = spendingKey[SpendingKeySize-1-i], spendingKey[i]
	}
	return hex.EncodeToString(spendingKey[:])
}
