package rpcservice

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/pruner"
)

func (blockService BlockService) Prune(shardIDs []byte, p *pruner.Pruner) (interface{}, *pruner.Pruner, error) {
	for _, shardID := range shardIDs {
		if int(shardID) > config.Param().ActiveShards {
			return nil, nil, fmt.Errorf("shardID is %v is invalid", shardID)
		}
		err := rawdbv2.StorePruneStatus(blockService.BlockChain.GetShardChainDatabase(shardID), shardID, rawdbv2.WaitingPruneStatus)
		if err != nil {
			return nil, nil, err
		}
		p.StatusMu.Lock()
		p.Statuses[int(shardID)] = rawdbv2.WaitingPruneStatus
		p.StatusMu.Unlock()
	}
	type Result struct {
		Message string `json:"Message"`
	}
	return Result{Message: "Success"}, p, nil
}
