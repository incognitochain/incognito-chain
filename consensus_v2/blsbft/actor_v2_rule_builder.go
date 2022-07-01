package blsbft

import (
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
)

var (
	VOTE_RULE_VOTE                 = "vote"
	VOTE_RULE_NO_VOTE              = "no-vote"
	HANDLE_VOTE_MESSAGE_COLLECT    = "collect-vote"
	HANDLE_VOTE_MESSAGE_NO_COLLECT = "no-collect-vote"
	CREATE_RULE_CREATE_ONLY        = "create-only"
	CREATE_RULE_NORMAL             = "create-repropose"
	HANDLE_PROPOSE_MESSAGE_NORMAL  = "handle-propose-message"
	HANDLE_PROPOSEMESSAGE_NO       = "no-handle-propose-message"
	INSERT_ONLY                    = "insert-only"
	INSERT_AND_BROADCAST           = "insert-and-broadcast"
	VALIDATOR_NO_VALIDATE          = "validator-no-validate"
	VALIDATOR_LEMMA2               = "validator-lemma2"
)

type builderContext struct {
	Lemma2Height      uint64
	VoteRule          string
	CreateRule        string
	HandleVoteRule    string
	HandleProposeRule string
	InsertRule        string
	ValidatorRule     string
}

var ActorRuleBuilderContext = &builderContext{
	VoteRule:          VOTE_RULE_VOTE,
	CreateRule:        CREATE_RULE_NORMAL,
	HandleVoteRule:    HANDLE_VOTE_MESSAGE_COLLECT,
	HandleProposeRule: HANDLE_PROPOSE_MESSAGE_NORMAL,
	InsertRule:        INSERT_AND_BROADCAST,
	ValidatorRule:     VALIDATOR_LEMMA2,
}

func SetBuilderContext(lemma2Height uint64) {
	ActorRuleBuilderContext.Lemma2Height = lemma2Height
}

type IActorV2RuleBuilder interface {
	SetProposeMessageRule(IProposeMessageRule)
	SetVoteRule(IVoteRule)
	SetCreateRule(ICreateNewBlockRule)
	SetValidatorRule(IConsensusValidator)
	SetHandleVoteMessageRule(IHandleVoteMessageRule)
	SetInsertBlockRule(IInsertBlockRule)
	ProposeMessageRule() IProposeMessageRule
	VoteRule() IVoteRule
	CreateRule() ICreateNewBlockRule
	ValidatorRule() IConsensusValidator
	HandleVoteMessageRule() IHandleVoteMessageRule
	InsertBlockRule() IInsertBlockRule
}

type ActorV2RuleBuilder struct {
	proposeRule     IProposeMessageRule
	voteRule        IVoteRule
	createBlockRule ICreateNewBlockRule
	validatorRule   IConsensusValidator
	handleVoteRule  IHandleVoteMessageRule
	insertRule      IInsertBlockRule
	logger          common.Logger
	chain           Chain
}

func NewActorV2RuleBuilder() *ActorV2RuleBuilder {
	return &ActorV2RuleBuilder{}
}

