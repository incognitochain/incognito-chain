package statedb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"math/big"
	"reflect"
)

type OnetimeAddressState struct {
	tokenID 	  common.Hash
	publicKey     []byte
	index 		  *big.Int
}

func (s OnetimeAddressState) Index() *big.Int {
	return s.index
}

func (s *OnetimeAddressState) SetIndex(index *big.Int) {
	s.index = index
}

func (s OnetimeAddressState) PublicKey() []byte {
	return s.publicKey
}

func (s *OnetimeAddressState) SetPublicKey(publicKey []byte) {
	s.publicKey = publicKey
}

func (s OnetimeAddressState) TokenID() common.Hash {
	return s.tokenID
}

func (s *OnetimeAddressState) SetTokenID(tokenID common.Hash) {
	s.tokenID = tokenID
}

func (s OnetimeAddressState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		TokenID common.Hash
		PublicKey     []byte
		Index 		  *big.Int
	}{
		TokenID: 		s.tokenID,
		PublicKey:      s.publicKey,
		Index: 			s.index,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (s *OnetimeAddressState) UnmarshalJSON(data []byte) error {
	temp := struct {
		TokenID 	common.Hash
		PublicKey   []byte
		Index 		*big.Int
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	s.tokenID = temp.TokenID
	s.publicKey = temp.PublicKey
	s.index = temp.Index
	return nil
}

func NewOnetimeAddressState() *OnetimeAddressState {
	return &OnetimeAddressState{}
}

func NewOnetimeAddressStateWithValue(tokenID common.Hash, publicKey []byte, index *big.Int) *OnetimeAddressState {
	return &OnetimeAddressState{tokenID: tokenID, publicKey: publicKey, index: index}
}

type OnetimeAddressObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version          int
	onetimeAddressHash  common.Hash
	onetimeAddressState *OnetimeAddressState
	objectType       int
	deleted          bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newOnetimeAddressObject(db *StateDB, hash common.Hash) *OnetimeAddressObject {
	return &OnetimeAddressObject{
		version:          	 defaultVersion,
		db:               	 db,
		onetimeAddressHash:  hash,
		onetimeAddressState: NewOnetimeAddressState(),
		objectType:       	 OnetimeAddressObjectType,
		deleted:          	 false,
	}
}

func newOnetimeAddressObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*OnetimeAddressObject, error) {
	var newOnetimeAddressState = NewOnetimeAddressState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newOnetimeAddressState)
		if err != nil {
			return nil, err
		}
	} else {
		newOnetimeAddressState, ok = data.(*OnetimeAddressState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidOnetimeAddressStateType, reflect.TypeOf(data))
		}
	}
	return &OnetimeAddressObject{
		version:          defaultVersion,
		onetimeAddressHash:  key,
		onetimeAddressState: newOnetimeAddressState,
		db:               db,
		objectType:       OnetimeAddressObjectType,
		deleted:          false,
	}, nil
}

func GenerateOnetimeAddressObjectKey(tokenID common.Hash, onetimeAddress []byte) common.Hash {
	if tokenID!=common.PRVCoinID{
		tokenID = common.ConfidentialAssetID
	}
	prefixHash := GetOnetimeAddressPrefix(tokenID)
	valueHash := common.HashH(onetimeAddress)
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (s OnetimeAddressObject) GetVersion() int {
	return s.version
}

// setError remembers the first non-nil error it is called with.
func (s *OnetimeAddressObject) SetError(err error) {
	if s.dbErr == nil {
		s.dbErr = err
	}
}

func (s OnetimeAddressObject) GetTrie(db DatabaseAccessWarper) Trie {
	return s.trie
}

func (s *OnetimeAddressObject) SetValue(data interface{}) error {
	newOnetimeAddressState, ok := data.(*OnetimeAddressState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidOnetimeAddressStateType, reflect.TypeOf(data))
	}
	s.onetimeAddressState = newOnetimeAddressState
	return nil
}

func (s OnetimeAddressObject) GetValue() interface{} {
	return s.onetimeAddressState
}

func (s OnetimeAddressObject) GetValueBytes() []byte {
	onetimeAddressState, ok := s.GetValue().(*OnetimeAddressState)
	if !ok {
		panic("wrong expected value type")
	}
	onetimeAddressStateBytes, err := json.Marshal(onetimeAddressState)
	if err != nil {
		panic(err.Error())
	}
	return onetimeAddressStateBytes
}

func (s OnetimeAddressObject) GetHash() common.Hash {
	return s.onetimeAddressHash
}

func (s OnetimeAddressObject) GetType() int {
	return s.objectType
}

// MarkDelete will delete an object in trie
func (s *OnetimeAddressObject) MarkDelete() {
	s.deleted = true
}

// Reset serial number into default
func (s *OnetimeAddressObject) Reset() bool {
	s.onetimeAddressState = NewOnetimeAddressState()
	return true
}

func (s OnetimeAddressObject) IsDeleted() bool {
	return s.deleted
}

// empty value or not
func (s OnetimeAddressObject) IsEmpty() bool {
	temp := NewOnetimeAddressState()
	return reflect.DeepEqual(temp, s.onetimeAddressState) || s.onetimeAddressState == nil
}
