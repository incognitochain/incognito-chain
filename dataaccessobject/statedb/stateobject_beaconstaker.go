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
	stakingTx           map[common.Hash]uint64
	unstaking           bool
	shardActiveTime     int
}

func NewBeaconStakerInfo() *BeaconStakerInfo {
	return &BeaconStakerInfo{}
}

func NewBeaconStakerInfoWithValue(
	rewardReceiver key.PaymentAddress,
	funderAddress key.PaymentAddress,
	stakingInfo map[common.Hash]uint64,
	beaconConfirmHeight uint64,
	shardActiveTime int,
	unstaking bool,

) *BeaconStakerInfo {
	return &BeaconStakerInfo{
		rewardReceiver:      rewardReceiver,
		funderAddress:       funderAddress,
		stakingTx:           stakingInfo,
		beaconConfirmHeight: beaconConfirmHeight,
		shardActiveTime:     shardActiveTime,
		unstaking:           unstaking,
	}
}

func (c BeaconStakerInfo) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		RewardReceiver      key.PaymentAddress
		FunderAddress       key.PaymentAddress
		Unstaking           bool
		StakingInfo         map[common.Hash]uint64
		BeaconConfirmHeight uint64
		ShardActiveTime     int
	}{
		RewardReceiver:      c.rewardReceiver,
		FunderAddress:       c.funderAddress,
		Unstaking:           c.unstaking,
		StakingInfo:         c.stakingTx,
		BeaconConfirmHeight: c.beaconConfirmHeight,
		ShardActiveTime:     c.shardActiveTime,
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
		Unstaking           bool
		StakingInfo         map[common.Hash]uint64
		BeaconConfirmHeight uint64
		ShardActiveTime     int
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	c.stakingTx = temp.StakingInfo
	c.rewardReceiver = temp.RewardReceiver
	c.beaconConfirmHeight = temp.BeaconConfirmHeight
	c.unstaking = temp.Unstaking
	c.funderAddress = temp.FunderAddress
	c.shardActiveTime = temp.ShardActiveTime
	return nil
}
func (s *BeaconStakerInfo) SetUnstaking() {
	s.unstaking = true
}

func (s *BeaconStakerInfo) IncreaseShardActiveTime() {
	s.shardActiveTime++
}

func (s *BeaconStakerInfo) AddStakingInfo(tx common.Hash, amount uint64) {
	s.stakingTx[tx] = amount
}

func (s BeaconStakerInfo) Unstaking() bool {
	return s.unstaking
}

func (s BeaconStakerInfo) RewardReceiver() key.PaymentAddress {
	return s.rewardReceiver
}

func (s BeaconStakerInfo) FunderAddress() key.PaymentAddress {
	return s.funderAddress
}

func (s BeaconStakerInfo) BeaconConfirmHeight() uint64 {
	return s.beaconConfirmHeight
}
func (s BeaconStakerInfo) ShardActiveTime() int {
	return s.shardActiveTime
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
	if err := SoValidation.ValidatePaymentAddressSanity(newStakerInfo.funderAddress); err != nil {
		return nil, fmt.Errorf("%+v, got err %+v, funderAddress %v", ErrInvalidPaymentAddressType, err, newStakerInfo.funderAddress)
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
	if err := SoValidation.ValidatePaymentAddressSanity(newStakerInfo.funderAddress); err != nil {
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
