package privacy

import (
	"errors"

	zkp "github.com/incognitochain/incognito-chain/privacy/privacy_v1/zeroknowledge"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2"

	"github.com/incognitochain/incognito-chain/privacy/coin"
	errhandler "github.com/incognitochain/incognito-chain/privacy/errorhandler"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/privacy_util"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v1/hybridencryption"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v1/schnorr"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v1/zeroknowledge/aggregatedrange"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2/bulletproofs"
	"github.com/incognitochain/incognito-chain/privacy/proof"
	"github.com/incognitochain/incognito-chain/privacy/proof/agg_interface"
)

type PrivacyError = errhandler.PrivacyError

var ErrCodeMessage = errhandler.ErrCodeMessage

// Public Constants
const (
	CStringBurnAddress    = "burningaddress"
	Ed25519KeySize        = operation.Ed25519KeySize
	CStringBulletProof    = operation.CStringBulletProof
	CommitmentRingSize    = privacy_util.CommitmentRingSize
	CommitmentRingSizeExp = privacy_util.CommitmentRingSizeExp

	PedersenSndIndex        = operation.PedersenSndIndex
	PedersenValueIndex      = operation.PedersenValueIndex
	PedersenShardIDIndex    = operation.PedersenShardIDIndex
	PedersenPrivateKeyIndex = operation.PedersenPrivateKeyIndex
	PedersenRandomnessIndex = operation.PedersenRandomnessIndex

	RingSize = privacy_util.RingSize
)

var PedCom = operation.PedCom

const (
	MaxSizeInfoCoin = coin.MaxSizeInfoCoin // byte
)

// Export as package privacy for other packages easily use it

type Point = operation.Point
type Scalar = operation.Scalar
type HybridCipherText = hybridencryption.HybridCipherText

// Point and Scalar operations
func RandomScalar() *Scalar {
	return operation.RandomScalar()
}

func HashToPoint(b []byte) *Point {
	return operation.HashToPoint(b)
}

type PublicKey = key.PublicKey
type ViewingKey = key.ViewingKey
type PrivateKey = key.PrivateKey
type PaymentInfo = key.PaymentInfo
type PaymentAddress = key.PaymentAddress

func GeneratePrivateKey(seed []byte) PrivateKey {
	return key.GeneratePrivateKey(seed)
}

func GeneratePaymentAddress(privateKey []byte) PaymentAddress {
	return key.GeneratePaymentAddress(privateKey)
}

func GenerateViewingKey(privateKey []byte) ViewingKey {
	return key.GenerateViewingKey(privateKey)
}

type SchnSignature = schnorr.SchnSignature
type SchnorrPublicKey = schnorr.SchnorrPublicKey
type SchnorrPrivateKey = schnorr.SchnorrPrivateKey

type Coin = coin.Coin
type CoinV1 = coin.CoinV1
type CoinV2 = coin.CoinV2
type CoinObject = coin.CoinObject

type Proof = proof.Proof
type ProofV1 = zkp.PaymentProof
type ProofV2 = privacy_v2.PaymentProofV2
type AggregatedRangeProof = agg_interface.AggregatedRangeProof
type AggregatedRangeProofV1 = aggregatedrange.AggregatedRangeProof
type AggregatedRangeProofV2 = bulletproofs.AggregatedRangeProof

var LoggerV1 = &zkp.Logger
var LoggerV2 = &privacy_v2.Logger

func NewProofWithVersion(version int8) Proof {
	var result Proof
	if version == 1 {
		result = &zkp.PaymentProof{}
	} else {
		result = &privacy_v2.PaymentProofV2{}
	}
	return result
}

func ArrayScalarToBytes(arr *[]*operation.Scalar) ([]byte, error) {
	scalarArr := *arr

	n := len(scalarArr)
	if n > 255 {
		return nil, errors.New("ArrayScalarToBytes: length of scalar array is too big")
	}
	b := make([]byte, 1)
	b[0] = byte(n)

	for _, sc := range scalarArr {
		b = append(b, sc.ToBytesS()...)
	}
	return b, nil
}

func ArrayScalarFromBytes(b []byte) (*[]*operation.Scalar, error) {
	if len(b) == 0 {
		return nil, errors.New("ArrayScalarFromBytes error: length of byte is 0")
	}
	n := int(b[0])
	if n*Ed25519KeySize+1 != len(b) {
		return nil, errors.New("ArrayScalarFromBytes error: length of byte is not correct")
	}
	scalarArr := make([]*operation.Scalar, n)
	offset := 1
	for i := 0; i < n; i += 1 {
		curByte := b[offset : offset+Ed25519KeySize]
		scalarArr[i] = new(operation.Scalar).FromBytesS(curByte)
		offset += Ed25519KeySize
	}
	return &scalarArr, nil
}
