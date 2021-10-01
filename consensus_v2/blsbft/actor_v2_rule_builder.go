package blsbft

import (
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
)

var (
	VOTERULE_VOTE                = "vote"
	VOTERULE_NO_VOTE             = "no-vote"
	HANDLEVOTEMESSAGE_COLLECT    = "collect-vote"
	HANDLEVOTEMESSAGE_NO_COLLECT = "no-collect-vote"
	CREATERULE_CREATE_ONLY       = "create"
	CREATERULE_NORMAL            = "create-repropose"
)

type builderContext struct {
	Lemma2Height   uint64
	VoteRule       string
	CreateRule     string
	HandleVoteRule string
}

var ActorV2BuilderContext = &builderContext{
	VoteRule:       VOTERULE_VOTE,
	CreateRule:     CREATERULE_NORMAL,
	HandleVoteRule: HANDLEVOTEMESSAGE_COLLECT,
}

func SetBuilderContext(lemma2Height uint64) {
	ActorV2BuilderContext.Lemma2Height = lemma2Height
}

type IActorV2RuleBuilder interface {
	SetProposeRule(IProposeRule)
	SetVoteRule(IVoteRule)
	SetCreateRule(ICreateNewBlockRule)
	SetValidatorRule(IConsensusValidator)
	SetHandleVoteMessageRule(IHandleVoteMessageRule)
	ProposeRule() IProposeRule
	VoteRule() IVoteRule
	CreateRule() ICreateNewBlockRule
	ValidatorRule() IConsensusValidator
	HandleVoteMessageRule() IHandleVoteMessageRule
}

type ActorV2RuleBuilder struct {
	proposeRule     IProposeRule
	voteRule        IVoteRule
	createBlockRule ICreateNewBlockRule
	validatorRule   IConsensusValidator
	handleVoteRule  IHandleVoteMessageRule
	logger          common.Logger
	chain           Chain
}

func NewActorV2RuleBuilder() *ActorV2RuleBuilder {
	return &ActorV2RuleBuilder{}
}

func (r *ActorV2RuleBuilder) SetProposeRule(rule IProposeRule) {
	r.proposeRule = rule
}

func (r *ActorV2RuleBuilder) SetVoteRule(rule IVoteRule) {
	r.voteRule = rule
}

func (r *ActorV2RuleBuilder) SetCreateRule(rule ICreateNewBlockRule) {
	r.createBlockRule = rule
}

func (r *ActorV2RuleBuilder) SetValidatorRule(validator IConsensusValidator) {
	r.validatorRule = validator
}

func (r *ActorV2RuleBuilder) SetHandleVoteMessageRule(rule IHandleVoteMessageRule) {
	r.handleVoteRule = rule
}

func (r *ActorV2RuleBuilder) ProposeRule() IProposeRule {
	return r.proposeRule
}

func (r *ActorV2RuleBuilder) VoteRule() IVoteRule {
	return r.voteRule
}

func (r *ActorV2RuleBuilder) CreateRule() ICreateNewBlockRule {
	return r.createBlockRule
}

func (r *ActorV2RuleBuilder) ValidatorRule() IConsensusValidator {
	return r.validatorRule
}

func (r *ActorV2RuleBuilder) HandleVoteMessageRule() IHandleVoteMessageRule {
	return r.handleVoteRule
}

type ActorV2RuleDirector struct {
	builder IActorV2RuleBuilder
}

func NewActorV2RuleDirector() *ActorV2RuleDirector {
	return &ActorV2RuleDirector{}
}

func (d *ActorV2RuleDirector) initRule(ctx *builderContext,
	bestViewHeight uint64,
	chain Chain,
	logger common.Logger,
) {

	builder := NewActorV2RuleBuilder()
	if bestViewHeight >= ctx.Lemma2Height {
		d.builder = d.makeLemma2Rule(builder, logger, chain)
	} else {
		d.builder = d.makeLemma1Rule(builder, logger, chain)
	}
}

func (d *ActorV2RuleDirector) updateRule(
	ctx *builderContext,
	builder IActorV2RuleBuilder,
	bestViewHeight uint64,
	chain Chain,
	logger common.Logger,
) {
	if bestViewHeight >= ctx.Lemma2Height {
		builder = d.makeLemma2Rule(builder, logger, chain)
		if ctx.VoteRule == VOTERULE_VOTE {
			builder.SetVoteRule(NewVoteRule(
				logger,
			))
		}

		if ctx.VoteRule == VOTERULE_NO_VOTE {
			builder.SetVoteRule(NewNoVoteRule(
				logger,
			))
		}

		if ctx.CreateRule == CREATERULE_CREATE_ONLY {
			builder.SetCreateRule(NewOnlyCreateBlockRule(
				logger,
				chain,
			))
		}

		if ctx.CreateRule == CREATERULE_NORMAL {
			builder.SetCreateRule(NewNormalCreateBlockRule(
				logger,
				chain,
			))
		}

		if ctx.HandleVoteRule == HANDLEVOTEMESSAGE_COLLECT {
			builder.SetHandleVoteMessageRule(NewHandleVoteMessage())
		}

		if ctx.HandleVoteRule == HANDLEVOTEMESSAGE_NO_COLLECT {
			builder.SetHandleVoteMessageRule(NewNoHandleVoteMessage())
		}
	} else {
		builder = d.makeLemma1Rule(builder, logger, chain)
	}

}

func (d *ActorV2RuleDirector) makeLemma1Rule(builder IActorV2RuleBuilder, logger common.Logger, chain Chain) IActorV2RuleBuilder {

	builder.SetProposeRule(NewProposeRuleLemma1(
		logger,
	))

	builder.SetCreateRule(NewNormalCreateBlockRule(
		logger,
		chain,
	))

	builder.SetVoteRule(NewVoteRule(
		logger,
	))

	builder.SetValidatorRule(NewConsensusValidatorV1(
		logger,
		chain,
	))

	builder.SetHandleVoteMessageRule(NewHandleVoteMessage())

	return builder
}

func (d *ActorV2RuleDirector) makeLemma2Rule(builder IActorV2RuleBuilder, logger common.Logger, chain Chain) IActorV2RuleBuilder {

	if builder.ProposeRule() == nil ||
		reflect.TypeOf(builder.ProposeRule()) != reflect.TypeOf(new(ProposeRuleLemma2)) {
		builder.SetProposeRule(NewProposeRuleLemma2(
			logger,
			make(map[string]map[int64]string),
			chain,
		))
	}
	builder.SetCreateRule(NewNormalCreateBlockRule(
		logger,
		chain,
	))

	builder.SetVoteRule(NewVoteRule(
		logger,
	))

	builder.SetValidatorRule(NewConsensusValidatorV2(
		logger,
		chain,
	))

	builder.SetHandleVoteMessageRule(NewHandleVoteMessage())

	return builder
}
