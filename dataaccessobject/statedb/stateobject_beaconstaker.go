package statedb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"reflect"
)

type StakingInfo struct {
	txID   common.Hash
	amount uint64
}

type BeaconStakerInfo struct {
	funderAddress       key.PaymentAddress
	rewardReceiver      key.PaymentAddress
	beaconConfirmHeight uint64
	stakingInfo         []StakingInfo
}

func NewBeaconStakerInfo() *BeaconStakerInfo {
	return &BeaconStakerInfo{}
}

func NewBeaconStakerInfoWithValue(
	rewardReceiver key.PaymentAddress,
	funderAddress key.PaymentAddress,
	autoStaking bool,
	txStakingIDs []common.Hash,
	beaconConfirmHeight uint64,
	stakingAmount uint64,
) *BeaconStakerInfo {
	return &BeaconStakerInfo{
		rewardReceiver:      rewardReceiver,
		funderAddress:       funderAddress,
		autoStaking:         autoStaking,
		txStakingIDs:        txStakingIDs,
		beaconConfirmHeight: beaconConfirmHeight,
		stakingAmount:       stakingAmount,
		delegatorReward:     uint64(0),
	}
}

func (c BeaconStakerInfo) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		RewardReceiver      key.PaymentAddress
		FunderAddress       key.PaymentAddress
		AutoStaking         bool
		TxStakingIDs        []common.Hash
		ShardID             byte
		NumberOfRound       int
		BeaconConfirmHeight uint64
		StakingAmount       uint64
	}{
		RewardReceiver:      c.rewardReceiver,
		FunderAddress:       c.funderAddress,
		TxStakingIDs:        c.txStakingIDs,
		AutoStaking:         c.autoStaking,
		BeaconConfirmHeight: c.beaconConfirmHeight,
		StakingAmount:       c.stakingAmount,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (c *BeaconStakerInfo) UnmarshalJSON(data []byte) error {
	temp := struct {
		RewardReceiver      key.PaymentAddress
		FunderAddress       key.PaymentAddress
		AutoStaking         bool
		TxStakingIDs        []common.Hash
		ShardID             byte
		NumberOfRound       int
		BeaconConfirmHeight uint64
		StakingAmount       uint64
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	c.txStakingIDs = temp.TxStakingIDs
	c.rewardReceiver = temp.RewardReceiver
	c.autoStaking = temp.AutoStaking
	c.beaconConfirmHeight = temp.BeaconConfirmHeight
	c.stakingAmount = temp.StakingAmount
	c.funderAddress = temp.FunderAddress
	return nil
}

func (s *BeaconStakerInfo) SetRewardReceiver(r key.PaymentAddress) {
	s.rewardReceiver = r
}

func (s *BeaconStakerInfo) SetTxStakingIDs(t []common.Hash) {
	s.txStakingIDs = t
}

func (s *BeaconStakerInfo) AddTxStakingID(t common.Hash) {
	s.txStakingIDs = append(s.txStakingIDs, t)
}

func (s *BeaconStakerInfo) SetAutoStaking(a bool) {
	s.autoStaking = a
}

func (s *BeaconStakerInfo) SetStakingAmount(a uint64) {
	s.stakingAmount = a
}

func (s *BeaconStakerInfo) SetFunderAddress(f key.PaymentAddress) {
	s.funderAddress = f
}
func (s *BeaconStakerInfo) AddStakingAmount(a uint64) {
	s.stakingAmount = a
}

func (s BeaconStakerInfo) RewardReceiver() key.PaymentAddress {
	return s.rewardReceiver
}

func (s BeaconStakerInfo) FunderAddress() key.PaymentAddress {
	return s.funderAddress
}

func (s BeaconStakerInfo) TxStakingIDs() []common.Hash {
	return s.txStakingIDs
}

func (s BeaconStakerInfo) AutoStaking() bool {
	return s.autoStaking
}

func (s BeaconStakerInfo) StakingAmount() uint64 {
	return s.stakingAmount
}

func (s BeaconStakerInfo) BeaconConfirmHeight() uint64 {
	return s.beaconConfirmHeight
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
		return nil, fmt.Errorf("%+v, got err %+v, staker key %v, payment %v", ErrInvalidPaymentAddressType, err, newStakerInfo.rewardReceiver, newStakerInfo.funderAddress)
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
