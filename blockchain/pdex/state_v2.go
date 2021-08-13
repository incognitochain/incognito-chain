package pdex

import (
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
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
	params                      *Params
	stakingPoolsState           map[string]*StakingPoolState // tokenID -> StakingPoolState
	nftIDs                      map[string]bool
	producer                    stateProducerV2
	processor                   stateProcessorV2
}

func newStateV2() *stateV2 {
	return &stateV2{
		params:                      NewParams(),
		waitingContributions:        make(map[string]rawdbv2.Pdexv3Contribution),
		deletedWaitingContributions: make(map[string]rawdbv2.Pdexv3Contribution),
		poolPairs:                   make(map[string]*PoolPairState),
		stakingPoolsState:           make(map[string]*StakingPoolState),
		nftIDs:                      make(map[string]bool),
	}
}

func newStateV2WithValue(
	waitingContributions map[string]rawdbv2.Pdexv3Contribution,
	deletedWaitingContributions map[string]rawdbv2.Pdexv3Contribution,
	poolPairs map[string]*PoolPairState,
	params *Params,
	stakingPoolsState map[string]*StakingPoolState,
	nftIDs map[string]bool,
) *stateV2 {
	return &stateV2{
		waitingContributions:        waitingContributions,
		deletedWaitingContributions: deletedWaitingContributions,
		poolPairs:                   poolPairs,
		stakingPoolsState:           stakingPoolsState,
		params:                      params,
		nftIDs:                      nftIDs,
	}
}

func initStateV2(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
) (*stateV2, error) {
	paramsState, err := statedb.GetPdexv3Params(stateDB)
	params := NewParamsWithValue(paramsState)
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
	nftIDs := make(map[string]bool)
	poolPairs := make(map[string]*PoolPairState)
	for poolPairID, poolPairState := range poolPairsStates {
		allShareStates, err := statedb.GetPdexv3Shares(stateDB, poolPairID, nftIDs)
		if err != nil {
			return nil, err
		}
		shares := make(map[string]map[uint64]*Share)
		for nftID, shareStates := range allShareStates {
			shares[nftID] = make(map[uint64]*Share)
			for beaconHeight, shareState := range shareStates {
				tradingFeesState, err := statedb.GetPdexv3TradingFees(
					stateDB, poolPairID, nftID, beaconHeight)
				if err != nil {
					return nil, err
				}
				tradingFees := make(map[string]uint64)
				for tradingFeesKey, tradingFeesValue := range tradingFeesState {
					tradingFees[tradingFeesKey] = tradingFeesValue.Amount()
				}
				shares[nftID][beaconHeight] = NewShareWithValue(
					shareState.Amount(),
					tradingFees, shareState.LastUpdatedBeaconHeight(),
				)
			}
		}
		// TODO: read order book from storage
		orderbook := Orderbook{}
		poolPair := NewPoolPairStateWithValue(poolPairState.Value(), shares, orderbook)
		poolPairs[poolPairID] = poolPair
	}

	return newStateV2WithValue(
		waitingContributions, make(map[string]rawdbv2.Pdexv3Contribution),
		poolPairs, params, nil, nftIDs,
	), nil
}

func (s *stateV2) Version() uint {
	return AmplifierVersion
}

