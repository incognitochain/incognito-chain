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
	unifiedTokenInfos map[common.Hash]map[uint]*Vault // unifiedTokenID -> tokenID -> vault
	producer          stateProducer
	processor         stateProcessor
}

func (s *State) UnifiedTokenInfos() map[common.Hash]map[uint]*Vault {
	return s.unifiedTokenInfos
}

func NewState() *State {
	return &State{
		unifiedTokenInfos: make(map[common.Hash]map[uint]*Vault),
	}
}

func NewStateWithValue(unifiedTokenInfos map[common.Hash]map[uint]*Vault) *State {
	return &State{
		unifiedTokenInfos: unifiedTokenInfos,
	}
}

func (s *State) Clone() *State {
	res := NewState()
	res.processor = stateProcessor{}
	res.producer = stateProducer{}
	for unifiedTokenID, vaults := range s.unifiedTokenInfos {
		res.unifiedTokenInfos[unifiedTokenID] = make(map[uint]*Vault)
		for tokenID, vault := range vaults {
			res.unifiedTokenInfos[unifiedTokenID][tokenID] = vault.Clone()
		}
	}
	return res
}

func (s *State) BuildInstructions(env StateEnvironment) ([][]string, error) {
	res := [][]string{}
	var err error

	/*for _, action := range env.UnshieldActions() {*/

	/*}*/

	for _, action := range env.ConvertActions() {
		inst := []string{}
		inst, s.unifiedTokenInfos, err = s.producer.convert(action, s.unifiedTokenInfos, env.StateDBs())
		if err != nil {
			return [][]string{}, NewBridgeAggErrorWithValue(FailToBuildModifyListToken, err)
		}
		res = append(res, inst)
	}

	/*for _, action := range env.ShieldActions() {*/

	/*}*/

	for _, action := range env.ModifyListTokensActions() {
		inst := []string{}
		inst, s.unifiedTokenInfos, err = s.producer.modifyListTokens(action, s.unifiedTokenInfos, env.StateDBs())
		if err != nil {
			return [][]string{}, NewBridgeAggErrorWithValue(FailToBuildModifyListToken, err)
		}
		res = append(res, inst)
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
		case metadataCommon.BridgeAggConvertTokenToUnifiedTokenRequestMeta:
			s.unifiedTokenInfos, err = s.processor.convert(*inst, s.unifiedTokenInfos, sDB)
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
			err := statedb.StoreBridgeAggUnifiedToken(
				sDB,
				unifiedTokenID,
				statedb.NewBridgeAggUnifiedTokenStateWithValue(unifiedTokenID),
			)
			if err != nil {
				return err
			}
		}
		for networkID, vault := range vaults {
			if stateChange.vaultChange[unifiedTokenID][networkID].IsChanged || stateChange.unifiedTokenID[unifiedTokenID] {
				err := statedb.StoreBridgeAggConvertedToken(
					sDB, unifiedTokenID, vault.tokenID,
					statedb.NewBridgeAggConvertedTokenStateWithValue(vault.tokenID, networkID),
				)
				if err != nil {
					return err
				}
			}
			if (stateChange.vaultChange[unifiedTokenID][networkID].IsReserveChanged ||
				stateChange.unifiedTokenID[unifiedTokenID]) && !vault.BridgeAggVaultState.IsEmpty() {
				err := statedb.StoreBridgeAggVault(
					sDB, unifiedTokenID, vault.tokenID,
					&vault.BridgeAggVaultState,
				)
				if err != nil {
					return err
				}

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
				if res.unifiedTokenInfos[unifiedTokenID] == nil {
					res.unifiedTokenInfos[unifiedTokenID] = make(map[uint]*Vault)
				}
				if compareVault, ok := compareVaults[tokenID]; !ok {
					res.unifiedTokenInfos[unifiedTokenID][tokenID] = vault
					if stateChange.vaultChange[unifiedTokenID] == nil {
						stateChange.vaultChange[unifiedTokenID] = make(map[uint]VaultChange)
					}
					stateChange.vaultChange[unifiedTokenID][tokenID] = VaultChange{
						IsChanged:        true,
						IsReserveChanged: true,
					}
				} else {
					temp, vaultChange, err := s.unifiedTokenInfos[unifiedTokenID][tokenID].GetDiff(compareVault)
					if err != nil {
						return nil, nil, err
					}
					if temp != nil {
						res.unifiedTokenInfos[unifiedTokenID][tokenID] = temp
						if stateChange.vaultChange[unifiedTokenID] == nil {
							stateChange.vaultChange[unifiedTokenID] = make(map[uint]VaultChange)
						}
						stateChange.vaultChange[unifiedTokenID][tokenID] = *vaultChange
					}
				}
			}
		}
	}
	return res, stateChange, nil
}
