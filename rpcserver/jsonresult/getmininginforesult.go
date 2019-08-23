package jsonresult

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/rpcserver"
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

func NewGetMiningInfoResult(config rpcserver.RpcServerConfig) *GetMiningInfoResult {
	result := &GetMiningInfoResult{}
	result.IsCommittee = true
	result.PoolSize = config.TxMemPool.Count()
	result.Chain = config.ChainParams.Name
	result.IsEnableMining = config.Server.IsEnableMining()
	result.BeaconHeight = config.BlockChain.BestState.Beacon.BeaconHeight

	role, shardID := config.BlockChain.BestState.Beacon.GetPubkeyRole(config.MiningPubKeyB58, 0)
	result.Role = role
	if role == common.SHARD_ROLE {
		result.ShardHeight = config.BlockChain.BestState.Shard[shardID].ShardHeight
		result.CurrentShardBlockTx = len(config.BlockChain.BestState.Shard[shardID].BestBlock.Body.Transactions)
		result.ShardID = int(shardID)
	} else if role == common.VALIDATOR_ROLE || role == common.PROPOSER_ROLE || role == common.PENDING_ROLE {
		result.ShardID = -1
	}
	if role == common.EmptyString {
		result.IsCommittee = false
	}
	return result
}
