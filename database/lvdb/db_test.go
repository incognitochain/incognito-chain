package lvdb_test

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/transaction"
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

// Block beacon index
func TestDb_StoreShardBlockBeaconIndex(t *testing.T) {
	if db != nil {
		beaconBlock := &blockchain.BeaconBlock{
			Header: blockchain.BeaconHeader{
				Version: 1,
				Height:  1,
			},
		}
		// test store block
		err := db.StoreBeaconBlockIndex(*beaconBlock.Hash(), beaconBlock.Header.Height)
		assert.Equal(t, err, nil)

		// test GetIndexOfBlock
		blockHeigh, err := db.GetIndexOfBeaconBlock(*beaconBlock.Hash())
		assert.Equal(t, err, nil)
		assert.Equal(t, blockHeigh, uint64(1))

		// GetBlockByIndex
		hash, err := db.GetBeaconBlockHashByIndex(1)
		assert.Equal(t, hash.String(), beaconBlock.Hash().String())

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

// Transaction index
func TestDb_StoreTxIndex(t *testing.T) {
	if db != nil {
		block := &blockchain.ShardBlock{
			Header: blockchain.ShardHeader{
				Version: 1,
				ShardID: 3,
				Height:  1,
			},
			Body: blockchain.ShardBody{
				Transactions: []metadata.Transaction{},
			},
		}
		block.Body.Transactions = append(block.Body.Transactions, &transaction.Tx{
			Version: 1,
			Info:    []byte("Test 1"),
		})
		block.Body.Transactions = append(block.Body.Transactions, &transaction.Tx{
			Version: 1,
			Info:    []byte("Test 2"),
		})
		err := db.StoreTransactionIndex(*block.Body.Transactions[1].Hash(), *block.Hash(), 1)
		assert.Equal(t, err, nil)

		blockHash, index, err := db.GetTransactionIndexById(*block.Body.Transactions[1].Hash())
		if err.(*database.DatabaseError) != nil {
			t.Error(err)
		}
		assert.Equal(t, blockHash, *block.Hash())
		assert.Equal(t, index, 1)
	} else {
		t.Error("DB is not open")
	}
}

// Best state of Prev
func TestDb_StorePrevBestState(t *testing.T) {
	if db != nil {
		bestState := blockchain.BestState{
			Beacon: &blockchain.BestStateBeacon{
				Epoch: 100,
			},
		}
		tempMarshal, err := json.Marshal(bestState.Beacon)
		assert.Equal(t, err, nil)
		err = db.StorePrevBestState(tempMarshal, true, 0)
		assert.Equal(t, err, nil)

		beaconInBytes, err := db.FetchPrevBestState(true, 0)
		assert.Equal(t, err, nil)
		temp := blockchain.BestStateBeacon{}
		json.Unmarshal(beaconInBytes, &temp)
		assert.Equal(t, bestState.Beacon.Epoch, temp.Epoch)
		err = db.CleanBackup(true, 0)
		_, err = db.FetchPrevBestState(true, 0)
		assert.NotEqual(t, err, nil)
	} else {
		t.Error("DB is not open")
	}
}

// Best state of shard chain
func TestDb_StoreShardBestState(t *testing.T) {
	if db != nil {
		besState := blockchain.BestState{
			Shard: make(map[byte]*blockchain.BestStateShard),
		}
		bestStateShard := blockchain.BestStateShard{
			Epoch: 100,
		}
		besState.Shard[0] = &bestStateShard
		err := db.StoreShardBestState(bestStateShard, 0)
		assert.Equal(t, err, nil)

		temp, err := db.FetchShardBestState(0)
		assert.Equal(t, err, nil)
		tempObject := blockchain.BestStateShard{}
		err = json.Unmarshal(temp, &tempObject)
		assert.Equal(t, err, nil)
		assert.Equal(t, tempObject.Epoch, bestStateShard.Epoch)

		err = db.CleanShardBestState()
		assert.Equal(t, err, nil)
		_, err = db.FetchShardBestState(0)
		assert.NotEqual(t, err, nil)
	} else {
		t.Error("DB is not open")
	}
}

// Best state of beacon chain
func TestDb_StoreBeaconBestState(t *testing.T) {
	if db != nil {
		bestState := blockchain.BestState{
			Beacon: &blockchain.BestStateBeacon{
				Epoch: 100,
			},
		}
		err := db.StoreBeaconBestState(bestState)
		assert.Equal(t, err, nil)
		temp, err := db.FetchBeaconBestState()
		assert.Equal(t, err, nil)
		tempObject := blockchain.BestState{}
		err = json.Unmarshal(temp, &tempObject)
		assert.Equal(t, err, nil)
		assert.Equal(t, tempObject.Beacon.Epoch, bestState.Beacon.Epoch)

		err = db.CleanBeaconBestState()
		assert.Equal(t, err, nil)
		_, err = db.FetchBeaconBestState()
		assert.NotEqual(t, err, nil)
	} else {
		t.Error("DB is not open")
	}
}

// Commitee with epoch
func TestDb_StoreCommitteeByHeight(t *testing.T) {
	if db != nil {
		block := blockchain.ShardBlock{
			Header: blockchain.ShardHeader{
				Height: 100,
			},
		}
		bestState := blockchain.BestState{
			Beacon: &blockchain.BestStateBeacon{
				Epoch:          100,
				ShardCommittee: make(map[byte][]string),
			},
		}
		bestState.Beacon.ShardCommittee[0] = make([]string, 0)
		bestState.Beacon.ShardCommittee[0] = append(bestState.Beacon.ShardCommittee[0], "committee1")
		bestState.Beacon.ShardCommittee[0] = append(bestState.Beacon.ShardCommittee[0], "committee2")
		err := db.StoreCommitteeByEpoch(block.Header.Height, bestState.Beacon.GetShardCommittee())
		assert.Equal(t, err, nil)

		shardCommittee := make(map[byte][]string)
		data, err := db.FetchCommitteeByEpoch(100)
		assert.Equal(t, err, nil)
		err = json.Unmarshal(data, &shardCommittee)
		assert.Equal(t, err, nil)
		assert.Equal(t, shardCommittee[0][0], "committee1")
		assert.Equal(t, shardCommittee[0][1], "committee2")

		has, err := db.HasCommitteeByEpoch(100)
		assert.Equal(t, has, true)
		assert.Equal(t, err, nil)

		err = db.DeleteCommitteeByEpoch(100)
		assert.Equal(t, err, nil)

		has, err = db.HasCommitteeByEpoch(100)
		assert.Equal(t, has, false)
		assert.Equal(t, err, nil)
	} else {
		t.Error("DB is not open")
	}
}

func TestDb_StoreSerialNumbers(t *testing.T) {
	if db != nil {
		serialNumber := make([][]byte, 0)
		ser1 := []byte{0, 1}
		ser2 := []byte{0, 2}
		serialNumber = append(serialNumber, ser1)
		serialNumber = append(serialNumber, ser2)
		tokenID := common.Hash{}
		err := db.StoreSerialNumbers(tokenID, serialNumber, 0)
		assert.Equal(t, err, nil)

		has, err := db.HasSerialNumber(tokenID, ser1, 0)
		assert.Equal(t, err, nil)
		assert.Equal(t, has, true)

		err = db.BackupSerialNumbersLen(tokenID, 0)
		assert.Equal(t, err, nil)

		err = db.RestoreSerialNumber(tokenID, 0, serialNumber)
		assert.Equal(t, err, nil)
		has, err = db.HasSerialNumber(tokenID, ser1, 0)
		assert.Equal(t, err, nil)
		assert.Equal(t, has, false)

		err = db.StoreSerialNumbers(tokenID, serialNumber, 0)
		assert.Equal(t, err, nil)
		has, err = db.HasSerialNumber(tokenID, ser1, 0)
		assert.Equal(t, err, nil)
		assert.Equal(t, has, true)

		err = db.CleanSerialNumbers()
		assert.Equal(t, err, nil)
		has, err = db.HasSerialNumber(tokenID, ser1, 0)
		assert.Equal(t, err, nil)
		assert.Equal(t, has, false)

	} else {
		t.Error("DB is not open")
	}
}

func TestDb_StoreCommitments(t *testing.T) {
	if db != nil {
		committments := make([][]byte, 0)
		cm1 := []byte{0, 1}
		cm2 := []byte{0, 2}
		committments = append(committments, cm1)
		committments = append(committments, cm2)
		tokenID := common.Hash{}
		publicKey := common.Hash{}

		err := db.StoreCommitments(tokenID, publicKey.GetBytes(), committments, 0)
		assert.Equal(t, err, nil)

		has, err := db.HasCommitment(tokenID, cm1, 0)
		assert.Equal(t, err, nil)
		assert.Equal(t, has, true)

		has, err = db.HasCommitmentIndex(tokenID, 0, 0)
		assert.Equal(t, err, nil)
		assert.Equal(t, has, true)

		len, err := db.GetCommitmentLength(tokenID, 0)
		assert.Equal(t, err, nil)
		assert.Equal(t, len.Int64(), int64(2))

		temp, err := db.GetCommitmentByIndex(tokenID, 1, 0)
		assert.Equal(t, err, nil)
		assert.Equal(t, temp, cm2)

		index, err := db.GetCommitmentIndex(tokenID, cm1, 0)
		assert.Equal(t, err, nil)
		assert.Equal(t, index.Uint64(), uint64(0))

		err = db.BackupCommitmentsOfPubkey(tokenID, 0, publicKey.GetBytes())
		assert.Equal(t, err, nil)

		err = db.RestoreCommitmentsOfPubkey(tokenID, 0, publicKey.GetBytes(), committments)
		assert.Equal(t, err, nil)
		has, err = db.HasSerialNumber(tokenID, cm1, 0)
		assert.Equal(t, err, nil)
		assert.Equal(t, has, false)

		err = db.CleanCommitments()
		assert.Equal(t, err, nil)
		has, err = db.HasCommitment(tokenID, cm1, 0)
		assert.Equal(t, err, nil)
		assert.Equal(t, has, false)
	} else {
		t.Error("DB is not open")
	}
}

// output
func TestDb_StoreOutputCoins(t *testing.T) {
	if db != nil {
		outputCoins := make([][]byte, 0)
		cm1 := []byte{0, 1}
		cm2 := []byte{0, 2}
		outputCoins = append(outputCoins, cm1)
		outputCoins = append(outputCoins, cm2)
		tokenID := common.Hash{}
		publicKey := common.Hash{}
		err := db.StoreOutputCoins(tokenID, publicKey.GetBytes(), outputCoins, 1)
		assert.Equal(t, err, nil)

		data, err := db.GetOutcoinsByPubkey(tokenID, publicKey.GetBytes(), 1)
		assert.Equal(t, err, nil)
		assert.Equal(t, len(data), 2)

	} else {
		t.Error("DB is not open")
	}
}

// SNDerivator
func TestDb_StoreSNDerivators(t *testing.T) {
	if db != nil {
		snd := make([][]byte, 0)
		snd1 := []byte{0, 1}
		snd2 := []byte{0, 2}
		snd = append(snd, snd1)
		snd = append(snd, snd2)
		tokenID := common.Hash{}

		err := db.StoreSNDerivators(tokenID, snd, 0)
		assert.Equal(t, err, nil)

		has, err := db.HasSNDerivator(tokenID, snd1, 0)
		assert.Equal(t, err, nil)
		assert.Equal(t, has, true)

		err = db.CleanSNDerivator()
		assert.Equal(t, err, nil)
		has, err = db.HasSerialNumber(tokenID, snd2, 0)
		assert.Equal(t, err, nil)
		assert.Equal(t, has, false)
	} else {
		t.Error("DB is not open")
	}
}

// Fee estimator
func TestDb_StoreFeeEstimator(t *testing.T) {
	if db != nil {
		feeEstimatorData := []byte{1, 2, 3, 4, 5}
		err := db.StoreFeeEstimator(feeEstimatorData, 1)
		assert.Equal(t, err, nil)
		data, err := db.GetFeeEstimator(1)
		assert.Equal(t, data, feeEstimatorData)
		assert.Equal(t, err, nil)
		db.CleanFeeEstimator()
		_, err = db.GetFeeEstimator(1)
		assert.NotEqual(t, err, nil)
	} else {
		t.Error("DB is not open")
	}
}
