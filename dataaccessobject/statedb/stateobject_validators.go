package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type Validators struct {
	shardID int
	role    int
	keys    []incognitokey.CommitteePublicKey
}

func NewValidators() *Validators {
	return &Validators{}
}

func NewValidatorsStateWithValue(
	shardID, role int,
	keys []incognitokey.CommitteePublicKey,
) *Validators {
	return &Validators{
		shardID: shardID,
		role:    role,
		keys:    keys,
	}
}

func (validators *Validators) ShardID() int {
	return validators.shardID
}

func (validators *Validators) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		ShardID int
		Role    int
		Keys    []incognitokey.CommitteePublicKey
	}{
		ShardID: validators.shardID,
		Role:    validators.role,
		Keys:    validators.keys,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (validators *Validators) UnmarshalJSON(data []byte) error {
	temp := struct {
		ShardID int
		Role    int
		Keys    []incognitokey.CommitteePublicKey
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	validators.shardID = temp.ShardID
	validators.role = temp.Role
	validators.keys = temp.Keys
	return nil
}

type ValidatorsObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version            int
	validatorsObjecKey common.Hash
	validators         *Validators
	objectType         int
	deleted            bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newValidatorsObject(db *StateDB, hash common.Hash) *ValidatorsObject {
	return &ValidatorsObject{
		version:            defaultVersion,
		db:                 db,
		validatorsObjecKey: hash,
		validators:         NewValidators(),
		objectType:         ValidatorsObjectType,
		deleted:            false,
	}
}

func newValidatorsObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*ValidatorsObject, error) {
	var newValidators = NewValidators()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newValidators)
		if err != nil {
			return nil, err
		}
	} else {
		newValidators, ok = data.(*Validators)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidValidators, reflect.TypeOf(data))
		}
	}
	return &ValidatorsObject{
		version:            defaultVersion,
		validatorsObjecKey: key,
		validators:         newValidators,
		db:                 db,
		objectType:         ValidatorsObjectType,
		deleted:            false,
	}, nil
}

func (validatorsObject *ValidatorsObject) GetVersion() int {
	return validatorsObject.version
}

// setError remembers the first non-nil error it is called with.
func (validatorsObject *ValidatorsObject) SetError(err error) {
	if validatorsObject.dbErr == nil {
		validatorsObject.dbErr = err
	}
}

func (validatorsObject ValidatorsObject) GetTrie(db DatabaseAccessWarper) Trie {
	return validatorsObject.trie
}

func (validatorsObject *ValidatorsObject) SetValue(data interface{}) error {
	newValidators, ok := data.(*Validators)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidCommitteeStateType, reflect.TypeOf(data))
	}
	validatorsObject.validators = newValidators
	return nil
}

func (validatorsObject *ValidatorsObject) GetValue() interface{} {
	return validatorsObject.validators
}

func (validatorsObject *ValidatorsObject) GetValueBytes() []byte {
	data := validatorsObject.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal all shard committee")
	}
	return value
}

func (validatorsObject *ValidatorsObject) GetHash() common.Hash {
	return validatorsObject.validatorsObjecKey
}

func (validatorsObject *ValidatorsObject) GetType() int {
	return validatorsObject.objectType
}

// MarkDelete will delete an object in trie
func (validatorsObject *ValidatorsObject) MarkDelete() {
	validatorsObject.deleted = true
}

// reset all shard committee value into default value
func (validatorsObject *ValidatorsObject) Reset() bool {
	validatorsObject.validators = NewValidators()
	return true
}

func (validatorsObject *ValidatorsObject) IsDeleted() bool {
	return validatorsObject.deleted
}

// value is either default or nil
func (validatorsObject *ValidatorsObject) IsEmpty() bool {
	temp := NewValidators()
	return reflect.DeepEqual(temp, validatorsObject.validators) || validatorsObject.validators == nil
}

func generateValidatorsObjectKey(role, shardID int) common.Hash {
	switch role {
	case NextEpochShardCandidate:
		return common.HashH(nextShardCandidatePrefix)
	case CurrentEpochShardCandidate:
		return common.HashH(currentShardCandidatePrefix)
	case NextEpochBeaconCandidate:
		return common.HashH(nextBeaconCandidatePrefix)
	case CurrentEpochBeaconCandidate:
		return common.HashH(currentBeaconCandidatePrefix)
	case SubstituteValidator:
		return common.HashH([]byte(string(substitutePrefix) + strconv.Itoa(shardID)))
	case CurrentValidator:
		return common.HashH([]byte(string(committeePrefix) + strconv.Itoa(shardID)))
	case SyncingValidators:
		return common.HashH([]byte(string(syncingValidatorsPrefix) + strconv.Itoa(shardID)))
	default:
		panic("role not exist: " + strconv.Itoa(role))
	}
	return common.Hash{}
}
