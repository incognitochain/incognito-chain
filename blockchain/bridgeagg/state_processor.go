package bridgeagg

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataBridge "github.com/incognitochain/incognito-chain/metadata/bridge"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type stateProcessor struct{}

func (sp *stateProcessor) modifyParam(
	inst metadataCommon.Instruction,
	state *State,
	sDB *statedb.StateDB,
) (*State, error) {
	var status byte
	var txReqID common.Hash
	var errorCode int

	if inst.Status != common.AcceptedStatusStr {
		Logger.log.Errorf("Invalid status of modify param instruction")
		return state, NewBridgeAggErrorWithValue(InvalidStatusError, errors.New("Invalid status of modify param instruction"))
	}

	// decode instruction
	contentBytes, err := base64.StdEncoding.DecodeString(inst.Content)
	if err != nil {
		Logger.log.Errorf("Can not decode instruction modify param: %v", err)
		return state, NewBridgeAggErrorWithValue(OtherError, err)
	}
	acceptedContent := metadataBridge.ModifyBridgeAggParamContentInst{}
	err = json.Unmarshal(contentBytes, &acceptedContent)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction modify param: %v", err)
		return state, NewBridgeAggErrorWithValue(OtherError, err)
	}

	// update vault state
	state = updateStateForModifyParam(state, acceptedContent.PercentFeeWithDec)

	txReqID = acceptedContent.TxReqID
	status = common.AcceptedStatusByte

	// store status
	modifyStatus := ModifyParamStatus{
		NewPercentFeeWithDec: acceptedContent.PercentFeeWithDec,
		Status:               status,
		ErrorCode:            errorCode,
	}
	modifyStatusBytes, _ := json.Marshal(modifyStatus)
	return state, statedb.TrackBridgeAggStatus(
		sDB,
		statedb.BridgeAggModifyParamStatusPrefix(),
		txReqID.Bytes(),
		modifyStatusBytes,
	)
}

