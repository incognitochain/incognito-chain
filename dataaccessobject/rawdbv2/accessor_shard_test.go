package rawdbv2_test

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"io/ioutil"
	"os"
	"testing"

	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/incdb"
	_ "github.com/incognitochain/incognito-chain/incdb/lvdb"
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
		if i != 0 && i != 1 {
			shardBlock.Header.PreviousBlockHash = shardBlocks[i-1].Header.Hash()
		} else {
			shardBlock.Header.PreviousBlockHash = common.Hash{}
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
func storeShardBlock() error {
	resetDatabaseShard()
	for i := 0; i < maxS; i++ {
		err := rawdbv2.StoreShardBlock(dbS, shardBlocks[i].Header.Hash(), shardBlocks[i])
		if err != nil {
			return err
		}
	}
	err := rawdbv2.StoreShardBlock(dbS, forkedShardBlock1.Header.Hash(), forkedShardBlock1)
	if err != nil {
		return err
	}
	err1 := rawdbv2.StoreShardBlock(dbS, forkedShardBlock2.Header.Hash(), forkedShardBlock2)
	if err1 != nil {
		return err1
	}
	for i := 0; i < maxS; i++ {
		err := rawdbv2.StoreFinalizedShardBlockHashByIndex(dbS, byte(0), uint64(i), shardBlocks[i].Header.Hash())
		if err != nil {
			return err
		}
	}
	return nil
}

func TestStoreShardBlock(t *testing.T) {
	resetDatabaseShard()
	for i := 0; i < maxS; i++ {
		err := rawdbv2.StoreShardBlock(dbS, shardBlocks[i].Header.Hash(), shardBlocks[i])
		if err != nil {
			t.Fatal(err)
		}
	}
	err := rawdbv2.StoreShardBlock(dbS, forkedShardBlock1.Header.Hash(), forkedShardBlock1)
	if err != nil {
		t.Fatal(err)
	}
	err1 := rawdbv2.StoreShardBlock(dbS, forkedShardBlock2.Header.Hash(), forkedShardBlock2)
	if err1 != nil {
		t.Fatal(err1)
	}
}

func TestStoreFinalizedShardBlockHashByIndex(t *testing.T) {
	resetDatabaseShard()
	for i := 0; i < maxS; i++ {
		err := rawdbv2.StoreFinalizedShardBlockHashByIndex(dbS, byte(0), uint64(i), shardBlocks[i].Header.Hash())
		if err != nil {
			t.Fatal(err)
		}
	}
	err := rawdbv2.StoreFinalizedShardBlockHashByIndex(dbS, byte(0), forkedShardBlock1.Header.Height, forkedShardBlock1.Header.Hash())
	if err != nil {
		t.Fatal(err)
	}
	err1 := rawdbv2.StoreFinalizedShardBlockHashByIndex(dbS, byte(0), forkedShardBlock2.Header.Height, forkedShardBlock2.Header.Hash())
	if err1 != nil {
		t.Fatal(err1)
	}
}

func TestHasShardBlock(t *testing.T) {
	err := storeShardBlock()
	if err != nil {
		t.Fatal(err)
	}
	has, err := rawdbv2.HasShardBlock(dbS, randomShardBlock1.Header.Hash())
	if has {
		t.Fatalf("Want block %+v but got nothing", randomShardBlock1.Header.Hash())
	}
	if err != nil {
		t.Fatal(err)
	}
	has1, err := rawdbv2.HasShardBlock(dbS, randomShardBlock2.Header.Hash())
	if has1 {
		t.Fatalf("Want block %+v but got nothing", randomShardBlock2.Header.Hash())
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestHasShardBlockForked(t *testing.T) {
	err := storeShardBlock()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < maxS; i++ {
		has, err := rawdbv2.HasShardBlock(dbS, shardBlocks[i].Header.Hash())
		if !has {
			t.Fatalf("Want block %+v but got nothing", shardBlocks[i].Header.Hash())
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	has, err := rawdbv2.HasShardBlock(dbS, forkedShardBlock1.Header.Hash())
	if !has {
		t.Fatalf("Want block %+v but got nothing", forkedShardBlock1.Header.Hash())
	}
	if err != nil {
		t.Fatal(err)
	}
	has1, err := rawdbv2.HasShardBlock(dbS, forkedShardBlock2.Header.Hash())
	if !has1 {
		t.Fatalf("Want block %+v but got nothing", forkedShardBlock2.Header.Hash())
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetShardBlockByHash(t *testing.T) {
	err := storeShardBlock()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < maxS; i++ {
		res, err := rawdbv2.GetShardBlockByHash(dbS, shardBlocks[i].Header.Hash())
		if err != nil {
			t.Fatal(err)
		}
		if len(res) == 0 {
			t.Fatalf("got nothing from hash %+v", shardBlocks[i].Header.Hash())
		}
	}
	res1, err1 := rawdbv2.GetShardBlockByHash(dbS, forkedShardBlock1.Header.Hash())
	if err1 != nil {
		t.Fatal(err1)
	}
	if len(res1) == 0 {
		t.Fatalf("got nothing from hash %+v", forkedShardBlock1.Header.Hash())
	}
	res2, err2 := rawdbv2.GetShardBlockByHash(dbS, forkedShardBlock2.Header.Hash())
	if err2 != nil {
		t.Fatal(err2)
	}
	if len(res2) == 0 {
		t.Fatalf("got nothing from hash %+v", forkedShardBlock2.Header.Hash())
	}

}

func TestGetShardBlockByIndex(t *testing.T) {
	err := storeShardBlock()
	if err != nil {
		t.Fatal(err)
	}
	hash, err := rawdbv2.GetFinalizedShardBlockHashByIndex(dbS, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	data, err := rawdbv2.GetShardBlockByHash(dbS, *hash)
	if err != nil {
		t.Fatal(err)
	}
	shardBlock := types.NewShardBlock()
	err = json.Unmarshal(data, shardBlock)
	if err != nil {
		t.Fatal(err)
	}
	if shardBlock.Header.Height != uint64(1) {
		t.Fatalf("want height %+v but got %+v", 1, shardBlock.Header.Height)
	}
	if shardBlock.Header.Hash().String() != hash.String() {
		t.Fatalf("want hash %+v but got %+v", shardBlock.Header.Hash(), *hash)
	}
}
