package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type AllShardStakersInfo struct {
	mapDelegate map[string]string
	hasCredit   map[string]bool
}

func NewAllShardStakersInfo() *AllShardStakersInfo {
	return &AllShardStakersInfo{}
}

func NewAllShardStakersInfoWithValue(
	mapDelegate map[string]string,
	hasCredit map[string]bool,
) *AllShardStakersInfo {
	return &AllShardStakersInfo{
		mapDelegate: mapDelegate,
		hasCredit:   hasCredit,
	}
}

func (c AllShardStakersInfo) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		MapDelegate map[string]string
		HasCredit   map[string]bool
	}{
		MapDelegate: c.mapDelegate,
		HasCredit:   c.hasCredit,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (c *AllShardStakersInfo) UnmarshalJSON(data []byte) error {
	temp := struct {
		MapDelegate         map[string]string
		AutoStaking         bool
		TxStakingID         common.Hash
		ShardID             byte
		NumberOfRound       int
		BeaconConfirmHeight uint64
		Delegate            string
		HasCredit           map[string]bool
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	c.mapDelegate = temp.MapDelegate
	c.hasCredit = temp.HasCredit
	return nil
}

func (s *AllShardStakersInfo) SetMapDelegate(m map[string]string) {
	s.mapDelegate = m
}

func (s *AllShardStakersInfo) SetHasCredit(m map[string]bool) {
	s.hasCredit = m
}

func (s AllShardStakersInfo) MapDelegate() map[string]string {
	return s.mapDelegate
}

func (s AllShardStakersInfo) HasCredit() map[string]bool {
	return s.hasCredit
}

type AllStakersObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version        int
	objectType     int
	objectKey      common.Hash
	allStakersInfo *AllShardStakersInfo
	deleted        bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newAllStakersObject(db *StateDB, hash common.Hash) *AllStakersObject {
	return &AllStakersObject{
		version:        defaultVersion,
		db:             db,
		objectKey:      hash,
		allStakersInfo: &AllShardStakersInfo{},
		objectType:     StakerObjectType,
		deleted:        false,
	}
}

func newAllStakersObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*AllStakersObject, error) {
	var newStakerInfo = NewAllShardStakersInfo()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newStakerInfo)
		if err != nil {
			return nil, err
		}
	} else {
		newStakerInfo, ok = data.(*AllShardStakersInfo)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidStakerInfoType, reflect.TypeOf(data))
		}
	}
	//TODO add SoValidation
	// if err := SoValidation.ValidatePaymentAddressSanity(newStakerInfo.mapDelegate); err != nil {
	// 	return nil, fmt.Errorf("%+v, got err %+v", ErrInvalidPaymentAddressType, err)
	// }
	return &AllStakersObject{
		version:        defaultVersion,
		objectKey:      key,
		allStakersInfo: newStakerInfo,
		db:             db,
		objectType:     StakerObjectType,
		deleted:        false,
	}, nil
}

func (c AllStakersObject) GetVersion() int {
	return c.version
}

// setError remembers the first non-nil error it is called with.
func (c *AllStakersObject) SetError(err error) {
	if c.dbErr == nil {
		c.dbErr = err
	}
}

func (c AllStakersObject) GetTrie(db DatabaseAccessWarper) Trie {
	return c.trie
}

func (c *AllStakersObject) SetValue(data interface{}) error {
	newStakerInfo, ok := data.(*AllShardStakersInfo)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidStakerInfoType, reflect.TypeOf(data))
	}
	//TODO add SoValidation
	// if err := SoValidation.ValidatePaymentAddressSanity(newStakerInfo.mapDelegate); err != nil {
	// 	return fmt.Errorf("%+v, got err %+v", ErrInvalidPaymentAddressType, err)
	// }
	c.allStakersInfo = newStakerInfo
	return nil
}

func (c AllStakersObject) GetValue() interface{} {
	return c.allStakersInfo
}

func (c AllStakersObject) GetValueBytes() []byte {
	data := c.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal all shard committee")
	}
	return value
}

func (c AllStakersObject) GetHash() common.Hash {
	return c.objectKey
}

func (c AllStakersObject) GetType() int {
	return c.objectType
}

// MarkDelete will delete an object in trie
func (c *AllStakersObject) MarkDelete() {
	c.deleted = true
}

// reset all shard committee value into default value
func (c *AllStakersObject) Reset() bool {
	c.allStakersInfo = NewAllShardStakersInfo()
	return true
}

func (c AllStakersObject) IsDeleted() bool {
	return c.deleted
}

// value is either default or nil
func (c AllStakersObject) IsEmpty() bool {
	temp := NewStakerInfo()
	return reflect.DeepEqual(temp, c.allStakersInfo) || c.allStakersInfo == nil
}
