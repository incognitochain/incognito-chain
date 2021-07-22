package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type Pdexv3Params struct {
	defaultFeeRateBPS               uint
	feeRateBPS                      map[string]uint
	prvDiscountPercent              uint
	limitProtocolFeePercent         uint
	limitStakingPoolRewardPercent   uint
	tradingProtocolFeePercent       uint
	tradingStakingPoolRewardPercent uint
	defaultStakingPoolsShare        uint
	stakingPoolsShare               map[string]uint
}

func (pp Pdexv3Params) DefaultFeeRateBPS() uint {
	return pp.defaultFeeRateBPS
}
func (pp Pdexv3Params) FeeRateBPS() map[string]uint {
	return pp.feeRateBPS
}
func (pp Pdexv3Params) PRVDiscountPercent() uint {
	return pp.prvDiscountPercent
}
func (pp Pdexv3Params) LimitProtocolFeePercent() uint {
	return pp.limitProtocolFeePercent
}
func (pp Pdexv3Params) LimitStakingPoolRewardPercent() uint {
	return pp.limitStakingPoolRewardPercent
}
func (pp Pdexv3Params) TradingProtocolFeePercent() uint {
	return pp.tradingProtocolFeePercent
}
func (pp Pdexv3Params) TradingStakingPoolRewardPercent() uint {
	return pp.tradingStakingPoolRewardPercent
}
func (pp Pdexv3Params) DefaultStakingPoolsShare() uint {
	return pp.defaultStakingPoolsShare
}
func (pp Pdexv3Params) StakingPoolsShare() map[string]uint {
	return pp.stakingPoolsShare
}

func (pp Pdexv3Params) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		DefaultFeeRateBPS               uint
		FeeRateBPS                      map[string]uint
		PRVDiscountPercent              uint
		LimitProtocolFeePercent         uint
		LimitStakingPoolRewardPercent   uint
		TradingProtocolFeePercent       uint
		TradingStakingPoolRewardPercent uint
		DefaultStakingPoolsShare        uint
		StakingPoolsShare               map[string]uint
	}{
		DefaultFeeRateBPS:               pp.defaultFeeRateBPS,
		FeeRateBPS:                      pp.feeRateBPS,
		PRVDiscountPercent:              pp.prvDiscountPercent,
		LimitProtocolFeePercent:         pp.limitProtocolFeePercent,
		LimitStakingPoolRewardPercent:   pp.limitStakingPoolRewardPercent,
		TradingProtocolFeePercent:       pp.tradingProtocolFeePercent,
		TradingStakingPoolRewardPercent: pp.tradingStakingPoolRewardPercent,
		DefaultStakingPoolsShare:        pp.defaultStakingPoolsShare,
		StakingPoolsShare:               pp.stakingPoolsShare,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (pp *Pdexv3Params) UnmarshalJSON(data []byte) error {
	temp := struct {
		DefaultFeeRateBPS               uint
		FeeRateBPS                      map[string]uint
		PRVDiscountPercent              uint
		LimitProtocolFeePercent         uint
		LimitStakingPoolRewardPercent   uint
		TradingProtocolFeePercent       uint
		TradingStakingPoolRewardPercent uint
		DefaultStakingPoolsShare        uint
		StakingPoolsShare               map[string]uint
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	pp.defaultFeeRateBPS = temp.DefaultFeeRateBPS
	pp.feeRateBPS = temp.FeeRateBPS
	pp.prvDiscountPercent = temp.PRVDiscountPercent
	pp.limitProtocolFeePercent = temp.LimitProtocolFeePercent
	pp.limitStakingPoolRewardPercent = temp.LimitStakingPoolRewardPercent
	pp.tradingProtocolFeePercent = temp.TradingProtocolFeePercent
	pp.tradingStakingPoolRewardPercent = temp.TradingStakingPoolRewardPercent
	pp.defaultStakingPoolsShare = temp.DefaultStakingPoolsShare
	pp.stakingPoolsShare = temp.StakingPoolsShare
	return nil
}

func NewPdexv3Params() *Pdexv3Params {
	return &Pdexv3Params{}
}

func NewPdexv3ParamsWithValue(
	defaultFeeRateBPS uint,
	feeRateBPS map[string]uint,
	prvDiscountPercent uint,
	limitProtocolFeePercent uint,
	limitStakingPoolRewardPercent uint,
	tradingProtocolFeePercent uint,
	tradingStakingPoolRewardPercent uint,
	defaultStakingPoolsShare uint,
	stakingPoolsShare map[string]uint,
) *Pdexv3Params {
	return &Pdexv3Params{
		defaultFeeRateBPS:               defaultFeeRateBPS,
		feeRateBPS:                      feeRateBPS,
		prvDiscountPercent:              prvDiscountPercent,
		limitProtocolFeePercent:         limitProtocolFeePercent,
		limitStakingPoolRewardPercent:   limitStakingPoolRewardPercent,
		tradingProtocolFeePercent:       tradingProtocolFeePercent,
		tradingStakingPoolRewardPercent: tradingStakingPoolRewardPercent,
		defaultStakingPoolsShare:        defaultStakingPoolsShare,
		stakingPoolsShare:               stakingPoolsShare,
	}
}

type Pdexv3ParamsObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version          int
	pdexv3ParamsHash common.Hash
	Pdexv3Params     *Pdexv3Params
	objectType       int
	deleted          bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPdexv3ParamsObject(db *StateDB, hash common.Hash) *Pdexv3ParamsObject {
	return &Pdexv3ParamsObject{
		version:          defaultVersion,
		db:               db,
		pdexv3ParamsHash: hash,
		Pdexv3Params:     NewPdexv3Params(),
		objectType:       Pdexv3ParamsObjectType,
		deleted:          false,
	}
}

func newPdexv3ParamsObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*Pdexv3ParamsObject, error) {
	var newPdexv3Params = NewPdexv3Params()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPdexv3Params)
		if err != nil {
			return nil, err
		}
	} else {
		newPdexv3Params, ok = data.(*Pdexv3Params)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3ParamsStateType, reflect.TypeOf(data))
		}
	}
	return &Pdexv3ParamsObject{
		version:          defaultVersion,
		pdexv3ParamsHash: key,
		Pdexv3Params:     newPdexv3Params,
		db:               db,
		objectType:       Pdexv3ParamsObjectType,
		deleted:          false,
	}, nil
}

