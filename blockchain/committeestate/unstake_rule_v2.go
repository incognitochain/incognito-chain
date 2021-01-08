package committeestate

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
)

type unstakeRuleV2 struct{}

func (unstakeRuleV2 *unstakeRuleV2) RemoveFromState(
	committeePublicKey incognitokey.CommitteePublicKey,
	autoStake map[string]bool,
	rewardReceivers map[string]privacy.PaymentAddress,
	stakingTx map[string]common.Hash,
	terms map[string]uint64,
	removedStakers []string,
	removedTerms []string,
) (
	map[string]bool,
	map[string]privacy.PaymentAddress,
	map[string]common.Hash,
	[]string, []string,
	error) {

	key, err := committeePublicKey.ToBase58()
	if err != nil {
		return autoStake, rewardReceivers, stakingTx, removedStakers, removedTerms, err
	}
	removedStakers = append(removedStakers, key)
	removedTerms = append(removedTerms, key)

	delete(rewardReceivers, committeePublicKey.GetIncKeyBase58())
	delete(autoStake, key)
	delete(stakingTx, key)
	delete(terms, key)

	return autoStake, rewardReceivers, stakingTx, removedStakers, removedTerms, nil
}

func NewUnstakeRuleV2() *unstakeRuleV2 {
	return &unstakeRuleV2{}
}

func (unstakeRuleV2 *unstakeRuleV2) Version() int {
	return unstakeRuleDCSVersion
}

func (unstakeRuleV2 *unstakeRuleV2) clone() UnstakeRule {
	return NewUnstakeRuleV2()
}
