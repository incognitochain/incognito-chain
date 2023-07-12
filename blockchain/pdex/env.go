package pdex

import (
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
)

type StateEnvBuilder interface {
	BuildContributionActions([][]string) StateEnvBuilder
	BuildPRVRequiredContributionActions([][]string) StateEnvBuilder
	BuildTradeActions([][]string) StateEnvBuilder
	BuildCrossPoolTradeActions([][]string) StateEnvBuilder
	BuildWithdrawalActions([][]string) StateEnvBuilder
	BuildFeeWithdrawalActions([][]string) StateEnvBuilder
	BuildPrevBeaconHeight(uint64) StateEnvBuilder
	BuildListTxs(map[byte][]metadata.Transaction) StateEnvBuilder
	BuildBeaconInstructions([][]string) StateEnvBuilder
	BuildStateDB(*statedb.StateDB) StateEnvBuilder
	BuildBCHeightBreakPointPrivacyV2(uint64) StateEnvBuilder
	BuildPdexv3BreakPoint(uint64) StateEnvBuilder
	BuildReward(uint64) StateEnvBuilder
	BuildIsInscriptionFeatureEnabled(bool) StateEnvBuilder
	Build() StateEnvironment
}

func NewStateEnvBuilder() StateEnvBuilder {
	return &stateEnvironment{}
}

type stateEnvironment struct {
	contributionActions            [][]string
	prvRequiredContributionActions [][]string
	tradeActions                   [][]string
	crossPoolTradeActions          [][]string
	withdrawalActions              [][]string
	feeWithdrawalActions           [][]string
	prevBeaconHeight               uint64
	beaconInstructions             [][]string
	listTxs                        map[byte][]metadata.Transaction
	stateDB                        *statedb.StateDB
	bcHeightBreakPointPrivacyV2    uint64
	reward                         uint64
	pdexv3BreakPoint               uint64
	isInscriptionEnabled           bool
}

func (env *stateEnvironment) BuildPdexv3BreakPoint(beaconHeight uint64) StateEnvBuilder {
	env.pdexv3BreakPoint = beaconHeight
	return env
}

func (env *stateEnvironment) BuildBCHeightBreakPointPrivacyV2(beaconHeight uint64) StateEnvBuilder {
	env.bcHeightBreakPointPrivacyV2 = beaconHeight
	return env
}

func (env *stateEnvironment) BuildContributionActions(actions [][]string) StateEnvBuilder {
	env.contributionActions = actions
	return env
}

func (env *stateEnvironment) BuildPRVRequiredContributionActions(actions [][]string) StateEnvBuilder {
	env.prvRequiredContributionActions = actions
	return env
}

func (env *stateEnvironment) BuildTradeActions(actions [][]string) StateEnvBuilder {
	env.tradeActions = actions
	return env
}

func (env *stateEnvironment) BuildCrossPoolTradeActions(actions [][]string) StateEnvBuilder {
	env.crossPoolTradeActions = actions
	return env
}

func (env *stateEnvironment) BuildWithdrawalActions(actions [][]string) StateEnvBuilder {
	env.withdrawalActions = actions
	return env
}

func (env *stateEnvironment) BuildFeeWithdrawalActions(actions [][]string) StateEnvBuilder {
	env.feeWithdrawalActions = actions
	return env
}

func (env *stateEnvironment) BuildPrevBeaconHeight(beaconHeight uint64) StateEnvBuilder {
	env.prevBeaconHeight = beaconHeight
	return env
}

func (env *stateEnvironment) BuildBeaconInstructions(beaconInstructions [][]string) StateEnvBuilder {
	env.beaconInstructions = beaconInstructions
	return env
}

func (env *stateEnvironment) BuildListTxs(listTxs map[byte][]metadata.Transaction) StateEnvBuilder {
	env.listTxs = listTxs
	return env
}

func (env *stateEnvironment) BuildStateDB(stateDB *statedb.StateDB) StateEnvBuilder {
	env.stateDB = stateDB
	return env
}

func (env *stateEnvironment) BuildReward(reward uint64) StateEnvBuilder {
	env.reward = reward
	return env
}

func (env *stateEnvironment) BuildIsInscriptionFeatureEnabled(isEnabled bool) StateEnvBuilder {
	env.isInscriptionEnabled = isEnabled
	return env
}

func (env *stateEnvironment) Build() StateEnvironment {
	return env
}

type StateEnvironment interface {
	ContributionActions() [][]string
	PRVRequiredContributionActions() [][]string
	TradeActions() [][]string
	CrossPoolTradeActions() [][]string
	WithdrawalActions() [][]string
	FeeWithdrawalActions() [][]string
	PrevBeaconHeight() uint64
	BeaconInstructions() [][]string
	ListTxs() map[byte][]metadata.Transaction
	StateDB() *statedb.StateDB
	BCHeightBreakPointPrivacyV2() uint64
	Pdexv3BreakPoint() uint64
	Reward() uint64
	IsInscriptionFeatureEnabled() bool
}

func (env *stateEnvironment) ContributionActions() [][]string {
	return env.contributionActions
}

func (env *stateEnvironment) PRVRequiredContributionActions() [][]string {
	return env.prvRequiredContributionActions
}

func (env *stateEnvironment) TradeActions() [][]string {
	return env.tradeActions
}

func (env *stateEnvironment) CrossPoolTradeActions() [][]string {
	return env.crossPoolTradeActions
}

func (env *stateEnvironment) WithdrawalActions() [][]string {
	return env.withdrawalActions
}

func (env *stateEnvironment) FeeWithdrawalActions() [][]string {
	return env.feeWithdrawalActions
}

func (env *stateEnvironment) PrevBeaconHeight() uint64 {
	return env.prevBeaconHeight
}

func (env *stateEnvironment) ListTxs() map[byte][]metadata.Transaction {
	return env.listTxs
}

func (env *stateEnvironment) BeaconInstructions() [][]string {
	return env.beaconInstructions
}

func (env *stateEnvironment) StateDB() *statedb.StateDB {
	return env.stateDB
}

func (env *stateEnvironment) BCHeightBreakPointPrivacyV2() uint64 {
	return env.bcHeightBreakPointPrivacyV2
}

func (env *stateEnvironment) Pdexv3BreakPoint() uint64 {
	return env.pdexv3BreakPoint
}

func (env *stateEnvironment) Reward() uint64 {
	return env.reward
}

func (env *stateEnvironment) IsInscriptionFeatureEnabled() bool {
	return env.isInscriptionEnabled
}
