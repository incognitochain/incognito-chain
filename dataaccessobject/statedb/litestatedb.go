package statedb

import (
	"bytes"
	"fmt"
	"log"
	"path"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/flatfile"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/pkg/errors"
)

type LiteStateDB struct {
	flatfile      *flatfile.FlatFileManager
	db            incdb.Database
	dbPrefix      string
	headStateNode *StateNode
	stateDB       *StateDB
}

func NewLiteStateDB(rootDir string, dbName string, rebuildRoot RebuildInfo, db incdb.Database) (*StateDB, error) {
	stateDB := &StateDB{
		mode:                common.STATEDB_LITE_MODE,
		stateObjects:        make(map[common.Hash]StateObject),
		stateObjectsPending: make(map[common.Hash]struct{}),
		stateObjectsDirty:   make(map[common.Hash]struct{}),
	}

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
		nil,
		stateDB,
	}

	//if init from empty root
	if rebuildRoot.rebuildRootHash.String() == common.EmptyRoot.String() {
		stateNode := NewStateNode()
		stateNode.aggregateHash = &common.EmptyRoot
		stateDB.liteStateDB.headStateNode = stateNode
		stateDB.liteStateDB.NewStateNode()
		return stateDB, nil
	}
	//else rebuild
	err = stateDB.liteStateDB.restoreStateNode(rebuildRoot)

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

func (stateDB *LiteStateDB) Commit() (common.Hash, uint64, error) {
	if stateDB.headStateNode.aggregateHash != nil {
		return common.Hash{}, 0, errors.New("Cannot commit twice")
	}

	h, err := stateDB.headStateNode.Commit()

	if h != nil {
		data, err := stateDB.headStateNode.Serialize()
		if err != nil {
			return common.Hash{}, 0, err
		}
		ffIndex, err := stateDB.flatfile.Append(data)
		if err != nil {
			return common.Hash{}, 0, err
		}
		return *h, ffIndex, nil
	}

	return common.Hash{}, 0, err

}

func (stateDB *LiteStateDB) Finalized(stateNodeHash common.Hash) error {

	//write finalize Key Value to disk
	// current head node is not commit yet! (When commit,it create new node, and we finalizd the old node)
	if stateDB.headStateNode.aggregateHash != nil {
		return errors.New("Not expected. Head node must not commit yet!")
	}

	stateNode := stateDB.headStateNode.previousLink
	for {
		if stateNode == nil {
			break
		}
		if stateNode.aggregateHash.String() == stateNodeHash.String() {
			err := stateNode.FlushFinalizedToDisk(stateDB.db, stateDB)
			stateNode.previousLink = nil
			if err != nil {
				return err
			}
			return err
		}
		stateNode = stateNode.previousLink
	}

	return errors.New("Cannot find finalized hash!")
}

func (stateDB *LiteStateDB) restoreStateNode(rebuildInfo RebuildInfo) error {
	root := rebuildInfo.rebuildRootHash
	ffIndex := rebuildInfo.rebuildFFIndex
	//get final commit
	pivotRootHash := rebuildInfo.pivotRootHash
	pivotRootIndex := rebuildInfo.pivotFFIndex

	if ffIndex < pivotRootIndex {
		return errors.New("Rebuild from root that before pivot point")
	}

	fmt.Println("Rebuild lite stateDB", ffIndex, pivotRootIndex, root.String(), pivotRootHash.String())
	dataChan, errChan, cancelReadStateNode := stateDB.flatfile.ReadRecently(uint64(ffIndex))
	stateNodeMap := map[common.Hash]*StateNode{}
	prevMap := map[common.Hash]*common.Hash{}
	for {
		if ffIndex < pivotRootIndex {
			break
		}
		select {
		case stateByte := <-dataChan:
			if len(stateByte) == 0 {
				if pivotRootHash.String() == common.EmptyRoot.String() {
					cancelReadStateNode()
					goto RESTORE_SUCCESS
				} else {
					return errors.New("Cannot rebuild " + root.String())
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
	if stateDB.headStateNode == nil || stateDB.headStateNode.aggregateHash.String() != root.String() {
		return errors.New("Cannot find root head " + rebuildInfo.String())
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
