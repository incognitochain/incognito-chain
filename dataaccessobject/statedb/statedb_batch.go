package statedb

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common/prque"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject"
	"github.com/incognitochain/incognito-chain/dataaccessobject/flatfile"
	"github.com/incognitochain/incognito-chain/incdb"
	errors2 "github.com/pkg/errors"
	"path"
	"time"
)

type BatchCommitConfig struct {
	triegc            *prque.Prque // Priority queue mapping block numbers to tries to gc
	blockTrieInMemory uint64
	trieNodeLimit     common.StorageSize
	trieImgsLimit     common.StorageSize
	flatFile          flatfile.FlatFile
}

func NewBatchCommitConfig(flatFile flatfile.FlatFile) *BatchCommitConfig {
	block_trie_in_memory := uint64(2000)
	trie_node_limit := common.StorageSize(1e9)
	trie_img_limit := common.StorageSize(1e9)

	if config.Param().BatchCommitSyncModeParam.BlockTrieInMemory != 0 {
		block_trie_in_memory = config.Param().BatchCommitSyncModeParam.BlockTrieInMemory
	}
	if config.Param().BatchCommitSyncModeParam.TrieNodeLimit != 0 {
		trie_node_limit = config.Param().BatchCommitSyncModeParam.TrieNodeLimit
	}
	if config.Param().BatchCommitSyncModeParam.TrieImgsLimit != 0 {
		trie_img_limit = config.Param().BatchCommitSyncModeParam.TrieImgsLimit
	}
	return &BatchCommitConfig{
		triegc:            prque.New(nil),
		flatFile:          flatFile,
		blockTrieInMemory: block_trie_in_memory,
		trieNodeLimit:     trie_node_limit,
		trieImgsLimit:     trie_img_limit,
	}
}

func InitBatchCommit(dbName string, db DatabaseAccessWarper, currentRootHash common.Hash, rebuildRootHash *RebuildInfo, pivotState *StateDB) (*StateDB, error) {

	ffDir := path.Join(db.TrieDB().GetPath(), fmt.Sprintf("batchstatedb_%v", dbName))
	ff, err := flatfile.NewFlatFile(ffDir, 5000)
	if err != nil {
		return nil, errors2.Wrap(err, "Cannot create flatfile")
	}
	batchCommitConfig := NewBatchCommitConfig(ff)

	//create new stateDB from beginning
	if rebuildRootHash == nil || rebuildRootHash.IsEmpty() {
		curRebuildInfo := NewRebuildInfo(currentRootHash, currentRootHash, int64(batchCommitConfig.flatFile.Size())-1, int64(batchCommitConfig.flatFile.Size())-1)
		stateDB, err := NewWithPrefixTrie(currentRootHash, db)
		if err != nil {
			return nil, err
		}
		stateDB.curRebuildInfo = curRebuildInfo
		stateDB.batchCommitConfig = batchCommitConfig
		return stateDB, nil
	}

	lastFFIndex := rebuildRootHash.lastFFIndex
	lastRootHash := rebuildRootHash.lastRootHash

	//else rebuild from state
	//if pivotState nil, rebuild it
	var pivotIndex = rebuildRootHash.pivotFFIndex
	if pivotState == nil {
		//rebuild pivotState
		pivotState, err = NewWithPrefixTrie(rebuildRootHash.pivotRootHash, db)
		pivotState.batchCommitConfig = NewBatchCommitConfig(ff)
		if err != nil {
			return nil, err
		}
	} else {
		pivotIndex = pivotState.curRebuildInfo.lastFFIndex //need to rebuild from this pivot rebuild ff index
	}

	//replay state object to rebuild to expected state
	buildTime := time.Now()
	stateSeries, err := pivotState.GetStateObjectFromBranch(uint64(lastFFIndex), int(pivotIndex))
	if err != nil {
		return nil, err
	}
	newState := pivotState.Copy()

	for i, stateObjects := range stateSeries {
		if i > 0 && i%1000 == 0 {
			dataaccessobject.Logger.Log.Infof("Building root: %v / %v commits", i, len(stateSeries))
		}
		//TODO: count size of object, if size > threshold then we call commit (cap node/image)
		for objKey, obj := range stateObjects {
			if err := newState.SetStateObject(obj.GetType(), objKey, obj.GetValue()); err != nil {
				return nil, err
			}
			if obj.IsDeleted() {
				newState.MarkDeleteStateObject(obj.GetType(), objKey)
			}
		}
	}
	newStateRoot := newState.IntermediateRoot(true)
	newState.db.TrieDB().Reference(newStateRoot, common.Hash{})
	newState.batchCommitConfig.triegc.Push(newStateRoot, -lastFFIndex)

	//check if we have expect rebuild root
	dataaccessobject.Logger.Log.Infof("Build root: %v (%v commits in %v) .Expected root %v", newStateRoot, len(stateSeries),
		time.Since(buildTime).Seconds(), rebuildRootHash)
	if newStateRoot.String() != lastRootHash.String() {
		return nil, errors.New("Cannot rebuild correct root")
	}

	newState.curRebuildInfo = rebuildRootHash.Copy()
	return newState, nil
}

