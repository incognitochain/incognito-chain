package pdex

import (
	"errors"
	"math/big"
	"sort"
	"strconv"

	v3 "github.com/incognitochain/incognito-chain/blockchain/pdex/v3utils"
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
	producer                    stateProducerV2
	processor                   stateProcessorV2
}

type StakingPoolState struct {
	liquidity        uint64
	stakers          map[string]uint64 // nfst -> amount staking
	currentStakingID uint64
}

type Order struct {
	id                string
	orderMatchingInfo v3.OrderMatchingInfo
	fee               uint64
}

type Contribution struct {
	poolPairID      string // only "" for the first contribution of pool
	receiverAddress string // receive nfct
	refundAddress   string // refund pToken
	tokenID         string
	tokenAmount     uint64
	amplifier       uint // only set for the first contribution
}

type PoolPairState struct {
	token0ID              string
	token1ID              string
	tokenReserve          v3.PoolReserve
	shares                map[string]uint64
	tradingFees           map[string]map[string]uint64
	currentContributionID uint64
	amplifier             uint
	orders                []*Order
}

// InsertOrder() appends a new order while keeping the list sorted (ascending by Token1Init / Token0Init)
func (p *PoolPairState) InsertOrder(ord *Order) {
	insertAt := func(lst []*Order, i int, newItem *Order) []*Order {
		if i == len(lst) {
			return append(lst, newItem)
		}
		lst = append(lst[:i+1], lst[i:]...)
		lst[i] = newItem
		return lst
	}
	index := sort.Search(len(p.orders), func(i int) bool {
		ordRate := big.NewInt(0).SetUint64(p.orders[i].orderMatchingInfo.Token0Rate)
		ordRate.Mul(ordRate, big.NewInt(0).SetUint64(ord.orderMatchingInfo.Token1Rate))
		myRate := big.NewInt(0).SetUint64(p.orders[i].orderMatchingInfo.Token1Rate)
		myRate.Mul(myRate, big.NewInt(0).SetUint64(ord.orderMatchingInfo.Token0Rate))
		// compare Token1Init / Token0Init of current order in the list to ord
		if ord.orderMatchingInfo.TradeDirection == v3.TradeDirectionSell0 {
			// orders selling token0 are iterated from start of list (buy the least token1), so we resolve equality of rate by putting the new one last
			return ordRate.Cmp(myRate) < 0
		} else {
			// orders selling token1 are iterated from end of list (buy the least token0), so we resolve equality of rate by putting the new one first
			return ordRate.Cmp(myRate) <= 0
		}
	})
	p.orders = insertAt(p.orders, index, ord)
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
		case metadata.PDexV3ModifyParamsMeta:
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
	addLiquidityTxs := []metadata.Transaction{}

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

	addLiquidityInstructions, err := s.producer.addLiquidity(
		addLiquidityTxs,
		env.BeaconHeight(),
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
