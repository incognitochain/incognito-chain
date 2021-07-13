package pdex

import (
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
)

type stateV2 struct {
	stateBase
	waitingContributions        map[string]Contribution
	deletedWaitingContributions map[string]Contribution
	poolPairs                   map[string]PoolPair
	params                      Params
	producer                    stateProducerV2
	processor                   stateProcessorV2
}

type Contribution struct {
	contributorAddressStr string // Can we replace this for privacy v2?
	token1ID              string
	token2ID              string
	token1Amount          uint64
	token2Amount          uint64
	minPrice              float64 // Compare price between token 1 and token 2
	maxPrice              float64
	txReqID               common.Hash
}

type PoolPair struct {
	token1ID        string
	token1Liquidity uint64
	token2ID        string
	token2Liquidity uint64
	baseAPY         float64
	fee             float64
	CurrentPrice    float64 // Compare price between token 1 and token 2
	CurrentTick     int
}

type Position struct {
	minPrice        float64
	maxPrice        float64
	token1ID        string
	token1Liquidity uint64
	token2ID        string
	token2Liquidity uint64
	apy             float64
	tradingFees     map[string]uint64
}

type Tick struct {
	Liquidity uint64
}

type Params struct {
	FeeRateBPS               map[string]int // map: pool ID -> fee rate (0.1% ~ 10 BPS)
	PRVDiscountPercent       int            // percent of fee that will be discounted if using PRV as the trading token fee (defaul: 25%)
	ProtocolFeePercent       int            // percent of fees that is rewarded for the core team (default: 0%)
	StakingPoolRewardPercent int            // percent of fees that is distributed for staking pools (PRV, PDEX, ..., default: 30%)
	StakingPoolsShare        map[string]int // map: staking tokenID -> pool staking share weight (default: pDEX pool - 1000)
}

func newStateV2() *stateV2 {
	return nil
}

func newStateV2WithValue() *stateV2 {
	return nil
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
		case metadata.PDEContributionMeta:
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
