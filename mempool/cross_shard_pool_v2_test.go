package mempool

import (
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

var (
	dbCrossShard          database.DatabaseInterface
	bestShardStateShard1  *blockchain.BestStateShard
	crossShardPoolMapTest = make(map[byte]*CrossShardPool_v2)
	crossShardBlock2      = &blockchain.CrossShardBlock{
		Header: blockchain.ShardHeader{
			ShardID:   0,
			Height:    2,
			Timestamp: time.Now().Unix() - 100,
		},
		ToShardID: 1,
	}
	crossShardBlock3WrongShard = &blockchain.CrossShardBlock{
		Header: blockchain.ShardHeader{
			ShardID:   0,
			Height:    3,
			Timestamp: time.Now().Unix(),
		},
		ToShardID: 0,
	}
	crossShardBlock3 = &blockchain.CrossShardBlock{
		Header: blockchain.ShardHeader{
			ShardID:       0,
			Height:        3,
			PrevBlockHash: crossShardBlock2.Header.Hash(),
		},
		ToShardID: 1,
	}
	crossShardBlock4 = &blockchain.CrossShardBlock{
		Header: blockchain.ShardHeader{
			ShardID:       0,
			Height:        4,
			PrevBlockHash: crossShardBlock3.Header.Hash(),
		},
		ToShardID: 1,
	}
	crossShardBlock5 = &blockchain.CrossShardBlock{
		Header: blockchain.ShardHeader{
			ShardID:       0,
			Height:        5,
			PrevBlockHash: crossShardBlock4.Header.Hash(),
		},
		ToShardID: 1,
	}
	crossShardBlock6 = &blockchain.CrossShardBlock{
		Header: blockchain.ShardHeader{
			ShardID:       0,
			Height:        6,
			PrevBlockHash: crossShardBlock5.Header.Hash(),
		},
		ToShardID: 1,
	}
	crossShardBlock7 = &blockchain.CrossShardBlock{
		Header: blockchain.ShardHeader{
			ShardID:       0,
			Height:        7,
			PrevBlockHash: crossShardBlock6.Header.Hash(),
		},
		ToShardID: 1,
	}
	crossShardBlock8 = &blockchain.CrossShardBlock{
		Header: blockchain.ShardHeader{
			ShardID:       0,
			Height:        8,
			PrevBlockHash: crossShardBlock6.Header.Hash(),
		},
		ToShardID: 1,
	}
	crossShardBlock9 = &blockchain.CrossShardBlock{
		Header: blockchain.ShardHeader{
			ShardID:       0,
			Height:        9,
			PrevBlockHash: crossShardBlock6.Header.Hash(),
		},
		ToShardID: 1,
	}
	pendingCrossShardBlocks = []*blockchain.CrossShardBlock{}
	validCrossShardBlocks   = []*blockchain.CrossShardBlock{}
)
var _ = func() (_ struct{}) {
	for i := 0; i < 255; i++ {
		shardID := byte(i)
		pool := new(CrossShardPool_v2)
		pool.shardID = shardID
		pool.validPool = make(map[byte][]*blockchain.CrossShardBlock)
		pool.pendingPool = make(map[byte][]*blockchain.CrossShardBlock)
		pool.mtx = new(sync.RWMutex)
		pool.db = dbCrossShard
		crossShardPoolMapTest[shardID] = pool
	}
	dbCrossShard, err = database.Open("leveldb", filepath.Join("./", "./testdatabase/crossshard"))
	if err != nil {
		panic("Could not open db connection")
	}
	err = dbCrossShard.StoreCrossShardNextHeight(byte(0), byte(1), 1, 3)
	if err != nil {
		panic("Could not store in db")
	}
	err = dbCrossShard.StoreCrossShardNextHeight(byte(0), byte(1), 3, 4)
	if err != nil {
		panic("Could not store in db")
	}
	err = dbCrossShard.StoreCrossShardNextHeight(byte(0), byte(1), 4, 5)
	if err != nil {
		panic("Could not store in db")
	}
	err = dbCrossShard.StoreCrossShardNextHeight(byte(0), byte(1), 5, 7)
	if err != nil {
		panic("Could not store in db")
	}
	bestShardStateShard1 = blockchain.InitBestStateShard(1, &blockchain.ChainTestParam)
	bestShardStateShard1.BestCrossShard[0] = 3
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()

func ResetCrossShardPoolTest() {
	for i := 0; i < 255; i++ {
		shardID := byte(i)
		pool := new(CrossShardPool_v2)
		pool.shardID = shardID
		pool.validPool = make(map[byte][]*blockchain.CrossShardBlock)
		pool.pendingPool = make(map[byte][]*blockchain.CrossShardBlock)
		pool.mtx = new(sync.RWMutex)
		pool.db = dbCrossShard
		crossShardPoolMapTest[shardID] = pool
	}
}
func TestCrossShardPoolv2InitCrossShardPool(t *testing.T) {
	InitCrossShardPool(make(map[byte]blockchain.CrossShardPool), dbCrossShard)
	if len(crossShardPoolMap) != 255 {
		t.Fatal("Fail to init")
	}
}
func TestCrossShardPoolv2GetNextCrossShardHeight(t *testing.T) {
	ResetCrossShardPoolTest()
	shardID := byte(0)
	toShardID := byte(1)
	var nextHeight uint64
	nextHeight = crossShardPoolMapTest[shardID].GetNextCrossShardHeight(shardID, toShardID, 1)
	if nextHeight != 3 {
		t.Fatal("Expect 3 but get ", nextHeight)
	}
	nextHeight = crossShardPoolMapTest[shardID].GetNextCrossShardHeight(shardID, toShardID, 3)
	if nextHeight != 4 {
		t.Fatal("Expect 4 but get ", nextHeight)
	}
	nextHeight = crossShardPoolMapTest[shardID].GetNextCrossShardHeight(shardID, toShardID, 4)
	if nextHeight != 5 {
		t.Fatal("Expect 5 but get ", nextHeight)
	}
	nextHeight = crossShardPoolMapTest[shardID].GetNextCrossShardHeight(shardID, toShardID, 5)
	if nextHeight != 7 {
		t.Fatal("Expect 7 but get ", nextHeight)
	}
	nextHeight = crossShardPoolMapTest[shardID].GetNextCrossShardHeight(shardID, toShardID, 7)
	if nextHeight != 0 {
		t.Fatal("Expect 0 but get ", nextHeight)
	}
	nextHeight = crossShardPoolMapTest[shardID].GetNextCrossShardHeight(shardID, toShardID, 10)
	if nextHeight != 0 {
		t.Fatal("Expect 0 but get ", nextHeight)
	}
}
func TestCrossShardPoolv2RemoveBlockByHeight(t *testing.T) {
	ResetCrossShardPoolTest()
	removeSinceBlockHeight := make(map[byte]uint64)
	fromShardID := byte(0)
	toShardID := byte(1)
	removeSinceBlockHeight[fromShardID] = 4
	crossShardPoolMapTest[toShardID].validPool[fromShardID] = append(crossShardPoolMapTest[toShardID].validPool[fromShardID], crossShardBlock2)
	crossShardPoolMapTest[toShardID].validPool[fromShardID] = append(crossShardPoolMapTest[toShardID].validPool[fromShardID], crossShardBlock5)
	crossShardPoolMapTest[toShardID].pendingPool[fromShardID] = append(crossShardPoolMapTest[toShardID].pendingPool[fromShardID], crossShardBlock3)
	crossShardPoolMapTest[toShardID].pendingPool[fromShardID] = append(crossShardPoolMapTest[toShardID].pendingPool[fromShardID], crossShardBlock6)
	if len(crossShardPoolMapTest[toShardID].pendingPool[fromShardID]) != 2 {
		t.Fatalf("expect pending pool has two block but get %+v", len(crossShardPoolMapTest[toShardID].pendingPool[fromShardID]))
	}
	if len(crossShardPoolMapTest[toShardID].validPool[fromShardID]) != 2 {
		t.Fatalf("expect valid pool has two block but get %+v", len(crossShardPoolMapTest[toShardID].validPool[fromShardID]))
	}
	crossShardPoolMapTest[toShardID].removeBlockByHeight(removeSinceBlockHeight)
	if len(crossShardPoolMapTest[toShardID].pendingPool[fromShardID]) != 1 {
		t.Fatalf("expect pending pool has two block but get %+v", len(crossShardPoolMapTest[toShardID].pendingPool[fromShardID]))
	}
	if len(crossShardPoolMapTest[toShardID].validPool[fromShardID]) != 1 {
		t.Fatalf("expect valid pool has two block but get %+v", len(crossShardPoolMapTest[toShardID].validPool[fromShardID]))
	}
	if crossShardPoolMapTest[toShardID].pendingPool[fromShardID][0].Header.Height != 6 {
		t.Fatalf("expect pending pool has block 6 but get %+v", crossShardPoolMapTest[toShardID].pendingPool[fromShardID][0].Header.Height)
	}
	if crossShardPoolMapTest[toShardID].validPool[fromShardID][0].Header.Height != 5 {
		t.Fatalf("expect valid pool has block 5 but get %+v", crossShardPoolMapTest[toShardID].validPool[fromShardID][0].Header.Height)
	}
}
func TestCrossShardPoolv2UpdatePool(t *testing.T) {
	ResetCrossShardPoolTest()
	fromShardID := byte(0)
	toShardID := byte(1)
	crossShardPoolMapTest[toShardID].validPool[fromShardID] = append(crossShardPoolMapTest[toShardID].validPool[fromShardID], crossShardBlock3)
	crossShardPoolMapTest[toShardID].pendingPool[fromShardID] = append(crossShardPoolMapTest[toShardID].pendingPool[fromShardID], crossShardBlock2)
	crossShardPoolMapTest[toShardID].pendingPool[fromShardID] = append(crossShardPoolMapTest[toShardID].pendingPool[fromShardID], crossShardBlock4)
	crossShardPoolMapTest[toShardID].pendingPool[fromShardID] = append(crossShardPoolMapTest[toShardID].pendingPool[fromShardID], crossShardBlock5)
	crossShardPoolMapTest[toShardID].pendingPool[fromShardID] = append(crossShardPoolMapTest[toShardID].pendingPool[fromShardID], crossShardBlock7)
	crossShardPoolMapTest[toShardID].pendingPool[fromShardID] = append(crossShardPoolMapTest[toShardID].pendingPool[fromShardID], crossShardBlock8)
	crossShardPoolMapTest[toShardID].pendingPool[fromShardID] = append(crossShardPoolMapTest[toShardID].pendingPool[fromShardID], crossShardBlock9)
	if len(crossShardPoolMapTest[toShardID].pendingPool[fromShardID]) != 6 {
		t.Fatalf("expect pending pool has two block but get %+v", len(crossShardPoolMapTest[toShardID].pendingPool[fromShardID]))
	}
	if len(crossShardPoolMapTest[toShardID].validPool[fromShardID]) != 1 {
		t.Fatalf("expect valid pool has two block but get %+v", len(crossShardPoolMapTest[toShardID].validPool[fromShardID]))
	}
	expectedHeight := crossShardPoolMapTest[toShardID].UpdatePool()
	if expectedHeight[0] != 0 {
		t.Fatalf("Expect height after update is 0 but get %+v", expectedHeight[0])
	}
	if len(crossShardPoolMapTest[toShardID].pendingPool[fromShardID]) != 2 {
		t.Fatalf("expect pending pool has two block but get %+v", len(crossShardPoolMapTest[toShardID].pendingPool[fromShardID]))
	}
	if len(crossShardPoolMapTest[toShardID].validPool[fromShardID]) != 4 {
		t.Fatalf("expect valid pool has two block but get %+v", len(crossShardPoolMapTest[toShardID].validPool[fromShardID]))
	}
	for index, block := range crossShardPoolMapTest[toShardID].validPool[fromShardID] {
		switch index {
		case 0:
			if block.Header.Height != 3 {
				t.Fatalf("Expect block height is 3 but get %+v", block.Header.Height)
			}
		case 1:
			if block.Header.Height != 4 {
				t.Fatalf("Expect block height is 4 but get %+v", block.Header.Height)
			}
		case 2:
			if block.Header.Height != 5 {
				t.Fatalf("Expect block height is 5 but get %+v", block.Header.Height)
			}
		case 3:
			if block.Header.Height != 7 {
				t.Fatalf("Expect block height is 7 but get %+v", block.Header.Height)
			}
		}
	}
	for index, block := range crossShardPoolMapTest[toShardID].pendingPool[fromShardID] {
		switch index {
		case 0:
			if block.Header.Height != 8 {
				t.Fatalf("Expect block height is 8 but get %+v", block.Header.Height)
			}
		case 1:
			if block.Header.Height != 9 {
				t.Fatalf("Expect block height is 9 but get %+v", block.Header.Height)
			}
		}
	}
}
func TestCrossShardPoolv2AddCrossShardBlock(t *testing.T) {
	ResetCrossShardPoolTest()
	fromShardID := byte(0)
	toShardID := byte(1)
	_, _, err1 := crossShardPoolMapTest[toShardID].AddCrossShardBlock(crossShardBlock3WrongShard)
	if err1 == nil {
		t.Fatalf("Expect WrongShardIDError but no error")
	} else {
		if err1.(*BlockPoolError).Code != ErrCodeMessage[WrongShardIDError].Code {
			t.Fatalf("Expect %+v error but get %+v", WrongShardIDError, err1)
		}
	}
	temp := make(map[byte]uint64)
	temp[0] = 4
	crossShardPoolMapTest[toShardID].crossShardState = temp
	_, _, err2 := crossShardPoolMapTest[toShardID].AddCrossShardBlock(crossShardBlock4)
	if err2 == nil {
		t.Fatalf("Expect WrongShardIDError but no error")
	} else {
		if err2.(*BlockPoolError).Code != ErrCodeMessage[OldBlockError].Code {
			t.Fatalf("Expect %+v error but get %+v", OldBlockError, err2)
		}
	}
	ResetCrossShardPoolTest()
	crossShardPoolMapTest[toShardID].validPool[fromShardID] = append(crossShardPoolMapTest[toShardID].validPool[fromShardID], crossShardBlock3)
	_, _, err3 := crossShardPoolMapTest[toShardID].AddCrossShardBlock(crossShardBlock3)
	if err3 == nil {
		t.Fatalf("Expect WrongShardIDError but no error")
	} else {
		if err3.(*BlockPoolError).Code != ErrCodeMessage[DuplicateBlockError].Code {
			t.Fatalf("Expect %+v error but get %+v", DuplicateBlockError, err3)
		}
	}
	crossShardPoolMapTest[toShardID].pendingPool[fromShardID] = append(crossShardPoolMapTest[toShardID].pendingPool[fromShardID], crossShardBlock4)
	_, _, err4 := crossShardPoolMapTest[toShardID].AddCrossShardBlock(crossShardBlock4)
	if err4 == nil {
		t.Fatalf("Expect WrongShardIDError but no error")
	} else {
		if err4.(*BlockPoolError).Code != ErrCodeMessage[DuplicateBlockError].Code {
			t.Fatalf("Expect %+v error but get %+v", DuplicateBlockError, err4)
		}
	}
}
