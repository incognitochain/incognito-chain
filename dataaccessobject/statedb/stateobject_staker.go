package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/privacy/key"

	"github.com/incognitochain/incognito-chain/common"
)

// @NOTE this struct is view object only
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

func NewStakerInfoSlashingVersion(committeePublicKey string, s *ShardStakerInfo) *StakerInfoSlashingVersion {
	return &StakerInfoSlashingVersion{
		committeePublicKey: committeePublicKey,
		rewardReceiver:     s.rewardReceiver,
		txStakingID:        s.txStakingID,
		autoStaking:        s.autoStaking,
	}
}

type ShardStakerInfo struct {
	rewardReceiver      key.PaymentAddress
	txStakingID         common.Hash
	autoStaking         bool
	beaconConfirmHeight uint64
	delegate            string
	hasCredit           bool
	activeInCommittee   int
}

func NewShardStakerInfo() *ShardStakerInfo {
	return &ShardStakerInfo{}
}

func NewShardStakerInfoWithValue(
	rewardReceiver key.PaymentAddress,
	autoStaking bool,
	txStakingID common.Hash,
	beaconConfirmHeight uint64,
	delegate string,
	hasCredit bool,
) *ShardStakerInfo {
	return &ShardStakerInfo{
		rewardReceiver:      rewardReceiver,
		autoStaking:         autoStaking,
		txStakingID:         txStakingID,
		beaconConfirmHeight: beaconConfirmHeight,
		delegate:            delegate,
		hasCredit:           hasCredit,
		activeInCommittee:   0,
	}
}

func (c ShardStakerInfo) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		RewardReceiver      key.PaymentAddress
		AutoStaking         bool
		TxStakingID         common.Hash
		ShardID             byte
		NumberOfRound       int
		BeaconConfirmHeight uint64
		Delegate            string
		HasCredit           bool
		ActiveTimes         int
	}{
		RewardReceiver:      c.rewardReceiver,
		TxStakingID:         c.txStakingID,
		AutoStaking:         c.autoStaking,
		BeaconConfirmHeight: c.beaconConfirmHeight,
		Delegate:            c.delegate,
		HasCredit:           c.hasCredit,
		ActiveTimes:         c.activeInCommittee,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (c *ShardStakerInfo) UnmarshalJSON(data []byte) error {
	temp := struct {
		RewardReceiver      key.PaymentAddress
		AutoStaking         bool
		TxStakingID         common.Hash
		ShardID             byte
		NumberOfRound       int
		BeaconConfirmHeight uint64
		Delegate            string
		HasCredit           bool
		ActiveTimes         int
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	c.txStakingID = temp.TxStakingID
	c.rewardReceiver = temp.RewardReceiver
	c.autoStaking = temp.AutoStaking
	c.beaconConfirmHeight = temp.BeaconConfirmHeight
	c.delegate = temp.Delegate
	c.hasCredit = temp.HasCredit
	c.activeInCommittee = temp.ActiveTimes
	return nil
}

func (s *ShardStakerInfo) SetRewardReceiver(r key.PaymentAddress) {
	s.rewardReceiver = r
}

func (s *ShardStakerInfo) SetTxStakingID(t common.Hash) {
	s.txStakingID = t
}

func (s *ShardStakerInfo) SetAutoStaking(a bool) {
	s.autoStaking = a
}

func (s *ShardStakerInfo) SetHasCredit(c bool) {
	s.hasCredit = c
}

func (s *ShardStakerInfo) SetDelegate(delegate string) {
	s.delegate = delegate
}

func (s *ShardStakerInfo) SetActiveTimesInCommittee(activeTimes int) {
	s.activeInCommittee = activeTimes
}

func (s *ShardStakerInfo) ActiveTimesInCommittee() int {
	return s.activeInCommittee
}

func (s ShardStakerInfo) RewardReceiver() key.PaymentAddress {
	return s.rewardReceiver
}

func (s ShardStakerInfo) TxStakingID() common.Hash {
	return s.txStakingID
}

func (s ShardStakerInfo) AutoStaking() bool {
	return s.autoStaking
}

func (s ShardStakerInfo) HasCredit() bool {
	return s.hasCredit
}

func (s ShardStakerInfo) BeaconConfirmHeight() uint64 {
	return s.beaconConfirmHeight
}

func (s ShardStakerInfo) Delegate() string {
	return s.delegate
}

type ShardStakerObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version             int
	stakerPublicKeyHash common.Hash
	stakerInfo          *ShardStakerInfo
	objectType          int
	deleted             bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newShardStakerObject(db *StateDB, hash common.Hash) *ShardStakerObject {
	return &ShardStakerObject{
		version:             defaultVersion,
		db:                  db,
		stakerPublicKeyHash: hash,
		stakerInfo:          &ShardStakerInfo{},
		objectType:          ShardStakerObjectType,
		deleted:             false,
	}
}

func newShardStakerObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*ShardStakerObject, error) {
	var newStakerInfo = NewShardStakerInfo()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newStakerInfo)
		if err != nil {
			return nil, err
		}
	} else {
		newStakerInfo, ok = data.(*ShardStakerInfo)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidStakerInfoType, reflect.TypeOf(data))
		}
	}
	if err := SoValidation.ValidatePaymentAddressSanity(newStakerInfo.rewardReceiver); err != nil {
		return nil, fmt.Errorf("%+v, got err %+v", ErrInvalidPaymentAddressType, err)
	}
	return &ShardStakerObject{
		version:             defaultVersion,
		stakerPublicKeyHash: key,
		stakerInfo:          newStakerInfo,
		db:                  db,
		objectType:          ShardStakerObjectType,
		deleted:             false,
	}, nil
}

