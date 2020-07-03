package blockchain

import (
	"github.com/incognitochain/incognito-chain/common"
	"sort"
)

type MapStringBool struct {
	data    map[string]bool
	hash    *common.Hash
	updated bool
}

func NewMapStringBool() *MapStringBool {
	return &MapStringBool{
		data:    make(map[string]bool),
		hash:    nil,
		updated: false,
	}
}

func (s *MapStringBool) LazyCopy() *MapStringBool {
	newCopy := *s
	s.updated = false
	return &newCopy
}

func (s *MapStringBool) copy() {
	prev := s.data
	s.data = make(map[string]bool)
	for k, v := range prev {
		s.data[k] = v
	}
	s.updated = false
}

func (s *MapStringBool) Remove(k string) {
	if !s.updated {
		s.copy()
	}
	delete(s.data, k)
	s.updated = true
	s.hash = nil
}

func (s *MapStringBool) Set(k string, v bool) {
	if !s.updated {
		s.copy()
	}
	s.data[k] = v
	s.updated = true
	s.hash = nil
}

func (s *MapStringBool) GetMap() map[string]bool {
	return s.data
}

func (s *MapStringBool) Get(k string) (bool, bool) {
	r, ok := s.data[k]
	return r, ok
}

func (s *MapStringBool) GenerateHash() (common.Hash, error) {
	if s.hash != nil {
		return *s.hash, nil
	}
	var keys []string
	var res []string
	for k := range s.data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		res = append(res, key)
		if s.data[key] {
			res = append(res, "true")
		} else {
			res = append(res, "false")
		}
	}
	return generateHashFromStringArray(res)
}
