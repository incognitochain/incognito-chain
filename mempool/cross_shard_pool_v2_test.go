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
	dbCrossShard database.DatabaseInterface
	crossShardPoolMapTest = make(map[byte]*CrossShardPool_v2)
	crossShardBlock2           = &blockchain.CrossShardBlock{
		Header: blockchain.ShardHeader{
			ShardID:   0,
			Height:    2,
			Timestamp: time.Now().Unix() - 100,
		},
		ToShardID: 1,
	}
	crossShardBlock2Forked = &blockchain.CrossShardBlock{
		Header: blockchain.ShardHeader{
			ShardID:   0,
			Height:    2,
			Timestamp: time.Now().Unix(),
		},
		ToShardID: 1,
	}
	crossShardBlock3 = &blockchain.CrossShardBlock{
		Header: blockchain.ShardHeader{
			ShardID:       0,
			Height:        3,
			PrevBlockHash: crossShardBlock2.Header.Hash(),
		},
		ToShardID: 1,
	}
	crossShardBlock3Forked = &blockchain.CrossShardBlock{
		Header: blockchain.ShardHeader{
			ShardID:       0,
			Height:        3,
			PrevBlockHash: crossShardBlock2Forked.Header.Hash(),
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
	pendingCrossShardBlocks = []*blockchain.CrossShardBlock{}
	validCrossShardBlocks   = []*blockchain.CrossShardBlock{}
)
var _ = func() (_ struct{}) {
	for i:=0; i < 255; i ++ {
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
	err = dbCrossShard.StoreCrossShardNextHeight(byte(0), byte(1), 1,3)
	if err != nil {
		panic("Could not store in db")
	}
	err = dbCrossShard.StoreCrossShardNextHeight(byte(0), byte(1), 3,4)
	if err != nil {
		panic("Could not store in db")
	}
	err = dbCrossShard.StoreCrossShardNextHeight(byte(0), byte(1), 4,5)
	if err != nil {
		panic("Could not store in db")
	}
	err = dbCrossShard.StoreCrossShardNextHeight(byte(0), byte(1), 5,7)
	if err != nil {
		panic("Could not store in db")
	}
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()
func ResetCrossShardPoolTest() {
	for i:=0; i<255; i++ {
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
}
func TestCrossShardPoolv2UpdatePool(t *testing.T) {
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