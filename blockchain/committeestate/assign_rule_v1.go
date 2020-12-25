package committeestate

import (
	"github.com/incognitochain/incognito-chain/instruction"
)

type AssignRuleV1 struct{}

//GenInstruction generate instruction by assignrule
func (a *AssignRuleV1) GenInstruction(
	validators []string, rand int64, assignOffset int, terms map[string]uint64, blockHeight uint64, shardID byte,
) (*instruction.AssignInstruction, error) {
	assignInstruction := &instruction.AssignInstruction{}
	return assignInstruction, nil
}
