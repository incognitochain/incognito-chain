package pdex

import (
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/incognitochain/incognito-chain/utils"
)

type stateV2 struct {
	stateBase
	waitingContributions        map[string]rawdbv2.Pdexv3Contribution
	deletedWaitingContributions map[string]rawdbv2.Pdexv3Contribution
	poolPairs                   map[string]*PoolPairState
	params                      Params
	stakingPoolStates           map[string]*StakingPoolState // tokenID -> StakingPoolState
	nftIDs                      map[string]uint64
	producer                    stateProducerV2
	processor                   stateProcessorV2
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
	MintNftRequireAmount            uint64          // amount prv for depositing to pdex
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
			MintNftRequireAmount:            InitMintNftRequireAmount,
		},
		waitingContributions:        make(map[string]rawdbv2.Pdexv3Contribution),
		deletedWaitingContributions: make(map[string]rawdbv2.Pdexv3Contribution),
		poolPairs:                   make(map[string]*PoolPairState),
		stakingPoolStates:           make(map[string]*StakingPoolState),
		nftIDs:                      make(map[string]uint64),
	}
}

func newStateV2WithValue(
	waitingContributions map[string]rawdbv2.Pdexv3Contribution,
	deletedWaitingContributions map[string]rawdbv2.Pdexv3Contribution,
	poolPairs map[string]*PoolPairState,
	params Params,
	stakingPoolStates map[string]*StakingPoolState,
	nftIDs map[string]uint64,
) *stateV2 {
	return &stateV2{
		waitingContributions:        waitingContributions,
		deletedWaitingContributions: deletedWaitingContributions,
		poolPairs:                   poolPairs,
		stakingPoolStates:           stakingPoolStates,
		params:                      params,
		nftIDs:                      nftIDs,
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
		MintNftRequireAmount:            stateObject.MintNftRequireAmount(),
	}
	if err != nil {
		return nil, err
	}
	waitingContributions, err := statedb.GetPdexv3WaitingContributions(stateDB)
	if err != nil {
		return nil, err
	}
	poolPairsStates, err := statedb.GetPdexv3PoolPairs(stateDB)
	if err != nil {
		return nil, err
	}
	poolPairs := make(map[string]*PoolPairState)
	for poolPairID, poolPairState := range poolPairsStates {
		shares := make(map[string]*Share)
		shareStates := make(map[string]statedb.Pdexv3ShareState)
		shareStates, err = statedb.GetPdexv3Shares(stateDB, poolPairID)
		if err != nil {
			return nil, err
		}
		for nftID, shareState := range shareStates {
			tradingFeesState, err := statedb.GetPdexv3TradingFees(stateDB, poolPairID, nftID)
			if err != nil {
				return nil, err
			}
			tradingFees := make(map[string]uint64)
			for tradingFeesKey, tradingFeesValue := range tradingFeesState {
				tradingFees[tradingFeesKey] = tradingFeesValue.Amount()
			}
			shares[nftID] = NewShareWithValue(
				shareState.Amount(),
				tradingFees, shareState.LastUpdatedBeaconHeight(),
			)
		}

		orderbook := &Orderbook{[]*Order{}}
		orderMap, err := statedb.GetPdexv3Orders(stateDB, poolPairState.PoolPairID())
		if err != nil {
			return nil, err
		}
		for _, item := range orderMap {
			v := item.Value()
			orderbook.InsertOrder(&v)
		}
		poolPair := NewPoolPairStateWithValue(
			poolPairState.Value(), shares, *orderbook,
		)
		poolPairs[poolPairID] = poolPair
	}

	nftIDs, err := statedb.GetPdexv3NftIDs(stateDB)
	if err != nil {
		return nil, err
	}
	stakingPoolStates := make(map[string]*StakingPoolState)
	for stakingPoolID := range params.StakingPoolsShare {
		stakerStates, liquidity, err := statedb.GetPdexv3Stakers(stateDB, stakingPoolID)
		if err != nil {
			return nil, err
		}
		stakers := make(map[string]*Staker)
		for nftID, stakerState := range stakerStates {
			rewards, err := statedb.GetPdexv3StakerRewards(stateDB, stakingPoolID, nftID)
			if err != nil {
				return nil, err
			}
			stakers[nftID] = NewStakerWithValue(stakerState.Liquidity(), stakerState.LastUpdatedBeaconHeight(), rewards)
		}
		stakingPoolStates[stakingPoolID] = NewStakingPoolStateWithValue(liquidity, stakers)
	}

	return newStateV2WithValue(
		waitingContributions, make(map[string]rawdbv2.Pdexv3Contribution),
		poolPairs, params, stakingPoolStates, nftIDs,
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

	for k, v := range s.stakingPoolStates {
		res.stakingPoolStates[k] = v.Clone()
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
	for k, v := range s.nftIDs {
		res.nftIDs[k] = v
	}
	res.producer = s.producer
	res.processor = s.processor

	return res
}

func (s *stateV2) Process(env StateEnvironment) error {
	s.processor.clearCache()
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
		case metadataCommon.Pdexv3UserMintNftRequestMeta:
			s.nftIDs, _, err = s.processor.userMintNft(env.StateDB(), inst, s.nftIDs)
			if err != nil {
				Logger.log.Debugf("process inst %s err %v:", inst, err)
				continue
			}
		case metadataCommon.Pdexv3ModifyParamsMeta:
			s.params, err = s.processor.modifyParams(
				env.StateDB(),
				env.BeaconHeight(),
				inst,
				s.params,
			)
		case metadataCommon.Pdexv3AddLiquidityRequestMeta:
			s.poolPairs,
				s.waitingContributions,
				s.deletedWaitingContributions, err = s.processor.addLiquidity(
				env.StateDB(),
				inst,
				env.BeaconHeight(),
				s.poolPairs,
				s.waitingContributions, s.deletedWaitingContributions,
			)
		case metadataCommon.Pdexv3WithdrawLiquidityRequestMeta:
			s.poolPairs, err = s.processor.withdrawLiquidity(env.StateDB(), inst, s.poolPairs)
		case metadataCommon.Pdexv3TradeRequestMeta:
			s.poolPairs, err = s.processor.trade(env.StateDB(), inst,
				s.poolPairs,
			)
		case metadataCommon.Pdexv3AddOrderRequestMeta:
			s.poolPairs, err = s.processor.addOrder(env.StateDB(), inst,
				s.poolPairs,
			)
		case metadataCommon.Pdexv3WithdrawOrderRequestMeta:
			s.poolPairs, err = s.processor.withdrawOrder(env.StateDB(), inst,
				s.poolPairs,
			)
		case metadataCommon.Pdexv3StakingRequestMeta:
			s.stakingPoolStates, _, err = s.processor.staking(env.StateDB(), inst, s.nftIDs, s.stakingPoolStates)
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
	withdrawLiquidityTxs := []metadata.Transaction{}
	modifyParamsTxs := []metadata.Transaction{}
	tradeTxs := []metadata.Transaction{}
	mintNftTxs := []metadata.Transaction{}
	addOrderTxs := []metadata.Transaction{}
	withdrawOrderTxs := []metadata.Transaction{}
	stakingTxs := []metadata.Transaction{}

	var err error
	pdexv3Txs := env.ListTxs()
	keys := []int{}

	for k := range pdexv3Txs {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, key := range keys {
		for _, tx := range pdexv3Txs[byte(key)] {
			switch tx.GetMetadataType() {
			case metadataCommon.Pdexv3UserMintNftRequestMeta:
				mintNftTxs = append(mintNftTxs, tx)
			case metadataCommon.Pdexv3AddLiquidityRequestMeta:
				addLiquidityTxs = append(addLiquidityTxs, tx)
			case metadataCommon.Pdexv3WithdrawLiquidityRequestMeta:
				withdrawLiquidityTxs = append(withdrawLiquidityTxs, tx)
			case metadataCommon.Pdexv3ModifyParamsMeta:
				modifyParamsTxs = append(modifyParamsTxs, tx)
			case metadataCommon.Pdexv3TradeRequestMeta:
				tradeTxs = append(tradeTxs, tx)
			case metadataCommon.Pdexv3AddOrderRequestMeta:
				addOrderTxs = append(addOrderTxs, tx)
			case metadataCommon.Pdexv3WithdrawOrderRequestMeta:
				withdrawOrderTxs = append(withdrawOrderTxs, tx)
			case metadataCommon.Pdexv3StakingRequestMeta:
				stakingTxs = append(stakingTxs, tx)
			}
		}
	}

	mintNftInstructions := [][]string{}
	mintNftInstructions, s.nftIDs, err = s.producer.userMintNft(mintNftTxs, s.nftIDs, env.BeaconHeight(), s.params.MintNftRequireAmount)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, mintNftInstructions...)

	withdrawLiquidityInstructions := [][]string{}
	withdrawLiquidityInstructions, s.poolPairs, err = s.producer.withdrawLiquidity(withdrawLiquidityTxs, s.poolPairs, s.nftIDs)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, withdrawLiquidityInstructions...)

	addLiquidityInstructions := [][]string{}
	addLiquidityInstructions, s.poolPairs, s.waitingContributions, err = s.producer.addLiquidity(
		addLiquidityTxs,
		env.BeaconHeight(),
		s.poolPairs,
		s.waitingContributions,
		s.nftIDs,
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
	tradeInstructions, s.poolPairs, err = s.producer.trade(
		tradeTxs,
		s.poolPairs,
		s.params,
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, tradeInstructions...)

	var addOrderInstructions [][]string
	addOrderInstructions, s.poolPairs, err = s.producer.addOrder(
		addOrderTxs,
		s.poolPairs,
		s.nftIDs,
		s.params,
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, addOrderInstructions...)

	var withdrawOrderInstructions [][]string
	withdrawOrderInstructions, s.poolPairs, err = s.producer.withdrawOrder(
		withdrawOrderTxs,
		s.poolPairs,
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, withdrawOrderInstructions...)

	var stakingInstructions [][]string
	stakingInstructions, s.stakingPoolStates, err = s.producer.staking(
		stakingTxs, s.nftIDs, s.stakingPoolStates, env.BeaconHeight(),
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, stakingInstructions...)

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
		s.params.MintNftRequireAmount,
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
		for nftID, share := range poolPairState.shares {
			if stateChange.shares[nftID] == nil {
				continue
			}
			if stateChange.shares[nftID].isChanged {
				nftID, err := common.Hash{}.NewHashFromStr(nftID)
				err = statedb.StorePdexv3Share(
					env.StateDB(), poolPairID,
					*nftID,
					share.amount, share.lastUpdatedBeaconHeight,
				)
				if err != nil {
					return err
				}
			}
			for tokenID, tradingFee := range share.tradingFees {
				if stateChange.shares[nftID].tokenIDs == nil {
					continue
				}
				if stateChange.shares[nftID].tokenIDs[tokenID] {
					err := statedb.StorePdexv3TradingFee(
						env.StateDB(), poolPairID, nftID, tokenID, tradingFee,
					)
					if err != nil {
						return err
					}
				}
			}
		}

		ordersByID := make(map[string]*Order)
		for _, ord := range poolPairState.orderbook.orders {
			ordersByID[ord.Id()] = ord
		}
		for orderID, changed := range stateChange.orderIDs {
			if changed {
				if order, exists := ordersByID[orderID]; exists {
					// update order in db
					orderState := statedb.NewPdexv3OrderStateWithValue(poolPairID, *order)
					err = statedb.StorePdexv3Order(env.StateDB(), *orderState)
					if err != nil {
						return err
					}
				} else {
					// delete order from db
					err = statedb.DeletePdexv3Order(env.StateDB(), poolPairID, orderID)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	err = statedb.StorePdexv3NftIDs(env.StateDB(), s.nftIDs)
	if err != nil {
		return err
	}
	for k, v := range s.stakingPoolStates {
		for nftID, staker := range v.stakers {
			if stateChange.stakingPool[k][nftID] == nil {
				continue
			}
			if stateChange.stakingPool[k][nftID].isChanged {
				nftHash, _ := common.Hash{}.NewHashFromStr(nftID)
				state := statedb.NewPdexv3StakerStateWithValue(
					*nftHash,
					staker.liquidity,
					staker.lastUpdatedBeaconHeight,
				)
				err = statedb.StorePdexv3Staker(env.StateDB(), k, nftID, state)
				if err != nil {
					return err
				}
			}
			if stateChange.stakingPool[k][nftID].tokenIDs == nil {
				continue
			}
			for tokenID := range stateChange.stakingPool[k][nftID].tokenIDs {
				tokenHash, err := common.Hash{}.NewHashFromStr(tokenID)
				if err != nil {
					return err
				}
				state := statedb.NewPdexv3StakerRewardStateWithValue(*tokenHash, staker.rewards[tokenID])
				err = statedb.StorePdexv3StakerReward(env.StateDB(), k, nftID, tokenID, state)
				if err != nil {
					return err
				}
			}
		}
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
			newStateChange = v.getDiff(k, m, newStateChange)
			res.poolPairs[k] = v.Clone()
		}
	}
	for k, v := range s.stakingPoolStates {
		if m, ok := compareStateV2.stakingPoolStates[k]; !ok || !reflect.DeepEqual(m, v) {
			newStateChange = v.getDiff(k, m, newStateChange)
			res.stakingPoolStates[k] = v.Clone()
		}
	}
	for k, v := range s.nftIDs {
		if m, ok := compareStateV2.nftIDs[k]; !ok || !reflect.DeepEqual(m, v) {
			res.nftIDs[k] = v
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
	metaData metadataPdexv3.AddLiquidityRequest, txReqID common.Hash, shardID byte,
) *rawdbv2.Pdexv3Contribution {
	tokenHash, _ := common.Hash{}.NewHashFromStr(metaData.TokenID())
	nftID := common.Hash{}
	if metaData.NftID() != utils.EmptyString {
		nftHash, _ := common.Hash{}.NewHashFromStr(metaData.NftID())
		nftID = *nftHash
	}
	return rawdbv2.NewPdexv3ContributionWithValue(
		metaData.PoolPairID(), metaData.OtaReceive(), metaData.OtaRefund(),
		*tokenHash, txReqID, nftID,
		metaData.TokenAmount(), metaData.Amplifier(),
		shardID,
	)
}

func (s *stateV2) WaitingContributions() []byte {
	temp := make(map[string]*rawdbv2.Pdexv3Contribution, len(s.waitingContributions))
	for k, v := range s.waitingContributions {
		temp[k] = v.Clone()
	}
	data, _ := json.Marshal(temp)
	return data
}

func (s *stateV2) PoolPairs() []byte {
	temp := make(map[string]*PoolPairState, len(s.poolPairs))
	for k, v := range s.poolPairs {
		temp[k] = v.Clone()
	}
	data, _ := json.Marshal(temp)
	return data
}

func (s *stateV2) TransformKeyWithNewBeaconHeight(beaconHeight uint64) {}

func (s *stateV2) NftIDs() map[string]uint64 {
	res := make(map[string]uint64)
	for k, v := range s.nftIDs {
		res[k] = v
	}
	return res
}
