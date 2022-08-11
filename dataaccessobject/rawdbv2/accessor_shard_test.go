package rawdbv2_test

import (
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/incdb"
	_ "github.com/incognitochain/incognito-chain/incdb/lvdb"
	"io/ioutil"
	"os"
)

var (
	shardBlocks       []*types.ShardBlock
	forkedShardBlock1 = types.NewShardBlock()
	forkedShardBlock2 = types.NewShardBlock()
	randomShardBlock1 = types.NewShardBlock()
	randomShardBlock2 = types.NewShardBlock()
	maxS              = 5
	dbS               incdb.Database
)
var _ = func() (_ struct{}) {
	dbSPath, err := ioutil.TempDir(os.TempDir(), "test_rawdbv2")
	if err != nil {
		panic(err)
	}
	dbS, err = incdb.Open("leveldb", dbSPath)
	if err != nil {
		panic(err)
	}
	for i := 0; i < maxS; i++ {
		shardBlock := types.NewShardBlock()
		shardBlock.Header.Height = uint64(i)
		shardBlock.Header.ShardID = 0
		shardBlock.Header.Round = 1
		shardBlock.Header.Version = 1
		if i != 0 {
			shardBlock.Header.PreviousBlockHash = shardBlocks[i-1].Header.Hash()
		}
		shardBlocks = append(shardBlocks, shardBlock)
	}

	forkedShardBlock1.Header.Height = 1
	forkedShardBlock1.Header.Version = 2
	forkedShardBlock1.Header.ShardID = 0
	forkedShardBlock1.Header.Round = 1
	forkedShardBlock2.Header.Height = 2
	forkedShardBlock2.Header.Version = 2
	forkedShardBlock2.Header.ShardID = 0
	forkedShardBlock2.Header.Round = 1
	forkedShardBlock2.Header.PreviousBlockHash = forkedShardBlock1.Header.Hash()

	randomShardBlock1.Header.Height = 10001
	randomShardBlock2.Header.Height = 10002
	return
}()

func resetDatabaseShard() {
	dbSPath, err := ioutil.TempDir(os.TempDir(), "test_rawdbv2")
	if err != nil {
		panic(err)
	}
	dbS, err = incdb.Open("leveldb", dbSPath)
	if err != nil {
		panic(err)
	}
}
