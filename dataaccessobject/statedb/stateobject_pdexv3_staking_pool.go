package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type Pdexv3StakingPoolState struct {
	tokenID     string
	totalAmount uint64
}

func (pp *Pdexv3StakingPoolState) TokenID() string {
	return pp.tokenID
}

func (pp *Pdexv3StakingPoolState) TotalAmount() uint64 {
	return pp.totalAmount
}

func (pp *Pdexv3StakingPoolState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		TokenID     string `json:"TokenID"`
		TotalAmount uint64 `json:"TotalAmount"`
	}{
		TokenID:     pp.tokenID,
		TotalAmount: pp.totalAmount,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (pp *Pdexv3StakingPoolState) UnmarshalJSON(data []byte) error {
	temp := struct {
		TokenID     string `json:"TokenID"`
		TotalAmount uint64 `json:"TotalAmount"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	pp.tokenID = temp.TokenID
	pp.totalAmount = temp.TotalAmount
	return nil
}

func (pp *Pdexv3StakingPoolState) Clone() *Pdexv3StakingPoolState {
	return &Pdexv3StakingPoolState{
		tokenID:     pp.tokenID,
		totalAmount: pp.totalAmount,
	}
}

func NewPdexv3StakingPoolState() *Pdexv3StakingPoolState {
	return &Pdexv3StakingPoolState{}
}

func NewPdexv3StakingPoolStateWithValue(
	tokenID string, totalAmount uint64,
) *Pdexv3StakingPoolState {
	return &Pdexv3StakingPoolState{
		tokenID:     tokenID,
		totalAmount: totalAmount,
	}
}

type Pdexv3StakingPoolObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *Pdexv3StakingPoolState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPdexv3StakingPoolObject(db *StateDB, hash common.Hash) *Pdexv3StakingPoolObject {
	return &Pdexv3StakingPoolObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewPdexv3StakingPoolState(),
		objectType: Pdexv3StakingPoolObjectType,
		deleted:    false,
	}
}

func newPdexv3StakingPoolObjectWithValue(db *StateDB, key common.Hash, data interface{}) (
	*Pdexv3StakingPoolObject, error,
) {
	var newPdexv3StakingPoolState = NewPdexv3StakingPoolState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPdexv3StakingPoolState)
		if err != nil {
			return nil, err
		}
	} else {
		newPdexv3StakingPoolState, ok = data.(*Pdexv3StakingPoolState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3StakingPoolStateType, reflect.TypeOf(data))
		}
	}
	return &Pdexv3StakingPoolObject{
		version:    defaultVersion,
		hash:       key,
		state:      newPdexv3StakingPoolState,
		db:         db,
		objectType: Pdexv3StakingPoolObjectType,
		deleted:    false,
	}, nil
}

func GeneratePdexv3StakingPoolObjectKey(tokenID string) common.Hash {
	prefixHash := GetPdexv3StakingPoolsPrefix()
	valueHash := common.HashH([]byte(tokenID))
	return common.BytesToHash(append(prefixHash, valueHash[:prefixKeyLength]...))
}

func (pp *Pdexv3StakingPoolObject) GetVersion() int {
	return pp.version
}

// setError remembers the first non-nil error it is called with.
func (pp *Pdexv3StakingPoolObject) SetError(err error) {
	if pp.dbErr == nil {
		pp.dbErr = err
	}
}

func (pp *Pdexv3StakingPoolObject) GetTrie(db DatabaseAccessWarper) Trie {
	return pp.trie
}

func (pp *Pdexv3StakingPoolObject) SetValue(data interface{}) error {
	newPdexv3StakingPoolState, ok := data.(*Pdexv3StakingPoolState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3StakingPoolStateType, reflect.TypeOf(data))
	}
	pp.state = newPdexv3StakingPoolState
	return nil
}

func (pp *Pdexv3StakingPoolObject) GetValue() interface{} {
	return pp.state
}

func (pp *Pdexv3StakingPoolObject) GetValueBytes() []byte {
	state, ok := pp.GetValue().(*Pdexv3StakingPoolState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(state)
	if err != nil {
		panic("failed to marshal pdexv3 staking pool state")
	}
	return value
}

func (pp *Pdexv3StakingPoolObject) GetHash() common.Hash {
	return pp.hash
}

func (pp *Pdexv3StakingPoolObject) GetType() int {
	return pp.objectType
}

// MarkDelete will delete an object in trie
func (pp *Pdexv3StakingPoolObject) MarkDelete() {
	pp.deleted = true
}

// reset all shard committee value into default value
func (pp *Pdexv3StakingPoolObject) Reset() bool {
	pp.state = NewPdexv3StakingPoolState()
	return true
}

func (pp *Pdexv3StakingPoolObject) IsDeleted() bool {
	return pp.deleted
}

// value is either default or nil
func (pp *Pdexv3StakingPoolObject) IsEmpty() bool {
	temp := NewPdexv3StakingPoolState()
	return reflect.DeepEqual(temp, pp.state) || pp.state == nil
}
