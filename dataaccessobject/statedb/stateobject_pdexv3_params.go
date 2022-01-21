package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type Pdexv3Params struct {
	defaultFeeRateBPS                 uint
	feeRateBPS                        map[string]uint
	prvDiscountPercent                uint
	tradingProtocolFeePercent         uint
	tradingStakingPoolRewardPercent   uint
	pdexRewardPoolPairsShare          map[string]uint
	stakingPoolsShare                 map[string]uint
	stakingRewardTokens               []common.Hash
	mintNftRequireAmount              uint64
	maxOrdersPerNft                   uint
	autoWithdrawOrderLimitAmount      uint
	minPRVReserveTradingRate          uint64
	defaultOrderTradingRewardRatioBPS uint
	orderTradingRewardRatioBPS        map[string]uint
	orderLiquidityMiningBPS           map[string]uint
	daoContributingPercent            uint
	miningRewardPendingBlocks         uint64
	orderMiningRewardRatioBPS         map[string]uint
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
func (pp Pdexv3Params) TradingProtocolFeePercent() uint {
	return pp.tradingProtocolFeePercent
}
func (pp Pdexv3Params) TradingStakingPoolRewardPercent() uint {
	return pp.tradingStakingPoolRewardPercent
}
func (pp Pdexv3Params) PDEXRewardPoolPairsShare() map[string]uint {
	return pp.pdexRewardPoolPairsShare
}
func (pp Pdexv3Params) StakingPoolsShare() map[string]uint {
	return pp.stakingPoolsShare
}
func (pp Pdexv3Params) StakingRewardTokens() []common.Hash {
	return pp.stakingRewardTokens
}
func (pp Pdexv3Params) MintNftRequireAmount() uint64 {
	return pp.mintNftRequireAmount
}
func (pp Pdexv3Params) MaxOrdersPerNft() uint {
	return pp.maxOrdersPerNft
}
func (pp Pdexv3Params) AutoWithdrawOrderLimitAmount() uint {
	return pp.autoWithdrawOrderLimitAmount
}

func (pp Pdexv3Params) MinPRVReserveTradingRate() uint64 {
	return pp.minPRVReserveTradingRate
}

func (pp Pdexv3Params) DefaultOrderTradingRewardRatioBPS() uint {
	return pp.defaultOrderTradingRewardRatioBPS
}

func (pp Pdexv3Params) OrderTradingRewardRatioBPS() map[string]uint {
	return pp.orderTradingRewardRatioBPS
}

func (pp Pdexv3Params) OrderLiquidityMiningBPS() map[string]uint {
	return pp.orderLiquidityMiningBPS
}

func (pp Pdexv3Params) DAOContributingPercent() uint {
	return pp.daoContributingPercent
}

func (pp Pdexv3Params) MiningRewardPendingBlocks() uint64 {
	return pp.miningRewardPendingBlocks
}

func (pp Pdexv3Params) OrderMiningRewardRatioBPS() map[string]uint {
	return pp.orderMiningRewardRatioBPS
}

func (pp Pdexv3Params) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		DefaultFeeRateBPS                 uint
		FeeRateBPS                        map[string]uint
		PRVDiscountPercent                uint
		TradingProtocolFeePercent         uint
		TradingStakingPoolRewardPercent   uint
		PDEXRewardPoolPairsShare          map[string]uint
		StakingPoolsShare                 map[string]uint
		StakingRewardTokens               []common.Hash
		MintNftRequireAmount              uint64
		MaxOrdersPerNft                   uint
		AutoWithdrawOrderLimitAmount      uint
		MinPRVReserveTradingRate          uint64
		DefaultOrderTradingRewardRatioBPS uint
		OrderTradingRewardRatioBPS        map[string]uint
		OrderLiquidityMiningBPS           map[string]uint
		DAOContributingPercent            uint
		MiningRewardPendingBlocks         uint64
		OrderMiningRewardRatioBPS         map[string]uint
	}{
		DefaultFeeRateBPS:                 pp.defaultFeeRateBPS,
		FeeRateBPS:                        pp.feeRateBPS,
		PRVDiscountPercent:                pp.prvDiscountPercent,
		TradingProtocolFeePercent:         pp.tradingProtocolFeePercent,
		TradingStakingPoolRewardPercent:   pp.tradingStakingPoolRewardPercent,
		PDEXRewardPoolPairsShare:          pp.pdexRewardPoolPairsShare,
		StakingPoolsShare:                 pp.stakingPoolsShare,
		StakingRewardTokens:               pp.stakingRewardTokens,
		MintNftRequireAmount:              pp.mintNftRequireAmount,
		MaxOrdersPerNft:                   pp.maxOrdersPerNft,
		AutoWithdrawOrderLimitAmount:      pp.autoWithdrawOrderLimitAmount,
		MinPRVReserveTradingRate:          pp.minPRVReserveTradingRate,
		DefaultOrderTradingRewardRatioBPS: pp.defaultOrderTradingRewardRatioBPS,
		OrderTradingRewardRatioBPS:        pp.orderTradingRewardRatioBPS,
		OrderLiquidityMiningBPS:           pp.orderLiquidityMiningBPS,
		DAOContributingPercent:            pp.daoContributingPercent,
		MiningRewardPendingBlocks:         pp.miningRewardPendingBlocks,
		OrderMiningRewardRatioBPS:         pp.orderMiningRewardRatioBPS,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (pp *Pdexv3Params) UnmarshalJSON(data []byte) error {
	temp := struct {
		DefaultFeeRateBPS                 uint
		FeeRateBPS                        map[string]uint
		PRVDiscountPercent                uint
		TradingProtocolFeePercent         uint
		TradingStakingPoolRewardPercent   uint
		PDEXRewardPoolPairsShare          map[string]uint
		StakingPoolsShare                 map[string]uint
		StakingRewardTokens               []common.Hash
		MintNftRequireAmount              uint64
		MaxOrdersPerNft                   uint
		AutoWithdrawOrderLimitAmount      uint
		MinPRVReserveTradingRate          uint64
		DefaultOrderTradingRewardRatioBPS uint
		OrderTradingRewardRatioBPS        map[string]uint
		OrderLiquidityMiningBPS           map[string]uint
		DAOContributingPercent            uint
		MiningRewardPendingBlocks         uint64
		OrderMiningRewardRatioBPS         map[string]uint
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	pp.defaultFeeRateBPS = temp.DefaultFeeRateBPS
	pp.feeRateBPS = temp.FeeRateBPS
	pp.prvDiscountPercent = temp.PRVDiscountPercent
	pp.tradingProtocolFeePercent = temp.TradingProtocolFeePercent
	pp.tradingStakingPoolRewardPercent = temp.TradingStakingPoolRewardPercent
	pp.pdexRewardPoolPairsShare = temp.PDEXRewardPoolPairsShare
	pp.stakingPoolsShare = temp.StakingPoolsShare
	pp.stakingRewardTokens = temp.StakingRewardTokens
	pp.mintNftRequireAmount = temp.MintNftRequireAmount
	pp.maxOrdersPerNft = temp.MaxOrdersPerNft
	pp.autoWithdrawOrderLimitAmount = temp.AutoWithdrawOrderLimitAmount
	pp.minPRVReserveTradingRate = temp.MinPRVReserveTradingRate
	pp.defaultOrderTradingRewardRatioBPS = temp.DefaultOrderTradingRewardRatioBPS
	if temp.OrderTradingRewardRatioBPS == nil {
		temp.OrderTradingRewardRatioBPS = make(map[string]uint)
	}
	pp.orderTradingRewardRatioBPS = temp.OrderTradingRewardRatioBPS
	if temp.OrderLiquidityMiningBPS == nil {
		temp.OrderLiquidityMiningBPS = make(map[string]uint)
	}
	pp.orderLiquidityMiningBPS = temp.OrderLiquidityMiningBPS
	pp.daoContributingPercent = temp.DAOContributingPercent
	pp.miningRewardPendingBlocks = temp.MiningRewardPendingBlocks
	if temp.OrderMiningRewardRatioBPS == nil {
		temp.OrderMiningRewardRatioBPS = make(map[string]uint)
	}
	pp.orderMiningRewardRatioBPS = temp.OrderMiningRewardRatioBPS
	return nil
}

func NewPdexv3Params() *Pdexv3Params {
	return &Pdexv3Params{}
}

func NewPdexv3ParamsWithValue(
	defaultFeeRateBPS uint,
	feeRateBPS map[string]uint,
	prvDiscountPercent uint,
	tradingProtocolFeePercent uint,
	tradingStakingPoolRewardPercent uint,
	pdexRewardPoolPairsShare map[string]uint,
	stakingPoolsShare map[string]uint,
	stakingRewardTokens []common.Hash,
	mintNftRequireAmount uint64,
	maxOrdersPerNft uint,
	autoWithdrawOrderLimitAmount uint,
	minPRVReserveTradingRate uint64,
	defaultOrderTradingRewardRatioBPS uint,
	orderTradingRewardRatioBPS map[string]uint,
	orderLiquidityMiningBPS map[string]uint,
	daoContributingPercent uint,
	miningRewardPendingBlocks uint64,
	orderMiningRewardRatioBPS map[string]uint,
) *Pdexv3Params {
	return &Pdexv3Params{
		defaultFeeRateBPS:                 defaultFeeRateBPS,
		feeRateBPS:                        feeRateBPS,
		prvDiscountPercent:                prvDiscountPercent,
		tradingProtocolFeePercent:         tradingProtocolFeePercent,
		tradingStakingPoolRewardPercent:   tradingStakingPoolRewardPercent,
		pdexRewardPoolPairsShare:          pdexRewardPoolPairsShare,
		stakingPoolsShare:                 stakingPoolsShare,
		stakingRewardTokens:               stakingRewardTokens,
		mintNftRequireAmount:              mintNftRequireAmount,
		maxOrdersPerNft:                   maxOrdersPerNft,
		autoWithdrawOrderLimitAmount:      autoWithdrawOrderLimitAmount,
		minPRVReserveTradingRate:          minPRVReserveTradingRate,
		defaultOrderTradingRewardRatioBPS: defaultOrderTradingRewardRatioBPS,
		orderTradingRewardRatioBPS:        orderTradingRewardRatioBPS,
		orderLiquidityMiningBPS:           orderLiquidityMiningBPS,
		daoContributingPercent:            daoContributingPercent,
		miningRewardPendingBlocks:         miningRewardPendingBlocks,
		orderMiningRewardRatioBPS:         orderMiningRewardRatioBPS,
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
