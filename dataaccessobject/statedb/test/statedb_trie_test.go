package test

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	_ "github.com/incognitochain/incognito-chain/incdb/lvdb"
	"reflect"
	"sync"
	"testing"
)

func TestGetAllNodeFromTrie(t *testing.T) {
	rootHashs := map[string]common.Hash{
		"ConsensusStateDBRootHash":   common.Hash{}.NewHashFromStr2("bcb542516541623873047bd56aa6155e8ec3376421ffd5086fc02dcb7ab1f699"),
		"TransactionStateDBRootHash": common.Hash{}.NewHashFromStr2("7ec2f50be79052c4b1f8ee57eec06d496935521343acbb009b9117cb1c2960d0"),
		"FeatureStateDBRootHash":     common.Hash{}.NewHashFromStr2("6f6e11615c053c23106daa2f9a37d894bbf7f93d2103c6b7ff9f6235fed32f6a"),
		"RewardStateDBRootHash":      common.Hash{}.NewHashFromStr2("05d7e3f1c8e9d66dd0d85e7f0de1d4eb6dbfc0208552e0eebebd8ce25f86e332"),
		"SlashStateDBRootHash":       common.Hash{}.NewHashFromStr2("21b463e3b52f6201c0ad6c991be0485b6ef8c092e64583ffa655cc1b171fe856"),
	}

	datadir := "../../../../inc-data/mainnet/fullnode/mainnet/block/shard0"
	diskDB, err := incdb.Open("leveldb", datadir)
	if err != nil {
		t.Fatal(err)
	}
	db := statedb.NewDatabaseAccessWarper(diskDB)

	wg := new(sync.WaitGroup)

	for name, root := range rootHashs {
		wg.Add(1)
		go func(name string, root common.Hash) {
			size := common.StorageSize(0)
			sDB, err := statedb.NewWithPrefixTrie(root, db)
			if err != nil {
				t.Fatal(err)
			}
			m := make(map[common.Hash]interface{})
			it := sDB.Trie().NodeIterator([]byte{})
			for it.Next(true) {
				m[it.Hash()] = it.Node()
			}
			for k, v := range m {
				if v != nil {
					sizeV, err := json.Marshal(v)
					if err != nil {
						t.Fatal(err)
					}
					//sizeV := uint64(reflect.TypeOf(v).Size())
					size += common.StorageSize(len(sizeV))
				}
				sizeK := uint64(reflect.TypeOf(k).Size())
				size += common.StorageSize(sizeK)
			}
			t.Log(fmt.Sprintf("%s:%d, size:%+v", name, len(m), size.String()))
			wg.Done()
		}(name, root)
	}
	wg.Wait()
}

func TestGetDataFromBlock(t *testing.T) {

	datadir := "../../../../inc-data/mainnet/fullnode/mainnet/block/shard0"
	diskDB, err := incdb.Open("leveldb", datadir)
	if err != nil {
		t.Fatal(err)
	}
	common.MaxShardNumber = 8
	size := common.StorageSize(0)
	sizeByte := common.StorageSize(0)
	hash := common.Hash{}.NewHashFromStr2("53891c0c41d2a56e1cbe9de3db68091a8fcb7aff9aa0d9134932477ef0c36b8d")
	height := uint64(1020001)

	for height != 1 {
		t.Log(height)
		shardBlockBytes, err := rawdbv2.GetShardBlockByHash(diskDB, hash)
		if err != nil {
			t.Fatal(err)
		}

		shardBlock := types.NewShardBlock()
		err = json.Unmarshal(shardBlockBytes, shardBlock)
		if err != nil {
			t.Fatal(err)
		}
		sizeByte += common.StorageSize(len(shardBlockBytes))

		sizeShardBlock := uint64(reflect.TypeOf(*shardBlock).Size())
		size += common.StorageSize(sizeShardBlock)

		hash = shardBlock.GetPrevHash()
		height = shardBlock.GetHeight()
	}
	t.Log("------------")
	t.Logf(fmt.Sprintf("size: %+v", size.String()))
	t.Logf(fmt.Sprintf("sizeBytes: %+v", sizeByte.String()))
}
