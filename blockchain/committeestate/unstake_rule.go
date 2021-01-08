package committeestate

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
)

type UnstakeRule interface {
	RemoveFromState(
		incognitokey.CommitteePublicKey,
		map[string]bool,
		map[string]privacy.PaymentAddress,
		map[string]common.Hash,
		map[string]uint64,
		[]string, []string,
	) (
		map[string]bool,
		map[string]privacy.PaymentAddress,
		map[string]common.Hash,
		[]string, []string,
		error)
	Version() int
}

func cloneUnstakeRuleByVersion(unstakeRule UnstakeRule) UnstakeRule {
	var res UnstakeRule
	if unstakeRule != nil {
		switch unstakeRule.Version() {
		case unstakeRuleSlashingVersion:
			res = unstakeRule.(*unstakeRuleV1).clone()
		case unstakeRuleDCSVersion:
			res = unstakeRule.(*unstakeRuleV2).clone()
		case unstakeRuleTestVersion:
			res = unstakeRule
		default:
			panic("Not implement this version yet")
		}
	}
	return res
}