func (sp *stateProcessor) convert(
	inst metadataCommon.Instruction,
	state *State,
	sDB *statedb.StateDB,
	updatingInfoByTokenID map[common.Hash]metadata.UpdatingInfo,
) (*State, map[common.Hash]metadata.UpdatingInfo, error) {
	var status byte
	var txReqID common.Hash
	var errorCode int
	convertPUnifiedAmount := uint64(0)
	reward := uint64(0)

	switch inst.Status {
	case common.AcceptedStatusStr:
		contentBytes, err := base64.StdEncoding.DecodeString(inst.Content)
		if err != nil {
			Logger.log.Errorf("Can not decode instruction convert: %v", err)
			return state, updatingInfoByTokenID, NewBridgeAggErrorWithValue(OtherError, err)
		}
		acceptedContent := metadataBridge.AcceptedConvertTokenToUnifiedToken{}
		err = json.Unmarshal(contentBytes, &acceptedContent)
		if err != nil {
			Logger.log.Errorf("Can not unmarshal instruction convert: %v", err)
			return state, updatingInfoByTokenID, NewBridgeAggErrorWithValue(OtherError, err)
		}

		unifiedTokenID := acceptedContent.UnifiedTokenID
		tokenID := acceptedContent.TokenID
		vaults, found := state.unifiedTokenVaults[unifiedTokenID]
		if !found {
			Logger.log.Errorf("UnifiedTokenID is not found %v", unifiedTokenID)
			return state, updatingInfoByTokenID, NewBridgeAggErrorWithValue(NotFoundUnifiedTokenIDError, errors.New("UnifiedTokenID is not found"))
		}
		vault, found := vaults[tokenID]
		if !found {
			Logger.log.Errorf("TokenID is not invalid %v", tokenID)
			return state, updatingInfoByTokenID, NewBridgeAggErrorWithValue(InvalidPTokenIDError, errors.New("TokenID is invalid"))
		}

		// update vault state
		vault, err = updateVaultForRefill(vault, acceptedContent.ConvertPUnifiedAmount, acceptedContent.Reward)
		if err != nil {
			Logger.log.Errorf("Update state for converting err %v", err)
			return state, updatingInfoByTokenID, NewBridgeAggErrorWithValue(ProcessUpdateStateError, err)
		}
		state.unifiedTokenVaults[unifiedTokenID][tokenID] = vault

		// update bridge token info
		// decrease ptoken amount
		updatingInfo, found := updatingInfoByTokenID[tokenID]
		if found {
			updatingInfo.DeductAmt += acceptedContent.ConvertPTokenAmount
		} else {
			updatingInfo = metadata.UpdatingInfo{
				CountUpAmt:    0,
				DeductAmt:     acceptedContent.ConvertPTokenAmount,
				TokenID:       tokenID,
				IsCentralized: false,
			}
		}
		updatingInfoByTokenID[tokenID] = updatingInfo

		// increase punifiedtoken amount
		updatingInfo, found = updatingInfoByTokenID[unifiedTokenID]
		if found {
			updatingInfo.CountUpAmt += acceptedContent.ConvertPUnifiedAmount
		} else {
			updatingInfo = metadata.UpdatingInfo{
				CountUpAmt:      acceptedContent.ConvertPUnifiedAmount,
				DeductAmt:       0,
				TokenID:         unifiedTokenID,
				ExternalTokenID: GetExternalTokenIDForUnifiedToken(),
				IsCentralized:   false,
			}
		}
		updatingInfoByTokenID[unifiedTokenID] = updatingInfo

		convertPUnifiedAmount = acceptedContent.ConvertPUnifiedAmount
		reward = acceptedContent.Reward
		txReqID = acceptedContent.TxReqID
		status = common.AcceptedStatusByte
	case common.RejectedStatusStr:
		rejectContent := metadataCommon.NewRejectContent()
		if err := rejectContent.FromString(inst.Content); err != nil {
			Logger.log.Errorf("Can not decode rejected instruction convert: %v", err)
			return state, updatingInfoByTokenID, NewBridgeAggErrorWithValue(OtherError, err)
		}
		errorCode = rejectContent.ErrorCode
		txReqID = rejectContent.TxReqID
		status = common.RejectedStatusByte
	default:
		return state, updatingInfoByTokenID, NewBridgeAggErrorWithValue(InvalidStatusError, errors.New("Can not recognize status"))
	}
	convertStatus := ConvertStatus{
		ConvertPUnifiedAmount: convertPUnifiedAmount,
		Reward:                reward,
		Status:                status,
		ErrorCode:             errorCode,
	}
	contentBytes, _ := json.Marshal(convertStatus)
	return state, updatingInfoByTokenID, statedb.TrackBridgeAggStatus(
		sDB,
		statedb.BridgeAggConvertStatusPrefix(),
		txReqID.Bytes(),
		contentBytes,
	)
}

