package pdex

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

type stateV2 struct {
	stateBase
	waitingContributions        map[string]Contribution
	deletedWaitingContributions map[string]Contribution
	poolPairs                   map[string]PoolPair
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
	return nil
}

func (s *stateV2) BuildInstructions(env StateEnvironment) ([][]string, error) {
	return nil, nil
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
