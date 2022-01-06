package statedb

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
	"log"
)

const PREFIX_LITESTATEDB = "litestatedb"

type LiteStateDB struct {
	db            incdb.Database
	headStateNode *stateNode
	stateDB       *StateDB
}

func (stateDB *LiteStateDB) Copy() *LiteStateDB {
	cpy := *stateDB.headStateNode
	cpy.stateObjects = make(map[common.Hash]StateObject)
	return &LiteStateDB{
		stateDB.db,
		&cpy,
		stateDB.stateDB,
	}
}

func (stateDB *LiteStateDB) CommitToDisk(dbWriter incdb.KeyValueWriter, stateNodeHash common.Hash) error {
	stateDB.headStateNode.Commit()
	log.Println("state Object len ", len(stateDB.headStateNode.stateObjects))
	log.Printf("state Object %+v %+v ", stateDB.headStateNode.aggregateHash.String(), stateNodeHash)

	//if stateNodeHash.String() == common.EmptyRoot.String() {
	//	err := stateDB.headStateNode.CommitToDisk(dbWriter)
	//	if err != nil {
	//		return err
	//	}
	//	return nil
	//}

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

func (stateDB *LiteStateDB) Finalized(stateNodeHash common.Hash) {
	//TODO: must consult finalized and best view to link headstate
	defer func() {
		newNode := NewStateNode()
		newNode.previousLink = stateDB.headStateNode
		stateDB.headStateNode = newNode
	}()

	stateNode := stateDB.headStateNode
	for {
		if stateNode == nil {
			return
		}
		if stateNode.aggregateHash.String() == stateNodeHash.String() {
			stateNode.previousLink = nil
			return
		}
	}

}

func (stateDB *LiteStateDB) Commit() (common.Hash, error) {
	h, err := stateDB.headStateNode.Commit()
	if h != nil {
		return *h, nil
	}
	return common.Hash{}, err

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