// Commit writes the state to the underlying in-memory trie database.
func (stateDB *StateDB) checkpoint(finalizedRebuild *RebuildInfo) (*RebuildInfo, error) {
	trieDB := stateDB.db.TrieDB()
	batchCommitConfig := stateDB.batchCommitConfig

	if finalizedRebuild == nil {
		return stateDB.curRebuildInfo, nil
	}

	//ifreach #commit threshold => write to disk
	finalViewIndex := finalizedRebuild.lastFFIndex
	//pivotFFIndex := stateDB.curRebuildInfo.pivotFFIndex
	if stateDB.curRebuildInfo.pivotFFIndex+int64(stateDB.batchCommitConfig.blockTrieInMemory) < finalViewIndex {
		//write the current roothash commit nodes to disk
		rootHash := stateDB.curRebuildInfo.lastRootHash
		rootIndex := stateDB.curRebuildInfo.lastFFIndex
		if err := stateDB.db.TrieDB().Commit(rootHash, false); err != nil {
			return nil, err
		}
		if stateDB.curRebuildInfo.pivotFFIndex > 0 {
			batchCommitConfig.flatFile.Truncate(uint64(stateDB.curRebuildInfo.pivotFFIndex))
		}
		stateDB.curRebuildInfo.pivotRootHash = rootHash
		stateDB.curRebuildInfo.pivotFFIndex = rootIndex
	}
	//dereference roothash of finalized commit, for GC reduce memory
	for !batchCommitConfig.triegc.Empty() {
		oldRootHash, number := batchCommitConfig.triegc.Pop() //the largest number will be pop, (so we get the smallest ffindex, until finalIndex)
		if -number >= finalViewIndex {
			batchCommitConfig.triegc.Push(oldRootHash, number)
			break
		}
		trieDB.Dereference(oldRootHash.(common.Hash))
	}
	return stateDB.curRebuildInfo.Copy(), nil
}