func (sp *stateProcessor) shield(
	inst metadataCommon.Instruction,
	state *State,
	sDB *statedb.StateDB,
	updatingInfoByTokenID map[common.Hash]metadata.UpdatingInfo,
) (*State, map[common.Hash]metadata.UpdatingInfo, error) {
	var status byte
	var txReqID common.Hash
	var errorCode int
	var shieldStatusData []ShieldStatusData
	switch inst.Status {
	case common.AcceptedStatusStr:
		// decode instruction
		contentBytes, err := base64.StdEncoding.DecodeString(inst.Content)
		if err != nil {
			Logger.log.Errorf("Can not decode content shield instruction %v", err)
			return state, updatingInfoByTokenID, NewBridgeAggErrorWithValue(OtherError, fmt.Errorf("Can not decode content shield instruction - Error %v", err))
		}
		Logger.log.Info("Processing inst content:", string(contentBytes))

		acceptedInst := metadataBridge.AcceptedInstShieldRequest{}
		err = json.Unmarshal(contentBytes, &acceptedInst)
		if err != nil {
			Logger.log.Errorf("Can not unmarshal content shield instruction %v", err)
			return state, updatingInfoByTokenID, NewBridgeAggErrorWithValue(OtherError, fmt.Errorf("Can not unmarshal content shield instruction - Error %v", err))
		}
		clonedVaults, err := state.CloneVaultsByUnifiedTokenID(acceptedInst.UnifiedTokenID)
		if err != nil {
			Logger.log.Errorf("Can not get vault by unifiedTokenID %v", err)
			return state, updatingInfoByTokenID, NewBridgeAggErrorWithValue(NotFoundUnifiedTokenIDError, fmt.Errorf("Can not get vault by unifiedTokenID %v", err))
		}
		totalShieldAmt := uint64(0)
		totalReward := uint64(0)
		for _, data := range acceptedInst.Data {
			vault, ok := clonedVaults[data.IncTokenID] // check available before
			if !ok {
				Logger.log.Errorf("Can not found vault with unifiedTokenID %v and incTokenID %v", acceptedInst.UnifiedTokenID, data.IncTokenID)
				return state, updatingInfoByTokenID, NewBridgeAggErrorWithValue(InvalidPTokenIDError,
					fmt.Errorf("Can not found vault with unifiedTokenID %v and incTokenID %v", acceptedInst.UnifiedTokenID, data.IncTokenID))
			}

			// update vault state
			clonedVaults[data.IncTokenID], err = updateVaultForRefill(vault, data.ShieldAmount, data.Reward)
			if err != nil {
				Logger.log.Errorf("Can not update vault state for shield request - Error %v", err)
				return state, updatingInfoByTokenID, NewBridgeAggErrorWithValue(ProcessUpdateStateError, fmt.Errorf("Can not update vault state for shield request - Error %v", err))
			}

			// store UniqTx in TxHashIssued
			insertEVMTxHashIssued := GetInsertTxHashIssuedFuncByNetworkID(data.NetworkID)
			if insertEVMTxHashIssued == nil {
				return state, updatingInfoByTokenID, NewBridgeAggErrorWithValue(InvalidNetworkIDError,
					fmt.Errorf("cannot find networkID %d", data.NetworkID))
			}
			err = insertEVMTxHashIssued(sDB, data.UniqTx)
			if err != nil {
				Logger.log.Warn("WARNING: an error occured while inserting EVM tx hash issued to leveldb: ", err)
				return state, updatingInfoByTokenID, NewBridgeAggErrorWithValue(StoreShieldExtTxError, err)
			}

			// new shielding data for storing status
			statusData := ShieldStatusData{
				Amount: data.ShieldAmount,
				Reward: data.Reward,
			}
			shieldStatusData = append(shieldStatusData, statusData)
			totalShieldAmt += data.ShieldAmount
			totalReward += data.Reward
		}
		mintAmt := totalShieldAmt + totalReward
		state.unifiedTokenVaults[acceptedInst.UnifiedTokenID] = clonedVaults
		txReqID = acceptedInst.TxReqID
		status = common.AcceptedStatusByte

		// update bridge token info
		updatingInfo, found := updatingInfoByTokenID[acceptedInst.UnifiedTokenID]
		if found {
			updatingInfo.CountUpAmt += mintAmt
		} else {
			updatingInfo = metadata.UpdatingInfo{
				CountUpAmt:      mintAmt,
				DeductAmt:       0,
				TokenID:         acceptedInst.UnifiedTokenID,
				ExternalTokenID: GetExternalTokenIDForUnifiedToken(),
				IsCentralized:   false,
			}
		}
		updatingInfoByTokenID[acceptedInst.UnifiedTokenID] = updatingInfo
	case common.RejectedStatusStr:
		rejectContent := metadataCommon.NewRejectContent()
		if err := rejectContent.FromString(inst.Content); err != nil {
			Logger.log.Errorf("Can not decode content rejected shield instruction %v", err)
			return state, updatingInfoByTokenID, NewBridgeAggErrorWithValue(OtherError, fmt.Errorf("Can not decode content rejected shield instruction - Error %v", err))
		}
		errorCode = rejectContent.ErrorCode
		txReqID = rejectContent.TxReqID
		status = common.RejectedStatusByte
		// track bridge tx req status
		err := statedb.TrackBridgeReqWithStatus(sDB, txReqID, common.BridgeRequestRejectedStatus)
		if err != nil {
			Logger.log.Warn("WARNING: an error occurred while tracking bridge request with rejected status to leveldb: ", err)
		}
	default:
		return state, updatingInfoByTokenID, NewBridgeAggErrorWithValue(InvalidNetworkIDError, errors.New("Can not recognize status"))
	}

	// track shield req status
	shieldStatus := ShieldStatus{
		Status:    status,
		ErrorCode: errorCode,
		Data:      shieldStatusData,
	}
	contentBytes, _ := json.Marshal(shieldStatus)
	return state, updatingInfoByTokenID, statedb.TrackBridgeAggStatus(
		sDB,
		statedb.BridgeAggShieldStatusPrefix(),
		txReqID.Bytes(),
		contentBytes,
	)
}

