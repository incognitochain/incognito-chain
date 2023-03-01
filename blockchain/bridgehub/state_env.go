package bridgehub

import (
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
)

type StateEnvironment struct {
	stakeActions    [][]string
	unstakeActions  [][]string
	shieldActions   [][]string
	unshieldActions [][]string

	slashActions                [][]string
	registerBridgeActions       [][]string
	swapBridgeValidatorsActions [][]string
	shieldBridgeActions         [][]string

	beaconHeight      uint64
	accumulatedValues *metadata.AccumulatedValues
	stateDBs          map[int]*statedb.StateDB
}

func NewStateEnvironment() *StateEnvironment {
	return &StateEnvironment{}
}

func (env *StateEnvironment) SetBeaconHeight(beaconHeight uint64) *StateEnvironment {
	env.beaconHeight = beaconHeight
	return env
}

func (env *StateEnvironment) SetAccumulatedValues(accumulatedValues *metadata.AccumulatedValues) *StateEnvironment {
	env.accumulatedValues = accumulatedValues
	return env
}

func (env *StateEnvironment) SetStateDBs(stateDBs map[int]*statedb.StateDB) *StateEnvironment {
	env.stateDBs = stateDBs
	return env
}

func (env *StateEnvironment) SetUnshieldActions(actions [][]string) *StateEnvironment {
	env.unshieldActions = actions
	return env
}

func (env *StateEnvironment) SetShieldActions(actions [][]string) *StateEnvironment {
	env.shieldActions = actions
	return env
}

func (env *StateEnvironment) SetUnstakeActions(actions [][]string) *StateEnvironment {
	env.unstakeActions = actions
	return env
}

func (env *StateEnvironment) SetStakeActions(actions [][]string) *StateEnvironment {
	env.stakeActions = actions
	return env
}

func (env *StateEnvironment) SetRegisterBridgeActions(actions [][]string) *StateEnvironment {
	env.registerBridgeActions = actions
	return env
}

func (env *StateEnvironment) SetSwapBridgeValidatorsActions(actions [][]string) *StateEnvironment {
	env.swapBridgeValidatorsActions = actions
	return env
}

func (env *StateEnvironment) SetSlashActions(actions [][]string) *StateEnvironment {
	env.slashActions = actions
	return env
}

func (env StateEnvironment) BeaconHeight() uint64 {
	return env.beaconHeight
}

func (env StateEnvironment) AccumulatedValues() *metadata.AccumulatedValues {
	return env.accumulatedValues
}

func (env StateEnvironment) StateDBs() map[int]*statedb.StateDB {
	return env.stateDBs
}

func (env StateEnvironment) StakeActions() [][]string {
	return env.stakeActions
}

func (env StateEnvironment) UnstakeActions() [][]string {
	return env.unstakeActions
}

func (env StateEnvironment) UnshieldActions() [][]string {
	return env.unshieldActions
}

func (env StateEnvironment) ShieldActions() [][]string {
	return env.shieldActions
}

func (env StateEnvironment) SlashActions() [][]string {
	return env.slashActions
}

func (env StateEnvironment) RegisterBridgeActions() [][]string {
	return env.registerBridgeActions
}

func (env StateEnvironment) SwapBridgeValidatorsActions() [][]string {
	return env.swapBridgeValidatorsActions
}
