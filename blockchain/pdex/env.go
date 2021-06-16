package pdex

import "github.com/incognitochain/incognito-chain/common"

type StateEnvBuilder interface {
	BuildShardInstructions(instructions [][]string) StateEnvBuilder
	BuildTxHashes(txHashes map[byte][]common.Hash) StateEnvBuilder
	Build() StateEnvironment
}

func NewStateEnvBuilder() StateEnvBuilder {
	return &stateEnvironment{}
}

type StateEnvironment interface {
	ShardInstructions() [][]string
	TxHashes() map[byte][]common.Hash
}

type stateEnvironment struct {
	shardInstructions [][]string
	txHashes          map[byte][]common.Hash
}

func (env *stateEnvironment) BuildShardInstructions(instructions [][]string) StateEnvBuilder {
	env.shardInstructions = instructions
	return env
}

func (env *stateEnvironment) BuildTxHashes(txHashes map[byte][]common.Hash) StateEnvBuilder {
	env.txHashes = txHashes
	return env
}

func (env *stateEnvironment) Build() StateEnvironment {
	return env
}

func (env *stateEnvironment) ShardInstructions() [][]string {
	return env.shardInstructions
}

func (env *stateEnvironment) TxHashes() map[byte][]common.Hash {
	return env.txHashes
}
