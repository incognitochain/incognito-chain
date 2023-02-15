package bridgehub

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataBriHub "github.com/incognitochain/incognito-chain/metadata/bridgehub"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type stateProducer struct{}

func (sp *stateProducer) registerBridge(
	contentStr string, state *BridgeHubState, sDBs map[int]*statedb.StateDB, shardID byte,
) ([][]string, *BridgeHubState, error) {
	Logger.log.Infof("[BriHub] Beacon producer - Handle register bridge request")

	// decode action
	action := metadataCommon.NewAction()
	meta := &metadataBriHub.RegisterBridgeRequest{}
	action.Meta = meta
	err := action.FromString(contentStr)
	if err != nil {
		Logger.log.Errorf("[BriHub] Beacon producer - Can not decode action register bridge from shard: %v - Error: %v", contentStr, err)
		return [][]string{}, state, nil
	}

	// don't need to verify the signature because it was verified in func ValidateSanityData

	// check number of validators
	if uint(len(meta.ValidatorPubKeys)) < state.params.MinNumberValidators() {
		inst, _ := buildBriHubRegisterBridgeInst(*meta, shardID, action.TxReqID, "", common.RejectedStatusStr, InvalidNumberValidatorError)
		return [][]string{inst}, state, nil
	}

	// check all ValidatorPubKeys staked or not
	for _, validatorPubKeyStr := range meta.ValidatorPubKeys {
		if state.stakingInfos[validatorPubKeyStr] < state.params.MinStakedAmountValidator() {
			inst, _ := buildBriHubRegisterBridgeInst(*meta, shardID, action.TxReqID, "", common.RejectedStatusStr, InvalidStakedAmountValidatorError)
			return [][]string{inst}, state, nil
		}
	}

	// check bridgeID existed or not
	bridgeIDBytes := append([]byte(meta.ExtChainID), []byte(meta.BridgePoolPubKey)...)
	bridgeID := common.HashH(bridgeIDBytes).String()

	if state.bridgeInfos[bridgeID] != nil {
		inst, _ := buildBriHubRegisterBridgeInst(*meta, shardID, action.TxReqID, bridgeID, common.RejectedStatusStr, BridgeIDExistedError)
		return [][]string{inst}, state, nil

	}

	// update state
	clonedState := state.Clone()
	clonedState.bridgeInfos[bridgeID] = &BridgeInfo{
		Info:          statedb.NewBridgeInfoStateWithValue(meta.ExtChainID, meta.ValidatorPubKeys, meta.BridgePoolPubKey, []string{}, ""),
		PTokenAmounts: map[string]*statedb.BridgeHubPTokenState{},
	}

	// build accepted instruction
	inst, _ := buildBriHubRegisterBridgeInst(*meta, shardID, action.TxReqID, bridgeID, common.AcceptedStatusStr, 0)
	return [][]string{inst}, clonedState, nil
}
