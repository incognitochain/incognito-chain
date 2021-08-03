package pdex

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"strconv"

	v2 "github.com/incognitochain/incognito-chain/blockchain/pdex/v2utils"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
)

type stateV2 struct {
	stateBase
	waitingContributions        map[string]rawdbv2.Pdexv3Contribution
	deletedWaitingContributions map[string]rawdbv2.Pdexv3Contribution
	poolPairs                   map[string]PoolPairState //
	orderbooks                  map[string]Orderbook
	params                      Params
	stakingPoolsState           map[string]StakingPoolState // tokenID -> StakingPoolState
	producer                    stateProducerV2
	processor                   stateProcessorV2
}

type Order = rawdbv2.Pdexv3Order

type Orderbook struct {
	orders []*Order
}

func (ob Orderbook) MarshalJSON() ([]byte, error) {
	temp := struct {
		Orders []*Order `json:"orders"`
	}{ob.orders}
	return json.Marshal(temp)
}

func (ob *Orderbook) UnmarshalJSON(data []byte) error {
	var temp struct {
		Orders []*Order `json:"orders"`
	}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	ob.orders = temp.Orders
	return nil
}

// InsertOrder() appends a new order while keeping the list sorted (ascending by Token1Rate / Token0Rate)
func (ob *Orderbook) InsertOrder(ord *Order) {
	insertAt := func(lst []*Order, i int, newItem *Order) []*Order {
		if i == len(lst) {
			return append(lst, newItem)
		}
		lst = append(lst[:i+1], lst[i:]...)
		lst[i] = newItem
		return lst
	}
	index := sort.Search(len(ob.orders), func(i int) bool {
		ordRate := big.NewInt(0).SetUint64(ob.orders[i].Token0Rate())
		ordRate.Mul(ordRate, big.NewInt(0).SetUint64(ord.Token1Rate()))
		myRate := big.NewInt(0).SetUint64(ob.orders[i].Token1Rate())
		myRate.Mul(myRate, big.NewInt(0).SetUint64(ord.Token0Rate()))
		// compare Token1Rate / Token0Rate of current order in the list to ord
		if ord.TradeDirection() == v2.TradeDirectionSell0 {
			// orders selling token0 are iterated from start of list (buy the least token1), so we resolve equality of rate by putting the new one last
			return ordRate.Cmp(myRate) < 0
		} else {
			// orders selling token1 are iterated from end of list (buy the least token0), so we resolve equality of rate by putting the new one first
			return ordRate.Cmp(myRate) <= 0
		}
	})
	ob.orders = insertAt(ob.orders, index, ord)
}

// NextOrder() returns the matchable order with the best rate that has any outstanding balance to sell
func (ob *Orderbook) NextOrder(tradeDirection byte) (*v2.MatchingOrder, string, error) {
	lstLen := len(ob.orders)
	switch tradeDirection {
	case v2.TradeDirectionSell0:
		for i := 0; i < lstLen; i++ {
			// only match a trade with an order of the opposite direction
			if ob.orders[i].TradeDirection() != tradeDirection && ob.orders[i].Token1Balance() > 0 {
				return &v2.MatchingOrder{ob.orders[i]}, ob.orders[i].Id(), nil
			}
		}
		// no active order
		return nil, "", nil
	case v2.TradeDirectionSell1:
		for i := lstLen - 1; i >= 0; i-- {
			if ob.orders[i].TradeDirection() != tradeDirection && ob.orders[i].Token0Balance() > 0 {
				return &v2.MatchingOrder{ob.orders[i]}, ob.orders[i].Id(), nil
			}
		}
		// no active order
		return nil, "", nil
	default:
		return nil, "", fmt.Errorf("Invalid trade direction %d", tradeDirection)
	}
}

type Params struct {
	DefaultFeeRateBPS               uint            // the default value if fee rate is not specific in FeeRateBPS (default 0.3% ~ 30 BPS)
	FeeRateBPS                      map[string]uint // map: pool ID -> fee rate (0.1% ~ 10 BPS)
	PRVDiscountPercent              uint            // percent of fee that will be discounted if using PRV as the trading token fee (default: 25%)
	LimitProtocolFeePercent         uint            // percent of fees from limit orders
	LimitStakingPoolRewardPercent   uint            // percent of fees from limit orders
	TradingProtocolFeePercent       uint            // percent of fees that is rewarded for the core team (default: 0%)
	TradingStakingPoolRewardPercent uint            // percent of fees that is distributed for staking pools (PRV, PDEX, ..., default: 10%)
	DefaultStakingPoolsShare        uint            // the default value of staking pool share weight (default - 0)
	StakingPoolsShare               map[string]uint // map: staking tokenID -> pool staking share weight
}

