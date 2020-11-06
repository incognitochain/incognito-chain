package devframework

import (
	"context"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/syncker"
)

type Execute struct {
	sim          *SimulationEngine
	appliedChain []int
}

func (exec *Execute) GenerateBlock(args ...interface{}) {
	args = append(args, exec)
	exec.sim.GenerateBlock(args...)
}

func (sim *SimulationEngine) ApplyChain(chain_array ...int) *Execute {
	return &Execute{
		sim,
		chain_array,
	}
}

type Syncker interface {
	GetCrossShardBlocksForShardProducer(toShard byte, limit map[byte][]uint64) map[byte][]interface{}
	GetCrossShardBlocksForShardValidator(toShard byte, list map[byte][]uint64) (map[byte][]interface{}, error)
	SyncMissingBeaconBlock(ctx context.Context, peerID string, fromHash common.Hash)
	SyncMissingShardBlock(ctx context.Context, peerID string, sid byte, fromHash common.Hash)
	Init(*syncker.SynckerManagerConfig)
	InsertCrossShardBlock(block *blockchain.CrossShardBlock)
}
