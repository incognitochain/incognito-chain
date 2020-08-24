package jsonresult

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
)

type RandomCommitmentAndPublicKeyResult struct {
	CommitmentIndices  []uint64 `json:"CommitmentIndices"`
	PublicKeys 			[]string `json:"PublicKeys"`
	Commitments        []string `json:"Commitments"`
}

func NewRandomCommitmentAndPublicKeyResult(commitmentIndices []uint64, publicKeys, commitments [][]byte) *RandomCommitmentAndPublicKeyResult {
	result := &RandomCommitmentAndPublicKeyResult{
		CommitmentIndices:   commitmentIndices,
	}
	tempCommitments := make([]string, 0)
	for _, commitment := range commitments {
		tempCommitments = append(tempCommitments, base58.Base58Check{}.Encode(commitment, common.ZeroByte))
	}
	result.Commitments = tempCommitments

	tempPublicKeys := make([]string, 0)
	for _, pubkey := range publicKeys {
		tempPublicKeys = append(tempPublicKeys, base58.Base58Check{}.Encode(pubkey, common.ZeroByte))
	}
	result.PublicKeys = tempPublicKeys

	return result
}