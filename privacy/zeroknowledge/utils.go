package zkp

import (
	"github.com/ninjadotorg/constant/common"
	"math"
	"math/big"

	"github.com/ninjadotorg/constant/privacy"
)

// GenerateChallengeFromByte get hash of n points in G append with input values
// return blake_2b(G[0]||G[1]||...||G[CM_CAPACITY-1]||<values>)
// G[i] is list of all generator point of Curve
func generateChallenge(values [][]byte) *big.Int {
	bytes := privacy.PedCom.G[0].Compress()
	for i := 1; i < len(privacy.PedCom.G); i++ {
		bytes = append(bytes, privacy.PedCom.G[i].Compress()...)
	}

	for i := 0; i < len(values); i++ {
		bytes = append(bytes, values[i]...)
	}

	hash := common.HashB(bytes)

	res := new(big.Int).SetBytes(hash)
	res.Mod(res, privacy.Curve.Params().N)
	return res
}

// EstimateProofSize returns the estimated size of the proof in bytes
func EstimateProofSize(nInput int, nOutput int, hasPrivacy bool) uint64 {
	if !hasPrivacy{
		FlagSize := 14 + 2*nInput + nOutput
		sizeSNNoPrivacyProof := nInput * privacy.SNNoPrivacyProofSize
		sizeInputCoins := nInput * privacy.InputCoinsNoPrivacySize
		sizeOutputCoins := nOutput * privacy.OutputCoinsNoPrivacySize

		sizeProof := uint64(FlagSize + sizeSNNoPrivacyProof + sizeInputCoins + sizeOutputCoins)
		return uint64(math.Ceil(float64(sizeProof) / 1024))
	}

	FlagSize := 14 + 7*nInput + 4*nOutput

	sizeOneOfManyProof := nInput * privacy.OneOfManyProofSize
	sizeSNPrivacyProof := nInput * privacy.SNPrivacyProofSize
	sizeComOutputMultiRangeProof := int(estimateMultiRangeProofSize(nOutput))


	sizeInputCoins := nInput * privacy.InputCoinsPrivacySize
	sizeOutputCoins := nOutput * privacy.OutputCoinsPrivacySize

	sizeComOutputValue := nOutput * privacy.CompressedPointSize
	sizeComOutputSND := nOutput * privacy.CompressedPointSize
	sizeComOutputShardID := nOutput * privacy.CompressedPointSize

	sizeComInputSK := privacy.CompressedPointSize
	sizeComInputValue := nInput * privacy.CompressedPointSize
	sizeComInputSND := nInput * privacy.CompressedPointSize
	sizeComInputShardID := privacy.CompressedPointSize

	sizeCommitmentIndices := nInput * privacy.CMRingSize * privacy.Uint64Size

	sizeProof := sizeOneOfManyProof + sizeSNPrivacyProof +
		sizeComOutputMultiRangeProof + sizeInputCoins + sizeOutputCoins +
		sizeComOutputValue + sizeComOutputSND + sizeComOutputShardID +
		sizeComInputSK + sizeComInputValue + sizeComInputSND + sizeComInputShardID +
		sizeCommitmentIndices + FlagSize

	return uint64(sizeProof)
}


