package statedb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

//@NOTE this struct is view object only
type StakerInfoSlashingVersion struct {
	committeePublicKey string
	rewardReceiver     key.PaymentAddress
	txStakingID        common.Hash
	autoStaking        bool
}

func NewStakerInfoSlashingVersionWithCommittee(committeePublicKey string) *StakerInfoSlashingVersion {
	return &StakerInfoSlashingVersion{committeePublicKey: committeePublicKey}
}

func (s StakerInfoSlashingVersion) CommitteePublicKey() string {
	return s.committeePublicKey
}

func (s StakerInfoSlashingVersion) RewardReceiver() key.PaymentAddress {
	return s.rewardReceiver
}

func (s StakerInfoSlashingVersion) TxStakingID() common.Hash {
	return s.txStakingID
}

func (s StakerInfoSlashingVersion) AutoStaking() bool {
	return s.autoStaking
}

func NewStakerInfoSlashingVersion(committeePublicKey string, s *StakerInfo) *StakerInfoSlashingVersion {
	return &StakerInfoSlashingVersion{
		committeePublicKey: committeePublicKey,
		rewardReceiver:     s.rewardReceiver,
		txStakingID:        s.txStakingID,
		autoStaking:        s.autoStaking,
	}
}

type StakerInfo struct {
	rewardReceiver      key.PaymentAddress
	txStakingID         common.Hash
	autoStaking         bool
	beaconConfirmHeight uint64
}

func NewStakerInfo() *StakerInfo {
	return &StakerInfo{}
}

func NewStakerInfoWithValue(
	rewardReceiver key.PaymentAddress,
	autoStaking bool,
	txStakingID common.Hash,
	beaconConfirmHeight uint64,
) *StakerInfo {
	return &StakerInfo{
		rewardReceiver:      rewardReceiver,
		autoStaking:         autoStaking,
		txStakingID:         txStakingID,
		beaconConfirmHeight: beaconConfirmHeight,
	}
}

func (c StakerInfo) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		RewardReceiver      key.PaymentAddress
		AutoStaking         bool
		TxStakingID         common.Hash
		ShardID             byte
		NumberOfRound       int
		BeaconConfirmHeight uint64
	}{
		RewardReceiver:      c.rewardReceiver,
		TxStakingID:         c.txStakingID,
		AutoStaking:         c.autoStaking,
		BeaconConfirmHeight: c.beaconConfirmHeight,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (c *StakerInfo) UnmarshalJSON(data []byte) error {
	temp := struct {
		RewardReceiver      key.PaymentAddress
		AutoStaking         bool
		TxStakingID         common.Hash
		ShardID             byte
		NumberOfRound       int
		BeaconConfirmHeight uint64
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	c.txStakingID = temp.TxStakingID
	c.rewardReceiver = temp.RewardReceiver
	c.autoStaking = temp.AutoStaking
	c.beaconConfirmHeight = temp.BeaconConfirmHeight
	return nil
}

func (s *StakerInfo) SetRewardReceiver(r key.PaymentAddress) {
	s.rewardReceiver = r
}

func (s *StakerInfo) SetTxStakingID(t common.Hash) {
	s.txStakingID = t
}

func (s *StakerInfo) SetAutoStaking(a bool) {
	s.autoStaking = a
}

func (s StakerInfo) RewardReceiver() key.PaymentAddress {
	return s.rewardReceiver
}

func (s StakerInfo) TxStakingID() common.Hash {
	return s.txStakingID
}

func (s StakerInfo) AutoStaking() bool {
	return s.autoStaking
}

func (s StakerInfo) BeaconConfirmHeight() uint64 {
	return s.beaconConfirmHeight
}

type StakerObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version             int
	stakerPublicKeyHash common.Hash
	stakerInfo          *StakerInfo
	objectType          int
	deleted             bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newStakerObject(db *StateDB, hash common.Hash) *StakerObject {
	return &StakerObject{
		version:             defaultVersion,
		db:                  db,
		stakerPublicKeyHash: hash,
		stakerInfo:          &StakerInfo{},
		objectType:          ShardStakerObjectType,
		deleted:             false,
	}
}

func newStakerObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*StakerObject, error) {
	var newStakerInfo = NewStakerInfo()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newStakerInfo)
		if err != nil {
			return nil, err
		}
	} else {
		newStakerInfo, ok = data.(*StakerInfo)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidStakerInfoType, reflect.TypeOf(data))
		}
	}
	if err := SoValidation.ValidatePaymentAddressSanity(newStakerInfo.rewardReceiver); err != nil {
		return nil, fmt.Errorf("%+v, got err %+v", ErrInvalidPaymentAddressType, err)
	}
	return &StakerObject{
		version:             defaultVersion,
		stakerPublicKeyHash: key,
		stakerInfo:          newStakerInfo,
		db:                  db,
		objectType:          ShardStakerObjectType,
		deleted:             false,
	}, nil
}

func (c StakerObject) GetVersion() int {
	return c.version
}

// setError remembers the first non-nil error it is called with.
func (c *StakerObject) SetError(err error) {
	if c.dbErr == nil {
		c.dbErr = err
	}
}

func (c StakerObject) GetTrie(db DatabaseAccessWarper) Trie {
	return c.trie
}

func (c *StakerObject) SetValue(data interface{}) error {
	newStakerInfo, ok := data.(*StakerInfo)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidStakerInfoType, reflect.TypeOf(data))
	}
	if err := SoValidation.ValidatePaymentAddressSanity(newStakerInfo.rewardReceiver); err != nil {
		return fmt.Errorf("%+v, got err %+v", ErrInvalidPaymentAddressType, err)
	}
	c.stakerInfo = newStakerInfo
	return nil
}

func (c StakerObject) GetValue() interface{} {
	return c.stakerInfo
}

func (c StakerObject) GetValueBytes() []byte {
	data := c.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal all shard committee")
	}
	return value
}

func (c StakerObject) GetHash() common.Hash {
	return c.stakerPublicKeyHash
}

func (c StakerObject) GetType() int {
	return c.objectType
}

// MarkDelete will delete an object in trie
func (c *StakerObject) MarkDelete() {
	c.deleted = true
}

// reset all shard committee value into default value
func (c *StakerObject) Reset() bool {
	c.stakerInfo = NewStakerInfo()
	return true
}

func (c StakerObject) IsDeleted() bool {
	return c.deleted
}

// value is either default or nil
func (c StakerObject) IsEmpty() bool {
	temp := NewStakerInfo()
	return reflect.DeepEqual(temp, c.stakerInfo) || c.stakerInfo == nil
}
