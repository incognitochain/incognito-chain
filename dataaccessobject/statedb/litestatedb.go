package statedb

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/flatfile"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/pkg/errors"
	"log"
)

type LiteStateDB struct {
	flatfile *flatfile.FlatFileManager
	db       incdb.Database
	dbPrefix string

	headStateNode *StateNode
	stateDB       *StateDB
}

func NewLiteStateDB(dir string, name string, root common.Hash, db incdb.Database) (*StateDB, error) {
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
	dbPrefix := "litestateDB-" + name

	stateDB.liteStateDB = &LiteStateDB{
		ff,
		db,
		dbPrefix,
		stateNode,
		stateDB,
	}

	if root.IsZeroValue() {
		return stateDB, nil
	}

	if !root.IsZeroValue() {
		err := stateDB.liteStateDB.restoreStateNode(root)
		if err != nil {
			return nil, err
		}
	}

	return stateDB, nil
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

	//if equal to previous link
	if stateDB.headStateNode.previousLink != nil && h.String() == stateDB.headStateNode.previousLink.aggregateHash.String() {
		stateDB.headStateNode.ffIndex = stateDB.headStateNode.previousLink.ffIndex
		ffIndexBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(ffIndexBytes, uint64(stateDB.headStateNode.ffIndex))
		b := append(ffIndexBytes, stateDB.headStateNode.previousLink.aggregateHash.Bytes()[8:]...)
		newHash := common.Hash{}
		newHash.SetBytes(b)
		fmt.Println(1, b)
		return newHash, nil
	}

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
		ffIndexBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(ffIndexBytes, uint64(ffIndex))
		b := append(ffIndexBytes, h.Bytes()[8:]...)
		newHash := common.Hash{}
		newHash.SetBytes(b)
		fmt.Println(2, b)
		return newHash, nil
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
		if bytes.Equal(stateNode.aggregateHash[8:], stateNodeHash[8:]) {
			err := stateNode.FlushFinalizedToDisk(dbWriter, stateDB)
			if err != nil {
				return err
			}
			stateNode.previousLink = nil
			return nil
		}
		dbWriter.Put(stateDB.GetFinalKey(), stateNodeHash.Bytes())
		stateNode = stateNode.previousLink
	}

	return errors.New("Cannot find finalized hash!")
}

func (stateDB *LiteStateDB) restoreStateNode(root common.Hash) error {
	//get final commit
	finalKey := stateDB.GetFinalKey()
	v, _ := stateDB.db.Get([]byte(finalKey))
	finalState := common.Hash{}
	if len(v) == 32 {
		finalState.SetBytes(v)
	}

	var ffIndex uint64
	err := binary.Read(bytes.NewBuffer(root.Bytes()[:8]), binary.LittleEndian, &ffIndex)
	if err != nil {
		return err
	}

	fmt.Println("===================> read from index", ffIndex, root.String())
	dataChan, errChan, cancelReadStateNode := stateDB.flatfile.ReadRecently(ffIndex)
	stateNodeMap := map[common.Hash]*StateNode{}
	prevMap := map[common.Hash]*common.Hash{}
	cnt := 0
	for {
		select {
		case stateByte := <-dataChan:
			cnt++
			if cnt > 5000 {
				return errors.New("Something wrong, retrieve more than 5000 state")
			}
			if len(stateByte) == 0 {
				e := common.Hash{}
				if finalState.String() == e.String() {
					fmt.Println("===================> 1")
					goto RESTORE_SUCCESS
				} else {
					fmt.Println("===================> 2")
					return errors.New("Cannot rebuild")
				}
			}
			stateNode, prevHash, err := stateDB.stateDB.DeSerializeFromStateNodeData(stateByte)
			if err != nil {
				return err
			}

			stateNodeMap[stateNode.GetHash()] = stateNode
			prevMap[stateNode.GetHash()] = prevHash
			fmt.Println("===================> 3", stateNode.aggregateHash.String())
			if bytes.Equal(stateNode.aggregateHash[8:], finalState[8:]) {

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
			fmt.Println("===================> SetPreviousLink ", stateNode.aggregateHash.String(), prevMap[hash].String())
			stateNode.SetPreviousLink(stateNodeMap[*prevMap[hash]])
		}
	}

	//point head to root
	for nodeHash, state := range stateNodeMap {
		if bytes.Equal(nodeHash.Bytes()[8:], root[8:]) {
			stateDB.headStateNode = state
			break
		}
	}
	if stateDB.headStateNode == nil {
		return errors.New("Cannot find root head")
	}
	fmt.Println("===================> set ", stateDB.headStateNode.aggregateHash.String())
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
	//log.Println("Insert key", object.GetHash().String())
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
