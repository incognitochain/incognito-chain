package statedb

import (
	"bytes"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/flatfile"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/pkg/errors"
	"log"
	"path"
)

type LiteStateDB struct {
	flatfile      *flatfile.FlatFileManager
	db            incdb.Database
	dbPrefix      string
	headStateNode *StateNode
	stateDB       *StateDB
}

func NewLiteStateDB(rootDir string, dbName string, rebuildRoot string, db incdb.Database) (*StateDB, error) {
	rebuildRootHash, _, rebuildRootIndex, err := getRebuildRootInfo(rebuildRoot)
	stateDB := &StateDB{
		mode:                common.STATEDB_LITE_MODE,
		stateObjects:        make(map[common.Hash]StateObject),
		stateObjectsPending: make(map[common.Hash]struct{}),
		stateObjectsDirty:   make(map[common.Hash]struct{}),
	}
	stateNode := NewStateNode()
	stateNode.aggregateHash = nil
	ffDir := path.Join(rootDir, fmt.Sprintf("lite_%v", dbName))

	ff, err := flatfile.NewFlatFile(ffDir, 5000)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot create flatfile for litestatedb")
	}
	dbPrefix := "litestateDB-" + dbName

	stateDB.liteStateDB = &LiteStateDB{
		ff,
		db,
		dbPrefix,
		stateNode,
		stateDB,
	}

	//if init from empty root
	if rebuildRoot == "" || rebuildRootHash.String() == common.EmptyRoot.String() {
		return stateDB, nil
	}
	//else rebuild
	err = stateDB.liteStateDB.restoreStateNode(rebuildRootHash, rebuildRootIndex)

	return stateDB, err
}

func (stateDB *LiteStateDB) GetFinalKey() []byte {
	final := "final-" + stateDB.dbPrefix
	return []byte(final)
}

func (stateDB *LiteStateDB) Copy() *LiteStateDB {
	cpy := *stateDB.headStateNode
	cpy.stateObjects = make(map[common.Hash]StateObject)

	return &LiteStateDB{
		stateDB.flatfile,
		stateDB.db,
		stateDB.dbPrefix,
		&cpy,
		stateDB.stateDB,
	}
}

func (stateDB *LiteStateDB) Commit() (common.Hash, error) {
	if stateDB.headStateNode.aggregateHash != nil {
		return common.Hash{}, errors.New("Cannot commit twice")
	}

	h, err := stateDB.headStateNode.Commit()

	if h != nil {
		data, err := stateDB.headStateNode.Serialize()
		if err != nil {
			return common.Hash{}, err
		}
		ffIndex, err := stateDB.flatfile.Append(data)
		if err != nil {
			return common.Hash{}, err
		}
		stateDB.headStateNode.ffIndex = uint64(ffIndex)
		return *h, nil
	}

	return common.Hash{}, err

}

func (stateDB *LiteStateDB) Finalized(stateNodeHash common.Hash) error {

	//write finalize Key Value to disk
	stateNode := stateDB.headStateNode.previousLink //current headstatenode dont have aggregatehash, as soon as we commit to calculate agghash, we create new head
	for {
		if stateNode == nil {
			break
		}
		if stateNode.aggregateHash.String() == stateNodeHash.String() {
			err := stateNode.FlushFinalizedToDisk(stateDB.db, stateDB)
			if err != nil {
				return err
			}
			stateNode.previousLink = nil
			err = StoreLatestPivotCommit(stateDB.db, stateDB.dbPrefix, fmt.Sprintf("%v-%v", stateNodeHash.String(), stateNode.ffIndex))
			return err
		}
		stateNode = stateNode.previousLink
	}

	return errors.New("Cannot find finalized hash!")
}

