package bridgeagg

import (
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
)

type StateEnvBuilder interface {
	BuildAccumulatedValues(*metadata.AccumulatedValues) StateEnvBuilder
	BuildUnshieldActions([][]string) StateEnvBuilder
	BuildShieldActions([][]string) StateEnvBuilder
	BuildConvertActions([][]string) StateEnvBuilder
	BuildModifyRewardReserveActions([][]string) StateEnvBuilder
	BuildStateDBs(map[int]*statedb.StateDB) StateEnvBuilder
	BuildBeaconHeight(uint64) StateEnvBuilder
	Build() StateEnvironment
}

func NewStateEnvBuilder() StateEnvBuilder {
	return &stateEnvironment{}
}

type stateEnvironment struct {
	accumulatedValues          *metadata.AccumulatedValues
	unshieldActions            [][]string
	shieldActions              [][]string
	convertActions             [][]string
	modifyRewardReserveActions [][]string
	beaconHeight               uint64
	stateDBs                   map[int]*statedb.StateDB
}

func (env *stateEnvironment) BuildBeaconHeight(beaconHeight uint64) StateEnvBuilder {
	env.beaconHeight = beaconHeight
	return env
}

func (env *stateEnvironment) BuildAccumulatedValues(accumulatedValues *metadata.AccumulatedValues) StateEnvBuilder {
	env.accumulatedValues = accumulatedValues
	return env
}

func (env *stateEnvironment) BuildUnshieldActions(actions [][]string) StateEnvBuilder {
	env.unshieldActions = actions
	return env
}

func (env *stateEnvironment) BuildShieldActions(actions [][]string) StateEnvBuilder {
	env.shieldActions = actions
	return env
}

func (env *stateEnvironment) BuildConvertActions(actions [][]string) StateEnvBuilder {
	env.convertActions = actions
	return env
}

func (env *stateEnvironment) BuildModifyRewardReserveActions(actions [][]string) StateEnvBuilder {
	env.modifyRewardReserveActions = actions
	return env
}

func (env *stateEnvironment) BuildStateDBs(stateDBs map[int]*statedb.StateDB) StateEnvBuilder {
	env.stateDBs = stateDBs
	return env
}

func (env *stateEnvironment) Build() StateEnvironment {
	return env
}

type StateEnvironment interface {
	BeaconHeight() uint64
	AccumulatedValues() *metadata.AccumulatedValues
	UnshieldActions() [][]string
	ShieldActions() [][]string
	ConvertActions() [][]string
	StateDBs() map[int]*statedb.StateDB
}

func (env *stateEnvironment) BeaconHeight() uint64 {
	return env.beaconHeight
}

func (env *stateEnvironment) AccumulatedValues() *metadata.AccumulatedValues {
	return env.accumulatedValues
}

func (env *stateEnvironment) UnshieldActions() [][]string {
	return env.unshieldActions
}

func (env *stateEnvironment) ShieldActions() [][]string {
	return env.shieldActions
}

func (env *stateEnvironment) ConvertActions() [][]string {
	return env.convertActions
}

func (env *stateEnvironment) ModifyRewardReserveActions() [][]string {
	return env.modifyRewardReserveActions
}

func (env *stateEnvironment) StateDBs() map[int]*statedb.StateDB {
	return env.stateDBs
}
