package bridgeagg

import "github.com/incognitochain/incognito-chain/utils"

type State struct {
}

func NewState() *State {
	return &State{}
}

func (s *State) BuildInstructions(
	contentStr string,
	shardID byte,
) ([][]string, error) {
	res := utils.EmptyStringMatrix
	return res, nil
}

func (s *State) Process() error {
	return nil
}

func (s *State) UpdateToDB() error {
	return nil
}

func (s *State) GetDiff(compareState *State) (*State, error) {
	res := NewState()
	return res, nil
}
