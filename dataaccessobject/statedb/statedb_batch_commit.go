package statedb

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
)

type TrieCommitEnvironment struct {
	ShardID            byte
	RawDB              incdb.KeyValueWriter
	NewPivotBlockHash  common.Hash
	CurrentBlockHeight uint64
	CurrentBlockHash   common.Hash
	IsWriteToDisk      bool
	IsForceWrite       bool
}

func NewTrieCommitEnvironment(shardID byte, rawDB incdb.KeyValueWriter, newPivotBlockHash common.Hash, currentBlockHeight uint64, currentBlockHash common.Hash, isWriteToDisk bool, isForceWrite bool) *TrieCommitEnvironment {
	return &TrieCommitEnvironment{ShardID: shardID, RawDB: rawDB, NewPivotBlockHash: newPivotBlockHash, CurrentBlockHeight: currentBlockHeight, CurrentBlockHash: currentBlockHash, IsWriteToDisk: isWriteToDisk, IsForceWrite: isForceWrite}
}

func NewForceTrieCommitEnvironment(shardID byte, rawDB incdb.KeyValueWriter, newPivotBlockHash common.Hash) *TrieCommitEnvironment {
	return &TrieCommitEnvironment{ShardID: shardID, RawDB: rawDB, NewPivotBlockHash: newPivotBlockHash, IsForceWrite: true}
}

var (
	fullSyncPivotBlockKey = []byte("LiteStateDB-Pivot-BlockHash-")
)

func GetCommitPivotKey(name string) []byte {
	temp := make([]byte, len(fullSyncPivotBlockKey))
	copy(temp, fullSyncPivotBlockKey)
	return append(temp, name...)
}

func GetLatestPivotCommit(reader incdb.KeyValueReader, pivotName string) (string, error) {
	value, err := reader.Get(GetCommitPivotKey(pivotName))
	if err != nil {
		return "", err
	}
	return string(value), err
}

func GetLatestPivotCommitInfo(db incdb.KeyValueReader, pivotName string) (*common.Hash, uint64, error) {
	var pivotIndex = uint64(0)

	var pivotRoot *common.Hash
	var err error
	pivotCommit, _ := GetLatestPivotCommit(db, pivotName)
	if len(pivotCommit) != 0 {
		pivotCommitInfo := strings.Split(pivotCommit, "-")

		if len(pivotCommitInfo) == 2 {
			pivotRoot, err = common.Hash{}.NewHashFromStr(pivotCommitInfo[0])
			pivotIndex, err = strconv.ParseUint(pivotCommitInfo[1], 10, 64)
			if err != nil {
				return nil, 0, errors.New("Cannot parse " + pivotCommit)
			}
		} else {
			return nil, 0, errors.New("Cannot parse " + pivotCommit)
		}
	}
	return pivotRoot, pivotIndex, nil
}

func StoreLatestPivotCommit(writer incdb.KeyValueWriter, pivotName, pivotInfo string) error {
	return writer.Put(GetCommitPivotKey(pivotName), []byte(pivotInfo))
}

func (stateDB *StateDB) GetStateObjectFromBranch(
	ffIndex uint64,
	pivotIndex int,
) ([]map[common.Hash]StateObject, error) {

	stateObjectSeries := []map[common.Hash]StateObject{}

	for {
		if ffIndex == uint64(pivotIndex) {
			break
		}
		data, err := stateDB.batchCommitConfig.flatFile.Read(ffIndex)
		if err != nil {
			return nil, err
		}
		var prevIndex int64
		err = binary.Read(bytes.NewBuffer(data[:8]), binary.LittleEndian, &prevIndex)
		if err != nil {
			return nil, err
		}
		stateObjects, err := MapByteDeserialize(stateDB, data[8:])
		if err != nil {
			return nil, err
		}
		//append revert
		tmp := []map[common.Hash]StateObject{}
		tmp = append(tmp, stateObjects)
		stateObjectSeries = append(tmp, stateObjectSeries...)
		// ffIndex = prevIndex
		if ffIndex == 0 {
			break
		}
		ffIndex = uint64(prevIndex)
	}

	return stateObjectSeries, nil
}