func (r *ActorV2RuleBuilder) SetProposeMessageRule(rule IProposeMessageRule) {
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

func (r *ActorV2RuleBuilder) SetInsertBlockRule(rule IInsertBlockRule) {
	r.insertRule = rule
}

func (r *ActorV2RuleBuilder) ProposeMessageRule() IProposeMessageRule {
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

func (r *ActorV2RuleBuilder) InsertBlockRule() IInsertBlockRule {
	return r.insertRule
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
		d.makeLemma2Rule(builder, logger, chain)

		if ctx.VoteRule == VOTE_RULE_NO_VOTE {
			if reflect.TypeOf(builder.VoteRule()) != reflect.TypeOf(new(NoVoteRule)) {
				builder.SetVoteRule(NewNoVoteRule(
					logger,
				))
			}
			logger.Debug("BLS BFT RULE", ctx.VoteRule, VOTE_RULE_NO_VOTE)
		}

		if ctx.CreateRule == CREATE_RULE_CREATE_ONLY {
			if reflect.TypeOf(builder.CreateRule()) != reflect.TypeOf(new(OnlyCreateBlockRule)) {
				builder.SetCreateRule(NewOnlyCreateBlockRule(
					logger,
					chain,
				))
			}
			logger.Debug("BLS BFT RULE", ctx.CreateRule, CREATE_RULE_CREATE_ONLY)
		}

		if ctx.HandleVoteRule == HANDLE_VOTE_MESSAGE_NO_COLLECT {
			if reflect.TypeOf(builder.HandleVoteMessageRule()) != reflect.TypeOf(new(NoHandleVoteMessage)) {
				builder.SetHandleVoteMessageRule(NewNoHandleVoteMessage())
			}
			logger.Debug("BLS BFT RULE", ctx.HandleVoteRule, HANDLE_VOTE_MESSAGE_NO_COLLECT)
		}

		if ctx.HandleProposeRule == HANDLE_PROPOSEMESSAGE_NO {
			if reflect.TypeOf(builder.ProposeMessageRule()) != reflect.TypeOf(new(NoHandleProposeMessageRule)) {
				builder.SetProposeMessageRule(NewNoHandleProposeMessageRule(
					logger,
				))
			}
			logger.Debug("BLS BFT RULE", ctx.HandleProposeRule, HANDLE_PROPOSEMESSAGE_NO)
		}

		if ctx.InsertRule == INSERT_ONLY {
			if reflect.TypeOf(builder.InsertBlockRule()) != reflect.TypeOf(new(InsertOnlyRule)) {
				builder.SetInsertBlockRule(NewInsertOnlyRule(
					chain,
					logger,
				))
			}
			logger.Debug("BLS BFT RULE", ctx.InsertRule, INSERT_ONLY)
		}

		if ctx.ValidatorRule == VALIDATOR_NO_VALIDATE {
			if reflect.TypeOf(builder.ValidatorRule()) != reflect.TypeOf(new(ConsensusValidatorNoValidate)) {
				builder.SetValidatorRule(NewConsensusValidatorNoValidate())
			}
			logger.Debug("BLS BFT RULE", ctx.ValidatorRule, VALIDATOR_NO_VALIDATE)
		}
	}
}

func (d *ActorV2RuleDirector) makeLemma1Rule(builder IActorV2RuleBuilder, logger common.Logger, chain Chain) IActorV2RuleBuilder {

	builder.SetProposeMessageRule(NewProposeRuleLemma1(
		logger,
	))

	builder.SetCreateRule(NewNormalCreateBlockRule(
		logger,
		chain,
	))

	builder.SetVoteRule(NewVoteRule(
		logger,
	))

	builder.SetValidatorRule(NewConsensusValidatorLemma1(
		logger,
		chain,
	))

	builder.SetHandleVoteMessageRule(NewHandleVoteMessage())

	builder.SetInsertBlockRule(NewInsertAndBroadcastRule(chain, logger))

	return builder
}

func (d *ActorV2RuleDirector) makeLemma2Rule(builder IActorV2RuleBuilder, logger common.Logger, chain Chain) IActorV2RuleBuilder {

	if builder.ProposeMessageRule() == nil ||
		reflect.TypeOf(builder.ProposeMessageRule()) != reflect.TypeOf(new(ProposeRuleLemma2)) {
		builder.SetProposeMessageRule(NewProposeRuleLemma2(
			logger,
			make(map[string]map[int64]string),
			chain,
		))
	}

	if reflect.TypeOf(builder.CreateRule()) != reflect.TypeOf(new(NormalCreateBlockRule)) {
		builder.SetCreateRule(NewNormalCreateBlockRule(
			logger,
			chain,
		))
	}

	if reflect.TypeOf(builder.VoteRule()) != reflect.TypeOf(new(VoteRule)) {
		builder.SetVoteRule(NewVoteRule(
			logger,
		))
	}

	if reflect.TypeOf(builder.ValidatorRule()) != reflect.TypeOf(new(ConsensusValidatorLemma2)) {
		builder.SetValidatorRule(NewConsensusValidatorLemma2(
			logger,
			chain,
		))
	}

	if reflect.TypeOf(builder.HandleVoteMessageRule()) != reflect.TypeOf(new(HandleVoteMessage)) {
		builder.SetHandleVoteMessageRule(NewHandleVoteMessage())
	}

	if reflect.TypeOf(builder.InsertBlockRule()) != reflect.TypeOf(new(InsertAndBroadcastRule)) {
		builder.SetInsertBlockRule(NewInsertAndBroadcastRule(chain, logger))
	}

	return builder
}
