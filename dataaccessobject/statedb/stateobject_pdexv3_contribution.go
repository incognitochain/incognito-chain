package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
)

type Pdexv3ContributionState struct {
	value    rawdbv2.Pdexv3Contribution
	pairHash string
}

func (pc *Pdexv3ContributionState) PairHash() string {
	return pc.pairHash
}

func (pc *Pdexv3ContributionState) SetValue(value rawdbv2.Pdexv3Contribution) {
	pc.value = value
}

func (pc *Pdexv3ContributionState) Value() rawdbv2.Pdexv3Contribution {
	return pc.value
}

func (pc *Pdexv3ContributionState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		PairHash string                      `json:"PairHash"`
		Value    *rawdbv2.Pdexv3Contribution `json:"Value"`
	}{
		PairHash: pc.pairHash,
		Value:    &pc.value,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (pc *Pdexv3ContributionState) UnmarshalJSON(data []byte) error {
	temp := struct {
		PairHash string                      `json:"PairHash"`
		Value    *rawdbv2.Pdexv3Contribution `json:"Value"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	pc.pairHash = temp.PairHash
	if temp.Value != nil {
		pc.value = *temp.Value
	}
	return nil
}

func (pc *Pdexv3ContributionState) Clone() *Pdexv3ContributionState {
	return &Pdexv3ContributionState{
		value:    *pc.value.Clone(),
		pairHash: pc.pairHash,
	}
}

func NewPdexv3ContributionState() *Pdexv3ContributionState {
	return &Pdexv3ContributionState{
		value: *rawdbv2.NewPdexv3Contribution(),
	}
}

func NewPdexv3ContributionStateWithValue(
	value rawdbv2.Pdexv3Contribution, pairHash string,
) *Pdexv3ContributionState {
	return &Pdexv3ContributionState{
		value:    value,
		pairHash: pairHash,
	}
}

type Pdexv3ContributionObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *Pdexv3ContributionState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPdexv3ContributionObject(db *StateDB, hash common.Hash) *Pdexv3ContributionObject {
	return &Pdexv3ContributionObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewPdexv3ContributionState(),
		objectType: Pdexv3ContributionObjectType,
		deleted:    false,
	}
}

func newPdexv3ContributionObjectWithValue(db *StateDB, key common.Hash, data interface{}) (
	*Pdexv3ContributionObject, error,
) {
	var newPdexv3ContributionState = NewPdexv3ContributionState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPdexv3ContributionState)
		if err != nil {
			return nil, err
		}
	} else {
		newPdexv3ContributionState, ok = data.(*Pdexv3ContributionState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3ContributionStateType, reflect.TypeOf(data))
		}
	}
	return &Pdexv3ContributionObject{
		version:    defaultVersion,
		hash:       key,
		state:      newPdexv3ContributionState,
		db:         db,
		objectType: Pdexv3ContributionObjectType,
		deleted:    false,
	}, nil
}

func GeneratePdexv3ContributionObjectKey(pairHash string) common.Hash {
	prefixHash := GetPdexv3WaitingContributionsPrefix()
	valueHash := common.HashH([]byte(pairHash))
	return common.BytesToHash(append(prefixHash, valueHash[:prefixKeyLength]...))
}

func (pc *Pdexv3ContributionObject) GetVersion() int {
	return pc.version
}

// setError remembers the first non-nil error it is called with.
func (pc *Pdexv3ContributionObject) SetError(err error) {
	if pc.dbErr == nil {
		pc.dbErr = err
	}
}

func (pc *Pdexv3ContributionObject) GetTrie(db DatabaseAccessWarper) Trie {
	return pc.trie
}

func (pc *Pdexv3ContributionObject) SetValue(data interface{}) error {
	newPdexv3ContributionState, ok := data.(*Pdexv3ContributionState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3ContributionStateType, reflect.TypeOf(data))
	}
	pc.state = newPdexv3ContributionState
	return nil
}

func (pc *Pdexv3ContributionObject) GetValue() interface{} {
	return pc.state
}

func (pc *Pdexv3ContributionObject) GetValueBytes() []byte {
	state, ok := pc.GetValue().(*Pdexv3ContributionState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(state)
	if err != nil {
		panic("failed to marshal pdexv3 contribution state")
	}
	return value
}

func (pc *Pdexv3ContributionObject) GetHash() common.Hash {
	return pc.hash
}

func (pc *Pdexv3ContributionObject) GetType() int {
	return pc.objectType
}

// MarkDelete will delete an object in trie
func (pc *Pdexv3ContributionObject) MarkDelete() {
	pc.deleted = true
}

// reset all shard committee value into default value
func (pc *Pdexv3ContributionObject) Reset() bool {
	pc.state = NewPdexv3ContributionState()
	return true
}

func (pc *Pdexv3ContributionObject) IsDeleted() bool {
	return pc.deleted
}

// value is either default or nil
func (pc *Pdexv3ContributionObject) IsEmpty() bool {
	temp := NewPdexv3ContributionState()
	return reflect.DeepEqual(temp, pc.state) || pc.state == nil
}
