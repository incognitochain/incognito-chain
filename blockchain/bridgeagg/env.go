package bridgeagg

import "github.com/incognitochain/incognito-chain/dataaccessobject/statedb"

type StateEnvBuilder interface {
	BuildUnshieldActions([]string) StateEnvBuilder
	BuildShieldActions([][]string) StateEnvBuilder
	BuildConvertActions([]string) StateEnvBuilder
	BuildModifyListTokenActions([]string) StateEnvBuilder
	BuildStateDBs(map[int]*statedb.StateDB) StateEnvBuilder
	Build() StateEnvironment
}

func NewStateEnvBuilder() StateEnvBuilder {
	return &stateEnvironment{}
}

type stateEnvironment struct {
	unshieldActions         []string
	shieldActions           [][]string
	convertActions          []string
	modifyListTokensActions []string
	stateDBs                map[int]*statedb.StateDB
}

func (env *stateEnvironment) BuildUnshieldActions(actions []string) StateEnvBuilder {
	env.unshieldActions = actions
	return env
}

func (env *stateEnvironment) BuildShieldActions(actions [][]string) StateEnvBuilder {
	env.shieldActions = actions
	return env
}

func (env *stateEnvironment) BuildConvertActions(actions []string) StateEnvBuilder {
	env.convertActions = actions
	return env
}

func (env *stateEnvironment) BuildModifyListTokenActions(actions []string) StateEnvBuilder {
	env.modifyListTokensActions = actions
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
	UnshieldActions() []string
	ShieldActions() [][]string
	ConvertActions() []string
	ModifyListTokensActions() []string
	StateDBs() map[int]*statedb.StateDB
}

func (env *stateEnvironment) UnshieldActions() []string {
	return env.unshieldActions
}

func (env *stateEnvironment) ShieldActions() [][]string {
	return env.shieldActions
}

func (env *stateEnvironment) ConvertActions() []string {
	return env.convertActions
}

func (env *stateEnvironment) ModifyListTokensActions() []string {
	return env.modifyListTokensActions
}

func (env *stateEnvironment) StateDBs() map[int]*statedb.StateDB {
	return env.stateDBs
}
