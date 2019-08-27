package jsonresult

import (
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
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

func NewGetMiningInfoResult(txMemPool mempool.TxPool, blChain blockchain.BlockChain, miningPubKeyB58 string, param blockchain.Params, isEnableMining bool) *GetMiningInfoResult {
	result := &GetMiningInfoResult{}
	result.IsCommittee = true
	result.PoolSize = txMemPool.Count()
	result.Chain = param.Name
	result.IsEnableMining = isEnableMining
	result.BeaconHeight = blChain.BestState.Beacon.BeaconHeight

	role, shardID := blChain.BestState.Beacon.GetPubkeyRole(miningPubKeyB58, 0)
	result.Role = role
	if role == common.SHARD_ROLE {
		result.ShardHeight = blChain.BestState.Shard[shardID].ShardHeight
		result.CurrentShardBlockTx = len(blChain.BestState.Shard[shardID].BestBlock.Body.Transactions)
		result.ShardID = int(shardID)
	} else if role == common.VALIDATOR_ROLE || role == common.PROPOSER_ROLE || role == common.PENDING_ROLE {
		result.ShardID = -1
	}
	if role == common.EmptyString {
		result.IsCommittee = false
	}
	return result
}
