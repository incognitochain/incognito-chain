package pdex

type stateV1 struct {
}

func (s *stateV1) Version() uint {
	return BasicVersion
}

func (s *stateV1) Clone() State {
	var state State
	return state
}

func (s *stateV1) Update(StateEnvironment) ([][]string, error) {
	return [][]string{}, nil
}

func (s *stateV1) Upgrade(StateEnvironment) State {
	var state State
	return state
}
