package committeestate

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
)

type unstakeRuleV1 struct{}

func (unstakeRuleV1 *unstakeRuleV1) RemoveFromState(
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

	delete(rewardReceivers, committeePublicKey.GetIncKeyBase58())
	delete(autoStake, key)
	delete(stakingTx, key)

	return autoStake, rewardReceivers, stakingTx, removedStakers, removedTerms, nil
}

func NewUnstakeRuleV1() *unstakeRuleV1 {
	return &unstakeRuleV1{}
}

func (unstakeRuleV1 *unstakeRuleV1) Version() int {
	return unstakeRuleSlashingVersion
}

func (unstakeRuleV1 *unstakeRuleV1) clone() UnstakeRule {
	return NewUnstakeRuleV1()
}
