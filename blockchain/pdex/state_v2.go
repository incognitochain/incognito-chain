package pdex

import (
	"errors"
	"sort"
	"strconv"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataPdexV3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
)

type stateV2 struct {
	stateBase
	waitingContributions        map[string]Contribution
	deletedWaitingContributions map[string]Contribution
	poolPairs                   map[string]PoolPairState //
	params                      Params
	stakingPoolsState           map[string]StakingPoolState // tokenID -> StakingPoolState
	orders                      map[int64][]Order
	producer                    stateProducerV2
	processor                   stateProcessorV2
}

type StakingPoolState struct {
	liquidity        uint64
	stakers          map[string]uint64 // nfst -> amount staking
	currentStakingID uint64
}

type Order struct {
	tick            int64
	tokenBuyID      string
	tokenBuyAmount  uint64
	tokenSellAmount uint64
	ota             string
	txRandom        string
	fee             uint64
}

type Params struct {
	FeeRateBPS               map[string]int // map: pool ID -> fee rate (0.1% ~ 10 BPS)
	PRVDiscountPercent       int            // percent of fee that will be discounted if using PRV as the trading token fee (defaul: 25%)
	ProtocolFeePercent       int            // percent of fees that is rewarded for the core team (default: 0%)
	StakingPoolRewardPercent int            // percent of fees that is distributed for staking pools (PRV, PDEX, ..., default: 30%)
	StakingPoolsShare        map[string]int // map: staking tokenID -> pool staking share weight (default: pDEX pool - 1000)
}

func newStateV2() *stateV2 {
	return &stateV2{
		stateBase:                   *newStateBase(),
		waitingContributions:        make(map[string]Contribution),
		deletedWaitingContributions: make(map[string]Contribution),
		poolPairs:                   make(map[string]PoolPairState),
		stakingPoolsState:           make(map[string]StakingPoolState),
		orders:                      make(map[int64][]Order),
	}
}

func newStateV2WithValue(
	waitingContributions map[string]Contribution,
	deletedWaitingContributions map[string]Contribution,
	poolPairs map[string]PoolPairState,
	params Params,
	stakingPoolsState map[string]StakingPoolState,
	orders map[int64][]Order,
) *stateV2 {
	return &stateV2{
		waitingContributions:        waitingContributions,
		deletedWaitingContributions: deletedWaitingContributions,
		poolPairs:                   poolPairs,
		params:                      params,
		stakingPoolsState:           stakingPoolsState,
		orders:                      orders,
	}
}

func initStateV2(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
) (*stateV2, error) {
	return nil, nil
}

func (s *stateV2) Version() uint {
	return RangeProvideVersion
}

func (s *stateV2) Clone() State {
	return nil
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
		if !metadata.IspDEXv3Type(metadataType) {
			continue // Not error, just not PDE instructions
		}
		switch metadataType {
		case metadata.PDexV3ModifyParamsMeta:
			s.params, err = s.processor.modifyParams(
				env.StateDB(),
				env.BeaconHeight(),
				inst,
				s.params,
			)
		case metadataCommon.PDexV3AddLiquidityMeta:
			s.poolPairs,
				s.waitingContributions,
				s.deletedWaitingContributions, err = s.processor.addLiquidity(
				env.StateDB(),
				inst,
				s.poolPairs,
				s.waitingContributions,
				s.deletedWaitingContributions,
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
	var err error

	allRemainTxs := env.AllRemainTxs()
	keys := []int{}

	for k := range allRemainTxs {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, key := range keys {
		for _, tx := range allRemainTxs[byte(key)] {
			// TODO: @pdex get metadata here and build instructions from transactions here
			switch tx.GetMetadataType() {
			case metadataCommon.PDexV3AddLiquidityMeta:
				_, ok := tx.GetMetadata().(*metadataPdexV3.AddLiquidity)
				if !ok {
					return instructions, errors.New("Can not parse add liquidity metadata")
				}
				addLiquidityTxs = append(addLiquidityTxs, tx)
			}
		}
	}

	addLiquidityInstructions, s.poolPairs, s.waitingContributions, err = s.producer.addLiquidity(
		addLiquidityTxs,
		s.poolPairs,
		s.waitingContributions,
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, addLiquidityInstructions...)

	// handle modify params
	modifyParamsInstructions, err := s.producer.modifyParams(
		env.ModifyParamsActions(),
		env.BeaconHeight(),
		s.params,
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

func (s *stateV2) StoreToDB(env StateEnvironment) error {
	return nil
}

func (s *stateV2) TransformKeyWithNewBeaconHeight(beaconHeight uint64) {

}

func (s *stateV2) ClearCache() {
	s.deletedWaitingContributions = make(map[string]Contribution)
}

func (s *stateV2) GetDiff(compareState State) (State, error) {
	return nil, nil
}

func (s *stateV2) Params() Params {
	return s.params
}
