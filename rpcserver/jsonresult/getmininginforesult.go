package jsonresult

import (
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/mempool"
)

type GetMiningInfoResult struct {
	ShardHeight         uint64 `json:"ShardHeight"`
	BeaconHeight        uint64 `json:"BeaconHeight"`
	CurrentShardBlockTx int    `json:"CurrentShardBlockTx"`
	PoolSize            int    `json:"PoolSize"`
	Chain               string `json:"Chain"`
	IsCommittee         bool   `json:"IsCommittee"`
	ShardID             int    `json:"ShardID"`
	Role                string `json:"Role"`
	IsEnableMining      bool   `json:"IsEnableMining"`
}

func NewGetMiningInfoResult(txMemPool mempool.TxPool, blChain blockchain.BlockChain, consensus interface{ GetUserRole() (string, int) }, param blockchain.Params, isMining bool) *GetMiningInfoResult {
	result := &GetMiningInfoResult{}
	result.IsCommittee = true
	result.PoolSize = txMemPool.Count()
	result.Chain = param.Name
	result.IsEnableMining = isEnableMining
	result.BeaconHeight = blChain.BestState.Beacon.BeaconHeight

	// role, shardID := httpServer.config.BlockChain.BestState.Beacon.GetPubkeyRole(httpServer.config.MiningPubKeyB58, 0)
	role, shardID := consensus.GetUserRole()
	result.Role = role

	switch shardID {
	case -2:
		result.ShardID = -2
		result.IsCommittee = false
	case -1:
		result.IsCommittee = true
	default:
		result.ShardHeight = blChain.BestState.Shard[byte(shardID)].ShardHeight
		result.CurrentShardBlockTx = len(blChain.BestState.Shard[byte(shardID)].BestBlock.Body.Transactions)
		result.ShardID = shardID
	}

	return result
}
