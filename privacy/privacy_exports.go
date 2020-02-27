package privacy

import (
	"github.com/incognitochain/incognito-chain/privacy/hybridencryption"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	C25519 "github.com/incognitochain/incognito-chain/privacy/operation/curve25519"
	"github.com/incognitochain/incognito-chain/privacy/pedersen"
)

// Public Constants
const (
	Ed25519KeySize        = 32
	AESKeySize            = 32
	CommitmentRingSize    = 8
	CommitmentRingSizeExp = 3
	CStringBulletProof    = "bulletproof"
	CStringBurnAddress    = "burningaddress"

	PedersenPrivateKeyIndex = byte(0x00)
	PedersenValueIndex      = byte(0x01)
	PedersenSndIndex        = byte(0x02)
	PedersenShardIDIndex    = byte(0x03)
	PedersenRandomnessIndex = byte(0x04)
)

var PedCom = pedersen.NewPedersenParams()

const (
	MaxSizeInfoCoin = 255 // byte
)

// Export as package privacy for other packages easily use it

type Point = operation.Point
type Scalar = operation.Scalar
type HybridCipherText = hybridencryption.HybridCipherText

// Point and Scalar operations
func RandomScalar() *Scalar {
	return operation.RandomScalar()
}

func RandomPoint() *Point {
	return operation.RandomPoint()
}

func CheckDuplicateScalarArray(arr []*Scalar) bool {
	return operation.CheckDuplicateScalarArray(arr)
}

func IsPointEqual(pa *Point, pb *Point) bool {
	return operation.IsPointEqual(pa, pb)
}

func HashToPoint(b []byte) *Point {
	return operation.HashToPoint(b)
}

func HashToScalar(data []byte) *Scalar {
	return operation.HashToScalar(data)
}

func Reverse(x C25519.Key) (result C25519.Key) {
	return operation.Reverse(x)
}

func HashToPointFromIndex(index int64, padStr string) *Point {
	return operation.HashToPointFromIndex(index, padStr)
}

// Hybrid encryption operations
func HybridEncrypt(msg []byte, publicKey *operation.Point) (ciphertext *HybridCipherText, err error) {
	return hybridencryption.HybridEncrypt(msg, publicKey)
}

func HybridDecrypt(ciphertext *HybridCipherText, privateKey *operation.Scalar) (msg []byte, err error) {
	return hybridencryption.HybridDecrypt(ciphertext, privateKey)
}
