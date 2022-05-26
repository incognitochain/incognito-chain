package bridgeagg

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataBridge "github.com/incognitochain/incognito-chain/metadata/bridge"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type Manager struct {
	state     *State
	producer  stateProducer
	processor stateProcessor
}

func NewManager() *Manager {
	return &Manager{
		state:     NewState(),
		producer:  stateProducer{},
		processor: stateProcessor{},
	}
}

func (m *Manager) Clone() *Manager {
	return &Manager{
		state: m.state.Clone(),
	}
}

func (m *Manager) State() *State {
	return m.state
}

func NewManagerWithValue(state *State) *Manager {
	return &Manager{
		state: state,
	}
}

func (m *Manager) BuildInstructions(env StateEnvironment) ([][]string, *metadata.AccumulatedValues, error) {
	res := [][]string{}
	insts := [][]string{}
	var err error
	ac := env.AccumulatedValues()

	// build instruction for convert actions
	for shardID, actions := range env.ConvertActions() {
		for _, action := range actions {
			insts, m.state, err = m.producer.convert(action, m.state, env.StateDBs(), byte(shardID))
			if err != nil {
				return [][]string{}, nil, NewBridgeAggErrorWithValue(FailToConvertTokenError, err)
			}
			res = append(res, insts...)
		}
	}

	// build instruction for shielding actions
	for shardID, actions := range env.ShieldActions() {
		for _, action := range actions {
			insts, m.state, ac, err = m.producer.shield(
				action, m.state, ac, byte(shardID), env.StateDBs(),
			)
			if err != nil {
				return [][]string{}, nil, NewBridgeAggErrorWithValue(FailToShieldError, err)
			}

			res = append(res, insts...)
		}
	}

	// build instruction for waiting unshielding reqs
	insts, m.state, err = m.producer.handleWaitingUnshieldReqs(m.state, env.BeaconHeight(), env.StateDBs()[common.BeaconChainID])
	if err != nil {
		return [][]string{}, nil, NewBridgeAggErrorWithValue(ProducerWaitingUnshieldError, err)
	}
	res = append(res, insts...)

	// build instruction for new unshielding actions
	for shardID, actions := range env.UnshieldActions() {
		for _, action := range actions {
			insts, m.state, err = m.producer.unshield(action, m.state, env.BeaconHeight(), byte(shardID), env.StateDBs()[common.BeaconChainID])
			if err != nil {
				return [][]string{}, nil, NewBridgeAggErrorWithValue(FailToUnshieldError, err)
			}
			res = append(res, insts...)
		}
	}
	Logger.log.Info("bridgeagg instructions:", res)
	return res, ac, nil
}

func (m *Manager) Process(insts [][]string, sDB *statedb.StateDB) error {
	updatingInfoByTokenID := map[common.Hash]metadata.UpdatingInfo{}
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
			m.state, err = m.processor.addToken(content, m.state, sDB)
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
			m.state, err = m.processor.convert(*inst, m.state, sDB)
		case metadataCommon.IssuingUnifiedTokenRequestMeta:
			m.state, updatingInfoByTokenID, err = m.processor.shield(*inst, m.state, sDB, updatingInfoByTokenID)
		case metadataCommon.BurningUnifiedTokenRequestMeta:
			m.state, updatingInfoByTokenID, err = m.processor.unshield(*inst, m.state, sDB, updatingInfoByTokenID)
		}
		if err != nil {
			return err
		}
	}

	for _, updatingInfo := range updatingInfoByTokenID {
		var updatingAmt uint64
		var updatingType string
		if updatingInfo.CountUpAmt > updatingInfo.DeductAmt {
			updatingAmt = updatingInfo.CountUpAmt - updatingInfo.DeductAmt
			updatingType = "+"
		}
		if updatingInfo.CountUpAmt < updatingInfo.DeductAmt {
			updatingAmt = updatingInfo.DeductAmt - updatingInfo.CountUpAmt
			updatingType = "-"
		}
		err := statedb.UpdateBridgeTokenInfo(
			sDB,
			updatingInfo.TokenID,
			updatingInfo.ExternalTokenID,
			updatingInfo.IsCentralized,
			updatingAmt,
			updatingType,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// TODO: 0xkraken
func (m *Manager) UpdateToDB(sDB *statedb.StateDB, newUnifiedTokens map[common.Hash]bool) error {
	// store new unifiedTokens
	for unifiedTokenID := range newUnifiedTokens {
		err := statedb.StoreBridgeAggUnifiedToken(
			sDB,
			unifiedTokenID,
			statedb.NewBridgeAggUnifiedTokenStateWithValue(unifiedTokenID),
		)
		if err != nil {
			return err
		}
	}

	// store updated vaults
	for unifiedTokenID, vaults := range m.state.unifiedTokenVaults {
		for tokenID, vault := range vaults {
			err := statedb.StoreBridgeAggVault(sDB, unifiedTokenID, tokenID, vault)
			if err != nil {
				return err
			}
		}
	}

	// store new waiting unshield reqs
	for unifiedTokenID, unshieldReqs := range m.state.newWaitingUnshieldReqs {
		for _, req := range unshieldReqs {
			err := statedb.StoreBridgeAggWaitingUnshieldReq(sDB, unifiedTokenID, req.GetUnshieldID(), req)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Manager) GetDiffState(state *State) (*State, map[common.Hash]bool, error) {
	return m.state.GetDiff(state)
}

func (m *Manager) BuildAddTokenInstruction(beaconHeight uint64, sDBs map[int]*statedb.StateDB, ac *metadata.AccumulatedValues, triggeredFeature map[string]uint64) ([][]string, *metadata.AccumulatedValues, error) {
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
		temp, m.state, ac, err = m.producer.addToken(m.state, beaconHeight, sDBs, ac, checkpoint)
		if err != nil {
			return res, nil, err
		}
		if len(temp) != 0 {
			res = append(res, temp)
		}
	}
	return res, ac, err
}

func (m *Manager) ClearCache() {
	m.processor.clearCache()
}

func (m *Manager) UnifiedTokenIDCached(txReqID common.Hash) (common.Hash, error) {
	if res, found := m.processor.UnshieldTxsCache[txReqID]; found {
		return res, nil
	} else {
		return common.Hash{}, fmt.Errorf("txID %s not found in cache", txReqID.String())
	}
}
