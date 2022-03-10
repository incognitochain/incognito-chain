package bridgeagg

import (
	"errors"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataBridgeAgg "github.com/incognitochain/incognito-chain/metadata/bridgeagg"
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

func (s *State) Clone() *State {
	res := NewState()
	res.processor = stateProcessor{}
	res.producer = stateProducer{}
	for unifiedTokenID, vaults := range s.unifiedTokenInfos {
		res.unifiedTokenInfos[unifiedTokenID] = make(map[common.Hash]*Vault)
		for tokenID, vault := range vaults {
			res.unifiedTokenInfos[unifiedTokenID][tokenID] = vault.Clone()
		}
	}
	return res
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
		if len(insts) < 2 {
			continue // Not error, just not bridgeagg instructions
		}
		metaType, err := strconv.Atoi(content[0])
		if err != nil {
			continue // Not error, just not bridgeagg instructions
		}
		if !metadataBridgeAgg.IsBridgeAggMetaType(metaType) {
			continue // Not error, just not bridgeagg instructions
		}
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

func (s *State) UpdateToDB(sDB *statedb.StateDB, stateChange *StateChange) error {
	for unifiedTokenID, vaults := range s.unifiedTokenInfos {
		if stateChange.unifiedTokenID[unifiedTokenID] {
			Logger.log.Info("[bridgeagg] Store unifiedTokenID", unifiedTokenID)
			err := statedb.StoreBridgeAggUnifiedToken(
				sDB,
				unifiedTokenID,
				statedb.NewBridgeAggUnifiedTokenStateWithValue(unifiedTokenID),
			)
			Logger.log.Info("[bridgeagg] err", err)
			if err != nil {
				return err
			}
		}
		for tokenID := range vaults {
			Logger.log.Info("[bridgeagg] Store convertedTokenID", unifiedTokenID, tokenID)
			err := statedb.StoreBridgeAggConvertedToken(
				sDB, unifiedTokenID, tokenID,
				statedb.NewBridgeAggConvertedTokenStateWithValue(tokenID),
			)
			Logger.log.Info("[bridgeagg] err", err)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *State) GetDiff(compareState *State) (*State, *StateChange, error) {
	res := NewState()
	stateChange := NewStateChange()
	if compareState == nil {
		return nil, nil, errors.New("compareState is nil")
	}
	for unifiedTokenID, vaults := range s.unifiedTokenInfos {
		if compareVaults, found := compareState.unifiedTokenInfos[unifiedTokenID]; !found {
			res.unifiedTokenInfos[unifiedTokenID] = vaults
			stateChange.unifiedTokenID[unifiedTokenID] = true
		} else {
			for tokenID, vault := range vaults {
				if compareVault, ok := compareVaults[tokenID]; !ok {
					res.unifiedTokenInfos[unifiedTokenID][tokenID] = vault
				} else {
					temp, err := s.unifiedTokenInfos[unifiedTokenID][tokenID].GetDiff(compareVault)
					if err != nil {
						return nil, nil, err
					}
					if temp != nil {
						if res.unifiedTokenInfos[unifiedTokenID] == nil {
							res.unifiedTokenInfos[unifiedTokenID] = make(map[common.Hash]*Vault)
						}
						res.unifiedTokenInfos[unifiedTokenID][tokenID] = temp
					}
				}
			}
		}
	}
	return res, stateChange, nil
}