func (sp *stateProcessor) unshield(
	inst metadataCommon.Instruction,
	state *State,
	sDB *statedb.StateDB,
	updatingInfoByTokenID map[common.Hash]metadata.UpdatingInfo,
	bridgeAggUnshieldTxIDs map[string]bool,
) (*State, map[common.Hash]metadata.UpdatingInfo, map[string]bool, error) {
	var txReqID common.Hash
	var errorCode int
	var unshieldStatusData []UnshieldStatusData

	statusByte, err := getStatusByteFromStatuStr(inst.Status)
	if err != nil {
		Logger.log.Errorf("Can not get status byte from status string: %v", err)
		return state, updatingInfoByTokenID, bridgeAggUnshieldTxIDs, NewBridgeAggErrorWithValue(InvalidNetworkIDError, fmt.Errorf("Can not get status byte from status string: %v", err))
	}

	if inst.Status == common.RejectedStatusStr {
		rejectContent := metadataCommon.NewRejectContent()
		if err := rejectContent.FromString(inst.Content); err != nil {
			Logger.log.Errorf("Can not decode content rejected unshield instruction %v", err)
			return state, updatingInfoByTokenID, bridgeAggUnshieldTxIDs,
				NewBridgeAggErrorWithValue(OtherError, fmt.Errorf("Can not decode content rejected unshield instruction - Error %v", err))
		}
		errorCode = rejectContent.ErrorCode
		txReqID = rejectContent.TxReqID
	} else {
		contentBytes, err := base64.StdEncoding.DecodeString(inst.Content)
		if err != nil {
			Logger.log.Errorf("Can not decode content unshield instruction %v", err)
			return state, updatingInfoByTokenID, bridgeAggUnshieldTxIDs,
				NewBridgeAggErrorWithValue(OtherError, fmt.Errorf("Can not decode content unshield instruction - Error %v", err))
		}
		Logger.log.Info("Processing inst content:", string(contentBytes))
		acceptedUnshieldReqInst := metadataBridge.AcceptedUnshieldRequestInst{}
		err = json.Unmarshal(contentBytes, &acceptedUnshieldReqInst)
		if err != nil {
			Logger.log.Errorf("Can not unmarshal unshield instruction: %v", err)
			return state, updatingInfoByTokenID, bridgeAggUnshieldTxIDs, NewBridgeAggErrorWithValue(OtherError, fmt.Errorf("Can not unmarshal content unshield instruction - Error %v", err))
		}

		txReqID = acceptedUnshieldReqInst.WaitingUnshieldReq.GetUnshieldID()
		unifiedTokenID := acceptedUnshieldReqInst.UnifiedTokenID
		waitingUnshieldReq := acceptedUnshieldReqInst.WaitingUnshieldReq
		statusStr := inst.Status

		// update state
		state, err := updateStateForUnshield(state, unifiedTokenID, waitingUnshieldReq, statusStr)
		if err != nil {
			Logger.log.Errorf("Update bridge agg state error: %v", err)
			return state, updatingInfoByTokenID, bridgeAggUnshieldTxIDs, NewBridgeAggErrorWithValue(ProcessUpdateStateError, err)
		}

		// track unshield status
		for _, data := range waitingUnshieldReq.GetData() {
			unshieldStatusData = append(unshieldStatusData, UnshieldStatusData{
				ReceivedAmount: data.BurningAmount - data.Fee,
				Fee:            data.Fee,
			})
		}

		// update bridge token info
		if statusStr == common.WaitingStatusStr || statusStr == common.AcceptedStatusStr {
			bridgeTokenExisted, err := statedb.IsBridgeTokenExistedByType(sDB, unifiedTokenID, false)
			if err != nil {
				Logger.log.Errorf("Check bridge token existed error: %v", err)
				return state, updatingInfoByTokenID, bridgeAggUnshieldTxIDs, NewBridgeAggErrorWithValue(CheckBridgeTokenExistedError, err)
			}
			if !bridgeTokenExisted {
				return state, updatingInfoByTokenID, bridgeAggUnshieldTxIDs,
					NewBridgeAggErrorWithValue(NotFoundUnifiedTokenIDError, fmt.Errorf("Not found bridge token %s", unifiedTokenID.String()))
			}
			var totalBurnAmt uint64
			for _, v := range waitingUnshieldReq.GetData() {
				totalBurnAmt += v.BurningAmount
			}

			updatingInfo, found := updatingInfoByTokenID[unifiedTokenID]
			if found {
				updatingInfo.DeductAmt += totalBurnAmt
			} else {
				updatingInfo = metadata.UpdatingInfo{
					CountUpAmt:      0,
					DeductAmt:       totalBurnAmt,
					TokenID:         unifiedTokenID,
					ExternalTokenID: GetExternalTokenIDForUnifiedToken(),
					IsCentralized:   false,
				}
			}
			updatingInfoByTokenID[unifiedTokenID] = updatingInfo
		}

		// add TxIDs into bridgeAggUnshieldTxIDs
		if statusStr == common.AcceptedStatusStr || statusStr == common.FilledStatusStr {
			for index, _ := range waitingUnshieldReq.GetData() {
				newTxReqID := common.HashH(append(txReqID.Bytes(), common.IntToBytes(index)...))
				bridgeAggUnshieldTxIDs[newTxReqID.String()] = true
			}
		}
	}

	unshieldStatus := UnshieldStatus{
		Status:    statusByte,
		Data:      unshieldStatusData,
		ErrorCode: errorCode,
	}
	contentBytes, _ := json.Marshal(unshieldStatus)
	return state, updatingInfoByTokenID, bridgeAggUnshieldTxIDs, statedb.TrackBridgeAggStatus(
		sDB,
		statedb.BridgeAggUnshieldStatusPrefix(),
		txReqID.Bytes(),
		contentBytes,
	)
}

