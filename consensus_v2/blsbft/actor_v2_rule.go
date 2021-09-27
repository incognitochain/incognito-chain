package blsbft

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
)

var (
	lemma2Height = uint64(1e9)
	IsNoVoteRule = false
)

type ActorV2RuleStrategy struct {
	proposeRule   IProposeRule
	voteRule      IVoteRule
	validatorRule IConsensusValidator
	logger        common.Logger
	chain         Chain
}

func NewMultiviewActorRule(logger common.Logger, chain Chain, bestViewHeight uint64) *ActorV2RuleStrategy {
	lemma2Height = config.Param().ConsensusParam.Lemma2Height
	r := &ActorV2RuleStrategy{logger: logger, chain: chain}
	r.SetStrategy(bestViewHeight)
	return r
}

func (r *ActorV2RuleStrategy) SetStrategy(bestViewHeight uint64) {

	if bestViewHeight >= lemma2Height {
		r.proposeRule = NewProposeRuleLemma2(
			r.logger,
			make(map[string]map[int64]string),
			r.chain,
		)
		r.validatorRule = NewConsensusValidatorV2(
			r.logger,
			r.chain,
		)
	} else {
		r.proposeRule = NewProposeRuleLemma1(
			r.logger,
		)
		r.validatorRule = NewConsensusValidatorV1(
			r.logger,
			r.chain,
		)
	}

	if !IsNoVoteRule {
		r.voteRule = NewVoteRule(
			r.logger,
		)
	} else {
		r.voteRule = NewNoVoteRule(
			r.logger,
		)
	}
}
