package statedb

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
	"strconv"
)

func GenerateBlockHashByIndexKey(chainName string, index uint64) common.Hash {
	h1 := common.HashH(append(blockHashByIndexPrefix, []byte(chainName)...))
	h2 := common.HashH([]byte(strconv.Itoa(int(index))))
	return common.BytesToHash(append(h1[:][:prefixHashKeyLength], h2[:][:prefixKeyLength]...))
}

func newBlockHashStateObject(db *StateDB, hash common.Hash) *BlockHashObject {
	return &BlockHashObject{
		db:           db,
		version:      defaultVersion,
		blockKeyHash: hash,
		blockHash:    &common.Hash{},
		objectType:   BlockHashObjectType,
		deleted:      false,
	}
}

func newBlockHashStateObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*BlockHashObject, error) {
	blockHashObject := new(BlockHashObject)
	blockHashObject.db = db
	blockHashObject.version = defaultVersion
	blockHashObject.blockKeyHash = key
	var dataBytes []byte
	var blockhash *common.Hash
	var err error
	var ok bool
	if dataBytes, ok = data.([]byte); ok {
		blockhash, err = common.Hash{}.NewHash(dataBytes)
		if err != nil {
			return nil, err
		}
	} else {
		blockhash, ok = data.(*common.Hash)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidBlockHashType, reflect.TypeOf(data))
		}
	}

	blockHashObject.blockHash = blockhash
	blockHashObject.objectType = BlockHashObjectType
	blockHashObject.deleted = false
	return blockHashObject, nil
}

type BlockHashObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version      int
	blockKeyHash common.Hash
	blockHash    *common.Hash
	objectType   int
	deleted      bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func (s *BlockHashObject) GetVersion() int {
	return s.version
}

func (s *BlockHashObject) GetValue() interface{} {
	return s.blockHash
}

func (s *BlockHashObject) GetValueBytes() []byte {
	return s.blockHash.Bytes()
}

func (s *BlockHashObject) GetHash() common.Hash {
	return s.blockKeyHash
}

func (s *BlockHashObject) GetType() int {
	return s.objectType
}

func (s *BlockHashObject) SetValue(hash interface{}) error {
	s.blockHash = hash.(*common.Hash)
	return nil
}

func (s *BlockHashObject) GetTrie(DatabaseAccessWarper) Trie {
	return s.trie
}

func (s *BlockHashObject) SetError(err error) {
	s.dbErr = err
}

func (s *BlockHashObject) MarkDelete() {
	s.deleted = true
}

func (s *BlockHashObject) IsDeleted() bool {
	return s.deleted
}

func (s *BlockHashObject) IsEmpty() bool {
	return reflect.DeepEqual(*s.blockHash, common.Hash{})

}

func (s *BlockHashObject) Reset() bool {
	s.blockHash = &common.Hash{}
	return true
}