func (sp *stateProcessor) addToken(inst []string, state *State, sDB *statedb.StateDB) (*State, error) {
	content := metadataBridge.AddToken{}
	err := content.FromStringSlice(inst)
	if err != nil {
		Logger.log.Errorf("Can not decode content add token instruction %v", err)
		return state, NewBridgeAggErrorWithValue(OtherError, fmt.Errorf("Can not decode content add token instruction - Error %v", err))
	}
	clonedState := state.Clone()
	for unifiedTokenID, vaults := range content.NewListTokens {
		if _, found := clonedState.unifiedTokenVaults[unifiedTokenID]; !found {
			clonedState.unifiedTokenVaults[unifiedTokenID] = make(map[common.Hash]*statedb.BridgeAggVaultState)
		}
		err = statedb.UpdateBridgeTokenInfo(sDB, unifiedTokenID, GetExternalTokenIDForUnifiedToken(), false, 0, "+")
		if err != nil {
			Logger.log.Errorf("Can not store bridge token %v", err)
			return state, NewBridgeAggErrorWithValue(StoreBridgeTokenError, err)
		}
		for tokenID, vault := range vaults {
			externalTokenID, err := GetExternalTokenIDByNetworkID(vault.ExternalTokenID, vault.NetworkID)
			if err != nil {
				Logger.log.Errorf("Can not get external token id %v", err)
				return state, NewBridgeAggErrorWithValue(InvalidNetworkIDError, err)
			}
			err = statedb.UpdateBridgeTokenInfo(sDB, tokenID, externalTokenID, false, 0, "+")
			if err != nil {
				Logger.log.Errorf("Can not store bridge token %v", err)
				return state, NewBridgeAggErrorWithValue(StoreBridgeTokenError, err)
			}
			v := statedb.NewBridgeAggVaultStateWithValue(0, 0, 0, 0, vault.ExternalDecimal, vault.NetworkID, tokenID)
			clonedState.unifiedTokenVaults[unifiedTokenID][tokenID] = v
		}
	}
	return clonedState, nil
}
