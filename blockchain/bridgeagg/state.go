package bridgeagg

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataBridge "github.com/incognitochain/incognito-chain/metadata/bridge"
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
		for networkID, vault := range vaults {
			res.unifiedTokenInfos[unifiedTokenID][networkID] = vault.Clone()
		}
	}
	return res
}

func (s *State) BuildInstructions(env StateEnvironment) ([][]string, *metadata.AccumulatedValues, error) {
	res := [][]string{}
	var err error
	ac := env.AccumulatedValues()

	for shardID, actions := range env.UnshieldActions() {
		for _, action := range actions {
			inst := [][]string{}
			inst, s.unifiedTokenInfos, err = s.producer.unshield(action, s.unifiedTokenInfos, env.BeaconHeight(), byte(shardID), env.StateDBs()[common.BeaconChainID])
			if err != nil {
				return [][]string{}, nil, NewBridgeAggErrorWithValue(FailToUnshieldError, err)
			}
			res = append(res, inst...)
		}
	}

	for shardID, actions := range env.ShieldActions() {
		for _, action := range actions {
			insts := [][]string{}
			insts, s.unifiedTokenInfos, ac, err = s.producer.shield(
				action, s.unifiedTokenInfos, ac, byte(shardID), env.StateDBs(),
			)
			if err != nil {
				return [][]string{}, nil, NewBridgeAggErrorWithValue(FailToShieldError, err)
			}

			res = append(res, insts...)
		}
	}

	for shardID, actions := range env.ConvertActions() {
		for _, action := range actions {
			insts := [][]string{}
			insts, s.unifiedTokenInfos, err = s.producer.convert(action, s.unifiedTokenInfos, env.StateDBs(), byte(shardID))
			if err != nil {
				return [][]string{}, nil, NewBridgeAggErrorWithValue(FailToConvertTokenError, err)
			}
			res = append(res, insts...)
		}
	}

	for shardID, actions := range env.ModifyRewardReserveActions() {
		for _, action := range actions {
			inst := []string{}
			inst, s.unifiedTokenInfos, err = s.producer.modifyRewardReserve(action, s.unifiedTokenInfos, env.StateDBs(), byte(shardID))
			if err != nil {
				return [][]string{}, nil, NewBridgeAggErrorWithValue(FailToBuildModifyRewardReserveError, err)
			}
			res = append(res, inst)
		}
	}

	Logger.log.Info("bridgeagg instructions:", res)

	return res, ac, nil
}

