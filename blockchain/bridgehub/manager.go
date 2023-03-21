package bridgehub

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
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
		state:     NewBridgeHubState(),
		producer:  stateProducer{},
		processor: stateProcessor{},
	}
}

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

func (m *Manager) BuildInstructions(env *StateEnvironment) ([][]string, *metadata.AccumulatedValues, error) {
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

	// build instruction for shield bridge hub actions
	for _, actions := range env.ShieldActions() {
		for _, action := range actions {
			insts, m.state, ac, err = m.producer.shield(action, m.state, ac, env.StateDBs(), statedb.IsBTCHubTxHashIssued)
			if err != nil {
				return [][]string{}, nil, err
			}
			res = append(res, insts...)
		}
	}

	// build instruction for stake btc hub actions
	for shardID, actions := range env.StakeActions() {
		for _, action := range actions {
			insts, m.state, err = m.producer.stake(action, m.state, env.StateDBs(), byte(shardID))
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
	// init bridge hub param if it's nil
	// it only runs one time when starting releasing
	if m.state.params == nil {
		m.InitBridgeHubParamDefault()
	}

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
			fmt.Printf("thachtb log state info 2 %+v \n", m.state)
		case metadataCommon.ShieldingBTCRequestMeta:
			m.state, updatingInfoByTokenID, err = m.processor.shield(*inst, m.state, sDB, updatingInfoByTokenID, statedb.InsertBTCHubTxHashIssued)
			fmt.Printf("thachtb log state info 2 %+v \n", m.state)
		case metadataCommon.StakePRVRequestMeta:
			m.state, updatingInfoByTokenID, err = m.processor.bridgeHubValidatorStake(*inst, m.state, sDB, updatingInfoByTokenID)
			fmt.Printf("thachtb log stake state info 2 %+v \n", m.state)
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

func (m *Manager) UpdateToDB(sDB *statedb.StateDB) error {
	// store new/updated bridge info
	for bridgeID, bridgeInfo := range m.state.bridgeInfos {
		// TODO: 0xkraken recheck this condition
		if bridgeInfo.Info != nil && bridgeInfo.Info.BriPubKey() != "" {
			err := statedb.StoreBridgeHubBridgeInfo(sDB, bridgeID, bridgeInfo.Info)
			if err != nil {
				return err
			}
		}

		if bridgeInfo.NetworkInfo != nil {
			for networkId, networkInfo := range bridgeInfo.NetworkInfo {
				err := statedb.StoreBridgeHubNetworkInfo(sDB, bridgeID, networkId, networkInfo)
				if err != nil {
					return err
				}
			}
		}
	}

	// store new param
	if m.state.params != nil {
		err := statedb.StoreBridgeHubParam(sDB, m.state.params)
		if err != nil {
			return err
		}
	}

	for k, v := range m.state.stakingInfos {
		err := statedb.StoreBridgeHubStaking(sDB, k, v)
		if err != nil {
			return err
		}
	}

	// TODO: coding for tokenPrices

	return nil
}

func (m *Manager) GetDiffState(state *BridgeHubState) (*BridgeHubState, error) {
	return m.state.GetDiff(state)
}

func (m *Manager) InitBridgeHubParamDefault() error {
	if m.state.params != nil {
		return errors.New("Can not set bridge agg param to valued param")
	}
	m.state.params = statedb.NewBridgeHubParamStateWithValue(
		config.Param().BridgeHubParam.MinNumberValidators,
		config.Param().BridgeHubParam.MinValidatorStakingAmount)
	return nil
}

func (m *Manager) BuildNewUnshieldBridgeHubInstructions(stateDB *statedb.StateDB, beaconHeight uint64, unshieldActionForProducers [][]string) ([][]string, error) {
	res := [][]string{}
	insts := [][]string{}

	// build instruction for new unshielding actions
	for _, a := range unshieldActionForProducers {
		metaType, err := strconv.Atoi(a[0])
		if err != nil {
			continue
		}
		switch metaType {
		case metadataCommon.BridgeHubUnshieldRequest:
			contentStr := a[1]
			insts, m.state, err = m.producer.unshield(contentStr, m.state, beaconHeight, stateDB)
			if err != nil {
				continue
			}
			if len(insts) > 0 {
				res = append(res, insts...)
			}
		default:
			continue
		}
	}

	return res, nil
}
