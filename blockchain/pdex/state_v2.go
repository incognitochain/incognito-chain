package pdex

import (
	"errors"
	"strconv"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
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
	txReqID         string
}

type Contribution struct {
	poolPairID     string // only "" for the first contribution of pool
	otaRefund      string // refund contributed token
	txRandomRefund string
	otaReceive     string // receive nfct
	txRandom       string
	tokenID        string
	tokenAmount    uint64
	amplifier      uint // only set for the first contribution
	txReqID        string
}

type PoolPairState struct {
	token0ID              string
	token1ID              string
	token0RealAmount      uint64
	token1RealAmount      uint64
	shares                map[string]uint64
	tradingFees           map[string]map[string]uint64
	currentContributionID uint64
	token0VirtualAmount   uint64
	token1VirtualAmount   uint64
	amplifier             uint
}

type Params struct {
	DefaultFeeRateBPS        uint            // the default value if fee rate is not specific in FeeRateBPS (default 0.3% ~ 30 BPS)
	FeeRateBPS               map[string]uint // map: pool ID -> fee rate (0.1% ~ 10 BPS)
	PRVDiscountPercent       uint            // percent of fee that will be discounted if using PRV as the trading token fee (default: 25%)
	ProtocolFeePercent       uint            // percent of fees that is rewarded for the core team (default: 0%)
	StakingPoolRewardPercent uint            // percent of fees that is distributed for staking pools (PRV, PDEX, ..., default: 10%)
	DefaultStakingPoolsShare uint            // the default value of staking pool share weight (default - 0)
	StakingPoolsShare        map[string]uint // map: staking tokenID -> pool staking share weight
}

func newStateV2() *stateV2 {
	return &stateV2{
		params: Params{
			DefaultFeeRateBPS:        InitFeeRateBPS,
			FeeRateBPS:               map[string]uint{},
			PRVDiscountPercent:       InitPRVDiscountPercent,
			ProtocolFeePercent:       InitProtocolFeePercent,
			StakingPoolRewardPercent: InitStakingPoolRewardPercent,
			DefaultStakingPoolsShare: InitStakingPoolsShare,
			StakingPoolsShare:        map[string]uint{},
		},
	}
}

func newStateV2WithValue(
	params Params,
) *stateV2 {
	return &stateV2{
		params: params,
	}
}

func initStateV2(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
) (*stateV2, error) {
	stateObject, err := statedb.GetPDexV3Params(stateDB)
	params := Params{
		DefaultFeeRateBPS:        stateObject.DefaultFeeRateBPS(),
		FeeRateBPS:               stateObject.FeeRateBPS(),
		PRVDiscountPercent:       stateObject.PRVDiscountPercent(),
		ProtocolFeePercent:       stateObject.ProtocolFeePercent(),
		StakingPoolRewardPercent: stateObject.StakingPoolRewardPercent(),
		DefaultStakingPoolsShare: stateObject.DefaultStakingPoolsShare(),
		StakingPoolsShare:        stateObject.StakingPoolsShare(),
	}
	if err != nil {
		return nil, err
	}
	return newStateV2WithValue(
		params,
	), nil
}

func (s *stateV2) Version() uint {
	return RangeProvideVersion
}

func (s *stateV2) Clone() State {
	res := newStateV2()

	res.params = s.params

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
		if !metadataCommon.IspDEXv3Type(metadataType) {
			continue // Not error, just not PDE instructions
		}
		switch metadataType {
		case metadataCommon.PDexV3ModifyParamsMeta:
			s.params, err = s.processor.modifyParams(
				env.StateDB(),
				env.BeaconHeight(),
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
	var err error

	// handle modify params
	var modifyParamsInstructions [][]string
	modifyParamsInstructions, s.params, err = s.producer.modifyParams(
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
	var err error

	err = statedb.StorePDexV3Params(
		env.StateDB(),
		s.params.DefaultFeeRateBPS,
		s.params.FeeRateBPS,
		s.params.PRVDiscountPercent,
		s.params.ProtocolFeePercent,
		s.params.StakingPoolRewardPercent,
		s.params.DefaultStakingPoolsShare,
		s.params.StakingPoolsShare,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *stateV2) TransformKeyWithNewBeaconHeight(beaconHeight uint64) {

}

func (s *stateV2) ClearCache() {
	s.deletedWaitingContributions = make(map[string]Contribution)
}

func (s *stateV2) GetDiff(compareState State) (State, error) {
	if compareState == nil {
		return nil, errors.New("compareState is nil")
	}

	res := newStateV2()

	res.params = s.params

	return res, nil

}

func (s *stateV2) Params() Params {
	return s.params
}

func (s *stateV2) Reader() StateReader {
	return s
}