// Commit writes the state to the underlying in-memory trie database.
func (stateDB *StateDB) BatchCommit(finalizedRebuild *RebuildInfo) (*RebuildInfo, error) {
	// Finalize any pending changes and merge everything into the tries
	changeObj := stateDB.stateObjects

	if len(stateDB.stateObjectsDirty) > 0 {
		stateDB.stateObjectsDirty = make(map[common.Hash]struct{})
	}

	// Write the account trie changes, measuing the amount of wasted time
	root, err := stateDB.trie.Commit(func(leaf []byte, parent common.Hash) error {
		return nil
	})

	if err != nil {
		return nil, err
	}

	if len(changeObj) == 0 {
		rebuildInfo := stateDB.curRebuildInfo.Copy()
		return rebuildInfo, nil
	}

	//cap max memory in stateDB
	//TODO: we need to test below case, when mem in stateDB exceed threshold, it will flush to disk, what is the effect here
	nodes, imgs := stateDB.db.TrieDB().Size()
	batchCommitConfig := stateDB.batchCommitConfig
	if nodes > batchCommitConfig.trieNodeLimit || imgs > batchCommitConfig.trieImgsLimit {
		stateDB.db.TrieDB().Cap(batchCommitConfig.trieNodeLimit - incdb.IdealBatchSize)
	}

	bytesSerialize := MapByteSerialize(changeObj)
	//append changeObj with previous ff index, to help retrieve stateObject of a branch

	preRebuildIndex := stateDB.curRebuildInfo.lastFFIndex
	var preRebuildIndexInt = make([]byte, 8)
	binary.LittleEndian.PutUint64(preRebuildIndexInt, uint64(preRebuildIndex))
	stateObjectsIndex, err := stateDB.batchCommitConfig.flatFile.Append(append(preRebuildIndexInt, bytesSerialize...))
	if err != nil {
		return nil, err
	}
	//log.Println("write index", stateObjectsIndex, preRebuildIndex, len(changeObj))
	//update current rebuild string
	stateDB.curRebuildInfo = NewRebuildInfo(root, stateDB.curRebuildInfo.pivotRootHash, int64(stateObjectsIndex), stateDB.curRebuildInfo.pivotFFIndex)

	// NOTE: reference current root hash, we will deference it  when we finalized pivot, so that GC will remove memory
	// push it into queue with index priority (we want to get smallest index first, so put negative to give small number have high priority)
	stateDB.db.TrieDB().Reference(root, common.Hash{})
	stateDB.batchCommitConfig.triegc.Push(root, -int64(stateObjectsIndex))
	stateDB.ClearObjects()

	return stateDB.checkpoint(finalizedRebuild)

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

const (
	TYPE_LENGTH   = 8
	STATUS_LENGTH = 2
	KEY_LENGTH    = 32
)

// ByteSerialize return a list of byte in format of:
// 0-7: for type
// 8-9: for status
// 10-41: for key
// 42-end: for value
func ByteSerialize(sob StateObject) []byte {

	res := []byte{}

	// first 8 byte for type
	var objTypeByte = make([]byte, 8)
	binary.LittleEndian.PutUint64(objTypeByte, uint64(sob.GetType()))
	res = append(res, objTypeByte[:]...)

	// next 2 to byte for status
	var isDeleteByte = make([]byte, 2)
	var bitDelete uint16
	if sob.IsDeleted() {
		bitDelete = 1
	}
	binary.LittleEndian.PutUint16(isDeleteByte, bitDelete)
	res = append(res, isDeleteByte[:]...)

	// next 32 byte is key
	res = append(res, sob.GetHash().Bytes()...)

	// the rest is value
	res = append(res, sob.GetValueBytes()...)

	return res
}

func ByteDeSerialize(stateDB *StateDB, sobByte []byte) (StateObject, error) {

	objTypeByte := sobByte[:TYPE_LENGTH]
	var objType uint64

	if err := binary.Read(bytes.NewBuffer(objTypeByte), binary.LittleEndian, &objType); err != nil {
		return nil, err
	}

	objStatusByte := sobByte[TYPE_LENGTH : TYPE_LENGTH+STATUS_LENGTH]
	var objStatusBit uint16
	var objStatus = false

	if err := binary.Read(bytes.NewBuffer(objStatusByte), binary.LittleEndian, &objStatusBit); err != nil {
		return nil, err
	}

	if objStatusBit == 1 {
		objStatus = true
	}

	objKeyByte := sobByte[TYPE_LENGTH+STATUS_LENGTH : TYPE_LENGTH+STATUS_LENGTH+KEY_LENGTH]
	objKey, err := common.Hash{}.NewHash(objKeyByte)
	if err != nil {
		return nil, err
	}

	objValue := sobByte[TYPE_LENGTH+STATUS_LENGTH+KEY_LENGTH:]

	sob, err := newStateObjectWithValue(stateDB, int(objType), *objKey, objValue)
	if err != nil {
		return nil, err
	}

	if objStatus {
		sob.MarkDelete()
	}

	return sob, nil
}

func MapByteSerialize(m map[common.Hash]StateObject) []byte {

	res := []byte{}

	for _, v := range m {
		b := ByteSerialize(v)
		offset := make([]byte, 4)
		binary.LittleEndian.PutUint32(offset, uint32(len(b)))
		res = append(res, offset...)
		res = append(res, b...)
	}

	return res
}

func MapByteDeserialize(stateDB *StateDB, data []byte) (map[common.Hash]StateObject, error) {

	m := make(map[common.Hash]StateObject)

	for len(data) > 0 {

		offsetByte := data[:4]
		offset := binary.LittleEndian.Uint32(offsetByte)
		data = data[4:]

		soByte := data[:offset]
		stateObject, err := ByteDeSerialize(stateDB, soByte)
		if err != nil {
			return m, err
		}
		data = data[offset:]

		m[stateObject.GetHash()] = stateObject
	}

	return m, nil
}
