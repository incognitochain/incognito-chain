package bridgeagg

import (
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataBridge "github.com/incognitochain/incognito-chain/metadata/bridge"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type Manager struct {
	state     State
	producer  stateProducer
	processor stateProcessor
}

func NewManager() *Manager {
	return &Manager{
		state:     *NewState(),
		producer:  stateProducer{},
		processor: stateProcessor{},
	}
}

func (m *Manager) Clone() *Manager {
	return &Manager{
		state: *m.state.Clone(),
	}
}

func (m *Manager) BuildInstructions(env StateEnvironment) ([][]string, *metadata.AccumulatedValues, error) {
	res := [][]string{}
	var err error
	ac := env.AccumulatedValues()

	for shardID, actions := range env.ConvertActions() {
		for _, action := range actions {
			insts := [][]string{}
			insts, s.unifiedTokenInfos, err = convertProducer(action, s.unifiedTokenInfos, env.StateDBs(), byte(shardID))
			if err != nil {
				return [][]string{}, nil, NewBridgeAggErrorWithValue(FailToConvertTokenError, err)
			}
			res = append(res, insts...)
		}
	}

	for shardID, actions := range env.ShieldActions() {
		for _, action := range actions {
			insts := [][]string{}
			insts, s, ac, err = shieldProducer(
				action, s, ac, byte(shardID), env.StateDBs(),
			)
			if err != nil {
				return [][]string{}, nil, NewBridgeAggErrorWithValue(FailToShieldError, err)
			}

			res = append(res, insts...)
		}
	}

	for shardID, actions := range env.UnshieldActions() {
		for _, action := range actions {
			inst := [][]string{}
			inst, s, err = unshieldProducer(action, s, env.BeaconHeight(), byte(shardID), env.StateDBs()[common.BeaconChainID])
			if err != nil {
				return [][]string{}, nil, NewBridgeAggErrorWithValue(FailToUnshieldError, err)
			}
			res = append(res, inst...)
		}
	}
	Logger.log.Info("bridgeagg instructions:", res)
	return res, ac, nil
}

func (m *Manager) Process(insts [][]string, sDB *statedb.StateDB) error {
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
			s.unifiedTokenInfos, err = addTokenProcessor(content, s.unifiedTokenInfos, sDB)
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
		case metadataCommon.BridgeAggConvertTokenToUnifiedTokenRequestMeta:
			s.unifiedTokenInfos, err = convertProcessor(*inst, s.unifiedTokenInfos, sDB)
			if err != nil {
				return err
			}
		case metadataCommon.IssuingUnifiedTokenRequestMeta:
			s, err = shieldProcessor(*inst, s, sDB)
			if err != nil {
				return err
			}
		case metadataCommon.BurningUnifiedTokenRequestMeta:
			s, err = unshieldProcessor(*inst, s, sDB)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Manager) UpdateToDB(sDB *statedb.StateDB, stateChange *StateChange) error {
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

func (m *Manager) GetDiffState(state *State) (*State, *StateChange, error) {
	return nil, nil, nil
}
