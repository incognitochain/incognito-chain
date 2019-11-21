package rawdb_test

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/core/rawdb"
	"github.com/incognitochain/incognito-chain/incdb"
	_ "github.com/incognitochain/incognito-chain/incdb/lvdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/stretchr/testify/assert"
)

var db incdb.Database

var _ = func() (_ struct{}) {
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_")
	if err != nil {
		log.Fatalf("failed to create temp dir: %+v", err)
	}
	log.Println(dbPath)
	db, err = incdb.Open("leveldb", dbPath)
	if err != nil {
		log.Fatalf("could not open db path: %s, %+v", dbPath, err)
	}
	incdb.Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()

func TestDb_Setup(t *testing.T) {
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_")
	if err != nil {
		t.Fatalf("failed to create temp dir: %+v", err)
	}
	t.Log(dbPath)
	db, err := incdb.Open("leveldb", dbPath)
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
		has, err := db.Has([]byte("a"))
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, has, true)

		err = db.Delete([]byte("a"))
		assert.Equal(t, nil, err)
		err = db.Delete([]byte("b"))
		assert.Equal(t, nil, err)
		has, err = db.Has([]byte("a"))
		assert.Equal(t, err, nil)
		assert.Equal(t, has, false)

		batchData := []incdb.BatchData{}
		batchData = append(batchData, incdb.BatchData{
			Key:   []byte("abc1"),
			Value: []byte("abc1"),
		})
		batchData = append(batchData, incdb.BatchData{
			Key:   []byte("abc2"),
			Value: []byte("abc2"),
		})
		err = db.PutBatch(batchData)
		assert.Equal(t, err, nil)
		v, err := db.Get([]byte("abc2"))
		assert.Equal(t, err, nil)
		assert.Equal(t, "abc2", string(v))
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
		err := rawdb.StoreShardBlock(db, block, *block.Hash(), block.Header.ShardID, nil)
		assert.Equal(t, err, nil)

		// test Fetch block
		fail, err := rawdb.FetchBlock(db, common.Hash{})
		assert.NotEqual(t, nil, err)
		assert.Equal(t, 0, len(fail))
		_, err = rawdb.FetchBlock(db, *block.Hash())
		assert.Equal(t, err, nil)
		// TODO
		//blockNew := blockchain.ShardBlock{}
		//err = json.Unmarshal(blockInBytes, &blockNew)
		//assert.IsEqualCommitteeKey(t, err, nil)
		//assert.IsEqualCommitteeKey(t, blockNew.Hash(), block.Hash())

		// has block
		has, err := rawdb.HasBlock(db, *block.Hash())
		assert.Equal(t, err, nil)
		assert.Equal(t, has, true)

		// delete block
		//err = db.DeleteBlock(*blockNew.Hash(), blockNew.Header.Height, blockNew.Header.ShardID)
		//assert.IsEqualCommitteeKey(t, err, nil)
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
		err := rawdb.StoreShardBlockIndex(db, *block.Hash(), block.Header.Height, block.Header.ShardID, nil)
		assert.Equal(t, err, nil)

		// test GetIndexOfBlock
		blockHeigh, shardID, err := rawdb.GetIndexOfBlock(db, *block.Hash())
		assert.Equal(t, err, nil)
		assert.Equal(t, blockHeigh, uint64(1))
		assert.Equal(t, shardID, uint8(3))

		// GetBlockByIndex
		hash, err := rawdb.GetBlockByIndex(db, 1, 3)
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
		err := rawdb.StoreBeaconBlock(db, beaconBlock, *beaconBlock.Hash(), nil)
		assert.Equal(t, err, nil)

		// test Fetch block
		blockInBytes, err := rawdb.FetchBeaconBlock(db, *beaconBlock.Hash())
		assert.Equal(t, err, nil)
		blockNew := blockchain.BeaconBlock{}
		err = json.Unmarshal(blockInBytes, &blockNew)
		assert.Equal(t, err, nil)
		assert.Equal(t, blockNew.Hash(), beaconBlock.Hash())

		// has block
		has, err := rawdb.HasBeaconBlock(db, *beaconBlock.Hash())
		assert.Equal(t, err, nil)
		assert.Equal(t, has, true)

		// delete block
		err = rawdb.DeleteBeaconBlock(db, *blockNew.Hash(), blockNew.Header.Height)
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
		err := rawdb.StoreBeaconBlockIndex(db, *beaconBlock.Hash(), beaconBlock.Header.Height)
		assert.Equal(t, err, nil)

		// test GetIndexOfBlock
		blockHeigh, err := rawdb.GetIndexOfBeaconBlock(db, *beaconBlock.Hash())
		assert.Equal(t, err, nil)
		assert.Equal(t, blockHeigh, uint64(1))

		// GetBlockByIndex
		hash, err := rawdb.GetBeaconBlockHashByIndex(db, 1)
		assert.Equal(t, hash.String(), beaconBlock.Hash().String())

	} else {
		t.Error("DB is not open")
	}
}

