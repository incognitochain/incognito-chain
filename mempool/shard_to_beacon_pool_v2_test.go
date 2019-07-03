package mempool

import (
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"sync"
	"testing"
	"time"
)

var (
	shardToBeaconPoolTest *ShardToBeaconPool
	shardToBeaconBlock2           = &blockchain.ShardToBeaconBlock{
		Header: blockchain.ShardHeader{
			ShardID: 0,
			Height: 2,
			Timestamp: time.Now().Unix()-100,
		},
	}
	shardToBeaconBlock2Forked           = &blockchain.ShardToBeaconBlock{
		Header: blockchain.ShardHeader{
			ShardID: 0,
			Height: 2,
			Timestamp: time.Now().Unix(),
		},
	}
	shardToBeaconBlock3 = &blockchain.ShardToBeaconBlock{
		Header: blockchain.ShardHeader{
			ShardID: 0,
			Height: 3,
			PrevBlockHash: shardToBeaconBlock2.Header.Hash(),
		},
	}
	shardToBeaconBlock3Forked = &blockchain.ShardToBeaconBlock{
		Header: blockchain.ShardHeader{
			ShardID: 0,
			Height: 3,
			PrevBlockHash: shardToBeaconBlock2Forked.Header.Hash(),
		},
	}
	shardToBeaconBlock4 = &blockchain.ShardToBeaconBlock{
		Header: blockchain.ShardHeader{
			ShardID: 0,
			Height: 4,
			PrevBlockHash: shardToBeaconBlock3.Header.Hash(),
		},
	}
	shardToBeaconBlock5 = &blockchain.ShardToBeaconBlock{
		Header: blockchain.ShardHeader{
			ShardID: 0,
			Height: 5,
			PrevBlockHash: shardToBeaconBlock4.Header.Hash(),
		},
	}
	shardToBeaconBlock6 = &blockchain.ShardToBeaconBlock{
		Header: blockchain.ShardHeader{
			ShardID: 0,
			Height: 6,
			PrevBlockHash: shardToBeaconBlock5.Header.Hash(),
		},
	}
	shardToBeaconBlock7 = &blockchain.ShardToBeaconBlock{
		Header: blockchain.ShardHeader{
			ShardID: 0,
			Height: 7,
			PrevBlockHash: shardToBeaconBlock6.Header.Hash(),
		},
	}
	shardToBeaconBlocks = []*blockchain.ShardToBeaconBlock{}
)
var InitShardToBeaconPoolTest = func() {
	shardToBeaconPoolTest = new(ShardToBeaconPool)
	shardToBeaconPoolTest.pool = make(map[byte][]*blockchain.ShardToBeaconBlock)
	// add to pool
	for i := 0; i < 255; i++ {
		shardID := byte(i)
		if shardToBeaconPoolTest.pool[shardID] == nil {
			shardToBeaconPoolTest.pool[shardID] = []*blockchain.ShardToBeaconBlock{}
		}
	}
	shardToBeaconPoolTest.mtx = new(sync.RWMutex)
	shardToBeaconPoolTest.latestValidHeightMutex = new(sync.RWMutex)
	shardToBeaconPoolTest.latestValidHeight = make(map[byte]uint64)
}
var ResetShardToBeaconPool = func() {
	shardToBeaconPool = new(ShardToBeaconPool)
	shardToBeaconPool.pool = make(map[byte][]*blockchain.ShardToBeaconBlock)
	shardToBeaconPool.mtx = new(sync.RWMutex)
	shardToBeaconPool.latestValidHeight = make(map[byte]uint64)
	shardToBeaconPool.latestValidHeightMutex = new(sync.RWMutex)
}
var _ = func() (_ struct{}) {
	for i:=0; i < 255; i++ {
		shardID := byte(i)
		bestShardHeight[shardID] = 1
		blockchain.SetBestStateShard(shardID, &blockchain.BestStateShard{
			ShardHeight:1,
		} )
	}
	blockchain.SetBestStateBeacon(&blockchain.BestStateBeacon{
		BestShardHeight: bestShardHeight,
	})
	InitShardToBeaconPool()
	InitShardToBeaconPoolTest()
	oldBlockHash := common.Hash{}
	for i := testLatestValidHeight + 1; i < MAX_VALID_SHARD_TO_BEACON_BLK_IN_POOL + MAX_INVALID_SHARD_TO_BEACON_BLK_IN_POOL + testLatestValidHeight+2; i++ {
		shardToBeaconBlock := &blockchain.ShardToBeaconBlock{
			Header: blockchain.ShardHeader{
				ShardID: 0,
				Height: uint64(i),
			},
		}
		if i != 0 {
			shardToBeaconBlock.Header.PrevBlockHash = oldBlockHash
		}
		oldBlockHash = shardToBeaconBlock.Header.Hash()
		shardToBeaconBlocks = append(shardToBeaconBlocks, shardToBeaconBlock)
	}
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()
func TestShardToBeaconPoolUpdateLatestShardState(t *testing.T) {
	shardToBeaconPoolTest.latestValidHeight = make(map[byte]uint64)
	for _, value := range shardToBeaconPoolTest.latestValidHeight {
		if value != 0 {
			t.Fatalf("Wrong Init value")
		}
	}
	shardToBeaconPoolTest.updateLatestShardState()
	for _, value := range shardToBeaconPoolTest.latestValidHeight {
		if value != 1 {
			t.Fatalf("Expect latestvalidheight %+v but get %+v", 1, shardToBeaconPoolTest.latestValidHeight)
		}
	}
	if pool, isOk := shardToBeaconPoolTest.pool[0]; isOk {
		pool = append(pool, shardToBeaconBlock2)
		pool = append(pool, shardToBeaconBlock3)
		pool = append(pool, shardToBeaconBlock4)
		pool = append(pool, shardToBeaconBlock6)
		shardToBeaconPoolTest.pool[0] = pool
		shardToBeaconPoolTest.updateLatestShardState()
		if latestValidHeight, isOk := shardToBeaconPoolTest.latestValidHeight[0]; isOk {
			if latestValidHeight != 4 {
				t.Fatalf("Expect latestvalidheight is 4 but get %+v", latestValidHeight)
			}
		} else {
			t.Fatalf("Fail to init shard to beacon pool")
		}
	} else {
		t.Fatalf("Fail to init shard to beacon pool")
	}
}
func TestShardToBeaconPoolRemovePendingBlock(t *testing.T) {
	InitShardToBeaconPool()
	if pool, isOk := shardToBeaconPoolTest.pool[0]; isOk {
		pool = append(pool, shardToBeaconBlock2)
		pool = append(pool, shardToBeaconBlock3)
		pool = append(pool, shardToBeaconBlock4)
		pool = append(pool, shardToBeaconBlock6)
		shardToBeaconPoolTest.pool[0] = pool
		lastHeight := make(map[byte]uint64)
		lastHeight[0] = 6
		shardToBeaconPoolTest.removePendingBlock(lastHeight)
		if len(shardToBeaconPoolTest.pool[0]) != 0 {
			t.Fatalf("Expect Pool from shard 0 to have 1 block but get %+v", len(shardToBeaconPoolTest.pool[0]))
		}
		pool = shardToBeaconPoolTest.pool[0]
		pool = append(pool, shardToBeaconBlock2)
		pool = append(pool, shardToBeaconBlock3)
		pool = append(pool, shardToBeaconBlock4)
		pool = append(pool, shardToBeaconBlock6)
		shardToBeaconPoolTest.pool[0] = pool
		lastHeight[0] = 4
		shardToBeaconPoolTest.removePendingBlock(lastHeight)
		if len(shardToBeaconPoolTest.pool[0]) != 1 {
			t.Fatalf("Expect Pool from shard 0 to have 1 block but get %+v", len(shardToBeaconPoolTest.pool[0]))
		}
		if shardToBeaconPoolTest.pool[0][0].Header.Height != shardToBeaconBlock6.Header.Height {
			t.Fatalf("Expect to have block %+v but get %+v", shardToBeaconBlock6.Header.Height, shardToBeaconPoolTest.pool[0][0].Header.Height)
		}
	} else {
		t.Fatalf("Fail to init shard to beacon pool")
	}
}
func TestShardToBeaconPoolSetShardState(t *testing.T) {
	InitShardToBeaconPool()
	if pool, isOk := shardToBeaconPoolTest.pool[0]; isOk {
		pool = append(pool, shardToBeaconBlock2)
		pool = append(pool, shardToBeaconBlock3)
		pool = append(pool, shardToBeaconBlock4)
		pool = append(pool, shardToBeaconBlock5)
		pool = append(pool, shardToBeaconBlock6)
		pool = append(pool, shardToBeaconBlock7)
		shardToBeaconPoolTest.pool[0] = pool
		lastHeight := make(map[byte]uint64)
		lastHeight[0] = 5
		lastHeight[1] = 0
		shardToBeaconPoolTest.SetShardState(lastHeight)
		if shardToBeaconPoolTest.latestValidHeight[0] != 7 {
			t.Fatalf("Expect latest valid height from shard 0 is 7 but get %+v ", shardToBeaconPoolTest.latestValidHeight[0])
		}
		if len(shardToBeaconPoolTest.pool[0]) != 2 {
			t.Fatalf("Expect block in pool from shard 0 is 2 but get %+v ", len(shardToBeaconPoolTest.pool[0]))
		}
		if shardToBeaconPoolTest.pool[0][0].Header.Height != shardToBeaconBlock6.Header.Height || shardToBeaconPoolTest.pool[0][1].Header.Height != shardToBeaconBlock7.Header.Height {
			t.Fatalf("Got unpexted block from pool")
		}
	} else {
		t.Fatalf("Fail to init shard to beacon pool")
	}
}