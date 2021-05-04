package mock

import (
	"context"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/syncker"
)

type Syncker struct {
	Syncker *syncker.SynckerManager
}

func (s *Syncker) InsertCrossShardBlock(blk *types.CrossShardBlock) {
	s.Syncker.InsertCrossShardBlock(blk)
}

func (s *Syncker) Init(config *syncker.SynckerManagerConfig) {
	s.Syncker = syncker.NewSynckerManager()
	s.Syncker.Init(&syncker.SynckerManagerConfig{
		Network:    config.Network,
		Blockchain: config.Blockchain,
		Consensus:  config.Consensus,
	})
	return
}

func (s *Syncker) GetCrossShardBlocksForShardProducer(view *blockchain.ShardBestState, limit map[byte][]uint64) map[byte][]interface{} {
	return s.Syncker.GetCrossShardBlocksForShardProducer(view, limit)
}

func (s *Syncker) GetCrossShardBlocksForShardValidator(view *blockchain.ShardBestState, list map[byte][]uint64) (map[byte][]interface{}, error) {
	return s.Syncker.GetCrossShardBlocksForShardProducer(view, list), nil
}

func (s *Syncker) SyncMissingBeaconBlock(ctx context.Context, peerID string, fromHash common.Hash) {
	return
}

func (s *Syncker) SyncMissingShardBlock(ctx context.Context, peerID string, sid byte, fromHash common.Hash) {
	return
}