//Crossshard
func TestDb_StoreCrossShardNextHeight(t *testing.T) {
	if db != nil {
		err := rawdb.StoreCrossShardNextHeight(db, 0, 1, 1, 2)
		assert.Equal(t, err, nil)

		err = rawdb.StoreCrossShardNextHeight(db, 0, 1, 2, 0)
		assert.Equal(t, err, nil)

		val, err := rawdb.FetchCrossShardNextHeight(db, 0, 1, 1)
		assert.Equal(t, err, nil)
		assert.Equal(t, uint64(val), uint64(2))

		err = rawdb.StoreCrossShardNextHeight(db, 0, 1, 2, 4)
		assert.Equal(t, err, nil)

		err = rawdb.StoreCrossShardNextHeight(db, 0, 1, 4, 0)
		assert.Equal(t, err, nil)

		err = rawdb.RestoreCrossShardNextHeights(db, 0, 1, 2)
		assert.Equal(t, err, nil)

		val, err = rawdb.FetchCrossShardNextHeight(db, 0, 1, 2)
		assert.Equal(t, err, nil)
		assert.Equal(t, uint64(val), uint64(0))

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
		err := rawdb.StoreTransactionIndex(db, *block.Body.Transactions[1].Hash(), *block.Hash(), 1, nil)
		assert.Equal(t, err, nil)

		blockHash, index, err := rawdb.GetTransactionIndexById(db, *block.Body.Transactions[1].Hash())
		if err != nil && err.(*rawdb.RawdbError) != nil {
			t.Error(err)
		}
		assert.Equal(t, blockHash, *block.Hash())
		assert.Equal(t, index, 1)

		err = rawdb.DeleteTransactionIndex(db, *block.Body.Transactions[1].Hash())
		assert.Equal(t, err, nil)

	} else {
		t.Error("DB is not open")
	}
}

// Best state of Prev
func TestDb_StorePrevBestState(t *testing.T) {
	if db != nil {
		bestState := blockchain.BestState{
			Beacon: &blockchain.BeaconBestState{
				Epoch: 100,
			},
		}
		tempMarshal, err := json.Marshal(bestState.Beacon)
		assert.Equal(t, err, nil)
		err = rawdb.StorePrevBestState(db, tempMarshal, true, 0)
		assert.Equal(t, err, nil)

		beaconInBytes, err := rawdb.FetchPrevBestState(db, true, 0)
		assert.Equal(t, err, nil)
		temp := blockchain.BeaconBestState{}
		json.Unmarshal(beaconInBytes, &temp)
		assert.Equal(t, bestState.Beacon.Epoch, temp.Epoch)
		err = rawdb.CleanBackup(db, true, 0)
		_, err = rawdb.FetchPrevBestState(db, true, 0)
		assert.NotEqual(t, err, nil)
	} else {
		t.Error("DB is not open")
	}
}