func (stateDB *LiteStateDB) restoreStateNode(root common.Hash, ffIndex int) error {
	//get final commit
	pivotRootHash, pivotRootIndex, err := GetLatestPivotCommitInfo(stateDB.db, stateDB.dbPrefix)
	if err != nil {
		return err
	}

	if ffIndex < pivotRootIndex {
		return errors.New("Rebuild from root that before pivot point")
	}

	dataChan, errChan, cancelReadStateNode := stateDB.flatfile.ReadRecently(uint64(ffIndex))
	stateNodeMap := map[common.Hash]*StateNode{}
	prevMap := map[common.Hash]*common.Hash{}
	for {
		select {
		case stateByte := <-dataChan:
			if len(stateByte) == 0 {
				e := common.Hash{}
				if pivotRootHash.String() == e.String() {
					goto RESTORE_SUCCESS
				} else {
					return errors.New("Cannot rebuild")
				}
			}

			stateNode, prevHash, err := stateDB.stateDB.DeSerializeFromStateNodeData(stateByte)
			if err != nil {
				return err
			}

			stateNodeMap[stateNode.GetHash()] = stateNode
			prevMap[stateNode.GetHash()] = prevHash

			if stateNode.aggregateHash.String() == pivotRootHash.String() {
				cancelReadStateNode()
				goto RESTORE_SUCCESS
			}
		case <-errChan:
			return errors.New("Read data return err")
		}

	}
RESTORE_SUCCESS:
	for hash, stateNode := range stateNodeMap {
		if prevMap[hash] != nil {
			stateNode.SetPreviousLink(stateNodeMap[*prevMap[hash]])
		}
	}

	//point head to root
	for nodeHash, state := range stateNodeMap {
		if bytes.Equal(nodeHash.Bytes(), root.Bytes()) {
			stateDB.headStateNode = state
			break
		}
	}

	if stateDB.headStateNode == nil {
		return errors.New("Cannot find root head")
	}
	stateDB.headStateNode = stateNodeMap[root]
	stateDB.NewStateNode()
	return nil
}

func (stateDB *LiteStateDB) NewStateNode() {
	newNode := NewStateNode()
	newNode.previousLink = stateDB.headStateNode
	stateDB.headStateNode = newNode
}

func (stateDB *LiteStateDB) GetStateObject(objectType int, addr common.Hash) (StateObject, error) {
	//search from mem first (incase getting fresh data value)
	stateNode := stateDB.headStateNode
	for {
		if stateNode == nil {
			break
		}
		if obj, ok := stateNode.stateObjects[addr]; ok {
			return obj, nil
		}
		stateNode = stateNode.previousLink
	}

	//then search from DB
	bytesResult, _ := stateDB.db.Get(append([]byte(stateDB.dbPrefix), addr.GetBytes()...))
	if len(bytesResult) > 0 {
		obj, err := newStateObjectWithValue(stateDB.stateDB, objectType, addr, bytesResult[1:])
		if err != nil {
			return nil, err
		}
		if bytesResult[0] == 1 {
			obj.MarkDelete()
		}
		return obj, nil
	}
	return nil, nil
}

func (stateDB *LiteStateDB) SetStateObject(object StateObject) {
	log.Println("Insert key", object.GetHash().String())
	if stateDB.headStateNode.aggregateHash != nil {
		log.Println("Warning: set state object after commit, will not calculate aggregate hash again, could break logic")
		panic("Set key after commit")
	}
	stateDB.headStateNode.stateObjects[object.GetHash()] = object
}

func (stateDB *LiteStateDB) MarkDeleteStateObject(objectType int, key common.Hash) bool {
	obj, _ := stateDB.GetStateObject(objectType, key)
	if obj != nil {
		obj.MarkDelete()
		stateDB.SetStateObject(obj)
		return true
	}
	return false
}

/*
Iterator Interface
*/
func (stateDB *LiteStateDB) NewIteratorwithPrefix(prefix []byte) incdb.Iterator {
	kvMap := map[string][]byte{}
	stateDB.headStateNode.replay(kvMap)

	iter := NewLiteStateDBIterator(stateDB.db, []byte(stateDB.dbPrefix), prefix, kvMap)
	return iter
}
