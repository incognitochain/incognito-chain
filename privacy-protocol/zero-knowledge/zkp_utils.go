package zkp

import (
	"fmt"
	"math"
	"math/big"

	"github.com/ninjadotorg/constant/privacy-protocol"

	"github.com/minio/blake2b-simd"
)

// GenerateChallengeFromPoint get hash of n points in G append with input values
// return blake_2b(G[0]||G[1]||...||G[CM_CAPACITY-1]||<values>)
// G[i] is list of all generator point of Curve
func GenerateChallengeFromPoint(values []*privacy.EllipticPoint) *big.Int {
	appendStr := privacy.PedCom.G[0].Compress()
	for i := 1; i < privacy.PedCom.Capacity; i++ {
		appendStr = append(appendStr, privacy.PedCom.G[i].Compress()...)
	}
	fmt.Printf("len values: %v\n", len(values))

	fmt.Printf("values[0]: %v\n", values[0])
	for i := 0; i < len(values); i++ {
		appendStr = append(appendStr, values[i].Compress()...)
	}
	hashFunc := blake2b.New256()
	hashFunc.Write(appendStr)
	hashValue := hashFunc.Sum(nil)
	result := big.NewInt(0)
	result.SetBytes(hashValue)
	result.Mod(result, privacy.Curve.Params().N)
	return result
}

// GenerateChallengeFromByte get hash of n points in G append with input values
// return blake_2b(G[0]||G[1]||...||G[CM_CAPACITY-1]||<values>)
// G[i] is list of all generator point of Curve
func GenerateChallengeFromByte(values [][]byte) *big.Int {
	appendStr := privacy.PedCom.G[0].Compress()
	for i := 1; i < privacy.PedCom.Capacity; i++ {
		appendStr = append(appendStr, privacy.PedCom.G[i].Compress()...)
	}
	for i := 0; i < len(values); i++ {
		appendStr = append(appendStr, values[i]...)
	}
	hashFunc := blake2b.New256()
	hashFunc.Write(appendStr)
	hashValue := hashFunc.Sum(nil)
	result := big.NewInt(0)
	result.SetBytes(hashValue)
	result.Mod(result, privacy.Curve.Params().N)
	return result
}

// EstimateProofSize returns the estimated size of the proof in kilobyte
func EstimateProofSize(inputCoins []*privacy.OutputCoin, payments []*privacy.PaymentInfo) uint64 {
	nInput := len(inputCoins)
	nOutput := len(payments)

	sizeComInputOpeningsProof := nInput * privacy.ComInputOpeningsProofSize
	sizeOneOfManyProof := nInput * privacy.OneOfManyProofSize
	sizeEqualityOfCommittedValProof := nInput * privacy.EqualityOfCommittedValProofSize
	sizeProductCommitmentProof := nInput * privacy.ProductCommitmentProofSize

	sizeComOutputOpeningsProof := nOutput * privacy.ComOutputOpeningsProofSize
	sizeComOutputMultiRangeProof := privacy.ComOutputMultiRangeProofSize
	sizeSumOutRangeProof := privacy.ComZeroProofSize
	sizeComZeroProof := privacy.ComZeroProofSize

	sizeInputCoins :=  nInput * privacy.InputCoinsPrivacySize
	sizeOutputCoins := nOutput * privacy.OutputCoinsPrivacySize

	sizeComOutputValue  := nOutput * privacy.CompressedPointSize
	sizeComOutputSND  := nOutput * privacy.CompressedPointSize
	sizeComOutputShardID  := nOutput * privacy.CompressedPointSize

	sizeProof := sizeComInputOpeningsProof + sizeOneOfManyProof + sizeEqualityOfCommittedValProof + sizeProductCommitmentProof +
		sizeComOutputOpeningsProof + sizeComOutputMultiRangeProof + sizeSumOutRangeProof + sizeComZeroProof + sizeInputCoins + sizeOutputCoins +
		sizeComOutputValue + sizeComOutputSND + sizeComOutputShardID

	return uint64(math.Ceil(float64(sizeProof) / 1024))
}