func (c ShardStakerObject) GetVersion() int {
	return c.version
}

// setError remembers the first non-nil error it is called with.
func (c *ShardStakerObject) SetError(err error) {
	if c.dbErr == nil {
		c.dbErr = err
	}
}

func (c ShardStakerObject) GetTrie(db DatabaseAccessWarper) Trie {
	return c.trie
}

func (c *ShardStakerObject) SetValue(data interface{}) error {
	newStakerInfo, ok := data.(*ShardStakerInfo)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidStakerInfoType, reflect.TypeOf(data))
	}
	if err := SoValidation.ValidatePaymentAddressSanity(newStakerInfo.rewardReceiver); err != nil {
		return fmt.Errorf("%+v, got err %+v", ErrInvalidPaymentAddressType, err)
	}
	c.stakerInfo = newStakerInfo
	return nil
}

func (c ShardStakerObject) GetValue() interface{} {
	return c.stakerInfo
}

func (c ShardStakerObject) GetValueBytes() []byte {
	data := c.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal all shard committee")
	}
	return value
}

func (c ShardStakerObject) GetHash() common.Hash {
	return c.stakerPublicKeyHash
}

func (c ShardStakerObject) GetType() int {
	return c.objectType
}

// MarkDelete will delete an object in trie
func (c *ShardStakerObject) MarkDelete() {
	c.deleted = true
}

// reset all shard committee value into default value
func (c *ShardStakerObject) Reset() bool {
	c.stakerInfo = NewShardStakerInfo()
	return true
}

func (c ShardStakerObject) IsDeleted() bool {
	return c.deleted
}

// value is either default or nil
func (c ShardStakerObject) IsEmpty() bool {
	temp := NewShardStakerInfo()
	return reflect.DeepEqual(temp, c.stakerInfo) || c.stakerInfo == nil
}

type BeaconStakerInfo struct {
	rewardReceiver      key.PaymentAddress
	txStakingIDs        []common.Hash
	autoStaking         bool
	beaconConfirmHeight uint64
	stakingAmount       uint64
	delegatorReward     uint64
}

func NewBeaconStakerInfo() *BeaconStakerInfo {
	return &BeaconStakerInfo{}
}

func NewBeaconStakerInfoWithValue(
	rewardReceiver key.PaymentAddress,
	autoStaking bool,
	txStakingIDs []common.Hash,
	beaconConfirmHeight uint64,
	stakingAmount uint64,
) *BeaconStakerInfo {
	return &BeaconStakerInfo{
		rewardReceiver:      rewardReceiver,
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
		AutoStaking         bool
		TxStakingIDs        []common.Hash
		ShardID             byte
		NumberOfRound       int
		BeaconConfirmHeight uint64
		StakingAmount       uint64
	}{
		RewardReceiver:      c.rewardReceiver,
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

func (s *BeaconStakerInfo) AddStakingAmount(a uint64) {
	s.stakingAmount = a
}

func (s BeaconStakerInfo) RewardReceiver() key.PaymentAddress {
	return s.rewardReceiver
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
		return nil, fmt.Errorf("%+v, got err %+v", ErrInvalidPaymentAddressType, err)
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
