package committeestate

import (
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
)

type AssignRuleV2 struct{}

//GenInstruction generate instruction by assignrule
func (a *AssignRuleV2) GenInstruction(
	validators []string, rand int64, assignOffset int, terms map[string]uint64, blockHeight uint64, shardID byte,
) (*instruction.AssignInstruction, error) {
	validKeys := []string{}
	for _, v := range validators {
		if terms[v]-syncTerm-blockHeight < 0 {
			break
		}
		validKeys = append(validKeys, v)
	}

	validKeysStruct, _ := incognitokey.CommitteeBase58KeyListToStruct(validKeys)
	assignInstruction := &instruction.AssignInstruction{
		ChainID:               int(shardID),
		ShardCandidates:       validKeys,
		ShardCandidatesStruct: validKeysStruct,
	}

	return assignInstruction, nil
}
