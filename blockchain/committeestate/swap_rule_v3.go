package committeestate

import (
	"github.com/incognitochain/incognito-chain/blockchain/signaturecounter"
	"github.com/incognitochain/incognito-chain/instruction"
)

//swapRuleV3 ...
type swapRuleV3 struct {
}

func NewSwapRuleV3() *swapRuleV3 {
	return &swapRuleV3{}
}

func (s *swapRuleV3) GenInstructions(
	shardID byte,
	committees, substitutes []string,
	minCommitteeSize, maxCommitteeSize, typeIns, numberOfFixedValidators int,
	penalty map[string]signaturecounter.Penalty,
) (*instruction.SwapShardInstruction, []string, []string, []string, []string) {
	return nil, nil, nil, nil, nil
}

func (s *swapRuleV3) getSwapOutOffset(lenSubstitute, lenCommittees, numberOfFixedValidators, minCommitteeSize int) int {
	return 0
}
