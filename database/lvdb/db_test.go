package lvdb_test

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/incognitochain/incognito-chain/database"
	_ "github.com/incognitochain/incognito-chain/database/lvdb"
)

var db database.DatabaseInterface

var _ = func() (_ struct{}) {
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_")
	if err != nil {
		log.Fatalf("failed to create temp dir: %+v", err)
	}
	log.Println(dbPath)
	db, err = database.Open("leveldb", dbPath)
	if err != nil {
		log.Fatalf("could not open db path: %s, %+v", dbPath, err)
	}
	return
}()

func TestDb_Setup(t *testing.T) {
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_")
	if err != nil {
		t.Fatalf("failed to create temp dir: %+v", err)
	}
	t.Log(dbPath)
	db, err := database.Open("leveldb", dbPath)
	if err != nil {
		t.Fatalf("could not open db path: %s, %+v", dbPath, err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("db.close %+v", err)
	}
	os.RemoveAll(dbPath)
}

func TestDb_Base(t *testing.T) {
	if db != nil {
		db.Put([]byte("a"), []byte{1})
		result, err := db.Get([]byte("a"))
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, result[0], []byte{1}[0])
		has, err := db.HasValue([]byte("a"))
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, has, true)

		err = db.Delete([]byte("a"))
		if err != nil {
			t.Error(err)
		}
		has, err = db.HasValue([]byte("a"))
		assert.Equal(t, err, nil)
		assert.Equal(t, has, false)
	} else {
		t.Error("DB is not open")
	}
}

// Process on Block data
func TestDb_StoreShardBlock(t *testing.T) {
	if db != nil {
		block := &blockchain.ShardBlock{
			Header: blockchain.ShardHeader{
				Version: 1,
				ShardID: 3,
				Height:  1,
			},
		}
		// test store block
		err := db.StoreShardBlock(block, *block.Hash(), block.Header.ShardID)
		assert.Equal(t, err, nil)

		// test Fetch block
		blockInBytes, err := db.FetchBlock(*block.Hash())
		assert.Equal(t, err, nil)
		blockNew := blockchain.ShardBlock{}
		err = json.Unmarshal(blockInBytes, &blockNew)
		assert.Equal(t, err, nil)
		assert.Equal(t, blockNew.Hash(), block.Hash())

		// has block
		has, err := db.HasBlock(*block.Hash())
		assert.Equal(t, err, nil)
		assert.Equal(t, has, true)

		// delete block
		err = db.DeleteBlock(*blockNew.Hash(), blockNew.Header.Height, blockNew.Header.ShardID)
		assert.Equal(t, err, nil)
	} else {
		t.Error("DB is not open")
	}
}

// Block index
func TestDb_StoreShardBlockIndex(t *testing.T) {
	if db != nil {
		block := &blockchain.ShardBlock{
			Header: blockchain.ShardHeader{
				Version: 1,
				ShardID: 3,
				Height:  1,
			},
		}
		// test store block
		err := db.StoreShardBlockIndex(*block.Hash(), block.Header.Height, block.Header.ShardID)
		assert.Equal(t, err, nil)

		// test GetIndexOfBlock
		blockHeigh, shardID, err := db.GetIndexOfBlock(*block.Hash())
		assert.Equal(t, err, nil)
		assert.Equal(t, blockHeigh, uint64(1))
		assert.Equal(t, shardID, uint8(3))

		// GetBlockByIndex
		hash, err := db.GetBlockByIndex(1, 3)
		assert.Equal(t, hash.String(), block.Hash().String())

	} else {
		t.Error("DB is not open")
	}
}

// Beacon
func TestDb_StoreBeaconBlock(t *testing.T) {
	if db != nil {
		beaconBlock := &blockchain.BeaconBlock{
			Header: blockchain.BeaconHeader{
				Version: 1,
				Height:  1,
			},
		}
		// test store block
		err := db.StoreBeaconBlock(beaconBlock, *beaconBlock.Hash())
		assert.Equal(t, err, nil)

		// test Fetch block
		blockInBytes, err := db.FetchBeaconBlock(*beaconBlock.Hash())
		assert.Equal(t, err, nil)
		blockNew := blockchain.BeaconBlock{}
		err = json.Unmarshal(blockInBytes, &blockNew)
		assert.Equal(t, err, nil)
		assert.Equal(t, blockNew.Hash(), beaconBlock.Hash())

		// has block
		has, err := db.HasBeaconBlock(*beaconBlock.Hash())
		assert.Equal(t, err, nil)
		assert.Equal(t, has, true)

		// delete block
		err = db.DeleteBeaconBlock(*blockNew.Hash(), blockNew.Header.Height)
		assert.Equal(t, err, nil)
	} else {
		t.Error("DB is not open")
	}
}

//Crossshard
func TestDb_StoreCrossShardNextHeight(t *testing.T) {
	if db != nil {
		err := db.StoreCrossShardNextHeight(0, 1, 1, 2)
		assert.Equal(t, err, nil)

		val, err := db.FetchCrossShardNextHeight(0, 1, 1)
		assert.Equal(t, err, nil)
		assert.Equal(t, uint64(val), uint64(2))

		//err = db.RestoreCrossShardNextHeights(0, 1, 2)
		// TODO: 0xbahamooth\
		//assert.Equal(t, err, nil)
	} else {
		t.Error("DB is not open")
	}
}
