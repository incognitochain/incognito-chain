package statedb

import (
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
	"log"
	"sort"
)

type stateNode struct {
	stateObjects  map[common.Hash]StateObject
	previousLink  *stateNode
	aggregateHash *common.Hash
	flushDB       bool
}

func NewStateNode() *stateNode {
	return &stateNode{
		stateObjects:  map[common.Hash]StateObject{},
		previousLink:  nil,
		aggregateHash: nil,
		flushDB:       false,
	}
}

func (s *stateNode) CommitToDisk(dbWriter incdb.KeyValueWriter) error {
	if s.previousLink == nil {
		if s.flushDB {
			return nil
		}

		for addr, obj := range s.stateObjects {
			log.Println("Write key", addr.String())
			err := dbWriter.Put(append([]byte(PREFIX_LITESTATEDB), addr.GetBytes()...), obj.GetValueBytes())
			if err != nil {
				return err
			}
		}
		s.flushDB = true

		return nil
	}

	err := s.previousLink.CommitToDisk(dbWriter)
	if err != nil {
		return err
	}

	return nil
}

func (s *stateNode) Commit() (*common.Hash, error) {
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
		prevAggHash := s.previousLink.aggregateHash.Bytes()
		for _, obj := range sortObjs {
			prevAggHash = append(prevAggHash, obj.GetHash().Bytes()...)
		}
		aggregateHash := common.Keccak256(prevAggHash)
		s.aggregateHash = &aggregateHash
	} else {
		s.aggregateHash = s.previousLink.aggregateHash
	}

	return s.aggregateHash, nil
}