func newStateV2() *stateV2 {
	return &stateV2{
		params: Params{
			DefaultFeeRateBPS:               InitFeeRateBPS,
			FeeRateBPS:                      map[string]uint{},
			PRVDiscountPercent:              InitPRVDiscountPercent,
			LimitProtocolFeePercent:         InitProtocolFeePercent,
			LimitStakingPoolRewardPercent:   InitStakingPoolRewardPercent,
			TradingProtocolFeePercent:       InitProtocolFeePercent,
			TradingStakingPoolRewardPercent: InitStakingPoolRewardPercent,
			DefaultStakingPoolsShare:        InitStakingPoolsShare,
			StakingPoolsShare:               map[string]uint{},
		},
		waitingContributions:        make(map[string]rawdbv2.Pdexv3Contribution),
		deletedWaitingContributions: make(map[string]rawdbv2.Pdexv3Contribution),
		poolPairs:                   make(map[string]PoolPairState),
		stakingPoolsState:           make(map[string]StakingPoolState),
	}
}

func newStateV2WithValue(
	waitingContributions map[string]rawdbv2.Pdexv3Contribution,
	deletedWaitingContributions map[string]rawdbv2.Pdexv3Contribution,
	poolPairs map[string]PoolPairState,
	params Params,
	stakingPoolsState map[string]StakingPoolState,
	orderbooks map[string]Orderbook,
) *stateV2 {
	return &stateV2{
		waitingContributions:        waitingContributions,
		deletedWaitingContributions: deletedWaitingContributions,
		poolPairs:                   poolPairs,
		stakingPoolsState:           stakingPoolsState,
		orderbooks:                  orderbooks,
		params:                      params,
	}
}

func initStateV2(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
) (*stateV2, error) {
	stateObject, err := statedb.GetPdexv3Params(stateDB)
	params := Params{
		DefaultFeeRateBPS:               stateObject.DefaultFeeRateBPS(),
		FeeRateBPS:                      stateObject.FeeRateBPS(),
		PRVDiscountPercent:              stateObject.PRVDiscountPercent(),
		LimitProtocolFeePercent:         stateObject.LimitProtocolFeePercent(),
		LimitStakingPoolRewardPercent:   stateObject.LimitStakingPoolRewardPercent(),
		TradingProtocolFeePercent:       stateObject.TradingProtocolFeePercent(),
		TradingStakingPoolRewardPercent: stateObject.TradingStakingPoolRewardPercent(),
		DefaultStakingPoolsShare:        stateObject.DefaultStakingPoolsShare(),
		StakingPoolsShare:               stateObject.StakingPoolsShare(),
	}
	if err != nil {
		return nil, err
	}
	waitingContributions, err := statedb.GetPdexv3WaitingContributions(stateDB)
	if err != nil {
		return nil, err
	}
	poolPairsState, err := statedb.GetPdexv3PoolPairs(stateDB)
	if err != nil {
		return nil, err
	}
	poolPairs := make(map[string]PoolPairState)
	for k, v := range poolPairsState {
		sharesState, err := statedb.GetPdexv3Shares(stateDB, k)
		if err != nil {
			return nil, err
		}
		shares := make(map[string]Share)
		for key, value := range sharesState {
			tradingFeesState, err := statedb.GetPdexv3TradingFees(stateDB, k, key)
			if err != nil {
				return nil, err
			}
			tradingFees := make(map[string]uint64)
			for tradingFeesKey, tradingFeesValue := range tradingFeesState {
				tradingFees[tradingFeesKey] = tradingFeesValue.Amount()
			}
			shares[k] = *NewShareWithValue(value.Amount(), tradingFees, value.LastUpdatedBeaconHeight())
		}
		poolPair := NewPoolPairStateWithValue(
			v.Value(), shares,
		)
		poolPairs[k] = *poolPair
	}

	return newStateV2WithValue(
		waitingContributions,
		make(map[string]rawdbv2.Pdexv3Contribution),
		poolPairs,
		params,
		nil, nil,
	), nil
}

func (s *stateV2) Version() uint {
	return AmplifierVersion
}

