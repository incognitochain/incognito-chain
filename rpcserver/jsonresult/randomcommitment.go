package jsonresult

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
)

type RandomCommitmentResult struct {
	CommitmentIndices  []uint64 `json:"CommitmentIndices"`
	MyCommitmentIndexs []uint64 `json:"MyCommitmentIndexs"`
	Commitments        []string `json:"Commitments"`
}

func NewRandomCommitmentResult(commitmentIndexs []uint64, myCommitmentIndexs []uint64, commitments [][]byte) *RandomCommitmentResult {
	result := &RandomCommitmentResult{
		CommitmentIndices:  commitmentIndexs,
		MyCommitmentIndexs: myCommitmentIndexs,
	}
	temp := []string{}
	for _, commitment := range commitments {
		temp = append(temp, base58.Base58Check{}.Encode(commitment, common.ZeroByte))
	}
	result.Commitments = temp
	return result
}
