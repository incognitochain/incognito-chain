package statedb

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/flatfile"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/pkg/errors"
	"log"
)

const PREFIX_LITESTATEDB = "litestatedb"

type LiteStateDB struct {
	flatfile      *flatfile.FlatFileManager
	db            incdb.Database
	headStateNode *StateNode
	stateDB       *StateDB
}

func NewLiteStateDB(dir string, root common.Hash, lastState common.Hash, db incdb.Database) (*StateDB, error) {
	stateDB := &StateDB{
		stateObjects:        make(map[common.Hash]StateObject),
		stateObjectsPending: make(map[common.Hash]struct{}),
		stateObjectsDirty:   make(map[common.Hash]struct{}),
	}
	stateNode := NewStateNode()
	stateNode.aggregateHash = nil
	ff, err := flatfile.NewFlatFile(dir, 5000)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot create flatfile for litestatedb")
	}

	stateDB.liteStateDB = &LiteStateDB{
		ff,
		db,
		stateNode,
		stateDB,
	}

	if root.IsZeroValue() {
		return stateDB, nil
	}

	if !root.IsZeroValue() {
		stateDB.liteStateDB.restoreStateNode(lastState, root)
	}

	return stateDB, nil
}

func (stateDB *LiteStateDB) Copy() *LiteStateDB {
	cpy := *stateDB.headStateNode
	cpy.stateObjects = make(map[common.Hash]StateObject)

	return &LiteStateDB{
		stateDB.flatfile,
		stateDB.db,
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
		_, err = stateDB.flatfile.Append(data)
		if err != nil {
			return common.Hash{}, err
		}
		return *h, nil
	}

	return common.Hash{}, err

}

func (stateDB *LiteStateDB) Finalized(dbWriter incdb.KeyValueWriter, stateNodeHash common.Hash) error {

	//write finalize Key Value to disk
	stateNode := stateDB.headStateNode.previousLink //current headstatenode dont have aggregatehash, as soon as we commit to calculate agghash, we create new head
	for {
		if stateNode == nil {
			break
		}
		if stateNode.aggregateHash.String() == stateNodeHash.String() {
			err := stateNode.FlushFinalizedToDisk(dbWriter)
			if err != nil {
				return err
			}
			stateNode.previousLink = nil
			return nil
		}
		stateNode = stateNode.previousLink
	}

	return errors.New("Cannot find finalized hash!")
}

func (stateDB *LiteStateDB) restoreStateNode(finalState common.Hash, root common.Hash) error {
	dataChan, errChan, cancelReadStateNode := stateDB.flatfile.ReadRecently()
	stateNodeMap := map[common.Hash]*StateNode{}
	prevMap := map[common.Hash]*common.Hash{}
	for {
		select {
		case stateByte := <-dataChan:
			if len(stateByte) == 0 {
				goto RESTORE_SUCCESS
			}
			stateNode, prevHash, err := stateDB.stateDB.DeSerializeFromStateNodeData(stateByte)
			if err != nil {
				return err
			}

			stateNodeMap[stateNode.GetHash()] = stateNode
			prevMap[stateNode.GetHash()] = prevHash

			if stateNode.aggregateHash.String() == finalState.String() {
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
	if _, ok := stateNodeMap[root]; !ok {
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
	bytesResult, _ := stateDB.db.Get(append([]byte(PREFIX_LITESTATEDB), addr.GetBytes()...))
	if len(bytesResult) > 0 {
		return newStateObjectWithValue(stateDB.stateDB, objectType, addr, bytesResult)
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
	//fmt.Println("kvMap", kvMap)
	iter := NewLiteStateDBIterator(stateDB.db, prefix, kvMap)
	return iter
}
