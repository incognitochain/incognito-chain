package mempool
//
//import (
//	"github.com/incognitochain/incognito-chain/common"
//	"testing"
//
//	"github.com/incognitochain/incognito-chain/blockchain"
//)
//
//var (
//	blk2             = blockchain.ShardToBeaconBlock{}
//	blk3             = blockchain.ShardToBeaconBlock{}
//	blk4             = blockchain.ShardToBeaconBlock{}
//	blk5             = blockchain.ShardToBeaconBlock{}
//	blk6             = blockchain.ShardToBeaconBlock{}
//	blks             = make([]blockchain.ShardToBeaconBlock, 100)
//	latestShardState = make(map[byte]uint64)
//	err              error
//)
//
//func Init() {
//	blk2.Header.Height = 2
//	blk3.Header.Height = 3
//	blk4.Header.Height = 4
//	blk5.Header.Height = 5
//	blk6.Header.Height = 6
//	for index, blk := range blks {
//		blk.Header.Height = uint64(index)
//	}
//	latestShardState[common.ZeroByte] = 1
//}
//
///*
//	Testcase:
//	- Init GetShardToBeaconPool
//	- Set Init Shard State
//	- Set arbitrary shard state
//	- Successfully add block in pool
//	- Add Duplicate block
//	- Add Too Old Block
//
//*/
//func TestGetShardToBeaconPool(t *testing.T) {
//	shardToBeaconPoolTest := GetShardToBeaconPool()
//	if shardToBeaconPoolTest == nil {
//		panic("ShardToBeaconPool Can't Be Nil via GetShardToBeaconPool")
//	}
//}
//func TestSetShardState(t *testing.T) {
//	Init()
//	shardToBeaconPoolTest := GetShardToBeaconPool()
//	shardToBeaconPoolTest.SetShardState(latestShardState)
//	if shardToBeaconPoolTest.latestValidHeight[common.ZeroByte] != 1 {
//		panic("LastestValidHeight Should Be 1")
//	}
//	latestShardState[common.ZeroByte] = 5
//	shardToBeaconPoolTest.SetShardState(latestShardState)
//	if shardToBeaconPoolTest.latestValidHeight[common.ZeroByte] != 5 {
//		panic("LastestValidHeight Should Be 5")
//	}
//}
//func TestAddShardToBeaconBlock(t *testing.T) {
//	Init()
//	shardToBeaconPoolTest := GetShardToBeaconPool()
//	shardToBeaconPoolTest.SetShardState(latestShardState)
//	shardToBeaconPoolTest.AddShardToBeaconBlock(blk2)
//	shardToBeaconPoolTest.AddShardToBeaconBlock(blk3)
//	if shardToBeaconPoolTest.latestValidHeight[common.ZeroByte] != 3 {
//		panic("LastestValidHeight Should Be 3")
//	}
//	if len(shardToBeaconPoolTest.pool[common.ZeroByte]) != 2 {
//		panic("Pool Should Have 2 Block")
//	}
//	shardToBeaconPoolTest.AddShardToBeaconBlock(blk5)
//	if shardToBeaconPoolTest.latestValidHeight[common.ZeroByte] != 3 {
//		panic("LastestValidHeight Should Be 3")
//	}
//	if len(shardToBeaconPoolTest.pool[common.ZeroByte]) != 3 {
//		panic("Pool Should Have 3 Block")
//	}
//	shardToBeaconPoolTest.AddShardToBeaconBlock(blk6)
//	if shardToBeaconPoolTest.latestValidHeight[common.ZeroByte] != 3 {
//		panic("LastestValidHeight Should Be 3")
//	}
//	t.Log(shardToBeaconPoolTest.pool[common.ZeroByte][0].Header.Height)
//	t.Log(shardToBeaconPoolTest.pool[common.ZeroByte][1].Header.Height)
//	t.Log(shardToBeaconPoolTest.pool[common.ZeroByte][2].Header.Height)
//	t.Log(shardToBeaconPoolTest.pool[common.ZeroByte][3].Header.Height)
//	if len(shardToBeaconPoolTest.pool[common.ZeroByte]) != 4 {
//		panic("Pool Should Have 4 Block")
//	}
//	//==============Duplicate==================
//	_, _, err = shardToBeaconPoolTest.AddShardToBeaconBlock(blk6)
//	if err.Error() != "receive duplicate block" {
//		panic(err.Error())
//	}
//	//==============Old Block==================
//	_, _, err = shardToBeaconPoolTest.AddShardToBeaconBlock(blk3)
//	if err.Error() != "receive old block" {
//		panic(err.Error())
//	}
//}
