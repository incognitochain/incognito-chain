package mempool

import (
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
	"sync"
	"testing"
	"time"
)

var (
	shardToBeaconPoolTest *ShardToBeaconPool
	shardToBeaconBlock2   = &blockchain.ShardToBeaconBlock{
		Header: blockchain.ShardHeader{
			ShardID:   0,
			Height:    2,
			Timestamp: time.Now().Unix() - 100,
		},
	}
	shardToBeaconBlock2Forked = &blockchain.ShardToBeaconBlock{
		Header: blockchain.ShardHeader{
			ShardID:   0,
			Height:    2,
			Timestamp: time.Now().Unix(),
		},
	}
	shardToBeaconBlock3 = &blockchain.ShardToBeaconBlock{
		Header: blockchain.ShardHeader{
			ShardID:       0,
			Height:        3,
			PrevBlockHash: shardToBeaconBlock2.Header.Hash(),
		},
	}
	shardToBeaconBlock3Forked = &blockchain.ShardToBeaconBlock{
		Header: blockchain.ShardHeader{
			ShardID:       0,
			Height:        3,
			PrevBlockHash: shardToBeaconBlock2Forked.Header.Hash(),
		},
	}
	shardToBeaconBlock4 = &blockchain.ShardToBeaconBlock{
		Header: blockchain.ShardHeader{
			ShardID:       0,
			Height:        4,
			PrevBlockHash: shardToBeaconBlock3.Header.Hash(),
		},
	}
	shardToBeaconBlock5 = &blockchain.ShardToBeaconBlock{
		Header: blockchain.ShardHeader{
			ShardID:       0,
			Height:        5,
			PrevBlockHash: shardToBeaconBlock4.Header.Hash(),
		},
	}
	shardToBeaconBlock6 = &blockchain.ShardToBeaconBlock{
		Header: blockchain.ShardHeader{
			ShardID:       0,
			Height:        6,
			PrevBlockHash: shardToBeaconBlock5.Header.Hash(),
		},
	}
	shardToBeaconBlock7 = &blockchain.ShardToBeaconBlock{
		Header: blockchain.ShardHeader{
			ShardID:       0,
			Height:        7,
			PrevBlockHash: shardToBeaconBlock6.Header.Hash(),
		},
	}
	validShardToBeaconBlocks   = []*blockchain.ShardToBeaconBlock{}
	pendingShardToBeaconBlocks = []*blockchain.ShardToBeaconBlock{}
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
	for i := 0; i < 255; i++ {
		shardID := byte(i)
		bestShardHeight[shardID] = 1
		blockchain.SetBestStateShard(shardID, &blockchain.BestStateShard{
			ShardHeight: 1,
		})
	}
	blockchain.SetBestStateBeacon(&blockchain.BestStateBeacon{
		BestShardHeight: bestShardHeight,
	})
	InitShardToBeaconPool()
	InitShardToBeaconPoolTest()
	oldBlockHash := common.Hash{}
	for i := 1; i < MAX_VALID_SHARD_TO_BEACON_BLK_IN_POOL+2; i++ {
		shardToBeaconBlock := &blockchain.ShardToBeaconBlock{
			Header: blockchain.ShardHeader{
				ShardID: 0,
				Height:  uint64(i),
			},
		}
		if i != 0 {
			shardToBeaconBlock.Header.PrevBlockHash = oldBlockHash
		}
		oldBlockHash = shardToBeaconBlock.Header.Hash()
		validShardToBeaconBlocks = append(validShardToBeaconBlocks, shardToBeaconBlock)
	}
	for i := MAX_VALID_SHARD_TO_BEACON_BLK_IN_POOL + 2; i < MAX_VALID_SHARD_TO_BEACON_BLK_IN_POOL+MAX_INVALID_SHARD_TO_BEACON_BLK_IN_POOL+3; i++ {
		shardToBeaconBlock := &blockchain.ShardToBeaconBlock{
			Header: blockchain.ShardHeader{
				ShardID: 0,
				Height:  uint64(i),
			},
		}
		if i != 0 {
			shardToBeaconBlock.Header.PrevBlockHash = oldBlockHash
		}
		oldBlockHash = shardToBeaconBlock.Header.Hash()
		pendingShardToBeaconBlocks = append(pendingShardToBeaconBlocks, shardToBeaconBlock)
	}
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()

func TestShardToBeaconPoolUpdateLatestShardState(t *testing.T) {
	InitShardToBeaconPoolTest()
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
	InitShardToBeaconPoolTest()
	if pool, isOk := shardToBeaconPoolTest.pool[0]; isOk {
		pool = append(pool, shardToBeaconBlock2)
		pool = append(pool, shardToBeaconBlock3)
		pool = append(pool, shardToBeaconBlock4)
		pool = append(pool, shardToBeaconBlock6)
		shardToBeaconPoolTest.pool[0] = pool
		lastHeight := make(map[byte]uint64)
		lastHeight[0] = 6
		shardToBeaconPoolTest.removeBlock(lastHeight)
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
		shardToBeaconPoolTest.removeBlock(lastHeight)
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
	InitShardToBeaconPoolTest()
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
func TestShardToBeaconPoolGetShardState(t *testing.T) {
	InitShardToBeaconPoolTest()
	if !reflect.DeepEqual(shardToBeaconPoolTest.GetShardState(), shardToBeaconPoolTest.latestValidHeight) {
		t.Fatalf("Get Shard State has something wrong!!!")
	}
}
func TestShardToBeaconPoolCheckLatestValidHeightValidity(t *testing.T) {
	InitShardToBeaconPoolTest()
	for shardID, _ := range shardToBeaconPoolTest.latestValidHeight {
		shardToBeaconPoolTest.latestValidHeight[shardID] = 0
	}
	for i := 0; i < 255; i++ {
		shardID := byte(i)
		if shardToBeaconPoolTest.latestValidHeight[shardID] != 0 {
			t.Fatalf("Expect latestvalidheight %+v from %+v but get %+v", 0, shardID, shardToBeaconPoolTest.latestValidHeight[shardID])
		}
		shardToBeaconPoolTest.checkLatestValidHeightValidity(shardID)
		if shardToBeaconPoolTest.latestValidHeight[shardID] != 1 {
			t.Fatalf("Expect latestvalidheight %+v from %+v but get %+v", 1, shardID, shardToBeaconPoolTest.latestValidHeight[shardID])
		}
	}
}
func TestShardToBeaconPoolAddShardToBeaconBlock(t *testing.T) {
	InitShardToBeaconPoolTest()
	// check old block
	shardToBeaconPoolTest.latestValidHeight[0] = 2
	_, _, err = shardToBeaconPoolTest.AddShardToBeaconBlock(shardToBeaconBlock2)
	if err == nil {
		t.Fatalf("Expect old block error but no error")
	} else {
		if err.(*BlockPoolError).Code != ErrCodeMessage[OldBlockError].Code {
			t.Fatalf("Expect error %+v but get %+v", ErrCodeMessage[OldBlockError], err.(*BlockPoolError))
		}
	}
	shardToBeaconPoolTest.latestValidHeight[0] = 3
	_, _, err = shardToBeaconPoolTest.AddShardToBeaconBlock(shardToBeaconBlock2)
	if err == nil {
		t.Fatal("Expect old block error but no error")
	} else {
		if err.(*BlockPoolError).Code != ErrCodeMessage[OldBlockError].Code {
			t.Fatalf("Expect error %+v but get %+v", ErrCodeMessage[OldBlockError], err.(*BlockPoolError))
		}
	}
	_, _, err := shardToBeaconPoolTest.AddShardToBeaconBlock(shardToBeaconBlock3)
	if err == nil {
		t.Fatalf("Expect old block error but no error")
	} else {
		if err.(*BlockPoolError).Code != ErrCodeMessage[OldBlockError].Code {
			t.Fatalf("Expect error %+v but get %+v", ErrCodeMessage[OldBlockError], err.(*BlockPoolError))
		}
	}
	InitShardToBeaconPoolTest()
	// check duplicate block
	shardToBeaconPoolTest.pool[0] = append(shardToBeaconPoolTest.pool[0], shardToBeaconBlock2)
	_, _, err = shardToBeaconPoolTest.AddShardToBeaconBlock(shardToBeaconBlock2Forked)
	if err == nil {
		t.Fatal("Expect duplicate block error but no error")
	} else {
		if err.(*BlockPoolError).Code != ErrCodeMessage[DuplicateBlockError].Code {
			t.Fatalf("Expect error %+v but get %+v", ErrCodeMessage[DuplicateBlockError], err.(*BlockPoolError))
		}
	}
	InitShardToBeaconPoolTest()
	// check valid pool capacity
	for index, block := range validShardToBeaconBlocks {
		if index < len(validShardToBeaconBlocks)-1 {
			shardToBeaconPoolTest.pool[0] = append(shardToBeaconPoolTest.pool[0], block)
			shardToBeaconPoolTest.latestValidHeight[0] = block.Header.Height
		} else {
			_, _, err = shardToBeaconPoolTest.AddShardToBeaconBlock(block)
			if err == nil {
				t.Fatal("Expect max pool size error but no error")
			} else {
				if err.(*BlockPoolError).Code != ErrCodeMessage[MaxPoolSizeError].Code {
					t.Fatalf("Expect error %+v but get %+v", ErrCodeMessage[MaxPoolSizeError], err.(*BlockPoolError))
				}
			}
		}
	}
	InitShardToBeaconPoolTest()
	// check pending pool capacity
	for index, block := range validShardToBeaconBlocks {
		if index < len(validShardToBeaconBlocks)-2 {
			shardToBeaconPoolTest.pool[0] = append(shardToBeaconPoolTest.pool[0], block)
			shardToBeaconPoolTest.latestValidHeight[0] = block.Header.Height
		}
	}
	for index, block := range pendingShardToBeaconBlocks {
		if index < len(pendingShardToBeaconBlocks)-1 {
			shardToBeaconPoolTest.pool[0] = append(shardToBeaconPoolTest.pool[0], block)
		}
	}
	_, _, err = shardToBeaconPoolTest.AddShardToBeaconBlock(pendingShardToBeaconBlocks[len(pendingShardToBeaconBlocks)-1])
	if err == nil {
		t.Fatal("Expect max pool size error but no error")
	} else {
		if err.(*BlockPoolError).Code != ErrCodeMessage[MaxPoolSizeError].Code {
			t.Fatalf("Expect error %+v but get %+v", ErrCodeMessage[MaxPoolSizeError], err.(*BlockPoolError))
		}
	}
	InitShardToBeaconPoolTest()
	// although 2 valid and pending == invalid pool is max, but if better is found then replace it
	for index, block := range validShardToBeaconBlocks {
		if index < len(validShardToBeaconBlocks)-2 {
			shardToBeaconPoolTest.pool[0] = append(shardToBeaconPoolTest.pool[0], block)
			shardToBeaconPoolTest.latestValidHeight[0] = block.Header.Height
		}
	}
	for index, block := range pendingShardToBeaconBlocks {
		if index < len(pendingShardToBeaconBlocks)-1 {
			shardToBeaconPoolTest.pool[0] = append(shardToBeaconPoolTest.pool[0], block)
		}
	}
	_, _, err = shardToBeaconPoolTest.AddShardToBeaconBlock(validShardToBeaconBlocks[len(validShardToBeaconBlocks)-1])
	if err != nil {
		t.Fatalf("Expect no error but get %+v", err)
	}
	InitShardToBeaconPoolTest()
	// block insert to pool will be reorder properly
	shardToBeaconPoolTest.pool[0] = append(shardToBeaconPoolTest.pool[0], shardToBeaconBlock2)
	shardToBeaconPoolTest.pool[0] = append(shardToBeaconPoolTest.pool[0], shardToBeaconBlock3)
	shardToBeaconPoolTest.pool[0] = append(shardToBeaconPoolTest.pool[0], shardToBeaconBlock5)
	_, _, err = shardToBeaconPoolTest.AddShardToBeaconBlock(shardToBeaconBlock4)
	if err != nil {
		t.Fatalf("Expect no error but get %+v", err)
	}
	if len(shardToBeaconPoolTest.pool[0]) != 4 {
		t.Fatalf("Expect pool to have 4 block but get %+v", len(shardToBeaconPoolTest.pool[0]))
	}
	for index, block := range shardToBeaconPoolTest.pool[0] {
		switch index {
		case 0:
			if block.Header.Height != 2 {
				t.Fatalf("Expect block 2 but get %+v ", block.Header.Height)
			}
		case 1:
			if block.Header.Height != 3 {
				t.Fatalf("Expect block 3 but get %+v ", block.Header.Height)
			}
		case 2:
			if block.Header.Height != 4 {
				t.Fatalf("Expect block 4 but get %+v ", block.Header.Height)
			}
		case 3:
			if block.Header.Height != 5 {
				t.Fatalf("Expect block 5 but get %+v ", block.Header.Height)
			}
		}
	}
}
func TestShardToBeaconPoolGetValidBlock(t *testing.T) {
	InitShardToBeaconPoolTest()
	shardToBeaconPoolTest.AddShardToBeaconBlock(shardToBeaconBlock2)
	shardToBeaconPoolTest.AddShardToBeaconBlock(shardToBeaconBlock3)
	shardToBeaconPoolTest.AddShardToBeaconBlock(shardToBeaconBlock4)
	shardToBeaconPoolTest.AddShardToBeaconBlock(shardToBeaconBlock5)

	shardToBeaconPoolTest.AddShardToBeaconBlock(shardToBeaconBlock7)
	limit := make(map[byte]uint64)
	limit[0] = 7
	blocks := shardToBeaconPoolTest.GetValidBlock(limit)
	if len(blocks[0]) != 4 {
		t.Fatalf("Expect pool to have 4 block but get %+v", len(blocks[0]))
	}
	for index, block := range blocks[0] {
		switch index {
		case 0:
			if block.Header.Height != 2 {
				t.Fatalf("Expect block 2 but get %+v ", block.Header.Height)
			}
		case 1:
			if block.Header.Height != 3 {
				t.Fatalf("Expect block 3 but get %+v ", block.Header.Height)
			}
		case 2:
			if block.Header.Height != 4 {
				t.Fatalf("Expect block 4 but get %+v ", block.Header.Height)
			}
		case 3:
			if block.Header.Height != 5 {
				t.Fatalf("Expect block 5 but get %+v ", block.Header.Height)
			}
		}
	}
	limit[0] = 4
	blocks = shardToBeaconPoolTest.GetValidBlock(limit)
	if len(blocks[0]) != 3 {
		t.Fatalf("Expect pool to have 3 block but get %+v", len(blocks[0]))
	}
	for index, block := range blocks[0] {
		switch index {
		case 0:
			if block.Header.Height != 2 {
				t.Fatalf("Expect block 2 but get %+v ", block.Header.Height)
			}
		case 1:
			if block.Header.Height != 3 {
				t.Fatalf("Expect block 3 but get %+v ", block.Header.Height)
			}
		case 2:
			if block.Header.Height != 4 {
				t.Fatalf("Expect block 4 but get %+v ", block.Header.Height)
			}
		}
	}
}
