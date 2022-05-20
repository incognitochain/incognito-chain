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

type stateProcessor struct {
	UnshieldTxsCache map[common.Hash]common.Hash
}

func (sp *stateProcessor) convert(
	inst metadataCommon.Instruction, state *State, sDB *statedb.StateDB,
) (*State, error) {
	var status byte
	var txReqID common.Hash
	var errorCode int
	switch inst.Status {
	case common.AcceptedStatusStr:
		contentBytes, err := base64.StdEncoding.DecodeString(inst.Content)
		if err != nil {
			return nil, err
		}
		Logger.log.Info("Processing inst content:", string(contentBytes))
		acceptedContent := metadataBridge.AcceptedConvertTokenToUnifiedToken{}
		err = json.Unmarshal(contentBytes, &acceptedContent)
		if err != nil {
			return nil, err
		}
		if vaults, found := state.unifiedTokenVaults[acceptedContent.UnifiedTokenID]; found {
			if vault, found := vaults[acceptedContent.TokenID]; found {
				vault, err = increaseVaultAmount(vault, acceptedContent.MintAmount)
				if err != nil {
					return nil, NewBridgeAggErrorWithValue(InvalidConvertAmountError, err)
				}
				state.unifiedTokenVaults[acceptedContent.UnifiedTokenID][acceptedContent.TokenID] = vault
			} else {
				return nil, NewBridgeAggErrorWithValue(NotFoundNetworkIDError, err)
			}
		} else {
			return nil, NewBridgeAggErrorWithValue(NotFoundTokenIDInNetworkError, err)
		}
		txReqID = acceptedContent.TxReqID
		status = common.AcceptedStatusByte
	case common.RejectedStatusStr:
		rejectContent := metadataCommon.NewRejectContent()
		if err := rejectContent.FromString(inst.Content); err != nil {
			return nil, err
		}
		errorCode = rejectContent.ErrorCode
		txReqID = rejectContent.TxReqID
		status = common.RejectedStatusByte
	default:
		return nil, errors.New("Can not recognize status")
	}
	convertStatus := ConvertStatus{
		Status:    status,
		ErrorCode: errorCode,
	}
	contentBytes, _ := json.Marshal(convertStatus)
	return state, statedb.TrackBridgeAggStatus(
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
			return state, updatingInfoByTokenID, NewBridgeAggErrorWithValue(ProcessShieldError, fmt.Errorf("Can not decode content shield instruction - Error %v", err))
		}
		Logger.log.Info("Processing inst content:", string(contentBytes))

		acceptedInst := metadataBridge.AcceptedInstShieldRequest{}
		err = json.Unmarshal(contentBytes, &acceptedInst)
		if err != nil {
			Logger.log.Errorf("Can not unmarshal content shield instruction %v", err)
			return state, updatingInfoByTokenID, NewBridgeAggErrorWithValue(ProcessShieldError, fmt.Errorf("Can not unmarshal content shield instruction - Error %v", err))
		}
		clonedVaults, err := state.CloneVaultsByUnifiedTokenID(acceptedInst.UnifiedTokenID)
		if err != nil {
			Logger.log.Errorf("Can not get vault by unifiedTokenID %v", err)
			return state, updatingInfoByTokenID, NewBridgeAggErrorWithValue(ProcessShieldError, fmt.Errorf("Can not get vault by unifiedTokenID %v", err))
		}
		totalShieldAmt := uint64(0)
		totalReward := uint64(0)
		for _, data := range acceptedInst.Data {
			vault, ok := clonedVaults[data.IncTokenID] // check available before
			if !ok {
				Logger.log.Errorf("Can not found vault with unifiedTokenID %v and incTokenID %v", acceptedInst.UnifiedTokenID, data.IncTokenID)
				return state, updatingInfoByTokenID, NewBridgeAggErrorWithValue(ProcessShieldError,
					fmt.Errorf("Can not found vault with unifiedTokenID %v and incTokenID %v", acceptedInst.UnifiedTokenID, data.IncTokenID))
			}

			// update vault state
			clonedVaults[data.IncTokenID], err = updateVaultForShielding(vault, data.ShieldAmount, data.Reward)
			if err != nil {
				Logger.log.Errorf("Can not update vault state for shield request - Error %v", err)
				return state, updatingInfoByTokenID, NewBridgeAggErrorWithValue(ProcessShieldError, fmt.Errorf("Can not update vault state for shield request - Error %v", err))
			}

			// store UniqTx in TxHashIssued
			insertEVMTxHashIssued := GetInsertTxHashIssuedFuncByNetworkID(data.NetworkID)
			if insertEVMTxHashIssued == nil {
				return state, updatingInfoByTokenID, NewBridgeAggErrorWithValue(ProcessShieldError,
					fmt.Errorf("cannot find networkID %d", data.NetworkID))
			}
			err = insertEVMTxHashIssued(sDB, data.UniqTx)
			if err != nil {
				Logger.log.Warn("WARNING: an error occured while inserting EVM tx hash issued to leveldb: ", err)
				return state, updatingInfoByTokenID, NewBridgeAggErrorWithValue(ProcessShieldError, err)
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
			return state, updatingInfoByTokenID, NewBridgeAggErrorWithValue(ProcessShieldError, fmt.Errorf("Can not decode content shield instruction - Error %v", err))
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
		return state, updatingInfoByTokenID, NewBridgeAggErrorWithValue(ProcessShieldError, errors.New("Can not recognize status"))
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
) (*State, error) {
	var status byte
	var txReqID common.Hash
	var errorCode int
	var unshieldStatusData []UnshieldStatusData
	switch inst.Status {
	case common.AcceptedStatusStr:
		contentBytes, err := base64.StdEncoding.DecodeString(inst.Content)
		if err != nil {
			return state, err
		}
		Logger.log.Info("Processing inst content:", string(contentBytes))
		acceptedContent := metadataBridge.AcceptedInstUnshieldRequest{}
		err = json.Unmarshal(contentBytes, &acceptedContent)
		if err != nil {
			return state, err
		}
		txReqID = acceptedContent.TxReqID
		for index, data := range acceptedContent.Data {
			vault := state.unifiedTokenVaults[acceptedContent.UnifiedTokenID][data.IncTokenID] // check available before
			// TODO: 0xkraken
			// err = vault.increaseCurrentRewardReserve(data.Fee)
			// if err != nil {
			// 	return unifiedTokenInfos, err
			// }
			vault, err = decreaseVaultAmount(vault, data.BurningAmount)
			if err != nil {
				return state, err
			}
			state.unifiedTokenVaults[acceptedContent.UnifiedTokenID][data.IncTokenID] = vault
			status = common.AcceptedStatusByte
			newTxReqID := common.HashH(append(txReqID.Bytes(), common.IntToBytes(index)...))
			sp.UnshieldTxsCache[newTxReqID] = acceptedContent.UnifiedTokenID
			unshieldStatusData = append(unshieldStatusData, UnshieldStatusData{
				ReceivedAmount: data.ReceivedAmount,
				Fee:            data.BurningAmount - data.ReceivedAmount,
			})
		}
	case common.RejectedStatusStr:
		rejectContent := metadataCommon.NewRejectContent()
		if err := rejectContent.FromString(inst.Content); err != nil {
			return state, err
		}
		errorCode = rejectContent.ErrorCode
		txReqID = rejectContent.TxReqID
		status = common.RejectedStatusByte
	default:
		return state, errors.New("Can not recognize status")
	}
	unshieldStatus := UnshieldStatus{
		Status:    status,
		Data:      unshieldStatusData,
		ErrorCode: errorCode,
	}
	contentBytes, _ := json.Marshal(unshieldStatus)
	return state, statedb.TrackBridgeAggStatus(
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
		return nil, err
	}
	for unifiedTokenID, vaults := range content.NewListTokens {
		if _, found := state.unifiedTokenVaults[unifiedTokenID]; !found {
			state.unifiedTokenVaults[unifiedTokenID] = make(map[common.Hash]*statedb.BridgeAggVaultState)
		}
		err = statedb.UpdateBridgeTokenInfo(sDB, unifiedTokenID, GetExternalTokenIDForUnifiedToken(), false, 0, "+")
		if err != nil {
			return nil, err
		}
		for tokenID, vault := range vaults {
			externalTokenID, err := getExternalTokenIDByNetworkID(vault.ExternalTokenID, vault.NetworkID)
			if err != nil {
				return nil, err
			}
			err = statedb.UpdateBridgeTokenInfo(sDB, tokenID, externalTokenID, false, 0, "+")
			if err != nil {
				return nil, err
			}
			v := statedb.NewBridgeAggVaultStateWithValue(0, 0, 0, 0, vault.ExternalDecimal, vault.NetworkID, tokenID)
			state.unifiedTokenVaults[unifiedTokenID][tokenID] = v
		}
	}
	return state, nil
}

func (sp *stateProcessor) clearCache() {
	sp.UnshieldTxsCache = make(map[common.Hash]common.Hash)
}
