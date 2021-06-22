package pdex

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

type StateEnvBuilder interface {
	BuildContributionActions([][]string) StateEnvBuilder
	BuildPRVRequiredContributionActions([][]string) StateEnvBuilder
	BuildTradeActions([][]string) StateEnvBuilder
	BuildCrossPoolTradeActions([][]string) StateEnvBuilder
	BuildWithdrawalActions([][]string) StateEnvBuilder
	BuildFeeWithdrawalActions([][]string) StateEnvBuilder
	BuildBeaconHeight(uint64) StateEnvBuilder
	BuildTxHashes([]common.Hash) StateEnvBuilder
	BuildBeaconInstructions([][]string) StateEnvBuilder
	BuildStateDB(*statedb.StateDB) StateEnvBuilder
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
	beaconHeight                   uint64
	beaconInstructions             [][]string
	txHashes                       []common.Hash
	stateDB                        *statedb.StateDB
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

func (env *stateEnvironment) BuildBeaconHeight(beaconHeight uint64) StateEnvBuilder {
	env.beaconHeight = beaconHeight
	return env
}

func (env *stateEnvironment) BuildBeaconInstructions(beaconInstructions [][]string) StateEnvBuilder {
	env.beaconInstructions = beaconInstructions
	return env
}

func (env *stateEnvironment) BuildTxHashes(txHashes []common.Hash) StateEnvBuilder {
	env.txHashes = txHashes
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
	BeaconHeight() uint64
	BeaconInstructions() [][]string
	TxHashes() []common.Hash
	StateDB() *statedb.StateDB
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

func (env *stateEnvironment) BeaconHeight() uint64 {
	return env.beaconHeight
}

func (env *stateEnvironment) TxHashes() []common.Hash {
	return env.txHashes
}

func (env *stateEnvironment) BeaconInstructions() [][]string {
	return env.beaconInstructions
}

func (env *stateEnvironment) StateDB() *statedb.StateDB {
	return env.stateDB
}