// Best state of shard chain
func TestDb_StoreShardBestState(t *testing.T) {
	if db != nil {
		besState := blockchain.BestState{
			Shard: make(map[byte]*blockchain.ShardBestState),
		}
		bestStateShard := blockchain.ShardBestState{
			Epoch: 100,
		}
		besState.Shard[0] = &bestStateShard
		err := rawdb.StoreShardBestState(db, bestStateShard, 0, nil)
		assert.Equal(t, err, nil)

		temp, err := rawdb.FetchShardBestState(db, 0)
		assert.Equal(t, err, nil)
		tempObject := blockchain.ShardBestState{}
		err = json.Unmarshal(temp, &tempObject)
		assert.Equal(t, err, nil)
		assert.Equal(t, tempObject.Epoch, bestStateShard.Epoch)

		err = rawdb.CleanShardBestState(db)
		assert.Equal(t, err, nil)
		_, err = rawdb.FetchShardBestState(db, 0)
		assert.NotEqual(t, err, nil)
	} else {
		t.Error("DB is not open")
	}
}

// Best state of beacon chain
func TestDb_StoreBeaconBestState(t *testing.T) {
	if db != nil {
		bestState := blockchain.BestState{
			Beacon: &blockchain.BeaconBestState{
				Epoch: 100,
			},
		}
		err := rawdb.StoreBeaconBestState(db, bestState, nil)
		assert.Equal(t, err, nil)
		temp, err := rawdb.FetchBeaconBestState(db)
		assert.Equal(t, err, nil)
		tempObject := blockchain.BestState{}
		err = json.Unmarshal(temp, &tempObject)
		assert.Equal(t, err, nil)
		assert.Equal(t, tempObject.Beacon.Epoch, bestState.Beacon.Epoch)

		err = rawdb.CleanBeaconBestState(db)
		assert.Equal(t, err, nil)
		_, err = rawdb.FetchBeaconBestState(db)
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
			Beacon: &blockchain.BeaconBestState{
				Epoch:          100,
				ShardCommittee: make(map[byte][]incognitokey.CommitteePublicKey),
			},
		}
		bestState.Beacon.ShardCommittee[0] = make([]incognitokey.CommitteePublicKey, 0)
		bestState.Beacon.ShardCommittee[0] = append(bestState.Beacon.ShardCommittee[0], incognitokey.CommitteePublicKey{MiningPubKey: map[string][]byte{common.BlsConsensus: []byte("committee1")}})
		bestState.Beacon.ShardCommittee[0] = append(bestState.Beacon.ShardCommittee[0], incognitokey.CommitteePublicKey{MiningPubKey: map[string][]byte{common.BlsConsensus: []byte("committee2")}})
		err := rawdb.StoreShardCommitteeByHeight(db, block.Header.Height, bestState.Beacon.GetShardCommittee())
		assert.Equal(t, err, nil)

		shardCommittee := make(map[byte][]incognitokey.CommitteePublicKey)
		data, err := rawdb.FetchShardCommitteeByHeight(db, block.Header.Height)
		assert.Equal(t, err, nil)
		err = json.Unmarshal(data, &shardCommittee)
		assert.Equal(t, err, nil)
		assert.Equal(t, shardCommittee[0][0].MiningPubKey[common.BlsConsensus], []byte("committee1"))
		assert.Equal(t, shardCommittee[0][1].MiningPubKey[common.BlsConsensus], []byte("committee2"))

		has, err := rawdb.HasShardCommitteeByHeight(db, block.Header.Height)
		assert.Equal(t, has, true)
		assert.Equal(t, err, nil)

		err = rawdb.DeleteCommitteeByHeight(db, block.Header.Height)
		assert.Equal(t, err, nil)

		has, err = rawdb.HasShardCommitteeByHeight(db, block.Header.Height)
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
		err := rawdb.StoreSerialNumbers(db, tokenID, serialNumber, 0)
		assert.Equal(t, err, nil)

		has, err := rawdb.HasSerialNumber(db, tokenID, ser1, 0)
		assert.Equal(t, err, nil)
		assert.Equal(t, has, true)

		err = rawdb.BackupSerialNumbersLen(db, tokenID, 0)
		assert.Equal(t, err, nil)

		err = rawdb.RestoreSerialNumber(db, tokenID, 0, serialNumber)
		assert.Equal(t, err, nil)
		has, err = rawdb.HasSerialNumber(db, tokenID, ser1, 0)
		assert.Equal(t, err, nil)
		assert.Equal(t, has, false)

		err = rawdb.StoreSerialNumbers(db, tokenID, serialNumber, 0)
		assert.Equal(t, err, nil)
		has, err = rawdb.HasSerialNumber(db, tokenID, ser1, 0)
		assert.Equal(t, err, nil)
		assert.Equal(t, has, true)

		err = rawdb.CleanSerialNumbers(db)
		assert.Equal(t, err, nil)
		has, err = rawdb.HasSerialNumber(db, tokenID, ser1, 0)
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

		err := rawdb.StoreCommitments(db, tokenID, publicKey.GetBytes(), committments, 0)
		assert.Equal(t, err, nil)

		has, err := rawdb.HasCommitment(db, tokenID, cm1, 0)
		assert.Equal(t, err, nil)
		assert.Equal(t, has, true)

		has, err = rawdb.HasCommitmentIndex(db, tokenID, 0, 0)
		assert.Equal(t, err, nil)
		assert.Equal(t, has, true)

		len, err := rawdb.GetCommitmentLength(db, tokenID, 0)
		assert.Equal(t, err, nil)
		assert.Equal(t, len.Int64(), int64(2))

		temp, err := rawdb.GetCommitmentByIndex(db, tokenID, 1, 0)
		assert.Equal(t, err, nil)
		assert.Equal(t, temp, cm2)

		index, err := rawdb.GetCommitmentIndex(db, tokenID, cm1, 0)
		assert.Equal(t, err, nil)
		assert.Equal(t, index.Uint64(), uint64(0))

		err = rawdb.BackupCommitmentsOfPublicKey(db, tokenID, 0, publicKey.GetBytes())
		assert.Equal(t, err, nil)

		err = rawdb.RestoreCommitmentsOfPubkey(db, tokenID, 0, publicKey.GetBytes(), committments)
		assert.Equal(t, err, nil)
		has, err = rawdb.HasSerialNumber(db, tokenID, cm1, 0)
		assert.Equal(t, err, nil)
		assert.Equal(t, has, false)

		err = rawdb.CleanCommitments(db)
		assert.Equal(t, err, nil)
		has, err = rawdb.HasCommitment(db, tokenID, cm1, 0)
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
		cm3 := []byte{0, 3}
		outputCoins = append(outputCoins, cm1)
		outputCoins = append(outputCoins, cm2)
		outputCoins = append(outputCoins, cm3)
		tokenID := common.Hash{}
		publicKey := common.Hash{}
		err := rawdb.StoreOutputCoins(db, tokenID, publicKey.GetBytes(), outputCoins, 1)
		assert.Equal(t, err, nil)

		data, err := rawdb.GetOutcoinsByPubkey(db, tokenID, publicKey.GetBytes(), 1)
		assert.Equal(t, err, nil)
		assert.NotEqual(t, 2, len(data))
		assert.Equal(t, 3, len(data))
		err = rawdb.DeleteOutputCoin(db, tokenID, publicKey.GetBytes(), outputCoins, 1)
		assert.Equal(t, err, nil)

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

		err := rawdb.StoreSNDerivators(db, tokenID, snd)
		assert.Equal(t, err, nil)

		has, err := rawdb.HasSNDerivator(db, tokenID, snd1)
		assert.Equal(t, nil, err)
		assert.Equal(t, true, has)

		err = rawdb.CleanSNDerivator(db)
		assert.Equal(t, err, nil)
		has, err = rawdb.HasSerialNumber(db, tokenID, snd2, 1)
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
		err := rawdb.StoreFeeEstimator(db, feeEstimatorData, 1)
		assert.Equal(t, err, nil)
		data, err := rawdb.GetFeeEstimator(db, 1)
		assert.Equal(t, data, feeEstimatorData)
		assert.Equal(t, err, nil)
		rawdb.CleanFeeEstimator(db)
		_, err = rawdb.GetFeeEstimator(db, 1)
		assert.NotEqual(t, err, nil)
	} else {
		t.Error("DB is not open")
	}
}

