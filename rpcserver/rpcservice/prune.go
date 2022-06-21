package rpcservice

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/pruner"
)

type PruneData struct {
	ShouldPruneByHash bool `json:"ShouldPruneByHash"`
}

func (blockService BlockService) Prune(pruneData map[byte]PruneData, p *pruner.Pruner) (interface{}, *pruner.Pruner, error) {
	for shardID, data := range pruneData {
		if int(shardID) > config.Param().ActiveShards {
			return nil, nil, fmt.Errorf("shardID is %v is invalid", shardID)
		}
		status := byte(rawdbv2.WaitingPruneByHeightStatus)
		if data.ShouldPruneByHash {
			status = byte(rawdbv2.WaitingPruneByHashStatus)
		}
		err := rawdbv2.StorePruneStatus(blockService.BlockChain.GetShardChainDatabase(shardID), shardID, status)
		if err != nil {
			return nil, nil, err
		}
		p.StatusMu.Lock()
		p.Statuses[int(shardID)] = status
		p.StatusMu.Unlock()
	}
	type Result struct {
		Message string `json:"Message"`
	}
	return Result{Message: "Success"}, p, nil
}
