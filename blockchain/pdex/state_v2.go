package pdex

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"

	"github.com/incognitochain/incognito-chain/blockchain/pdex/v2utils"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
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
	poolPairs                   map[string]*PoolPairState
	params                      *Params
	stakingPoolStates           map[string]*StakingPoolState // tokenID -> StakingPoolState
	nftIDs                      map[string]uint64
	producer                    stateProducerV2
	processor                   stateProcessorV2

	// cached state
	nftAssetTags *v2utils.NFTAssetTagsCache
}

func (s *stateV2) readConfig() {
	s.params = s.params.readConfig()
	s.stakingPoolStates = make(map[string]*StakingPoolState)
	for k := range s.params.StakingPoolsShare {
		s.stakingPoolStates[k] = NewStakingPoolState()
	}
}

func NewStatev2() *stateV2 { return newStateV2() }

func newStateV2() *stateV2 {
	return &stateV2{
		params:                      NewParams(),
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
	params *Params,
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

func (s *stateV2) Version() uint {
	return AmplifierVersion
}

func (s *stateV2) Clone() State {
	res := newStateV2()
	res.params = s.params.Clone()

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
	beaconHeight := env.PrevBeaconHeight() + 1

	for _, poolPair := range s.poolPairs {
		// reset staking pool rewards
		poolPair.stakingPoolFees = map[common.Hash]uint64{}
		poolPair.stakingPoolFees[common.PRVCoinID] = 0
		poolPair.stakingPoolFees[poolPair.state.Token0ID()] = 0
		poolPair.stakingPoolFees[poolPair.state.Token1ID()] = 0

		// init order rewards
		if poolPair.orderRewards == nil {
			poolPair.orderRewards = map[string]*OrderReward{}
		}
		// init making volume
		if poolPair.makingVolume == nil {
			poolPair.makingVolume = map[common.Hash]*MakingVolume{}
		}
	}

	var err error

	s.poolPairs, err = unlockLmLockedShareAmount(s.poolPairs, s.params, beaconHeight)
	if err != nil {
		return err
	}

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
		case metadataCommon.Pdexv3MintBlockRewardMeta:
			s.poolPairs, err = s.processor.mintBlockReward(
				env.StateDB(),
				inst,
				s.poolPairs,
			)
		case metadataCommon.Pdexv3UserMintNftRequestMeta:
			s.nftIDs, _, err = s.processor.userMintNft(env.StateDB(), inst, s.nftIDs, s.nftAssetTags)
			if err != nil {
				continue
			}
		case metadataCommon.Pdexv3ModifyParamsMeta:
			s.params, s.stakingPoolStates, err = s.processor.modifyParams(
				env.StateDB(),
				inst,
				s.params,
				s.stakingPoolStates,
			)
		case metadataCommon.Pdexv3AddLiquidityRequestMeta:
			s.poolPairs,
				s.waitingContributions,
				s.deletedWaitingContributions, err = s.processor.addLiquidity(
				env.StateDB(),
				inst,
				beaconHeight,
				s.poolPairs,
				s.waitingContributions, s.deletedWaitingContributions,
				s.params,
			)
		case metadataCommon.Pdexv3WithdrawLiquidityRequestMeta:
			s.poolPairs, err = s.processor.withdrawLiquidity(
				env.StateDB(), inst, s.poolPairs, beaconHeight, s.params.MiningRewardPendingBlocks,
			)
		case metadataCommon.Pdexv3TradeRequestMeta:
			s.poolPairs, err = s.processor.trade(env.StateDB(), inst,
				s.poolPairs, s.params,
			)
		case metadataCommon.Pdexv3WithdrawLPFeeRequestMeta:
			s.poolPairs, err = s.processor.withdrawLPFee(
				env.StateDB(),
				inst,
				s.poolPairs,
			)
		case metadataCommon.Pdexv3WithdrawProtocolFeeRequestMeta:
			s.poolPairs, err = s.processor.withdrawProtocolFee(
				env.StateDB(),
				inst,
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
		case metadataCommon.Pdexv3DistributeStakingRewardMeta:
			s.stakingPoolStates, err = s.processor.distributeStakingReward(
				env.StateDB(), inst, s.stakingPoolStates,
			)
		case metadataCommon.Pdexv3StakingRequestMeta:
			s.stakingPoolStates, _, err = s.processor.staking(
				env.StateDB(), inst, s.nftIDs, s.stakingPoolStates, beaconHeight,
			)
		case metadataCommon.Pdexv3UnstakingRequestMeta:
			s.stakingPoolStates, _, err = s.processor.unstaking(
				env.StateDB(), inst, s.nftIDs, s.stakingPoolStates, beaconHeight,
			)

		case metadataCommon.Pdexv3WithdrawStakingRewardRequestMeta:
			s.stakingPoolStates, err = s.processor.withdrawStakingReward(
				env.StateDB(),
				inst,
				s.stakingPoolStates,
			)
		case metadataCommon.Pdexv3DistributeMiningOrderRewardMeta:
			s.poolPairs, err = s.processor.distributeMiningOrderReward(
				env.StateDB(),
				inst,
				s.poolPairs,
			)
		default:
			Logger.log.Debug("Can not process this metadata")
		}
		if err != nil {
			return err
		}
	}
	if s.params.IsZeroValue() {
		s.readConfig()
	}

	return nil
}

func (s *stateV2) BuildInstructions(env StateEnvironment) ([][]string, error) {
	instructions := [][]string{}
	addLiquidityTxs := []metadata.Transaction{}
	withdrawLPFeeTxs := []metadata.Transaction{}
	withdrawlProtocolFeeTxs := []metadata.Transaction{}
	withdrawLiquidityTxs := []metadata.Transaction{}
	modifyParamsTxs := []metadata.Transaction{}
	tradeTxs := []metadata.Transaction{}
	mintNftTxs := []metadata.Transaction{}
	addOrderTxs := []metadata.Transaction{}
	withdrawOrderTxs := []metadata.Transaction{}
	stakingTxs := []metadata.Transaction{}
	unstakingTxs := []metadata.Transaction{}
	withdrawStakingRewardTxs := []metadata.Transaction{}

	beaconHeight := env.PrevBeaconHeight() + 1

	var err error
	listTxs := env.ListTxs()
	keys := []int{}

	for k := range listTxs {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, key := range keys {
		for _, tx := range listTxs[byte(key)] {
			Logger.log.Infof("tx %v prepare for build instruction %v:", tx.Hash().String(), tx.GetMetadata())
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
			case metadataCommon.Pdexv3WithdrawLPFeeRequestMeta:
				withdrawLPFeeTxs = append(withdrawLPFeeTxs, tx)
			case metadataCommon.Pdexv3WithdrawProtocolFeeRequestMeta:
				withdrawlProtocolFeeTxs = append(withdrawlProtocolFeeTxs, tx)
			case metadataCommon.Pdexv3AddOrderRequestMeta:
				addOrderTxs = append(addOrderTxs, tx)
			case metadataCommon.Pdexv3WithdrawOrderRequestMeta:
				withdrawOrderTxs = append(withdrawOrderTxs, tx)
			case metadataCommon.Pdexv3StakingRequestMeta:
				stakingTxs = append(stakingTxs, tx)
			case metadataCommon.Pdexv3UnstakingRequestMeta:
				unstakingTxs = append(unstakingTxs, tx)
			case metadataCommon.Pdexv3WithdrawStakingRewardRequestMeta:
				withdrawStakingRewardTxs = append(withdrawStakingRewardTxs, tx)
			}
		}
	}

	orderCountByNftID := make(map[string]uint)
	for _, poolPair := range s.poolPairs {
		// reset staking pool rewards
		poolPair.stakingPoolFees = map[common.Hash]uint64{}
		poolPair.stakingPoolFees[common.PRVCoinID] = 0
		poolPair.stakingPoolFees[poolPair.state.Token0ID()] = 0
		poolPair.stakingPoolFees[poolPair.state.Token1ID()] = 0

		// get order count per NftID
		for _, ord := range poolPair.orderbook.orders {
			// increment counter by NftID (orderCountByNftID[ord.NftID()] is 0 if no entry)
			orderCountByNftID[ord.NftID().String()] = orderCountByNftID[ord.NftID().String()] + 1
		}

		// init order rewards
		if poolPair.orderRewards == nil {
			poolPair.orderRewards = map[string]*OrderReward{}
		}
		// init making volume
		if poolPair.makingVolume == nil {
			poolPair.makingVolume = map[common.Hash]*MakingVolume{}
		}
	}

	s.poolPairs, err = unlockLmLockedShareAmount(s.poolPairs, s.params, beaconHeight)
	if err != nil {
		return instructions, err
	}

	var withdrawLPFeeInstructions [][]string
	withdrawLPFeeInstructions, s.poolPairs, err = s.producer.withdrawLPFee(
		withdrawLPFeeTxs,
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

	withdrawLiquidityInstructions := [][]string{}
	withdrawLiquidityInstructions, s.poolPairs, err = s.producer.withdrawLiquidity(
		withdrawLiquidityTxs, s.poolPairs, s.nftIDs, beaconHeight, s.params.MiningRewardPendingBlocks,
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, withdrawLiquidityInstructions...)

	var withdrawOrderInstructions [][]string
	withdrawOrderInstructions, s.poolPairs, err = s.producer.withdrawOrder(
		withdrawOrderTxs,
		s.poolPairs,
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, withdrawOrderInstructions...)

	var unstakingInstructions [][]string
	unstakingInstructions, s.stakingPoolStates, err = s.producer.unstaking(
		unstakingTxs, s.nftIDs, s.stakingPoolStates, beaconHeight,
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, unstakingInstructions...)

	var withdrawStakingRewardInstructions [][]string
	withdrawStakingRewardInstructions, s.stakingPoolStates, err = s.producer.withdrawStakingReward(
		withdrawStakingRewardTxs,
		s.stakingPoolStates,
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, withdrawStakingRewardInstructions...)

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

	var matchedWithdrawInstructions [][]string
	sortedPairIDs := []string{}
	matchedWithdrawInstructions, s.poolPairs, sortedPairIDs, err = s.producer.withdrawAllMatchedOrders(
		s.poolPairs, s.params.AutoWithdrawOrderLimitAmount,
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, matchedWithdrawInstructions...)

	var distributingInstruction [][]string
	distributingInstruction, s.stakingPoolStates, err = s.producer.distributeStakingReward(
		s.poolPairs,
		s.params,
		s.stakingPoolStates,
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, distributingInstruction...)

	addLiquidityInstructions := [][]string{}
	addLiquidityInstructions, s.poolPairs, s.waitingContributions, err = s.producer.addLiquidity(
		addLiquidityTxs,
		beaconHeight,
		s.poolPairs,
		s.waitingContributions,
		s.nftIDs,
		s.params,
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, addLiquidityInstructions...)

	var stakingInstructions [][]string
	stakingInstructions, s.stakingPoolStates, err = s.producer.staking(
		stakingTxs, s.nftIDs, s.stakingPoolStates, beaconHeight,
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, stakingInstructions...)

	var addOrderInstructions [][]string
	addOrderInstructions, s.poolPairs, err = s.producer.addOrder(
		addOrderTxs,
		s.poolPairs,
		s.nftIDs,
		s.params,
		orderCountByNftID,
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, addOrderInstructions...)

	// mint PDEX token at the pDex v3 checkpoint block
	if beaconHeight == config.Param().PDexParams.Pdexv3BreakPointHeight {
		mintPDEXGenesisInstructions, err := s.producer.mintPDEXGenesis()
		if err != nil {
			return instructions, err
		}
		instructions = append(instructions, mintPDEXGenesisInstructions...)
	}

	if env.Reward() > 0 {
		var mintInstructions [][]string
		mintInstructions, s.poolPairs, err = s.producer.mintReward(
			common.PRVCoinID,
			env.Reward(),
			s.params,
			s.poolPairs,
			true,
		)
		if err != nil {
			return instructions, err
		}
		instructions = append(instructions, mintInstructions...)
	}

	mintNftInstructions := [][]string{}
	burningPRVAmount := uint64(0)
	mintNftInstructions, s.nftIDs, burningPRVAmount, err = s.producer.userMintNft(
		mintNftTxs, s.nftIDs, s.nftAssetTags, beaconHeight, s.params.MintNftRequireAmount)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, mintNftInstructions...)

	if burningPRVAmount > 0 {
		var mintInstructions [][]string
		mintInstructions, s.poolPairs, err = s.producer.mintReward(
			common.PRVCoinID,
			burningPRVAmount,
			s.params,
			s.poolPairs,
			false,
		)
		if err != nil {
			return instructions, err
		}
		instructions = append(instructions, mintInstructions...)
	}

	if env.Reward() > 0 {
		withdrawOrderRewardInstructions := [][]string{}
		withdrawOrderRewardInstructions, s.poolPairs, err = s.producer.withdrawPendingOrderRewards(
			s.poolPairs, s.params.AutoWithdrawOrderRewardLimitAmount, sortedPairIDs,
		)
		instructions = append(instructions, withdrawOrderRewardInstructions...)
	}

	// handle modify params
	var modifyParamsInstructions [][]string
	modifyParamsInstructions, s.params, s.stakingPoolStates, err = s.producer.modifyParams(
		modifyParamsTxs,
		beaconHeight,
		s.params,
		s.poolPairs,
		s.stakingPoolStates,
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, modifyParamsInstructions...)

	return instructions, nil
}

func (s *stateV2) StoreToDB(env StateEnvironment, stateChange *v2utils.StateChange) error {
	var err error

	if !s.params.IsZeroValue() {
		err = statedb.StorePdexv3Params(
			env.StateDB(),
			s.params.DefaultFeeRateBPS,
			s.params.FeeRateBPS,
			s.params.PRVDiscountPercent,
			s.params.TradingProtocolFeePercent,
			s.params.TradingStakingPoolRewardPercent,
			s.params.PDEXRewardPoolPairsShare,
			s.params.StakingPoolsShare,
			s.params.StakingRewardTokens,
			s.params.MintNftRequireAmount,
			s.params.MaxOrdersPerNft,
			s.params.AutoWithdrawOrderLimitAmount,
			s.params.MinPRVReserveTradingRate,
			s.params.DefaultOrderTradingRewardRatioBPS,
			s.params.OrderTradingRewardRatioBPS,
			s.params.OrderLiquidityMiningBPS,
			s.params.DAOContributingPercent,
			s.params.MiningRewardPendingBlocks,
			s.params.OrderMiningRewardRatioBPS,
			s.params.AutoWithdrawOrderRewardLimitAmount,
		)
		if err != nil {
			return err
		}
	}
	err = s.updateWaitingContributionsToDB(env)
	if err != nil {
		return err
	}
	err = s.updatePoolPairsToDB(env, stateChange)
	if err != nil {
		return err
	}
	err = statedb.StorePdexv3NftIDs(env.StateDB(), s.nftIDs)
	if err != nil {
		return err
	}
	return s.updateStakingPoolToDB(env, stateChange)
}

func (s *stateV2) ClearCache() {
	s.deletedWaitingContributions = make(map[string]rawdbv2.Pdexv3Contribution)
}

func (s *stateV2) GetDiff(
	compareState State, stateChange *v2utils.StateChange,
) (State, *v2utils.StateChange, error) {
	newStateChange := stateChange
	if compareState == nil {
		return nil, newStateChange, errors.New("compareState is nil")
	}
	res := newStateV2()
	compareStateV2, ok := compareState.(*stateV2)
	if !ok {
		return nil, newStateChange, errors.New("compareState is not stateV2")
	}

	if !reflect.DeepEqual(s.params, compareStateV2.params) {
		res.params = s.params.Clone()
	} else {
		res.params = NewParams()
	}

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
			poolPairChange := v2utils.NewPoolPairChange()
			poolPairChange, newStateChange = v.getDiff(k, m, poolPairChange, newStateChange)
			newStateChange.PoolPairs[k] = poolPairChange
			res.poolPairs[k] = v.Clone()
		}
	}
	for k, v := range s.stakingPoolStates {
		if m, ok := compareStateV2.stakingPoolStates[k]; !ok || !reflect.DeepEqual(m, v) {
			stakingPoolChange := v2utils.NewStakingChange()
			stakingPoolChange = v.getDiff(k, m, stakingPoolChange)
			newStateChange.StakingPools[k] = stakingPoolChange
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

func (s *stateV2) Params() *Params {
	return s.params
}

func (s *stateV2) Reader() StateReader {
	return s
}

func NewContributionWithMetaData(
	metaData metadataPdexv3.AddLiquidityRequest, txReqID common.Hash, shardID byte,
) (*rawdbv2.Pdexv3Contribution, error) {
	tokenHash, err := common.Hash{}.NewHashFromStr(metaData.TokenID())
	if err != nil {
		return nil, err
	}

	accessID := common.Hash{}
	var accessOTA []byte
	var otaReceivers map[common.Hash]string
	if metaData.AccessOption.UseNft() {
		accessID = *metaData.AccessOption.NftID
	} else {
		otaReceivers = make(map[common.Hash]string)
		for k, v := range metaData.OtaReceivers() {
			otaReceivers[k], _ = v.String()
		}
		if metaData.AccessOption.AccessID != nil {
			accessID = *metaData.AccessOption.AccessID
		} else {
			if otaReceiver, found := metaData.OtaReceivers()[common.PdexAccessCoinID]; found {
				accessID = metadataPdexv3.GenAccessID(otaReceiver)
				accessOTA, err = metadataPdexv3.GenAccessOTA(otaReceiver)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return rawdbv2.NewPdexv3ContributionWithValue(
		metaData.PoolPairID(), metaData.OtaReceiver(),
		*tokenHash, txReqID, accessID,
		metaData.TokenAmount(), metaData.Amplifier(),
		shardID, accessOTA, otaReceivers,
	), nil
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

func (s *stateV2) StakingPools() map[string]*StakingPoolState {
	res := make(map[string]*StakingPoolState)
	for k, v := range s.stakingPoolStates {
		res[k] = v.Clone()
	}
	return res
}

func (s *stateV2) updateWaitingContributionsToDB(env StateEnvironment) error {
	deletedWaitingContributionsKeys := []string{}
	for k := range s.deletedWaitingContributions {
		deletedWaitingContributionsKeys = append(deletedWaitingContributionsKeys, k)
	}
	err := statedb.DeletePdexv3WaitingContributions(env.StateDB(), deletedWaitingContributionsKeys)
	if err != nil {
		return err
	}
	return statedb.StorePdexv3WaitingContributions(env.StateDB(), s.waitingContributions)
}

func (s *stateV2) updatePoolPairsToDB(env StateEnvironment, stateChange *v2utils.StateChange) error {
	var err error
	for poolPairID, poolPairState := range s.poolPairs {
		poolPairChange, found := stateChange.PoolPairs[poolPairID]
		if !found || poolPairChange == nil {
			continue
		}
		err = poolPairState.updateToDB(env, poolPairID, poolPairChange)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *stateV2) updateStakingPoolToDB(env StateEnvironment, stateChange *StateChange) error {
	for stakingPoolID, stakingPoolState := range s.stakingPoolStates {
		stakingPoolChange, found := stateChange.StakingPools[stakingPoolID]
		if !found || stakingPoolChange == nil {
			continue
		}
		err := stakingPoolState.updateToDB(env, stakingPoolID, stakingPoolChange)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *stateV2) Validator() StateValidator {
	return s
}

func (s *stateV2) IsValidNftID(nftID string) (bool, error) {
	if _, found := s.nftIDs[nftID]; !found {
		return false, fmt.Errorf("%v nftID can not be found", nftID)
	}
	return true, nil
}

func (s *stateV2) IsValidPoolPairID(poolPairID string) (bool, error) {
	if poolPair, found := s.poolPairs[poolPairID]; poolPair == nil || !found {
		return false, fmt.Errorf("%v pool pair id is not valid", poolPairID)
	}
	return true, nil
}

func (s *stateV2) IsValidMintNftRequireAmount(amount uint64) (bool, error) {
	if s.params.MintNftRequireAmount != amount {
		return false, fmt.Errorf("Expect mint nft require amount by %v but got %v",
			s.params.MintNftRequireAmount, amount)
	}
	return true, nil
}

func (s *stateV2) IsValidStakingPool(stakingPoolID string) (bool, error) {
	if stakingPool, found := s.stakingPoolStates[stakingPoolID]; stakingPool == nil || !found {
		return false, fmt.Errorf("Can not find stakingPoolID %s", stakingPoolID)
	}
	return true, nil
}

func (s *stateV2) IsValidUnstakingAmount(tokenID, nftID string, unstakingAmount uint64) (bool, error) {
	stakingPoolState, found := s.stakingPoolStates[tokenID]
	if !found || stakingPoolState == nil {
		return false, fmt.Errorf("Can not find stakingPoolID %s", tokenID)
	}
	staker, found := stakingPoolState.Stakers()[nftID]
	if !found || staker == nil {
		return false, fmt.Errorf("Can not find nftID %s", nftID)
	}
	if staker.Liquidity() < unstakingAmount {
		return false, errors.New("unstakingAmount > current staker liquidity")
	}
	if staker.Liquidity() == 0 || unstakingAmount == 0 {
		return false, errors.New("unstakingAmount or staker.Liquidity is 0")
	}
	return true, nil
}

func (s *stateV2) IsValidShareAmount(poolPairID, nftID string, shareAmount uint64) (bool, error) {
	poolPair, found := s.poolPairs[poolPairID]
	if !found || poolPair == nil {
		return false, fmt.Errorf("Can't not find pool pair ID %s", poolPairID)
	}
	share, found := poolPair.Shares()[nftID]
	if !found || share == nil {
		return false, fmt.Errorf("Can't not find nftID %s", nftID)
	}
	if share.Amount() < shareAmount {
		return false, errors.New("shareAmount > current share amount")
	}
	if shareAmount == 0 || share.Amount() == 0 {
		return false, errors.New("share amount or share.Amount() is 0")
	}
	return true, nil
}

func (s *stateV2) IsValidStaker(stakingPoolID, stakerID string) (bool, error) {
	stakingPool, found := s.stakingPoolStates[stakingPoolID]
	if stakingPool == nil || !found {
		return false, fmt.Errorf("Can not find stakingPoolID %s", stakingPoolID)
	}
	staker, found := stakingPool.stakers[stakerID]
	if staker == nil || !found {
		return false, fmt.Errorf("Can not find stakerID %s", stakerID)
	}
	return true, nil
}

func (s *stateV2) IsValidLP(poolPairID, lpID string) (bool, error) {
	poolPair, found := s.poolPairs[poolPairID]
	if !found || poolPair == nil {
		return false, fmt.Errorf("Can't not find pool pair ID %s", poolPairID)
	}
	share, found := poolPair.Shares()[lpID]
	if !found || share == nil {
		return false, fmt.Errorf("Can't not find lpID %s", lpID)
	}
	return true, nil
}

func (s *stateV2) NFTAssetTags() (map[string]*common.Hash, error) {
	if s.nftAssetTags == nil {
		var err error
		s.nftAssetTags, err = s.nftAssetTags.FromIDs(s.nftIDs)
		if err != nil {
			return nil, fmt.Errorf("NFTAssetTags missing from pdex state - %v", err)
		}
	}
	return *s.nftAssetTags, nil
}

func (s *stateV2) IsValidAccessOTA(checker metadataCommon.Pdexv3ExtendedAccessID) (bool, error) {
	var accessOTA []byte
	switch checker.MetadataType {
	case metadataCommon.Pdexv3WithdrawLiquidityRequestMeta, metadataCommon.Pdexv3WithdrawLPFeeRequestMeta:
		poolPair, found := s.poolPairs[checker.PoolID]
		if !found || poolPair == nil {
			return false, fmt.Errorf("Cannot find pool pair ID %s", checker.PoolID)
		}
		share, found := poolPair.Shares()[checker.AccessID.String()]
		if !found || share == nil {
			return false, fmt.Errorf("Cannot find accessID %s", checker.AccessID.String())
		}
		accessOTA = share.accessOTA
	case metadataCommon.Pdexv3UnstakingRequestMeta, metadataCommon.Pdexv3WithdrawStakingRewardRequestMeta:
		stakingPool, found := s.stakingPoolStates[checker.PoolID]
		if stakingPool == nil || !found {
			return false, fmt.Errorf("Cannot find stakingPoolID %s", checker.PoolID)
		}
		staker, found := stakingPool.stakers[checker.AccessID.String()]
		if staker == nil || !found {
			return false, fmt.Errorf("Cannot find stakerID %s", checker.AccessID.String())
		}
		accessOTA = staker.accessOTA
	case metadataCommon.Pdexv3WithdrawOrderRequestMeta:
		poolPair, found := s.poolPairs[checker.PoolID]
		if !found || poolPair == nil {
			return false, fmt.Errorf("Cannot find pool pair ID %s", checker.PoolID)
		}
		index := -1
		for i, order := range poolPair.Orderbook().Orders() {
			if order.Id() == checker.OrderID {
				index = i
				accessOTA = order.AccessOTA()
				if order.NftID().String() != checker.AccessID.String() {
					return false, fmt.Errorf("Expect accessID %s of order not %s", order.NftID().String(), checker.AccessID.String())
				}
				break
			}
		}
		if index == -1 {
			return false, fmt.Errorf("Cannot find orderID %v", checker.OrderID)
		}
	}
	if !bytes.Equal(accessOTA, checker.AccessOTA) {
		return false, fmt.Errorf("AccessOTA Expect %v but receive %v", accessOTA, checker.AccessOTA)
	}
	return true, nil
}
