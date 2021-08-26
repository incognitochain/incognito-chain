package pdex

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

func (s *stateBase) StoreToDB(env StateEnvironment, stateChagne *StateChange) error {
	var err error
	return err
}

func (s *stateBase) BuildInstructions(env StateEnvironment) ([][]string, error) {
	panic("Implement this function")
}

func (s *stateBase) Upgrade(StateEnvironment) State {
	panic("Implement this fucntion")
}

func (s *stateBase) TransformKeyWithNewBeaconHeight(beaconHeight uint64) {
	panic("Implement this fucntion")
}

func (s *stateBase) ClearCache() {
	panic("Implement this fucntion")
}

func (s *stateBase) GetDiff(compareState State, stateChange *StateChange) (State, *StateChange, error) {
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
