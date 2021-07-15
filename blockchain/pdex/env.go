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
	BuildModifyParamsActions([][]string) StateEnvBuilder
	BuildBeaconHeight(uint64) StateEnvBuilder
	BuildTransactions([]metadata.Transaction) StateEnvBuilder
	BuildBeaconInstructions([][]string) StateEnvBuilder
	BuildStateDB(*statedb.StateDB) StateEnvBuilder
	BuildBCHeightBreakPointPrivacyV2(uint64) StateEnvBuilder
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
	modifyParamsActions            [][]string
	beaconHeight                   uint64
	beaconInstructions             [][]string
	transactions                   []metadata.Transaction
	stateDB                        *statedb.StateDB
	bcHeightBreakPointPrivacyV2    uint64
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

func (env *stateEnvironment) BuildModifyParamsActions(actions [][]string) StateEnvBuilder {
	env.modifyParamsActions = actions
	return env
}

func (env *stateEnvironment) BuildBeaconHeight(beaconHeight uint64) StateEnvBuilder {
	env.beaconHeight = beaconHeight
	return env
}

func (env *stateEnvironment) BuildBeaconInstructions(beaconInstructions [][]string) StateEnvBuilder {
	env.beaconInstructions = beaconInstructions
	return env
}

func (env *stateEnvironment) BuildTransactions(transactions []metadata.Transaction) StateEnvBuilder {
	env.transactions = transactions
	return env
}

func (env *stateEnvironment) BuildStateDB(stateDB *statedb.StateDB) StateEnvBuilder {
	env.stateDB = stateDB
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
	ModifyParamsActions() [][]string
	BeaconHeight() uint64
	BeaconInstructions() [][]string
	Transactions() []metadata.Transaction
	StateDB() *statedb.StateDB
	BCHeightBreakPointPrivacyV2() uint64
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

func (env *stateEnvironment) ModifyParamsActions() [][]string {
	return env.modifyParamsActions
}

func (env *stateEnvironment) BeaconHeight() uint64 {
	return env.beaconHeight
}

func (env *stateEnvironment) Transactions() []metadata.Transaction {
	return env.transactions
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
