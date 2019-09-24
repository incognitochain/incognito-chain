package utils

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/zeroknowledge/aggregaterange"
)

// GenerateChallengeFromByte get hash of n points in G append with input values
// return blake_2b(G[0]||G[1]||...||G[CM_CAPACITY-1]||<values>)
// G[i] is list of all generator point of Curve
func GenerateChallenge(values [][]byte) []byte {
	bytes := []byte{}
	for i := 0; i < len(privacy.PedCom.G); i++ {
		bytes = append(bytes, privacy.ArrayToSlice(privacy.PedCom.G[i].ToBytes())...)
	}

	for i := 0; i < len(values); i++ {
		bytes = append(bytes, values[i]...)
	}

	hash := privacy.HashToScalar(bytes)

	//res := new(big.Int).SetBytes(hash)
	//res.Mod(res, privacy.Curve.Params().N)
	return privacy.ArrayToSlice(hash.ToBytes())
}

// EstimateProofSize returns the estimated size of the proof in bytes
func EstimateProofSize(nInput int, nOutput int, hasPrivacy bool) uint64 {
	if !hasPrivacy {
		FlagSize := 14 + 2*nInput + nOutput
		sizeSNNoPrivacyProof := nInput * SnNoPrivacyProofSize
		sizeInputCoins := nInput * inputCoinsNoPrivacySize
		sizeOutputCoins := nOutput * outputCoinsNoPrivacySize

		sizeProof := uint64(FlagSize + sizeSNNoPrivacyProof + sizeInputCoins + sizeOutputCoins)
		return uint64(sizeProof)
	}

	FlagSize := 14 + 7*nInput + 4*nOutput

	sizeOneOfManyProof := nInput * OneOfManyProofSize
	sizeSNPrivacyProof := nInput * SnPrivacyProofSize
	sizeComOutputMultiRangeProof := int(aggregaterange.EstimateMultiRangeProofSize(nOutput))

	sizeInputCoins := nInput * inputCoinsPrivacySize
	sizeOutputCoins := nOutput * outputCoinsPrivacySize

	sizeComOutputValue := nOutput * privacy.CompressedEllipticPointSize
	sizeComOutputSND := nOutput * privacy.CompressedEllipticPointSize
	sizeComOutputShardID := nOutput * privacy.CompressedEllipticPointSize

	sizeComInputSK := privacy.CompressedEllipticPointSize
	sizeComInputValue := nInput * privacy.CompressedEllipticPointSize
	sizeComInputSND := nInput * privacy.CompressedEllipticPointSize
	sizeComInputShardID := privacy.CompressedEllipticPointSize

	sizeCommitmentIndices := nInput * privacy.CommitmentRingSize * common.Uint64Size

	sizeProof := sizeOneOfManyProof + sizeSNPrivacyProof +
		sizeComOutputMultiRangeProof + sizeInputCoins + sizeOutputCoins +
		sizeComOutputValue + sizeComOutputSND + sizeComOutputShardID +
		sizeComInputSK + sizeComInputValue + sizeComInputSND + sizeComInputShardID +
		sizeCommitmentIndices + FlagSize

	return uint64(sizeProof)
}