// Custom token
func TestDb_StoreCustomToken(t *testing.T) {
	tokenID := common.Hash{}
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8}

	err := rawdb.StorePrivacyToken(db, tokenID, data)
	assert.Equal(t, err, nil)

	dataTemp, err := rawdb.ListPrivacyToken(db)
	assert.Equal(t, err, nil)
	assert.Equal(t, len(dataTemp), 1)

	err = rawdb.DeletePrivacyToken(db, tokenID)
	assert.Equal(t, err, nil)

	dataTemp, err = rawdb.ListPrivacyToken(db)
	assert.Equal(t, err, nil)
	assert.Equal(t, len(dataTemp), 0)

	err = rawdb.StorePrivacyToken(db, tokenID, data)
	assert.Equal(t, err, nil)

	has := rawdb.PrivacyTokenIDExisted(db, tokenID)
	assert.Equal(t, true, has)

	err = rawdb.StorePrivacyTokenTx(db, tokenID, 0, 1, 0, data)
	assert.Equal(t, err, nil)

	temp, err := rawdb.PrivacyTokenTxs(db, tokenID)
	assert.Equal(t, 1, len(temp))

	err = rawdb.DeletePrivacyTokenTx(db, tokenID, 0, 0, 1)
	assert.Equal(t, err, nil)

	temp, err = rawdb.PrivacyTokenTxs(db, tokenID)
	assert.Equal(t, 0, len(temp))

	// custom token payment address
}

