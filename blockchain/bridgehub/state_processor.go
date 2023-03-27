package bridgehub

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

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
		clonedState.bridgeInfos[contentInst.BridgePoolPubKey] = &BridgeInfo{
			Info:        statedb.NewBridgeInfoStateWithValue(contentInst.ValidatorPubKeys, contentInst.BridgePoolPubKey, []string{}, ""),
			NetworkInfo: newBridgeHubNetworkInfo(contentInst.VaultAddress),
		}
	}

	// track status
	trackStatus := RegisterBridgeStatus{
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

// updatingInfoByTokenID, err = blockchain.processIssuingBridgeReq(curView, inst, updatingInfoByTokenID, statedb.InsertFTMTxHashIssued, false)
func (sp *stateProcessor) shield(
	inst metadataCommon.Instruction,
	state *BridgeHubState,
	sDB *statedb.StateDB,
	updatingInfoByTokenID map[common.Hash]metadata.UpdatingInfo,
	insertEVMTxHashIssued func(*statedb.StateDB, []byte) error,
) (*BridgeHubState, map[common.Hash]metadata.UpdatingInfo, error) {
	var txReqID common.Hash
	if inst.Status == common.RejectedStatusStr {
		rejectContent := metadataCommon.NewRejectContent()
		if err := rejectContent.FromString(inst.Content); err != nil {
			Logger.log.Errorf("Can not decode content rejected shield instruction %v", err)
			return state, updatingInfoByTokenID, NewBridgeHubErrorWithValue(OtherError, fmt.Errorf("Can not decode content rejected shield btc hub instruction - Error %v", err))
		}
		txReqID = rejectContent.TxReqID
		// track bridge tx req status
		err := statedb.TrackBridgeReqWithStatus(sDB, txReqID, common.BridgeRequestRejectedStatus)
		if err != nil {
			Logger.log.Warn("WARNING: an error occurred while tracking bridge request with rejected status to leveldb: ", err)
		}
	} else if inst.Status == common.AcceptedStatusStr {
		contentBytes, err := base64.StdEncoding.DecodeString(inst.Content)
		if err != nil {
			Logger.log.Errorf("Can not decode content shield instruction %v", err)
			return state, updatingInfoByTokenID, NewBridgeHubErrorWithValue(OtherError, fmt.Errorf("Can not decode content shield instruction - Error %v", err))
		}
		Logger.log.Info("Processing inst content:", string(contentBytes))

		acceptedInst := metadataBridgeHub.ShieldingBTCAcceptedInst{}
		err = json.Unmarshal(contentBytes, &acceptedInst)
		if err != nil {
			Logger.log.Errorf("Can not unmarshal content shield instruction %v", err)
			return state, updatingInfoByTokenID, NewBridgeHubErrorWithValue(OtherError, fmt.Errorf("Can not unmarshal content shield instruction - Error %v", err))
		}

		// update state
		clonedState := state.Clone()
		if clonedState.bridgeInfos[acceptedInst.BridgePoolPubKey] == nil || clonedState.bridgeInfos[acceptedInst.BridgePoolPubKey].NetworkInfo[acceptedInst.ExtChainID] == nil {
			Logger.log.Errorf("Can not unmarshal content shield instruction %v", err)
			return state, updatingInfoByTokenID, NewBridgeHubErrorWithValue(OtherError, fmt.Errorf("Can not unmarshal content shield instruction - Error %v", err))
		}
		clonedState.bridgeInfos[acceptedInst.BridgePoolPubKey].NetworkInfo[acceptedInst.ExtChainID].PTokens[(&common.Hash{}).NewHash2(acceptedInst.UniqTx)] += acceptedInst.IssuingAmount

		txReqID = acceptedInst.TxReqID
		// track bridge tx req status
		err = statedb.TrackBridgeReqWithStatus(sDB, txReqID, common.BridgeRequestAcceptedStatus)
		if err != nil {
			Logger.log.Warn("WARNING: an error occurred while tracking bridge request with rejected status to leveldb: ", err)
		}
		// update tx issued
		err = insertEVMTxHashIssued(sDB, acceptedInst.UniqTx)
		if err != nil {
			Logger.log.Warn("WARNING: an error occured while inserting BTC hub tx hash issued to leveldb: ", err)
			return state, updatingInfoByTokenID, NewBridgeHubErrorWithValue(StoreShieldExtTxError, err)
		}

		// track shield req status
		updatingInfo, found := updatingInfoByTokenID[acceptedInst.IncTokenID]
		if found {
			updatingInfo.CountUpAmt += acceptedInst.IssuingAmount
		} else {
			updatingInfo = metadata.UpdatingInfo{
				CountUpAmt:    acceptedInst.IssuingAmount,
				DeductAmt:     0,
				TokenID:       acceptedInst.IncTokenID,
				IsCentralized: false,
			}
		}
		updatingInfoByTokenID[acceptedInst.IncTokenID] = updatingInfo
	} else {
		return state, updatingInfoByTokenID, NewBridgeHubErrorWithValue(InvalidBTCShieldStatus, errors.New("Can not recognize status"))
	}

	return state, updatingInfoByTokenID, nil
}

func (sp *stateProcessor) bridgeHubValidatorStake(
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
	contentInst := metadataBridgeHub.StakePRVRequestContentInst{}
	err = json.Unmarshal(contentBytes, &contentInst)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction stake bridge: %v", err)
		return state, updatingInfoByTokenID, NewBridgeHubErrorWithValue(OtherError, err)
	}

	// update state
	// TODO: 0xcryptolover add more information
	// handle fail status
	clonedState := state.Clone()
	if inst.Status == common.AcceptedStatusStr {
		_, found := clonedState.stakingInfos[contentInst.BridgePubKey]
		if !found {
			clonedState.stakingInfos[contentInst.BridgePubKey] = &statedb.BridgeStakingInfoState{}
		}
		clonedState.stakingInfos[contentInst.BridgePubKey].SetStakingAmount(clonedState.stakingInfos[contentInst.BridgePubKey].StakingAmount() + contentInst.StakeAmount)
		clonedState.stakingInfos[contentInst.BridgePubKey].SetTokenID(contentInst.TokenID)
		clonedState.stakingInfos[contentInst.BridgePubKey].SetBridgePubKey(contentInst.BridgePubKey)
	}

	// track status
	trackStatus := StakeBridgeStatus{
		StakeAmount:      contentInst.StakeAmount,
		TokenID:          contentInst.TokenID,
		Status:           status,
		BridgePoolPubKey: contentInst.BridgePoolPubKey,
		BridgePubKey:     contentInst.BridgePubKey,
		ErrorCode:        errorCode,
	}
	trackStatusBytes, _ := json.Marshal(trackStatus)
	txHash := &common.Hash{}
	txHash, _ = txHash.NewHashFromStr(contentInst.TxReqID)

	return clonedState, updatingInfoByTokenID, statedb.TrackBridgeHubStatus(
		sDB,
		statedb.BridgeHubStakeBridgeStatusPrefix(),
		txHash.Bytes(),
		trackStatusBytes,
	)
}
