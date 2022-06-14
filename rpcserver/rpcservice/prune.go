package rpcservice

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
)

func (blockService BlockService) Prune(shardIDs []byte) (interface{}, error) {
	for _, shardID := range shardIDs {
		if int(shardID) > config.Param().ActiveShards {
			return nil, fmt.Errorf("shardID is %v is invalid", shardID)
		}
		err := rawdbv2.StorePruneStatus(blockService.BlockChain.GetShardChainDatabase(shardID), shardID, rawdbv2.PendingPruneStatus)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}
