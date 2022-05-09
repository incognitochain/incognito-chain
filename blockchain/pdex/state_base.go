package pdex

import "github.com/incognitochain/incognito-chain/blockchain/pdex/v2utils"

type stateBase struct {
}

func newStateBase() *stateBase {
	return &stateBase{}
}

func newStateBaseWithValue() *stateBase {
	return &stateBase{}
}

//Version of state
func (s *stateBase) Version() uint {
	panic("Implement this fucntion")
}

func (s *stateBase) Clone() State {
	res := newStateBase()

	return res
}

func (s *stateBase) Process(env StateEnvironment) error {
	return nil
}

func (s *stateBase) StoreToDB(env StateEnvironment, stateChagne *v2utils.StateChange) error {
	var err error
	return err
}

func (s *stateBase) BuildInstructions(env StateEnvironment) ([][]string, error) {
	panic("Implement this function")
}

func (s *stateBase) TransformKeyWithNewBeaconHeight(beaconHeight uint64) {
	panic("Implement this fucntion")
}

func (s *stateBase) ClearCache() {
	panic("Implement this fucntion")
}

func (s *stateBase) GetDiff(compareState State, stateChange *v2utils.StateChange) (State, *v2utils.StateChange, error) {
	panic("Implement this fucntion")
}

func (s *stateBase) Params() *Params {
	panic("Implement this fucntion")
}

func (s *stateBase) PoolPairs() []byte {
	panic("Implement this fucntion")
}

func (s *stateBase) WaitingContributions() []byte {
	panic("Implement this fucntion")
}

func (s *stateBase) Shares() map[string]uint64 {
	panic("Implement this fucntion")
}

func (s *stateBase) TradingFees() map[string]uint64 {
	panic("Implement this fucntion")
}

func (s *stateBase) NftIDs() map[string]uint64 {
	panic("Implement this fucntion")
}

func (s *stateBase) Reader() StateReader {
	panic("Implement this fucntion")
}

func (s *stateBase) StakingPools() map[string]*StakingPoolState {
	panic("Implement this fucntion")
}

func (s *stateBase) IsValidNftID(nftID string) error {
	panic("Implement this fucntion")
}

func (s *stateBase) IsValidPoolPairID(poolPairID string) error {
	panic("Implement this fucntion")
}

func (s *stateBase) IsValidMintNftRequireAmount(amount uint64) error {
	panic("Implement this fucntion")
}

func (s *stateBase) IsValidStakingPool(tokenID string) error {
	panic("Implement this fucntion")
}

func (s *stateBase) IsValidUnstakingAmount(tokenID, nftID string, unstakingAmount uint64) error {
	panic("Implement this fucntion")
}

func (s *stateBase) IsValidShareAmount(poolPairID, nftID string, shareAmount uint64) error {
	panic("Implement this fucntion")
}

func (s *stateBase) Validator() StateValidator {
	panic("Implement this function")
}