func TestDb_StorePrivacyCustomTokenCrossShard(t *testing.T) {
	tokenID := common.Hash{}
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8}
	err := rawdb.StorePrivacyTokenCrossShard(db, tokenID, data)
	assert.Equal(t, nil, err)

	result, err := rawdb.ListPrivacyTokenCrossShard(db)
	assert.Equal(t, nil, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, data, result[0])

	has := rawdb.PrivacyTokenIDCrossShardExisted(db, tokenID)
	assert.Equal(t, true, has)

	err = rawdb.DeletePrivacyTokenCrossShard(db, tokenID)
	assert.Equal(t, nil, err)
}

func TestDb_StoreIncomingCrossShard(t *testing.T) {
	err := rawdb.StoreIncomingCrossShard(db, 0, 1, 1000, common.Hash{}, nil)
	assert.Equal(t, nil, err)

	err = rawdb.HasIncomingCrossShard(db, 0, 1, common.Hash{})
	assert.Equal(t, nil, err)

	height, err := rawdb.GetIncomingCrossShard(db, 0, 1, common.Hash{})
	assert.Equal(t, nil, err)
	assert.Equal(t, uint64(1000), uint64(height))

	err = rawdb.DeleteIncomingCrossShard(db, 0, 1, common.Hash{})
	assert.Equal(t, nil, err)
}

func TestDb_StoreAcceptedShardToBeacon(t *testing.T) {
	err := rawdb.StoreAcceptedShardToBeacon(db, 0, 1000, common.Hash{})
	assert.Equal(t, nil, err)

	err = rawdb.HasAcceptedShardToBeacon(db, 0, common.Hash{})
	assert.Equal(t, nil, err)

	err = rawdb.HasAcceptedShardToBeacon(db, 1, common.Hash{})
	assert.NotEqual(t, nil, err)

	height, err := rawdb.GetAcceptedShardToBeacon(db, 0, common.Hash{})
	assert.Equal(t, nil, err)
	assert.Equal(t, uint64(1000), uint64(height))

	height, err = rawdb.GetAcceptedShardToBeacon(db, 1, common.Hash{})
	assert.NotEqual(t, nil, err)

	err = rawdb.DeleteAcceptedShardToBeacon(db, 0, common.Hash{})
	assert.Equal(t, nil, err)
	err = rawdb.HasAcceptedShardToBeacon(db, 0, common.Hash{})
	assert.NotEqual(t, nil, err)
}
