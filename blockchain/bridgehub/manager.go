package bridgehub

import (
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataBridgeHub "github.com/incognitochain/incognito-chain/metadata/bridgehub"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type Manager struct {
	state     *BridgeHubState
	producer  stateProducer
	processor stateProcessor
}

func NewManager() *Manager {
	return &Manager{
		state:     NewBrigdeHubState(),
		producer:  stateProducer{},
		processor: stateProcessor{},
	}
}

// TODO 0xkraken: implement
func (m *Manager) Clone() *Manager {
	return &Manager{
		state: m.state.Clone(),
	}
}

func (m *Manager) State() *BridgeHubState {
	return m.state
}

func NewManagerWithValue(state *BridgeHubState) *Manager {
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
	for shardID, actions := range env.RegisterBridgeActions() {
		for _, action := range actions {
			insts, m.state, err = m.producer.registerBridge(action, m.state, env.StateDBs(), byte(shardID))
			if err != nil {
				return [][]string{}, nil, err
			}
			res = append(res, insts...)
		}
	}

	// build instruction for shield btc hub actions
	for _, actions := range env.RegisterBridgeActions() {
		for _, action := range actions {
			insts, m.state, ac, err = m.producer.shield(action, m.state, ac, env.StateDBs(), statedb.IsBTCHubTxHashIssued)
			if err != nil {
				return [][]string{}, nil, err
			}
			res = append(res, insts...)
		}
	}

	// TODO: add more funcs to handle action ...

	return res, ac, nil
}

func (m *Manager) Process(insts [][]string, sDB *statedb.StateDB) error {
	// init bridge agg param if it's nil

	// TODO: 0xkraken
	// if m.state.params == nil {
	// 	m.InitBridgeAggParamDefault()
	// }

	// process insts
	updatingInfoByTokenID := map[common.Hash]metadata.UpdatingInfo{}
	for _, content := range insts {
		if len(content) == 0 {
			continue // Empty instruction
		}
		metaType, err := strconv.Atoi(content[0])
		if err != nil {
			continue // Not error, just not bridgeagg instructions
		}
		if !metadataBridgeHub.IsBridgeHubMetaType(metaType) {
			continue // Not error, just not bridgeagg instructions
		}
		if len(content) != 4 {
			continue // Not error, just not bridgeagg instructions
		}

		inst := metadataCommon.NewInstruction()
		if err := inst.FromStringSlice(content); err != nil {
			return err
		}

		switch inst.MetaType {
		case metadataCommon.BridgeHubRegisterBridgeMeta:
			m.state, updatingInfoByTokenID, err = m.processor.registerBridge(*inst, m.state, sDB, updatingInfoByTokenID)
		case metadataCommon.ShieldingBTCRequestMeta:
			m.state, updatingInfoByTokenID, err = m.processor.shield(*inst, m.state, sDB, updatingInfoByTokenID, statedb.InsertBTCHubTxHashIssued)
			// TODO: add more ...
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

// func (m *Manager) UpdateToDB(sDB *statedb.StateDB, newUnifiedTokens map[common.Hash]bool) error {
// 	// store new unifiedTokens
// 	for unifiedTokenID := range newUnifiedTokens {
// 		err := statedb.StoreBridgeAggUnifiedToken(
// 			sDB,
// 			unifiedTokenID,
// 			statedb.NewBridgeAggUnifiedTokenStateWithValue(unifiedTokenID),
// 		)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	// store updated vaults
// 	for unifiedTokenID, vaults := range m.state.unifiedTokenVaults {
// 		for tokenID, vault := range vaults {
// 			err := statedb.StoreBridgeAggVault(sDB, unifiedTokenID, tokenID, vault)
// 			if err != nil {
// 				return err
// 			}
// 		}
// 	}

// 	// store new waiting unshield reqs
// 	for unifiedTokenID, unshieldReqs := range m.state.newWaitingUnshieldReqs {
// 		for _, req := range unshieldReqs {
// 			err := statedb.StoreBridgeAggWaitingUnshieldReq(sDB, unifiedTokenID, req.GetUnshieldID(), req)
// 			if err != nil {
// 				return err
// 			}
// 		}
// 	}

// 	// store new param
// 	if m.state.param != nil {
// 		err := statedb.StoreBridgeAggParam(sDB, m.state.param)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	// delete waiting unshield reqs
// 	err := statedb.DeleteBridgeAggWaitingUnshieldReqs(sDB, m.state.deletedWaitingUnshieldReqKeyHashes)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (m *Manager) GetDiffState(state *State) (*State, map[common.Hash]bool, error) {
// 	return m.state.GetDiff(state)
// }

// func (m *Manager) InitBridgeAggParamDefault() error {
// 	if m.state.param != nil {
// 		return errors.New("Can not set bridge agg param to valued param")
// 	}
// 	m.state.param = statedb.NewBridgeAggParamStateWithValue(config.Param().BridgeAggParam.DefaultPercentFeeWithDecimal)
// 	return nil
// }
