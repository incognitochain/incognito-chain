package pruner

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/trie"
)

func pruneByHeight(db incdb.Database, shardID int, stateBloom *trie.StateBloom, height uint64) (uint64, uint64, error) {
	h, err := rawdbv2.GetFinalizedShardBlockHashByIndex(db, byte(shardID), height)
	if err != nil {
		return 0, 0, err
	}
	data, err := rawdbv2.GetShardRootsHash(db, byte(shardID), *h)
	sRH := &blockchain.ShardRootHashv2{}
	if err = json.Unmarshal(data, sRH); err != nil {
		return 0, 0, err
	}
	return pruneTxStateDB(db, stateBloom, sRH)

}

func pruneTxStateDB(db incdb.Database, stateBloom *trie.StateBloom, sRH *blockchain.ShardRootHashv2) (uint64, uint64, error) {
	sDB, err := statedb.NewWithPrefixTrie(sRH.TransactionStateDBRootHash.GetRootHash(), statedb.NewDatabaseAccessWarper(db))
	if err != nil {
		return 0, 0, nil
	}
	keysShouldBeRemoved, err := sDB.Retrieve(false, true, stateBloom, false)
	if err != nil {
		return 0, 0, err
	}
	storage, node, err := removeNodes(db, keysShouldBeRemoved)
	if err != nil {
		return 0, 0, err
	}
	return storage, node, nil
}

// removeNodes after removeNodes keys map will be reset to empty value
func removeNodes(db incdb.Database, keysShouldBeRemoved map[common.Hash]struct{}) (uint64, uint64, error) {
	var storage, count uint64

	batch := db.NewBatch()
	for key := range keysShouldBeRemoved {
		temp, _ := db.Get(key.Bytes())
		storage += uint64(len(temp) + len(key.Bytes()))
		if err := batch.Delete(key.Bytes()); err != nil {
			return 0, 0, err
		}
		if batch.ValueSize() >= incdb.IdealBatchSize {
			if err := batch.Write(); err != nil {
				return 0, 0, err
			}
			batch.Reset()
		}
		count++
	}
	if batch.ValueSize() > 0 {
		batch.Write()
		batch.Reset()
	}

	return storage, count, nil
}
