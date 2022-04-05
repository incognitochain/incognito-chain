package statedb

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
	"sort"
)

type StateNode struct {
	stateObjects    map[common.Hash]StateObject
	previousLink    *StateNode
	aggregateHash   *common.Hash //!= nil, aslo mean commit from mem to disk
	finalizedCommit bool
}

func (s StateNode) GetHash() common.Hash {
	return *s.aggregateHash
}

func (s *StateNode) SetPreviousLink(stateNode *StateNode) {
	s.previousLink = stateNode
}

type SerializeObject struct {
	O [][]byte     // stateobjects bytes value
	H *common.Hash //aggregate hash
	P *common.Hash //previous hash
}

func init() {
	gob.Register(SerializeObject{})
}
func NewStateNode() *StateNode {
	return &StateNode{
		stateObjects:  map[common.Hash]StateObject{},
		previousLink:  nil,
		aggregateHash: nil,
	}
}

func (s *StateNode) Serialize() ([]byte, error) {
	sobj := SerializeObject{}
	for _, obj := range s.stateObjects {
		var objTypeByte = make([]byte, 8)
		binary.LittleEndian.PutUint64(objTypeByte, uint64(obj.GetType()))
		byteValue := append(objTypeByte[:], obj.GetHash().Bytes()...)
		if obj.IsDeleted() {
			byteValue = append(byteValue, 1)
		} else {
			byteValue = append(byteValue, 0)
		}

		byteValue = append(byteValue, obj.GetValueBytes()...)
		sobj.O = append(sobj.O, byteValue)
	}
	sobj.H = s.aggregateHash

	if s.previousLink != nil {
		sobj.P = s.previousLink.aggregateHash
	}
	var cachebuffer bytes.Buffer
	enc := gob.NewEncoder(&cachebuffer)
	err := enc.Encode(sobj)
	return cachebuffer.Bytes(), err
}

func (stateDB *StateDB) DeSerializeFromStateNodeData(data []byte) (*StateNode, *common.Hash, error) {
	var sobj = SerializeObject{}
	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&sobj)
	if err != nil {
		return nil, nil, err
	}
	stateObject := map[common.Hash]StateObject{}
	for _, objectByte := range sobj.O {
		var objType uint64
		err := binary.Read(bytes.NewBuffer(objectByte[:8]), binary.LittleEndian, &objType)
		if err != nil {
			return nil, nil, err
		}
		key, _ := common.Hash{}.NewHash(objectByte[8:40])

		obj, err := newStateObjectWithValue(stateDB, int(objType), *key, objectByte[41:])
		if err != nil {
			return nil, nil, err
		}
		stateObject[*key] = obj
		if objectByte[40] != 0 {
			obj.MarkDelete()
		}
	}
	stateNode := StateNode{}
	stateNode.stateObjects = stateObject
	stateNode.aggregateHash = sobj.H
	return &stateNode, sobj.P, nil
}

func (sobj *SerializeObject) DeSerialize(data []byte) error {
	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&sobj)
	return err
}

func (s *StateNode) FlushFinalizedToDisk(dbWriter incdb.KeyValueWriter, liteStateDB *LiteStateDB) error {
	if s.previousLink != nil {
		err := s.previousLink.FlushFinalizedToDisk(dbWriter, liteStateDB)
		if err != nil {
			return err
		}
	}

	if s.finalizedCommit {
		return nil
	}

	for addr, obj := range s.stateObjects {
		//log.Println("Write key", addr.String(), obj.IsDeleted())
		deleteByte := byte(0)
		if obj.IsDeleted() {
			deleteByte = 1
		}

		err := dbWriter.Put(append([]byte(liteStateDB.dbPrefix), addr.GetBytes()...), append([]byte{deleteByte}, obj.GetValueBytes()...))
		if err != nil {
			return err
		}
	}
	s.finalizedCommit = true

	return nil

}

func (s *StateNode) replay(kv map[string][]byte) error {
	if s.previousLink != nil && s.finalizedCommit == false {
		err := s.previousLink.replay(kv)
		if err != nil {
			return err
		}
	}

	if s.finalizedCommit {
		return nil
	}

	for addr, obj := range s.stateObjects {
		kv[string(addr[:])] = obj.GetValueBytes()
	}

	return nil

}

func (s *StateNode) Commit() (*common.Hash, error) {

	if s.aggregateHash != nil {
		return s.aggregateHash, nil
	}

	if s.previousLink != nil && s.previousLink.aggregateHash == nil {
		prevAggregateHash, _ := s.previousLink.Commit()
		if prevAggregateHash == nil {
			return nil, errors.New("Previous aggregate hash is nil!")
		}
	}

	sortObjs := []StateObject{}
	for _, obj := range s.stateObjects {
		sortObjs = append(sortObjs, obj)
	}

	sort.Slice(sortObjs, func(i int, j int) bool {
		x := sortObjs[i].GetHash()
		y := sortObjs[j].GetHash()
		cmp, _ := x.Cmp(&y)
		if cmp > 1 {
			return true
		}
		return false
	})

	if len(sortObjs) > 0 {
		prevAggHash := common.EmptyRoot.Bytes()
		if s.previousLink != nil {
			prevAggHash = s.previousLink.aggregateHash.Bytes()
		}

		for _, obj := range sortObjs {
			prevAggHash = append(prevAggHash, obj.GetHash().Bytes()...)
			if obj.IsDeleted() {
				prevAggHash = append(prevAggHash, 1)
			} else {
				prevAggHash = append(prevAggHash, 0)
			}
		}

		aggregateHash := common.Keccak256(prevAggHash)
		s.aggregateHash = &aggregateHash
	} else {
		s.aggregateHash = s.previousLink.aggregateHash
	}

	return s.aggregateHash, nil
}