func (s *stateV2) Clone() State {
	res := newStateV2()

	res.params = s.params
	clonedFeeRateBPS := map[string]uint{}
	for k, v := range s.params.FeeRateBPS {
		clonedFeeRateBPS[k] = v
	}
	clonedStakingPoolsShare := map[string]uint{}
	for k, v := range s.params.StakingPoolsShare {
		clonedStakingPoolsShare[k] = v
	}
	res.params.FeeRateBPS = clonedFeeRateBPS
	res.params.StakingPoolsShare = clonedStakingPoolsShare

	for k, v := range s.stakingPoolsState {
		res.stakingPoolsState[k] = v.Clone()
	}
	for k, v := range s.waitingContributions {
		res.waitingContributions[k] = *v.Clone()
	}
	for k, v := range s.deletedWaitingContributions {
		res.deletedWaitingContributions[k] = *v.Clone()
	}
	for k, v := range s.poolPairs {
		res.poolPairs[k] = v.Clone()
	}

	res.producer = s.producer
	res.processor = s.processor

	return res
}

func (s *stateV2) Process(env StateEnvironment) error {
	for _, inst := range env.BeaconInstructions() {
		if len(inst) < 2 {
			continue // Not error, just not PDE instructions
		}
		metadataType, err := strconv.Atoi(inst[0])
		if err != nil {
			continue // Not error, just not PDE instructions
		}
		if !metadataCommon.IsPdexv3Type(metadataType) {
			continue // Not error, just not PDE instructions
		}
		switch metadataType {
		case metadataCommon.Pdexv3ModifyParamsMeta:
			s.params, err = s.processor.modifyParams(
				env.StateDB(),
				env.BeaconHeight(),
				inst,
				s.params,
			)
		case metadataCommon.Pdexv3AddLiquidityResponseMeta:
			s.poolPairs,
				s.waitingContributions,
				s.deletedWaitingContributions, err = s.processor.addLiquidity(
				env.StateDB(),
				inst,
				env.BeaconHeight(),
				s.poolPairs,
				s.waitingContributions,
				s.deletedWaitingContributions,
			)
		case metadataCommon.Pdexv3TradeRequestMeta:
			s.poolPairs, s.orderbooks, err = s.processor.trade(env.StateDB(), inst,
				s.poolPairs, s.orderbooks,
			)
		default:
			Logger.log.Debug("Can not process this metadata")
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *stateV2) BuildInstructions(env StateEnvironment) ([][]string, error) {
	instructions := [][]string{}
	addLiquidityTxs := []metadata.Transaction{}
	addLiquidityInstructions := [][]string{}
	modifyParamsTxs := []metadata.Transaction{}
	tradeTxs := []metadata.Transaction{}

	var err error
	pdexv3Txs := env.ListTxs()
	keys := []int{}

	for k := range pdexv3Txs {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, key := range keys {
		for _, tx := range pdexv3Txs[byte(key)] {
			// TODO: @pdex get metadata here and build instructions from transactions here
			switch tx.GetMetadataType() {
			case metadataCommon.Pdexv3AddLiquidityRequestMeta:
				_, ok := tx.GetMetadata().(*metadataPdexv3.AddLiquidity)
				if !ok {
					return instructions, errors.New("Can not parse add liquidity metadata")
				}
				addLiquidityTxs = append(addLiquidityTxs, tx)
			case metadataCommon.Pdexv3ModifyParamsMeta:
				_, ok := tx.GetMetadata().(*metadataPdexv3.ParamsModifyingRequest)
				if !ok {
					return instructions, errors.New("Can not parse params modifying metadata")
				}
				modifyParamsTxs = append(modifyParamsTxs, tx)
			case metadataCommon.Pdexv3TradeRequestMeta:
				tradeTxs = append(tradeTxs, tx)
			}
		}
	}

	addLiquidityInstructions, s.poolPairs, s.waitingContributions, err = s.producer.addLiquidity(
		addLiquidityTxs,
		env.BeaconHeight(),
		s.poolPairs,
		s.waitingContributions,
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, addLiquidityInstructions...)

	// handle modify params
	var modifyParamsInstructions [][]string
	modifyParamsInstructions, s.params, err = s.producer.modifyParams(
		modifyParamsTxs,
		env.BeaconHeight(),
		s.params,
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, modifyParamsInstructions...)

	var tradeInstructions [][]string
	tradeInstructions, s.poolPairs, s.orderbooks, err = s.producer.trade(
		tradeTxs,
		s.poolPairs,
		s.orderbooks,
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, tradeInstructions...)

	return instructions, nil
}

func (s *stateV2) Upgrade(env StateEnvironment) State {
	return nil
}

func (s *stateV2) StoreToDB(env StateEnvironment, stateChange *StateChange) error {
	err := statedb.StorePdexv3Params(
		env.StateDB(),
		s.params.DefaultFeeRateBPS,
		s.params.FeeRateBPS,
		s.params.PRVDiscountPercent,
		s.params.LimitProtocolFeePercent,
		s.params.LimitStakingPoolRewardPercent,
		s.params.TradingProtocolFeePercent,
		s.params.TradingStakingPoolRewardPercent,
		s.params.DefaultStakingPoolsShare,
		s.params.StakingPoolsShare,
	)
	if err != nil {
		return err
	}
	deletedWaitingContributionsKeys := []string{}
	for k := range s.deletedWaitingContributions {
		deletedWaitingContributionsKeys = append(deletedWaitingContributionsKeys, k)
	}
	err = statedb.DeletePdexv3WaitingContributions(env.StateDB(), deletedWaitingContributionsKeys)
	if err != nil {
		return err
	}
	err = statedb.StorePdexv3WaitingContributions(env.StateDB(), s.waitingContributions)
	if err != nil {
		return err
	}

	for poolPairID, poolPairState := range s.poolPairs {
		if stateChange.poolPairIDs[poolPairID] {
			err := statedb.StorePdexv3PoolPair(env.StateDB(), poolPairID, poolPairState.state)
			if err != nil {
				return err
			}
		}
		for nfctID, share := range poolPairState.shares {
			if stateChange.nfctIDs[nfctID] {
				nfctIDHash, err := common.Hash{}.NewHashFromStr(nfctID)
				share := statedb.NewPdexv3ShareStateWithValue(*nfctIDHash, share.amount, share.lastUpdatedBeaconHeight)
				err = statedb.StorePdexv3Share(env.StateDB(), poolPairID, *share)
				if err != nil {
					return err
				}
			}
			for tokenID, tradingFee := range share.tradingFees {
				if stateChange.tokenIDs[tokenID] {
					err := statedb.StorePdexv3TradingFee(
						env.StateDB(), poolPairID, nfctID, tokenID, tradingFee,
					)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	err = statedb.StorePdexv3StakingPools()
	if err != nil {
		return err
	}

	return nil
}

func (s *stateV2) ClearCache() {
	s.deletedWaitingContributions = make(map[string]rawdbv2.Pdexv3Contribution)
}

func (s *stateV2) GetDiff(compareState State, stateChange *StateChange) (State, *StateChange, error) {
	newStateChange := stateChange
	if compareState == nil {
		return nil, newStateChange, errors.New("compareState is nil")
	}

	res := newStateV2()
	compareStateV2 := compareState.(*stateV2)

	res.params = s.params
	clonedFeeRateBPS := map[string]uint{}
	for k, v := range s.params.FeeRateBPS {
		clonedFeeRateBPS[k] = v
	}
	clonedStakingPoolsShare := map[string]uint{}
	for k, v := range s.params.StakingPoolsShare {
		clonedStakingPoolsShare[k] = v
	}
	res.params.FeeRateBPS = clonedFeeRateBPS
	res.params.StakingPoolsShare = clonedStakingPoolsShare

	for k, v := range s.waitingContributions {
		if m, ok := compareStateV2.waitingContributions[k]; !ok || !reflect.DeepEqual(m, v) {
			res.waitingContributions[k] = *v.Clone()
		}
	}
	for k, v := range s.deletedWaitingContributions {
		if m, ok := compareStateV2.deletedWaitingContributions[k]; !ok || !reflect.DeepEqual(m, v) {
			res.deletedWaitingContributions[k] = *v.Clone()
		}
	}
	for k, v := range s.poolPairs {
		if m, ok := compareStateV2.poolPairs[k]; !ok || !reflect.DeepEqual(m, v) {
			newStateChange = v.getDiff(k, &m, newStateChange)
			res.poolPairs[k] = v.Clone()
		}
	}
	for k, v := range s.stakingPoolsState {
		if m, ok := compareStateV2.stakingPoolsState[k]; !ok || !reflect.DeepEqual(m, v) {
			res.stakingPoolsState[k] = v.Clone()
		}
	}

	return res, newStateChange, nil

}

func (s *stateV2) Params() Params {
	return s.params
}

func (s *stateV2) Reader() StateReader {
	return s
}

func NewContributionWithMetaData(
	metaData metadataPdexv3.AddLiquidity, txReqID common.Hash, shardID byte,
) *rawdbv2.Pdexv3Contribution {
	tokenHash, _ := common.Hash{}.NewHashFromStr(metaData.TokenID())

	return rawdbv2.NewPdexv3ContributionWithValue(
		metaData.PoolPairID(), metaData.ReceiveAddress(), metaData.RefundAddress(),
		*tokenHash, txReqID, metaData.TokenAmount(), metaData.Amplifier(),
		shardID,
	)
}
