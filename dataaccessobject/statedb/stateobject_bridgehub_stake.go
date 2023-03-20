package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type BridgeStakingInfoState struct {
	validator     string
	stakingAmount uint64
}

func (b BridgeStakingInfoState) StakingAmount() uint64 {
	return b.stakingAmount
}

func (b *BridgeStakingInfoState) SetStakingAmount(stakingAmount uint64) {
	b.stakingAmount = stakingAmount
}

func (b BridgeStakingInfoState) Validator() string {
	return b.validator
}

func (b *BridgeStakingInfoState) SetValidator(validator string) {
	b.validator = validator
}

func (b BridgeStakingInfoState) Clone() *BridgeStakingInfoState {
	return &BridgeStakingInfoState{
		stakingAmount: b.stakingAmount,
		validator:     b.validator,
	}
}

func (b *BridgeStakingInfoState) IsDiff(compareParam *BridgeStakingInfoState) bool {
	if compareParam == nil {
		return true
	}
	return b.stakingAmount != compareParam.stakingAmount || b.validator != compareParam.validator
}

func (b BridgeStakingInfoState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Validator     string
		StakingAmount uint64
	}{
		Validator:     b.validator,
		StakingAmount: b.stakingAmount,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (b *BridgeStakingInfoState) UnmarshalJSON(data []byte) error {
	temp := struct {
		Validator     string
		StakingAmount uint64
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	b.validator = temp.Validator
	b.stakingAmount = temp.StakingAmount
	return nil
}

func NewBridgeStakingInfoState() *BridgeStakingInfoState {
	return &BridgeStakingInfoState{}
}

func NewBridgeStakingInfoStateWithValue(validator string, stakingAmount uint64) *BridgeStakingInfoState {
	return &BridgeStakingInfoState{
		validator:     validator,
		stakingAmount: stakingAmount,
	}
}

type BridgeHubStakingInfoObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *BridgeStakingInfoState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newBridgeHubStakingInfoObject(db *StateDB, hash common.Hash) *BridgeHubStakingInfoObject {
	return &BridgeHubStakingInfoObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewBridgeStakingInfoState(),
		objectType: BridgeHubStakerInfoObjectType,
		deleted:    false,
	}
}

func newBridgeHubStakingInfoObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*BridgeHubStakingInfoObject, error) {
	var newBridgeHubParam = NewBridgeStakingInfoState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgeHubParam)
		if err != nil {
			return nil, err
		}
	} else {
		newBridgeHubParam, ok = data.(*BridgeStakingInfoState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeHubStakingTxStateType, reflect.TypeOf(data))
		}
	}
	return &BridgeHubStakingInfoObject{
		version:    defaultVersion,
		hash:       key,
		state:      newBridgeHubParam,
		db:         db,
		objectType: BridgeHubStakerInfoObjectType,
		deleted:    false,
	}, nil
}

func GenerateBridgeHubStakingInfoObjectKey(validatorKey string) common.Hash {
	prefixHash := GetBridgeHubStakingTxPrefix()
	valueHash := common.HashH([]byte(validatorKey))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (t BridgeHubStakingInfoObject) GetVersion() int {
	return t.version
}

// setError remembers the first non-nil error it is called with.
func (t *BridgeHubStakingInfoObject) SetError(err error) {
	if t.dbErr == nil {
		t.dbErr = err
	}
}

func (t BridgeHubStakingInfoObject) GetTrie(db DatabaseAccessWarper) Trie {
	return t.trie
}

func (t *BridgeHubStakingInfoObject) SetValue(data interface{}) error {
	newBridgeHubParam, ok := data.(*BridgeStakingInfoState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeHubStakingTxStateType, reflect.TypeOf(data))
	}
	t.state = newBridgeHubParam
	return nil
}

func (t BridgeHubStakingInfoObject) GetValue() interface{} {
	return t.state
}

func (t BridgeHubStakingInfoObject) GetValueBytes() []byte {
	BridgeStakingInfoState, ok := t.GetValue().(*BridgeStakingInfoState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(BridgeStakingInfoState)
	if err != nil {
		panic("failed to marshal BridgeStakingInfoState")
	}
	return value
}

func (t BridgeHubStakingInfoObject) GetHash() common.Hash {
	return t.hash
}

func (t BridgeHubStakingInfoObject) GetType() int {
	return t.objectType
}

// MarkDelete will delete an object in trie
func (t *BridgeHubStakingInfoObject) MarkDelete() {
	t.deleted = true
}

// reset all shard committee value into default value
func (t *BridgeHubStakingInfoObject) Reset() bool {
	t.state = NewBridgeStakingInfoState()
	return true
}

func (t BridgeHubStakingInfoObject) IsDeleted() bool {
	return t.deleted
}

// value is either default or nil
func (t BridgeHubStakingInfoObject) IsEmpty() bool {
	return t.state == nil
}
