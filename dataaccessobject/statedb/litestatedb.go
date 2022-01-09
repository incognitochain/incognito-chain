package statedb

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/syncker/flatfile"
	"github.com/pkg/errors"
	"log"
	"path"
)

const PREFIX_LITESTATEDB = "litestatedb"

type LiteStateDB struct {
	flatfile      *flatfile.FlatFileManager
	db            incdb.Database
	headStateNode *StateNode
	stateDB       *StateDB
}

func GetFlatFileDatabase(dir string, sid int) (*flatfile.FlatFileManager, error) {
	p := path.Join(dir, fmt.Sprintf("block/commit_state_%v", sid))
	ff, err := flatfile.NewFlatFile(p, 100)
	return ff, err
}

func NewLiteStateDB(dir string, shardID int, db incdb.Database) (*StateDB, error) {
	stateDB := &StateDB{
		stateObjects:        make(map[common.Hash]StateObject),
		stateObjectsPending: make(map[common.Hash]struct{}),
		stateObjectsDirty:   make(map[common.Hash]struct{}),
	}

	stateNode := NewStateNode()
	stateNode.aggregateHash = nil
	ff, err := GetFlatFileDatabase(dir, shardID)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot create flatfile for litestatedb")
	}

	stateDB.liteStateDB = &LiteStateDB{
		ff,
		db,
		stateNode,
		stateDB,
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

func (stateDB *LiteStateDB) CommitToDisk(dbWriter incdb.KeyValueWriter, stateNodeHash common.Hash) error {

	//write finalize Key Value to disk
	stateNode := stateDB.headStateNode
	for {
		if stateNode == nil {
			return nil
		}

		if stateNode.aggregateHash.String() == stateNodeHash.String() {
			err := stateNode.CommitToDisk(dbWriter)
			if err != nil {
				return err
			}
		}
		stateNode = stateNode.previousLink
	}

}

func RestoreStateNode(dir string, shardID int, finalHash common.Hash, db incdb.Database) (*StateDB, map[common.Hash]*StateNode, error) {
	ff, err := GetFlatFileDatabase(dir, shardID)
	if err != nil {
		return nil, nil, err
	}

	dataChan, _, cancelReadStateNode := ff.ReadRecently()
	stateNodeMap := map[common.Hash]*StateNode{}
	prevMap := map[common.Hash]*common.Hash{}
	stateDB, err := NewLiteStateDB(dir, int(shardID), db)
	if err != nil {
		return nil, nil, err
	}
	for {
		stateByte := <-dataChan
		stateNode, prevHash, err := stateDB.DeSerializeFromStateNodeData(stateByte)
		if err != nil {
			return nil, nil, err
		}
		stateNode.SetTmpCommit(true)
		stateNodeMap[stateNode.GetHash()] = stateNode
		prevMap[stateNode.GetHash()] = prevHash

		// dataChan return from latest to oldest, finalHash is the oldest view
		// as soon as return data is oldest, we have enough state node to rebuild!
		if stateNode.GetHash().String() == finalHash.String() {
			stateNode.SetFinalize(true)
			cancelReadStateNode()
			break
		}
	}

	for hash, stateNode := range stateNodeMap {
		if prevMap[hash] != nil {
			stateNode.SetPreviousLink(stateNodeMap[*prevMap[hash]])
		}
	}

	return stateDB, stateNodeMap, nil
}

func (stateDB *LiteStateDB) NewStateNode() {
	newNode := NewStateNode()
	newNode.previousLink = stateDB.headStateNode
	stateDB.headStateNode = newNode
}

func (stateDB *LiteStateDB) Finalized(stateNodeHash common.Hash) {
	defer stateDB.NewStateNode()

	stateNode := stateDB.headStateNode
	for {
		if stateNode == nil {
			return
		}
		if stateNode.aggregateHash.String() == stateNodeHash.String() {
			stateNode.previousLink = nil
			return
		}
		stateNode = stateNode.previousLink
	}

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
	return stateDB.db.NewIteratorWithPrefix(append([]byte("litestadb"), prefix...))
}
