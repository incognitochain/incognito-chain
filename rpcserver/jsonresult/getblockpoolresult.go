package jsonresult

import "github.com/incognitochain/incognito-chain/mempool"

type Blocks struct {
	Pending []uint64 `json:"Pending"`
	Valid   []uint64 `json:"Valid"`
	Latest  uint64   `json:"Latest"`
}

func NewBlocksFromShardPool(shardPool *mempool.ShardPool) *Blocks {
	temp := &Blocks{Valid: shardPool.GetValidBlockHeight(), Pending: shardPool.GetPendingBlockHeight(), Latest: shardPool.GetShardState()}
	return temp
}

func NewBlocksFromBeaconPool(beaconPool *mempool.BeaconPool) *Blocks {
	temp := &Blocks{Valid: beaconPool.GetValidBlockHeight(), Pending: beaconPool.GetPendingBlockHeight(), Latest: beaconPool.GetBeaconState()}
	return temp
}

type BlockHeights struct {
	ShardID         byte     `json:"ShardID"`
	BlockHeightList []uint64 `json:"BlockHeightList"`
}

type CrossShardPoolResult struct {
	PendingBlockHeight []BlockHeights `json:"InexecutableBlock"`
	ValidBlockHeight   []BlockHeights `json:"ExecutableBlock"`
}

func NewCrossShardPoolResult(allValidBlockHeight map[byte][]uint64, allPendingBlockHeight map[byte][]uint64) *CrossShardPoolResult {
	var index = 0
	crossShardPoolResult := CrossShardPoolResult{}
	crossShardPoolResult.ValidBlockHeight = make([]BlockHeights, len(allValidBlockHeight))
	crossShardPoolResult.PendingBlockHeight = make([]BlockHeights, len(allPendingBlockHeight))
	index = 0
	for shardID, blockHeights := range allValidBlockHeight {
		crossShardPoolResult.ValidBlockHeight[index].ShardID = shardID
		crossShardPoolResult.ValidBlockHeight[index].BlockHeightList = blockHeights
	}
	index = 0
	for shardID, blockHeights := range allPendingBlockHeight {
		crossShardPoolResult.PendingBlockHeight[index].ShardID = shardID
		crossShardPoolResult.PendingBlockHeight[index].BlockHeightList = blockHeights
	}
	return &crossShardPoolResult
}

type ShardToBeaconPoolResult struct {
	PendingBlockHeight []BlockHeights `json:"InexecutableBlock"`
	ValidBlockHeight   []BlockHeights `json:"ExecutableBlock"`
}

func NewShardToBeaconPoolResult(allBlockHeight map[byte][]uint64, allLatestBlockHeight map[byte]uint64) *ShardToBeaconPoolResult {
	shardToBeaconPoolResult := &ShardToBeaconPoolResult{}
	shardToBeaconPoolResult.ValidBlockHeight = make([]BlockHeights, len(allBlockHeight))
	shardToBeaconPoolResult.PendingBlockHeight = make([]BlockHeights, len(allBlockHeight))
	index := 0
	for shardID, blockHeights := range allBlockHeight {
		latestBlockHeight := allLatestBlockHeight[shardID]
		for _, blockHeight := range blockHeights {
			if blockHeight <= latestBlockHeight {
				shardToBeaconPoolResult.ValidBlockHeight[index].BlockHeightList = append(shardToBeaconPoolResult.ValidBlockHeight[index].BlockHeightList, blockHeight)
				shardToBeaconPoolResult.ValidBlockHeight[index].ShardID = shardID
			} else {
				shardToBeaconPoolResult.PendingBlockHeight[index].BlockHeightList = append(shardToBeaconPoolResult.PendingBlockHeight[index].BlockHeightList, blockHeight)
				shardToBeaconPoolResult.PendingBlockHeight[index].ShardID = shardID
			}
		}
		index++
	}
	return shardToBeaconPoolResult
}

type ShardBlockPoolResult struct {
	ShardID            byte     `json:"ShardID"`
	ValidBlockHeight   []uint64 `json:"ExecutableBlock"`
	PendingBlockHeight []uint64 `json:"InexecutableBlock"`
}

type BeaconBlockPoolResult struct {
	ValidBlockHeight   []uint64 `json:"ExecutableBlock"`
	PendingBlockHeight []uint64 `json:"InexecutableBlock"`
}
