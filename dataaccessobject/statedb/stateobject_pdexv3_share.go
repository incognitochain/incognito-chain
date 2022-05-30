package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type Pdexv3ShareState struct {
	nftID          common.Hash
	amount         uint64
	lmLockedAmount uint64
}

func (ps *Pdexv3ShareState) NftID() common.Hash {
	return ps.nftID
}

func (ps *Pdexv3ShareState) Amount() uint64 {
	return ps.amount
}

func (ps *Pdexv3ShareState) LmLockedAmount() uint64 {
	return ps.lmLockedAmount
}

func (ps *Pdexv3ShareState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		NftID          common.Hash `json:"NftID"`
		Amount         uint64      `json:"Amount"`
		LmLockedAmount uint64      `json:"LmLockedAmount,omitempty"`
	}{
		NftID:          ps.nftID,
		Amount:         ps.amount,
		LmLockedAmount: ps.lmLockedAmount,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (ps *Pdexv3ShareState) UnmarshalJSON(data []byte) error {
	temp := struct {
		NftID          common.Hash `json:"NftID"`
		Amount         uint64      `json:"Amount"`
		LmLockedAmount uint64      `json:"LmLockedAmount,omitempty"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	ps.nftID = temp.NftID
	ps.amount = temp.Amount
	ps.lmLockedAmount = temp.LmLockedAmount
	return nil
}

func NewPdexv3ShareState() *Pdexv3ShareState {
	return &Pdexv3ShareState{}
}

func NewPdexv3ShareStateWithValue(
	nftID common.Hash, amount, lmLockedAmount uint64,
) *Pdexv3ShareState {
	return &Pdexv3ShareState{
		nftID:          nftID,
		amount:         amount,
		lmLockedAmount: lmLockedAmount,
	}
}

func (ps *Pdexv3ShareState) Clone() *Pdexv3ShareState {
	return &Pdexv3ShareState{
		nftID:          ps.nftID,
		amount:         ps.amount,
		lmLockedAmount: ps.lmLockedAmount,
	}
}

type Pdexv3ShareObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *Pdexv3ShareState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPdexv3ShareObject(db *StateDB, hash common.Hash) *Pdexv3ShareObject {
	return &Pdexv3ShareObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewPdexv3ShareState(),
		objectType: Pdexv3ShareObjectType,
		deleted:    false,
	}
}

func newPdexv3ShareObjectWithValue(db *StateDB, key common.Hash, data interface{}) (
	*Pdexv3ShareObject, error,
) {
	var newPdexv3ShareState = NewPdexv3ShareState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPdexv3ShareState)
		if err != nil {
			return nil, err
		}
	} else {
		newPdexv3ShareState, ok = data.(*Pdexv3ShareState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3ShareStateType, reflect.TypeOf(data))
		}
	}
	return &Pdexv3ShareObject{
		version:    defaultVersion,
		hash:       key,
		state:      newPdexv3ShareState,
		db:         db,
		objectType: Pdexv3ShareObjectType,
		deleted:    false,
	}, nil
}

func generatePdexv3ShareObjectPrefix(poolPairID string) []byte {
	b := append(GetPdexv3SharesPrefix(), []byte(poolPairID)...)
	h := common.HashH(b)
	return h[:prefixHashKeyLength]
}

func GeneratePdexv3ShareObjectKey(poolPairID, nftID string) common.Hash {
	prefixHash := generatePdexv3ShareObjectPrefix(poolPairID)
	valueHash := common.HashH([]byte(nftID))
	return common.BytesToHash(append(prefixHash, valueHash[:prefixKeyLength]...))
}

func (ps *Pdexv3ShareObject) GetVersion() int {
	return ps.version
}

// setError remembers the first non-nil error it is called with.
func (ps *Pdexv3ShareObject) SetError(err error) {
	if ps.dbErr == nil {
		ps.dbErr = err
	}
}

func (ps *Pdexv3ShareObject) GetTrie(db DatabaseAccessWarper) Trie {
	return ps.trie
}

func (ps *Pdexv3ShareObject) SetValue(data interface{}) error {
	newPdexv3ShareState, ok := data.(*Pdexv3ShareState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3ShareStateType, reflect.TypeOf(data))
	}
	ps.state = newPdexv3ShareState
	return nil
}

func (ps *Pdexv3ShareObject) GetValue() interface{} {
	return ps.state
}

func (ps *Pdexv3ShareObject) GetValueBytes() []byte {
	state, ok := ps.GetValue().(*Pdexv3ShareState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(state)
	if err != nil {
		panic("failed to marshal pdexv3 share state")
	}
	return value
}

func (ps *Pdexv3ShareObject) GetHash() common.Hash {
	return ps.hash
}

func (ps *Pdexv3ShareObject) GetType() int {
	return ps.objectType
}

// MarkDelete will delete an object in trie
func (ps *Pdexv3ShareObject) MarkDelete() {
	ps.deleted = true
}

// reset all shard committee value into default value
func (ps *Pdexv3ShareObject) Reset() bool {
	ps.state = NewPdexv3ShareState()
	return true
}

func (ps *Pdexv3ShareObject) IsDeleted() bool {
	return ps.deleted
}

// value is either default or nil
func (ps *Pdexv3ShareObject) IsEmpty() bool {
	temp := NewPdexv3ShareState()
	return reflect.DeepEqual(temp, ps.state) || ps.state == nil
}
