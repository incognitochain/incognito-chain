package bridgehub

import (
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataBridgeHub "github.com/incognitochain/incognito-chain/metadata/bridgehub"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type stateProcessor struct{}

func (sp *stateProcessor) registerBridge(
	inst metadataCommon.Instruction,
	state *BridgeHubState,
	sDB *statedb.StateDB,
	updatingInfoByTokenID map[common.Hash]metadata.UpdatingInfo,
) (*BridgeHubState, map[common.Hash]metadata.UpdatingInfo, error) {
	var status byte
	var errorCode int

	var err error
	contentBytes := []byte{}

	// extract inst content bytes
	switch inst.Status {
	case common.AcceptedStatusStr:
		contentBytes, err = base64.StdEncoding.DecodeString(inst.Content)
		if err != nil {
			Logger.log.Errorf("Can not decode instruction convert: %v", err)
			return state, updatingInfoByTokenID, NewBridgeHubErrorWithValue(OtherError, err)
		}
		status = common.AcceptedStatusByte
	case common.RejectedStatusStr:
		rejectContent := metadataCommon.NewRejectContent()
		if err := rejectContent.FromString(inst.Content); err != nil {
			Logger.log.Errorf("Can not decode rejected instruction convert: %v", err)
			return state, updatingInfoByTokenID, NewBridgeHubErrorWithValue(OtherError, err)
		}
		contentBytes = rejectContent.Data
		status = common.RejectedStatusByte
		errorCode = rejectContent.ErrorCode
	default:
		return state, updatingInfoByTokenID, NewBridgeHubErrorWithValue(InvalidStatusError, errors.New("Can not recognize status"))
	}

	// unmarshal inst content
	contentInst := metadataBridgeHub.RegisterBridgeContentInst{}
	err = json.Unmarshal(contentBytes, &contentInst)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction register bridge: %v", err)
		return state, updatingInfoByTokenID, NewBridgeHubErrorWithValue(OtherError, err)
	}

	// update state
	// TODO: 0xkraken: if chainID is BTC, init pToken with pBTC ID from portal v4
	clonedState := state.Clone()
	if inst.Status == common.AcceptedStatusStr {
		clonedState.bridgeInfos[contentInst.BridgeID] = &BridgeInfo{
			Info:          statedb.NewBridgeInfoStateWithValue(contentInst.BridgeID, contentInst.ExtChainID, contentInst.ValidatorPubKeys, contentInst.BridgePoolPubKey, []string{}, ""),
			PTokenAmounts: map[string]*statedb.BridgeHubPTokenState{},
		}
	}

	// track status
	trackStatus := RegisterBridgeStatus{
		BridgeID:         contentInst.BridgeID,
		ExtChainID:       contentInst.ExtChainID,
		BridgePoolPubKey: contentInst.BridgePoolPubKey,
		ValidatorPubKeys: contentInst.ValidatorPubKeys,
		VaultAddress:     contentInst.VaultAddress,
		Signature:        contentInst.Signature,
		Status:           status,
		ErrorCode:        errorCode,
	}
	trackStatusBytes, _ := json.Marshal(trackStatus)
	txHash := &common.Hash{}
	txHash, _ = txHash.NewHashFromStr(contentInst.TxReqID)

	return clonedState, updatingInfoByTokenID, statedb.TrackBridgeHubStatus(
		sDB,
		statedb.BridgeHubRegisterBridgeStatusPrefix(),
		txHash.Bytes(),
		trackStatusBytes,
	)
}
