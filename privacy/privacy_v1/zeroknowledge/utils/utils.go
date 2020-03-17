package utils

import (
	"math"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/privacy_util"
)

// GenerateChallengeFromByte get hash of n points in G append with input values
// return blake_2b(G[0]||G[1]||...||G[CM_CAPACITY-1]||<values>)
// G[i] is list of all generator point of Curve
func GenerateChallenge(values [][]byte) *operation.Scalar {
	bytes := []byte{}
	for i := 0; i < len(operation.PedCom.G); i++ {
		bytes = append(bytes, operation.PedCom.G[i].ToBytesS()...)
	}

	for i := 0; i < len(values); i++ {
		bytes = append(bytes, values[i]...)
	}

	hash := operation.HashToScalar(bytes)
	//res := new(big.Int).SetBytes(hash)
	//res.Mod(res, operation.Curve.Params().N)
	return hash
}

// EstimateProofSize returns the estimated size of the proof in bytes
func EstimateProofSize(nInput int, nOutput int, hasPrivacy bool) uint64 {
	if !hasPrivacy {
		FlagSize := 14 + 2*nInput + nOutput
		sizeSNNoPrivacyProof := nInput * SnNoPrivacyProofSize
		sizeInputCoins := nInput * inputCoinsNoPrivacySize
		sizeOutputCoins := nOutput * OutputCoinsNoPrivacySize

		sizeProof := uint64(FlagSize + sizeSNNoPrivacyProof + sizeInputCoins + sizeOutputCoins)
		return uint64(sizeProof)
	}

	FlagSize := 14 + 7*nInput + 4*nOutput

	sizeOneOfManyProof := nInput * OneOfManyProofSize
	sizeSNPrivacyProof := nInput * SnPrivacyProofSize
	sizeComOutputMultiRangeProof := int(EstimateMultiRangeProofSize(nOutput))

	sizeInputCoins := nInput * inputCoinsPrivacySize
	sizeOutputCoins := nOutput * outputCoinsPrivacySize

	sizeComOutputValue := nOutput * operation.Ed25519KeySize
	sizeComOutputSND := nOutput * operation.Ed25519KeySize
	sizeComOutputShardID := nOutput * operation.Ed25519KeySize

	sizeComInputSK := operation.Ed25519KeySize
	sizeComInputValue := nInput * operation.Ed25519KeySize
	sizeComInputSND := nInput * operation.Ed25519KeySize
	sizeComInputShardID := operation.Ed25519KeySize

	sizeCommitmentIndices := nInput * privacy_util.CommitmentRingSize * common.Uint64Size

	sizeProof := sizeOneOfManyProof + sizeSNPrivacyProof +
		sizeComOutputMultiRangeProof + sizeInputCoins + sizeOutputCoins +
		sizeComOutputValue + sizeComOutputSND + sizeComOutputShardID +
		sizeComInputSK + sizeComInputValue + sizeComInputSND + sizeComInputShardID +
		sizeCommitmentIndices + FlagSize

	return uint64(sizeProof)
}

func pad(num int) int {
	if num == 1 || num == 2 {
		return num
	}
	tmp := 2
	for i := 2; ; i++ {
		tmp *= 2
		if tmp >= num {
			num = tmp
			break
		}
	}
	return num
}

// estimateMultiRangeProofSize estimate multi range proof size
func EstimateMultiRangeProofSize(nOutput int) uint64 {
	return uint64((nOutput+2*int(math.Log2(float64(maxExp*pad(nOutput))))+5)*operation.Ed25519KeySize + 5*operation.Ed25519KeySize + 2)
}
