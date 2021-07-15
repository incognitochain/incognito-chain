package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type PDexV3Params struct {
	defaultFeeRateBPS        uint
	feeRateBPS               map[string]uint
	prvDiscountPercent       uint
	protocolFeePercent       uint
	stakingPoolRewardPercent uint
	defaultStakingPoolsShare uint
	stakingPoolsShare        map[string]uint
}

func (pp PDexV3Params) DefaultFeeRateBPS() uint {
	return pp.defaultFeeRateBPS
}
func (pp PDexV3Params) FeeRateBPS() map[string]uint {
	return pp.feeRateBPS
}
func (pp PDexV3Params) PRVDiscountPercent() uint {
	return pp.prvDiscountPercent
}
func (pp PDexV3Params) ProtocolFeePercent() uint {
	return pp.protocolFeePercent
}
func (pp PDexV3Params) StakingPoolRewardPercent() uint {
	return pp.stakingPoolRewardPercent
}
func (pp PDexV3Params) DefaultStakingPoolsShare() uint {
	return pp.defaultStakingPoolsShare
}
func (pp PDexV3Params) StakingPoolsShare() map[string]uint {
	return pp.stakingPoolsShare
}

func (pp PDexV3Params) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		DefaultFeeRateBPS        uint
		FeeRateBPS               map[string]uint
		PRVDiscountPercent       uint
		ProtocolFeePercent       uint
		StakingPoolRewardPercent uint
		DefaultStakingPoolsShare uint
		StakingPoolsShare        map[string]uint
	}{
		DefaultFeeRateBPS:        pp.defaultFeeRateBPS,
		FeeRateBPS:               pp.feeRateBPS,
		PRVDiscountPercent:       pp.prvDiscountPercent,
		ProtocolFeePercent:       pp.protocolFeePercent,
		StakingPoolRewardPercent: pp.stakingPoolRewardPercent,
		DefaultStakingPoolsShare: pp.defaultStakingPoolsShare,
		StakingPoolsShare:        pp.stakingPoolsShare,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (pp *PDexV3Params) UnmarshalJSON(data []byte) error {
	temp := struct {
		DefaultFeeRateBPS        uint
		FeeRateBPS               map[string]uint
		PRVDiscountPercent       uint
		ProtocolFeePercent       uint
		StakingPoolRewardPercent uint
		DefaultStakingPoolsShare uint
		StakingPoolsShare        map[string]uint
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	pp.defaultFeeRateBPS = temp.DefaultFeeRateBPS
	pp.feeRateBPS = temp.FeeRateBPS
	pp.prvDiscountPercent = temp.PRVDiscountPercent
	pp.protocolFeePercent = temp.ProtocolFeePercent
	pp.stakingPoolRewardPercent = temp.StakingPoolRewardPercent
	pp.defaultStakingPoolsShare = temp.DefaultStakingPoolsShare
	pp.stakingPoolsShare = temp.StakingPoolsShare
	return nil
}

func NewPDexV3Params() *PDexV3Params {
	return &PDexV3Params{}
}

func NewPDexV3ParamsWithValue(
	defaultFeeRateBPS uint,
	feeRateBPS map[string]uint,
	prvDiscountPercent uint,
	protocolFeePercent uint,
	stakingPoolRewardPercent uint,
	defaultStakingPoolsShare uint,
	stakingPoolsShare map[string]uint,
) *PDexV3Params {
	return &PDexV3Params{
		defaultFeeRateBPS:        defaultFeeRateBPS,
		feeRateBPS:               feeRateBPS,
		prvDiscountPercent:       prvDiscountPercent,
		protocolFeePercent:       protocolFeePercent,
		stakingPoolRewardPercent: stakingPoolRewardPercent,
		defaultStakingPoolsShare: defaultStakingPoolsShare,
		stakingPoolsShare:        stakingPoolsShare,
	}
}

type PDexV3ParamsObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version          int
	pDexV3ParamsHash common.Hash
	PDexV3Params     *PDexV3Params
	objectType       int
	deleted          bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPDexV3ParamsObject(db *StateDB, hash common.Hash) *PDexV3ParamsObject {
	return &PDexV3ParamsObject{
		version:          defaultVersion,
		db:               db,
		pDexV3ParamsHash: hash,
		PDexV3Params:     NewPDexV3Params(),
		objectType:       PDexV3ParamsObjectType,
		deleted:          false,
	}
}

func newPDexV3ParamsObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*PDexV3ParamsObject, error) {
	var newPDexV3Params = NewPDexV3Params()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPDexV3Params)
		if err != nil {
			return nil, err
		}
	} else {
		newPDexV3Params, ok = data.(*PDexV3Params)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPDexV3ParamsStateType, reflect.TypeOf(data))
		}
	}
	return &PDexV3ParamsObject{
		version:          defaultVersion,
		pDexV3ParamsHash: key,
		PDexV3Params:     newPDexV3Params,
		db:               db,
		objectType:       PDexV3ParamsObjectType,
		deleted:          false,
	}, nil
}

func GeneratePDexV3ParamsObjectKey() common.Hash {
	prefixHash := GetPDexV3ParamsPrefix()
	return common.BytesToHash(prefixHash)
}

func (t PDexV3ParamsObject) GetVersion() int {
	return t.version
}

// setError remembers the first non-nil error it is called with.
func (t *PDexV3ParamsObject) SetError(err error) {
	if t.dbErr == nil {
		t.dbErr = err
	}
}

func (t PDexV3ParamsObject) GetTrie(db DatabaseAccessWarper) Trie {
	return t.trie
}

func (t *PDexV3ParamsObject) SetValue(data interface{}) error {
	newPDexV3Params, ok := data.(*PDexV3Params)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPDexV3ParamsStateType, reflect.TypeOf(data))
	}
	t.PDexV3Params = newPDexV3Params
	return nil
}

func (t PDexV3ParamsObject) GetValue() interface{} {
	return t.PDexV3Params
}

func (t PDexV3ParamsObject) GetValueBytes() []byte {
	PDexV3Params, ok := t.GetValue().(*PDexV3Params)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(PDexV3Params)
	if err != nil {
		panic("failed to marshal pdex v3 params state")
	}
	return value
}

func (t PDexV3ParamsObject) GetHash() common.Hash {
	return t.pDexV3ParamsHash
}

func (t PDexV3ParamsObject) GetType() int {
	return t.objectType
}

// MarkDelete will delete an object in trie
func (t *PDexV3ParamsObject) MarkDelete() {
	t.deleted = true
}

// reset all shard committee value into default value
func (t *PDexV3ParamsObject) Reset() bool {
	t.PDexV3Params = NewPDexV3Params()
	return true
}

func (t PDexV3ParamsObject) IsDeleted() bool {
	return t.deleted
}

// value is either default or nil
func (t PDexV3ParamsObject) IsEmpty() bool {
	temp := NewPDexV3Params()
	return reflect.DeepEqual(temp, t.PDexV3Params) || t.PDexV3Params == nil
}
