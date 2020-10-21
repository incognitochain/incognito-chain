package mock

import (
	"context"

	"github.com/incognitochain/incognito-chain/common"
)

type Syncker struct{}

func (s *Syncker) GetCrossShardBlocksForShardProducer(toShard byte, list map[byte][]uint64) map[byte][]interface{} {
	var result map[byte][]interface{}
	result = make(map[byte][]interface{})
	return result
}

func (s *Syncker) GetCrossShardBlocksForShardValidator(toShard byte, list map[byte][]uint64) (map[byte][]interface{}, error) {
	var result map[byte][]interface{}
	result = make(map[byte][]interface{})
	return result, nil
}

func (s *Syncker) SyncMissingBeaconBlock(ctx context.Context, peerID string, fromHash common.Hash) {
	return
}

func (s *Syncker) SyncMissingShardBlock(ctx context.Context, peerID string, sid byte, fromHash common.Hash) {
	return
}
