package jsonresult

import (
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/config"
)

type GetMiningInfoResult struct {
	ShardHeight         uint64 `json:"ShardHeight"`
	BeaconHeight        uint64 `json:"BeaconHeight"`
	CurrentShardBlockTx int    `json:"CurrentShardBlockTx"`
	PoolSize            int    `json:"PoolSize"`
	Chain               string `json:"Chain"`
	ShardID             int    `json:"ShardID"`
	Layer               string `json:"Layer"`
	Role                string `json:"Role"`
	MiningPublickey     string `json:"MiningPublickey"`
	IsEnableMining      bool   `json:"IsEnableMining"`
}

func NewGetMiningInfoResult(
	txMemPoolSize int,
	blChain blockchain.BlockChain,
	consensus blockchain.ConsensusEngine,
) *GetMiningInfoResult {
	result := &GetMiningInfoResult{}
	result.PoolSize = txMemPoolSize
	result.Chain = config.Param().Name
	result.IsEnableMining = config.Config().EnableMining
	result.BeaconHeight = blChain.GetBeaconBestState().BeaconHeight

	// role, shardID := httpServer.config.BlockChain.BestState.Beacon.GetPubkeyRole(httpServer.config.MiningPubKeyB58, 0)
	layer, role, shardID := consensus.GetUserRole()
	result.Role = role
	result.Layer = layer
	result.ShardID = shardID
	result.MiningPublickey, _ = consensus.GetCurrentMiningPublicKey()
	if shardID >= 0 {
		result.ShardHeight = blChain.GetBestStateShard(byte(shardID)).ShardHeight
		result.CurrentShardBlockTx = len(blChain.GetBestStateShard(byte(shardID)).BestBlock.Body.Transactions)
	}

	return result
}