func (s *stateV2) Clone() State {
	res := newStateV2()
	res.params = s.params.Clone()

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
	for k, v := range s.nftIDs {
		res.nftIDs[k] = v
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
		case metadataCommon.Pdexv3MintPDEXBlockRewardMeta:
			s.poolPairs, err = s.processor.mintPDEX(
				env.StateDB(),
				inst,
				s.poolPairs,
			)
		case metadataCommon.Pdexv3AddLiquidityRequestMeta:
			s.poolPairs,
				s.waitingContributions,
				s.deletedWaitingContributions, s.nftIDs, err = s.processor.addLiquidity(
				env.StateDB(),
				inst,
				env.BeaconHeight(),
				s.poolPairs,
				s.waitingContributions, s.deletedWaitingContributions, s.nftIDs,
			)
		case metadataCommon.Pdexv3TradeRequestMeta:
			s.poolPairs, err = s.processor.trade(env.StateDB(), inst,
				s.poolPairs,
			)
		case metadataCommon.Pdexv3WithdrawLPFeeRequestMeta:
			s.poolPairs, err = s.processor.withdrawLPFee(
				env.StateDB(),
				inst,
				env.BeaconHeight(),
				s.poolPairs,
			)
		case metadataCommon.Pdexv3WithdrawProtocolFeeRequestMeta:
			s.poolPairs, err = s.processor.withdrawProtocolFee(
				env.StateDB(),
				inst,
				s.poolPairs,
			)
		case metadataCommon.Pdexv3ModifyParamsMeta:
			s.params, err = s.processor.modifyParams(
				env.StateDB(),
				inst,
				s.params,
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
	withdrawLPFeeTxs := []metadata.Transaction{}
	withdrawlProtocolFeeTxs := []metadata.Transaction{}
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
			switch tx.GetMetadataType() {
			case metadataCommon.Pdexv3AddLiquidityRequestMeta:
				addLiquidityTxs = append(addLiquidityTxs, tx)
			case metadataCommon.Pdexv3TradeRequestMeta:
				tradeTxs = append(tradeTxs, tx)
			case metadataCommon.Pdexv3WithdrawLPFeeRequestMeta:
				withdrawLPFeeTxs = append(withdrawLPFeeTxs, tx)
			case metadataCommon.Pdexv3WithdrawProtocolFeeRequestMeta:
				withdrawlProtocolFeeTxs = append(withdrawlProtocolFeeTxs, tx)
			case metadataCommon.Pdexv3ModifyParamsMeta:
				modifyParamsTxs = append(modifyParamsTxs, tx)
			}
		}
	}

	addLiquidityInstructions, s.poolPairs, s.waitingContributions, s.nftIDs, err = s.producer.addLiquidity(
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

	pdexBlockRewards := uint64(0)
	// mint PDEX token at the pDex v3 checkpoint block
	if env.BeaconHeight() == config.Param().PDexParams.Pdexv3BreakPointHeight {
		mintPDEXGenesis, err := s.producer.mintPDEXGenesis()
		if err != nil {
			return instructions, err
		}
		instructions = append(instructions, mintPDEXGenesis...)
	} else if env.BeaconHeight() > config.Param().PDexParams.Pdexv3BreakPointHeight {
		intervalLength := uint64(MintingBlocks / DecayIntervals)
		decayIntevalIdx := (env.BeaconHeight() - config.Param().PDexParams.Pdexv3BreakPointHeight) / intervalLength
		if decayIntevalIdx < DecayIntervals {
			curIntervalReward := PDEXRewardFirstInterval
			for i := uint64(0); i < decayIntevalIdx; i++ {
				curIntervalReward -= curIntervalReward * DecayRateBPS / BPS
			}
			pdexBlockRewards = curIntervalReward / intervalLength
		}
	}

	if pdexBlockRewards > 0 {
		var mintInstructions [][]string
		mintInstructions, s.poolPairs, err = s.producer.mintPDEX(
			pdexBlockRewards,
			s.params,
			s.poolPairs,
		)
		if err != nil {
			return instructions, err
		}
		instructions = append(instructions, mintInstructions...)
	}

	var tradeInstructions [][]string
	tradeInstructions, s.poolPairs, err = s.producer.trade(
		tradeTxs,
		s.poolPairs,
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, tradeInstructions...)

	var withdrawLPFeeInstructions [][]string
	withdrawLPFeeInstructions, s.poolPairs, err = s.producer.withdrawLPFee(
		withdrawLPFeeTxs,
		env.StateDB(),
		env.BeaconHeight(),
		s.poolPairs,
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, withdrawLPFeeInstructions...)

	var withdrawProtocolFeeInstructions [][]string
	withdrawProtocolFeeInstructions, s.poolPairs, err = s.producer.withdrawProtocolFee(
		withdrawlProtocolFeeTxs,
		s.poolPairs,
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, withdrawProtocolFeeInstructions...)

	// handle modify params: at the end of beacon block
	var modifyParamsInstructions [][]string
	modifyParamsInstructions, s.params, err = s.producer.modifyParams(
		modifyParamsTxs,
		env.BeaconHeight(),
		s.params,
		s.poolPairs,
		s.stakingPoolsState,
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, modifyParamsInstructions...)

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
		s.params.PDEXRewardPoolPairsShare,
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
		for nftID, share := range poolPairState.shares {
			for height, v := range share {
				if stateChange.shares[nftID] == nil || stateChange.shares[nftID][height] == nil {
					continue
				}
				if stateChange.shares[nftID][height].isChanged {
					nftID, err := common.Hash{}.NewHashFromStr(nftID)
					err = statedb.StorePdexv3Share(
						env.StateDB(), poolPairID,
						*nftID, env.BeaconHeight(),
						v.amount, v.lastUpdatedBeaconHeight,
					)
					if err != nil {
						return err
					}
				}
				for tokenID, tradingFee := range v.tradingFees {
					if stateChange.shares[nftID][height].tokenIDs == nil {
						continue
					}
					if stateChange.shares[nftID][height].tokenIDs[tokenID] {
						err := statedb.StorePdexv3TradingFee(
							env.StateDB(), poolPairID, nftID, tokenID, env.BeaconHeight(), tradingFee,
						)
						if err != nil {
							return err
						}
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
			newStateChange = v.getDiff(k, m, newStateChange)
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

func (s *stateV2) Params() *Params {
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
		metaData.PoolPairID(), metaData.ReceiveAddress(), metaData.RefundAddress(),
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
