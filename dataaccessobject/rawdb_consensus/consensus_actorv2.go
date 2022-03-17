package rawdb_consensus

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
	"strings"
)

func GetAllProposeHistory(db incdb.Database, chainID int) (map[int64]struct{}, error) {

	res := make(map[int64]struct{})

	prefix := GetProposeHistoryPrefix(chainID)
	it := db.NewIteratorWithPrefix(prefix)
	defer it.Release()
	for it.Next() {
		key := make([]byte, len(it.Key()))
		copy(key, it.Key())
		keys := strings.Split(string(key), string(splitter))
		tempTimeSlot := keys[2]
		timeSlot, err := common.BytesToUint64([]byte(tempTimeSlot))
		if err != nil {
			return nil, err
		}
		res[int64(timeSlot)] = struct{}{}
	}

	return res, nil
}

func StoreProposeHistory(db incdb.Database, chainID int, currentTimeSlot int64) error {

	key := GetProposeHistoryKey(chainID, uint64(currentTimeSlot))
	value := []byte{}

	if err := db.Put(key, value); err != nil {
		return err
	}

	return nil
}

func DeleteProposeHistory(db incdb.Database, chainID int, currentTimeSlot int64) error {

	key := GetProposeHistoryKey(chainID, uint64(currentTimeSlot))

	if err := db.Delete(key); err != nil {
		return err
	}

	return nil
}

func GetAllReceiveBlockByHeight(db incdb.Database, chainID int) (map[uint64][]byte, map[uint64]int, error) {

	res := make(map[uint64][]byte)
	res2 := make(map[uint64]int)

	prefix := GetReceiveBlockByHeightPrefix(chainID)
	it := db.NewIteratorWithPrefix(prefix)
	defer it.Release()
	for it.Next() {
		key := make([]byte, len(it.Key()))
		copy(key, it.Key())
		keys := strings.Split(string(key), string(splitter))
		tempHeight := keys[2]
		height, err := common.BytesToUint64([]byte(tempHeight))
		if err != nil {
			return nil, nil, err
		}
		value := make([]byte, len(it.Value()))
		copy(value, it.Value())
		values := strings.Split(string(value), string(splitter))
		res[height] = []byte(values[1])
		res2[height] = common.BytesToInt([]byte(values[0]))
	}

	return res, res2, nil
}

func StoreReceiveBlockByHeight(db incdb.Database, chainID int, height uint64, numberOfBlock int, value []byte) error {

	key := GetReceiveBlockByHeightKey(chainID, height)
	value = getReceiveBlockByHeightValue(numberOfBlock, value)

	if err := db.Put(key, value); err != nil {
		return err
	}

	return nil
}

func getReceiveBlockByHeightValue(numberOfBlock int, value []byte) []byte {
	res := common.IntToBytes(numberOfBlock)
	res = append(res, splitter...)
	res = append(res, value...)
	return res
}

func DeleteReceiveBlockByHeight(db incdb.Database, chainID int, height uint64) error {

	key := GetReceiveBlockByHeightKey(chainID, height)

	if err := db.Delete(key); err != nil {
		return err
	}

	return nil
}

func GetAllReceiveBlockByHash(db incdb.Database, chainID int) (map[string][]byte, error) {

	res := make(map[string][]byte)

	prefix := GetReceiveBlockByHashPrefix(chainID)
	it := db.NewIteratorWithPrefix(prefix)
	defer it.Release()
	for it.Next() {
		key := make([]byte, len(it.Key()))
		copy(key, it.Key())
		keys := strings.Split(string(key), string(splitter))
		blockHash := keys[2]
		value := make([]byte, len(it.Value()))
		copy(value, it.Value())
		res[blockHash] = value
	}

	return res, nil
}

func StoreReceiveBlockByHash(db incdb.Database, chainID int, blockHash string, value []byte) error {

	key := GetReceiveBlockByHashKey(chainID, blockHash)

	if err := db.Put(key, value); err != nil {
		return err
	}

	return nil
}

func DeleteReceiveBlockByHash(db incdb.Database, chainID int, blockHash string) error {

	key := GetReceiveBlockByHashKey(chainID, blockHash)

	if err := db.Delete(key); err != nil {
		return err
	}

	return nil
}

func GetAllVoteHistory(db incdb.Database, chainID int) (map[uint64][]byte, error) {

	res := make(map[uint64][]byte)

	prefix := GetVoteHistoryPrefix(chainID)
	it := db.NewIteratorWithPrefix(prefix)
	defer it.Release()
	for it.Next() {
		key := make([]byte, len(it.Key()))
		copy(key, it.Key())
		keys := strings.Split(string(key), string(splitter))
		tempHeight := keys[2]
		height, err := common.BytesToUint64([]byte(tempHeight))
		if err != nil {
			return nil, err
		}
		value := make([]byte, len(it.Value()))
		copy(value, it.Value())
		res[height] = value
	}

	return res, nil
}

func StoreVoteHistory(db incdb.Database, chainID int, height uint64, value []byte) error {

	key := GetVoteHistoryKey(chainID, height)

	if err := db.Put(key, value); err != nil {
		return err
	}

	return nil
}

func DeleteVoteHistory(db incdb.Database, chainID int, height uint64) error {

	key := GetVoteHistoryKey(chainID, height)

	if err := db.Delete(key); err != nil {
		return err
	}

	return nil
}