func GeneratePdexv3ParamsObjectKey() common.Hash {
	prefixHash := GetPdexv3ParamsPrefix()
	return common.HashH(prefixHash)
}

func (t Pdexv3ParamsObject) GetVersion() int {
	return t.version
}

// setError remembers the first non-nil error it is called with.
func (t *Pdexv3ParamsObject) SetError(err error) {
	if t.dbErr == nil {
		t.dbErr = err
	}
}

func (t Pdexv3ParamsObject) GetTrie(db DatabaseAccessWarper) Trie {
	return t.trie
}

func (t *Pdexv3ParamsObject) SetValue(data interface{}) error {
	newPdexv3Params, ok := data.(*Pdexv3Params)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3ParamsStateType, reflect.TypeOf(data))
	}
	t.Pdexv3Params = newPdexv3Params
	return nil
}

func (t Pdexv3ParamsObject) GetValue() interface{} {
	return t.Pdexv3Params
}

func (t Pdexv3ParamsObject) GetValueBytes() []byte {
	Pdexv3Params, ok := t.GetValue().(*Pdexv3Params)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(Pdexv3Params)
	if err != nil {
		panic("failed to marshal pdex v3 params state")
	}
	return value
}

func (t Pdexv3ParamsObject) GetHash() common.Hash {
	return t.pdexv3ParamsHash
}

func (t Pdexv3ParamsObject) GetType() int {
	return t.objectType
}

// MarkDelete will delete an object in trie
func (t *Pdexv3ParamsObject) MarkDelete() {
	t.deleted = true
}

// reset all shard committee value into default value
func (t *Pdexv3ParamsObject) Reset() bool {
	t.Pdexv3Params = NewPdexv3Params()
	return true
}

func (t Pdexv3ParamsObject) IsDeleted() bool {
	return t.deleted
}

// value is either default or nil
func (t Pdexv3ParamsObject) IsEmpty() bool {
	temp := NewPdexv3Params()
	return reflect.DeepEqual(temp, t.Pdexv3Params) || t.Pdexv3Params == nil
}