func (s *State) Process(insts [][]string, sDB *statedb.StateDB) error {
	for _, content := range insts {
		if len(content) == 0 {
			continue // Empty instruction
		}
		metaType, err := strconv.Atoi(content[0])
		if err != nil {
			continue // Not error, just not bridgeagg instructions
		}
		if !metadataBridge.IsBridgeAggMetaType(metaType) {
			continue // Not error, just not bridgeagg instructions
		}
		if metaType == metadataCommon.BridgeAggAddTokenMeta {
			s.unifiedTokenInfos, err = s.processor.addToken(content, s.unifiedTokenInfos, sDB)
			if err != nil {
				return err
			}
			continue
		}
		if len(content) != 4 {
			continue // Not error, just not bridgeagg instructions
		}

		inst := metadataCommon.NewInstruction()
		if err := inst.FromStringSlice(content); err != nil {
			return err
		}

		switch inst.MetaType {
		case metadataCommon.BridgeAggModifyRewardReserveMeta:
			s.unifiedTokenInfos, err = s.processor.modifyRewardReserve(*inst, s.unifiedTokenInfos, sDB)
			if err != nil {
				return err
			}
		case metadataCommon.BridgeAggConvertTokenToUnifiedTokenRequestMeta:
			s.unifiedTokenInfos, err = s.processor.convert(*inst, s.unifiedTokenInfos, sDB)
			if err != nil {
				return err
			}
		case metadataCommon.IssuingUnifiedTokenRequestMeta:
			s.unifiedTokenInfos, err = s.processor.shield(*inst, s.unifiedTokenInfos, sDB)
			if err != nil {
				return err
			}
		case metadataCommon.BurningUnifiedTokenRequestMeta:
			s.unifiedTokenInfos, err = s.processor.unshield(*inst, s.unifiedTokenInfos, sDB)
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
		for tokenID, vault := range vaults {
			if stateChange.vaultChange[unifiedTokenID][tokenID].IsChanged || stateChange.unifiedTokenID[unifiedTokenID] {
				err := statedb.StoreBridgeAggVault(sDB, unifiedTokenID, tokenID, &vault.BridgeAggVaultState)
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
					res.unifiedTokenInfos[unifiedTokenID] = make(map[common.Hash]*Vault)
				}
				if compareVault, ok := compareVaults[tokenID]; !ok {
					res.unifiedTokenInfos[unifiedTokenID][tokenID] = vault
					if stateChange.vaultChange[unifiedTokenID] == nil {
						stateChange.vaultChange[unifiedTokenID] = make(map[common.Hash]VaultChange)
					}
					stateChange.vaultChange[unifiedTokenID][tokenID] = VaultChange{
						IsChanged: true,
					}
				} else {
					temp, vaultChange, err := s.unifiedTokenInfos[unifiedTokenID][tokenID].GetDiff(compareVault)
					if err != nil {
						return nil, nil, err
					}
					if temp != nil {
						res.unifiedTokenInfos[unifiedTokenID][tokenID] = temp
						if stateChange.vaultChange[unifiedTokenID] == nil {
							stateChange.vaultChange[unifiedTokenID] = make(map[common.Hash]VaultChange)
						}
						stateChange.vaultChange[unifiedTokenID][tokenID] = *vaultChange
					}
				}
			}
		}
	}
	return res, stateChange, nil
}

func (s *State) ClearCache() {
	s.processor.clearCache()
}

func (s *State) UnifiedTokenIDCached(txReqID common.Hash) (common.Hash, error) {
	if res, found := s.processor.UnshieldTxsCache[txReqID]; found {
		return res, nil
	} else {
		return common.Hash{}, fmt.Errorf("txID %s not found in cache", txReqID.String())
	}
}

func (s *State) BuildAddTokenInstruction(beaconHeight uint64, sDBs map[int]*statedb.StateDB, ac *metadata.AccumulatedValues, triggeredFeature map[string]uint64) ([][]string, *metadata.AccumulatedValues, error) {
	res := [][]string{}
	temp := []string{}
	var err error

	checkpoints := []uint64{}
	validCheckpoints := []uint64{}
	for k := range triggeredFeature {
		if len(k) > len(unifiedTokenTriggerPrefix) {
			if bytes.Equal([]byte(k[:len(unifiedTokenTriggerPrefix)]), []byte(unifiedTokenTriggerPrefix)) {
				checkpoint, err := strconv.ParseUint(k[len(unifiedTokenTriggerPrefix):], 10, 64)
				if err != nil {
					return [][]string{}, ac, err
				}
				checkpoints = append(checkpoints, checkpoint)
			}
		}
	}
	if len(checkpoints) == 0 {
		return [][]string{}, ac, nil
	}
	sort.Slice(checkpoints, func(i, j int) bool {
		return checkpoints[i] < checkpoints[j]
	})

	// after beacon generate autoenablefeature instruction, TriggerFeature will mark the height of the trigger time.
	// we only need to process once at the mark time (block after trigger)
	//ex: trigger at 8, block 9 will process punified config
	for _, checkpoint := range checkpoints {
		if beaconHeight == triggeredFeature[unifiedTokenTriggerPrefix+strconv.FormatUint(checkpoint, 10)]+1 {
			validCheckpoints = append(validCheckpoints, checkpoint)
		}
	}
	if len(validCheckpoints) == 0 {
		return [][]string{}, ac, nil
	}
	for _, checkpoint := range validCheckpoints {
		temp, s.unifiedTokenInfos, ac, err = s.producer.addToken(s.unifiedTokenInfos, beaconHeight, sDBs, ac, checkpoint)
		if err != nil {
			return res, nil, err
		}
		if len(temp) != 0 {
			res = append(res, temp)
		}
	}
	return res, ac, err
}
