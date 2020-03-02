package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type BlackListProducerState struct {
	// base58 string of committee public key
	producerCommitteePublicKey string
	punishedEpoches            uint8
	beaconHeight               uint64
}

func NewBlackListProducerStateWithValue(producerCommitteePublicKey string, punishedEpoches uint8, beaconHeight uint64) *BlackListProducerState {
	return &BlackListProducerState{producerCommitteePublicKey: producerCommitteePublicKey, punishedEpoches: punishedEpoches, beaconHeight: beaconHeight}
}

func (bl BlackListProducerState) BeaconHeight() uint64 {
	return bl.beaconHeight
}

func (bl *BlackListProducerState) SetBeaconHeight(beaconHeight uint64) {
	bl.beaconHeight = beaconHeight
}

func (bl BlackListProducerState) PunishedEpoches() uint8 {
	return bl.punishedEpoches
}

func (bl *BlackListProducerState) SetPunishedEpoches(punishedEpoches uint8) {
	bl.punishedEpoches = punishedEpoches
}

func (bl BlackListProducerState) ProducerCommitteePublicKey() string {
	return bl.producerCommitteePublicKey
}

func (bl *BlackListProducerState) SetProducerCommitteePublicKey(producerCommitteePublicKey string) {
	bl.producerCommitteePublicKey = producerCommitteePublicKey
}

func NewBlackListProducerState() *BlackListProducerState {
	return &BlackListProducerState{}
}

func (bl BlackListProducerState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		ProducerCommitteePublicKey string
		PunishedEpoches            uint8
		BeaconHeight               uint64
	}{
		ProducerCommitteePublicKey: bl.producerCommitteePublicKey,
		PunishedEpoches:            bl.punishedEpoches,
		BeaconHeight:               bl.beaconHeight,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (bl *BlackListProducerState) UnmarshalJSON(data []byte) error {
	temp := struct {
		ProducerCommitteePublicKey string
		PunishedEpoches            uint8
		BeaconHeight               uint64
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	bl.producerCommitteePublicKey = temp.ProducerCommitteePublicKey
	bl.punishedEpoches = temp.PunishedEpoches
	bl.beaconHeight = temp.BeaconHeight
	return nil
}

type BlackListProducerObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version                int
	committeePublicKeyHash common.Hash
	blackListProducerState *BlackListProducerState
	objectType             int
	deleted                bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newBlackListProducerObject(db *StateDB, hash common.Hash) *BlackListProducerObject {
	return &BlackListProducerObject{
		version:                defaultVersion,
		db:                     db,
		committeePublicKeyHash: hash,
		blackListProducerState: NewBlackListProducerState(),
		objectType:             BlackListProducerObjectType,
		deleted:                false,
	}
}

func newBlackListProducerObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*BlackListProducerObject, error) {
	var newBlackListProducerState = NewBlackListProducerState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBlackListProducerState)
		if err != nil {
			return nil, err
		}
	} else {
		newBlackListProducerState, ok = data.(*BlackListProducerState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidBlackListProducerStateType, reflect.TypeOf(data))
		}
	}
	return &BlackListProducerObject{
		version:                defaultVersion,
		committeePublicKeyHash: key,
		blackListProducerState: newBlackListProducerState,
		db:                     db,
		objectType:             BlackListProducerObjectType,
		deleted:                false,
	}, nil
}

func GenerateBlackListProducerObjectKey(committeePublicKey string) common.Hash {
	prefixHash := GetBlackListProducerPrefix()
	valueHash := common.HashH([]byte(committeePublicKey))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (bl BlackListProducerObject) GetVersion() int {
	return bl.version
}

// setError remembers the first non-nil error it is called with.
func (bl *BlackListProducerObject) SetError(err error) {
	if bl.dbErr == nil {
		bl.dbErr = err
	}
}

func (bl BlackListProducerObject) GetTrie(db DatabaseAccessWarper) Trie {
	return bl.trie
}

func (bl *BlackListProducerObject) SetValue(data interface{}) error {
	var newBlackListProducerState = NewBlackListProducerState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBlackListProducerState)
		if err != nil {
			return err
		}
	} else {
		newBlackListProducerState, ok = data.(*BlackListProducerState)
		if !ok {
			return fmt.Errorf("%+v, got type %+v", ErrInvalidBlackListProducerStateType, reflect.TypeOf(data))
		}
	}
	bl.blackListProducerState = newBlackListProducerState
	return nil
}

func (bl BlackListProducerObject) GetValue() interface{} {
	return bl.blackListProducerState
}

func (bl BlackListProducerObject) GetValueBytes() []byte {
	data := bl.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal blas list producer state")
	}
	return []byte(value)
}

func (bl BlackListProducerObject) GetHash() common.Hash {
	return bl.committeePublicKeyHash
}

func (bl BlackListProducerObject) GetType() int {
	return bl.objectType
}

// MarkDelete will delete an object in trie
func (bl *BlackListProducerObject) MarkDelete() {
	bl.deleted = true
}

func (bl *BlackListProducerObject) Reset() bool {
	bl.blackListProducerState = NewBlackListProducerState()
	return true
}

func (bl BlackListProducerObject) IsDeleted() bool {
	return bl.deleted
}

// value is either default or nil
func (bl BlackListProducerObject) IsEmpty() bool {
	temp := NewBlackListProducerState()
	return reflect.DeepEqual(temp, bl.blackListProducerState) || bl.blackListProducerState == nil
}
