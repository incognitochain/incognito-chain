package statedb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"reflect"
)

const (
	LOCKING_BY_UNSTAKE = 1
	LOCKING_BY_SLASH   = 2
)

type StakingInfo struct {
	txID   common.Hash
	amount uint64
}

type BeaconStakerInfo struct {
	rewardReceiver      key.PaymentAddress
	beaconConfirmHeight uint64
	stakingTx           map[common.Hash]uint64
	unstaking           bool
	shardActiveTime     int
	finishSync          bool
	lockingEpoch        uint64
	lockingReason       int
}

func NewBeaconStakerInfo() *BeaconStakerInfo {
	return &BeaconStakerInfo{}
}
func NewBeaconStakerInfoWithValue(rewardReceiver key.PaymentAddress, beaconConfirmHeight uint64, stakingTx map[common.Hash]uint64) *BeaconStakerInfo {
	return &BeaconStakerInfo{rewardReceiver: rewardReceiver, beaconConfirmHeight: beaconConfirmHeight, stakingTx: stakingTx}
}
func (c BeaconStakerInfo) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		RewardReceiver      key.PaymentAddress
		Unstaking           bool
		StakingInfo         map[common.Hash]uint64
		BeaconConfirmHeight uint64
		ShardActiveTime     int
		LockingEpoch        uint64
		LockingReason       int
	}{
		RewardReceiver:      c.rewardReceiver,
		Unstaking:           c.unstaking,
		StakingInfo:         c.stakingTx,
		BeaconConfirmHeight: c.beaconConfirmHeight,
		ShardActiveTime:     c.shardActiveTime,
		LockingEpoch:        c.lockingEpoch,
		LockingReason:       c.lockingReason,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (c *BeaconStakerInfo) UnmarshalJSON(data []byte) error {
	temp := struct {
		RewardReceiver      key.PaymentAddress
		Unstaking           bool
		StakingInfo         map[common.Hash]uint64
		BeaconConfirmHeight uint64
		ShardActiveTime     int
		LockingEpoch        uint64
		LockingReason       int
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	c.stakingTx = temp.StakingInfo
	c.rewardReceiver = temp.RewardReceiver
	c.beaconConfirmHeight = temp.BeaconConfirmHeight
	c.unstaking = temp.Unstaking
	c.shardActiveTime = temp.ShardActiveTime
	c.lockingEpoch = temp.LockingEpoch
	c.lockingReason = temp.LockingReason
	return nil
}
func (s *BeaconStakerInfo) SetUnstaking() {
	s.unstaking = true
}
func (s *BeaconStakerInfo) SetLocking(epoch uint64, reason int) {
	s.lockingEpoch = epoch
	s.lockingReason = reason
}

func (s *BeaconStakerInfo) SetFinishSync() {
	s.finishSync = true
}
func (s *BeaconStakerInfo) FinishSync() bool {
	return s.finishSync
}
func (s *BeaconStakerInfo) IncreaseShardActiveTime() {
	s.shardActiveTime++
}
func (s *BeaconStakerInfo) ResetShardActiveTime() {
	s.shardActiveTime = 0
}

func (s *BeaconStakerInfo) AddStakingInfo(tx common.Hash, amount uint64) {
	s.stakingTx[tx] = amount
}

func (s BeaconStakerInfo) TotalStakingAmount() uint64 {
	total := uint64(0)
	for _, amount := range s.stakingTx {
		total += amount
	}
	return total
}

func (s BeaconStakerInfo) Unstaking() bool {
	return s.unstaking
}

func (s BeaconStakerInfo) RewardReceiver() key.PaymentAddress {
	return s.rewardReceiver
}

func (s BeaconStakerInfo) BeaconConfirmHeight() uint64 {
	return s.beaconConfirmHeight
}
func (s BeaconStakerInfo) ShardActiveTime() int {
	return s.shardActiveTime
}
func (s BeaconStakerInfo) LockingEpoch() uint64 {
	return s.lockingEpoch
}
func (s BeaconStakerInfo) LockingReason() int {
	return s.lockingReason
}
func (s BeaconStakerInfo) StakingInfo() map[common.Hash]uint64 {
	res := map[common.Hash]uint64{}
	for k, v := range s.stakingTx {
		res[k] = v
	}
	return s.stakingTx
}

type BeaconStakerObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version             int
	stakerPublicKeyHash common.Hash
	stakerInfo          *BeaconStakerInfo
	objectType          int
	deleted             bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newBeaconStakerObject(db *StateDB, hash common.Hash) *BeaconStakerObject {
	return &BeaconStakerObject{
		version:             defaultVersion,
		db:                  db,
		stakerPublicKeyHash: hash,
		stakerInfo:          &BeaconStakerInfo{},
		objectType:          BeaconStakerObjectType,
		deleted:             false,
	}
}

func newBeaconStakerObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*BeaconStakerObject, error) {
	var newStakerInfo = NewBeaconStakerInfo()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newStakerInfo)
		if err != nil {
			return nil, err
		}
	} else {
		newStakerInfo, ok = data.(*BeaconStakerInfo)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidStakerInfoType, reflect.TypeOf(data))
		}
	}
	if err := SoValidation.ValidatePaymentAddressSanity(newStakerInfo.rewardReceiver); err != nil {
		return nil, fmt.Errorf("%+v, got err %+v, rewardReceiver %v", ErrInvalidPaymentAddressType, err, newStakerInfo.rewardReceiver)
	}
	return &BeaconStakerObject{
		version:             defaultVersion,
		stakerPublicKeyHash: key,
		stakerInfo:          newStakerInfo,
		db:                  db,
		objectType:          BeaconStakerObjectType,
		deleted:             false,
	}, nil
}

func (c BeaconStakerObject) GetVersion() int {
	return c.version
}

// setError remembers the first non-nil error it is called with.
func (c *BeaconStakerObject) SetError(err error) {
	if c.dbErr == nil {
		c.dbErr = err
	}
}

func (c BeaconStakerObject) GetTrie(db DatabaseAccessWarper) Trie {
	return c.trie
}

func (c *BeaconStakerObject) SetValue(data interface{}) error {
	newStakerInfo, ok := data.(*BeaconStakerInfo)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidStakerInfoType, reflect.TypeOf(data))
	}
	if err := SoValidation.ValidatePaymentAddressSanity(newStakerInfo.rewardReceiver); err != nil {
		return fmt.Errorf("%+v, got err %+v", ErrInvalidPaymentAddressType, err)
	}
	c.stakerInfo = newStakerInfo
	return nil
}

func (c BeaconStakerObject) GetValue() interface{} {
	return c.stakerInfo
}

func (c BeaconStakerObject) GetValueBytes() []byte {
	data := c.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal all shard committee")
	}
	return value
}

func (c BeaconStakerObject) GetHash() common.Hash {
	return c.stakerPublicKeyHash
}

func (c BeaconStakerObject) GetType() int {
	return c.objectType
}

// MarkDelete will delete an object in trie
func (c *BeaconStakerObject) MarkDelete() {
	c.deleted = true
}

// reset all shard committee value into default value
func (c *BeaconStakerObject) Reset() bool {
	c.stakerInfo = NewBeaconStakerInfo()
	return true
}

func (c BeaconStakerObject) IsDeleted() bool {
	return c.deleted
}

// value is either default or nil
func (c BeaconStakerObject) IsEmpty() bool {
	temp := NewBeaconStakerInfo()
	return reflect.DeepEqual(temp, c.stakerInfo) || c.stakerInfo == nil
}
