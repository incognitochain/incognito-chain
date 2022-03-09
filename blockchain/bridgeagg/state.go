package bridgeagg

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type State struct {
	unifiedTokenInfos map[common.Hash]map[common.Hash]*Vault // unifiedTokenID -> tokenID -> vault
	producer          stateProducer
	processor         stateProcessor
}

func (s *State) UnifiedTokenInfos() map[common.Hash]map[common.Hash]*Vault {
	return s.unifiedTokenInfos
}

func NewState() *State {
	return &State{
		unifiedTokenInfos: make(map[common.Hash]map[common.Hash]*Vault),
	}
}

func NewStateWithValue(unifiedTokenInfos map[common.Hash]map[common.Hash]*Vault) *State {
	return &State{
		unifiedTokenInfos: unifiedTokenInfos,
	}
}

func (s *State) BuildInstructions(
	metaType int,
	contentStr string,
	shardID byte,
	sDBs map[int]*statedb.StateDB,
) ([][]string, error) {
	res := [][]string{}
	var err error
	switch metaType {
	case metadataCommon.BridgeAggModifyListTokenMeta:
		res, s.unifiedTokenInfos, err = s.producer.modifyListTokens(
			contentStr, shardID, s.unifiedTokenInfos, sDBs,
		)
		if err != nil {
			return [][]string{}, NewBridgeAggErrorWithValue(FailToBuildModifyListToken, err)
		}
	}
	return res, nil
}

func (s *State) Process(insts [][]string, sDB *statedb.StateDB) error {
	for _, content := range insts {
		var err error
		inst := metadataCommon.NewInstruction()
		if err := inst.FromStringSlice(content); err != nil {
			return err
		}
		switch inst.MetaType {
		case metadataCommon.BridgeAggModifyListTokenMeta:
			s.unifiedTokenInfos, err = s.processor.modifyListTokens(*inst, s.unifiedTokenInfos, sDB)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *State) UpdateToDB() error {
	return nil
}

func (s *State) GetDiff(compareState *State) (*State, error) {
	res := NewState()
	return res, nil
}
