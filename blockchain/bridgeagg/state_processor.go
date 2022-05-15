package bridgeagg

import (
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
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
	var errorCode uint
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
		if vaults, found := state.unifiedTokenInfos[acceptedContent.UnifiedTokenID]; found {
			if vault, found := vaults[acceptedContent.TokenID]; found {
				vault, err = increaseVaultAmount(vault, acceptedContent.MintAmount)
				if err != nil {
					return nil, NewBridgeAggErrorWithValue(InvalidConvertAmountError, err)
				}
				state.unifiedTokenInfos[acceptedContent.UnifiedTokenID][acceptedContent.TokenID] = vault
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
) (*State, error) {
	var status byte
	var txReqID common.Hash
	var errorCode uint
	var suffix []byte
	var shieldStatusData []ShieldStatusData
	switch inst.Status {
	case common.AcceptedStatusStr:
		contentBytes, err := base64.StdEncoding.DecodeString(inst.Content)
		if err != nil {
			return nil, err
		}
		Logger.log.Info("Processing inst content:", string(contentBytes))
		acceptedInst := metadataBridge.AcceptedShieldRequest{}
		err = json.Unmarshal(contentBytes, &acceptedInst)
		if err != nil {
			return nil, err
		}
		for _, data := range acceptedInst.Data {
			vault := state.unifiedTokenInfos[acceptedInst.UnifiedTokenID][data.IncTokenID] // check available before
			statusData := ShieldStatusData{}
			if acceptedInst.IsReward {
				// TODO: 0xkraken
				// err = vault.decreaseCurrentRewardReserve(data.IssuingAmount)
				// if err != nil {
				// 	return unifiedTokenInfos, err
				// }
				statusData.Reward = data.IssuingAmount
			} else {
				vault, err = increaseVaultAmount(vault, data.IssuingAmount)
				if err != nil {
					return nil, err
				}
				statusData.Amount = data.IssuingAmount
			}
			shieldStatusData = append(shieldStatusData, statusData)
			state.unifiedTokenInfos[acceptedInst.UnifiedTokenID][data.IncTokenID] = vault
		}
		txReqID = acceptedInst.TxReqID
		status = common.AcceptedStatusByte
		suffix = append(txReqID.Bytes(), common.BoolToByte(acceptedInst.IsReward))
	case common.RejectedStatusStr:
		rejectContent := metadataCommon.NewRejectContent()
		if err := rejectContent.FromString(inst.Content); err != nil {
			return nil, err
		}
		errorCode = rejectContent.ErrorCode
		txReqID = rejectContent.TxReqID
		status = common.RejectedStatusByte
		suffix = txReqID.Bytes()
	default:
		return nil, errors.New("Can not recognize status")
	}
	shieldStatus := ShieldStatus{
		Status:    status,
		ErrorCode: errorCode,
		Data:      shieldStatusData,
	}
	contentBytes, _ := json.Marshal(shieldStatus)
	return state, statedb.TrackBridgeAggStatus(
		sDB,
		statedb.BridgeAggShieldStatusPrefix(),
		suffix,
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
	var errorCode uint
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
			vault := state.unifiedTokenInfos[acceptedContent.UnifiedTokenID][data.IncTokenID] // check available before
			// TODO: 0xkraken
			// err = vault.increaseCurrentRewardReserve(data.Fee)
			// if err != nil {
			// 	return unifiedTokenInfos, err
			// }
			vault, err = decreaseVaultAmount(vault, data.BurningAmount)
			if err != nil {
				return state, err
			}
			state.unifiedTokenInfos[acceptedContent.UnifiedTokenID][data.IncTokenID] = vault
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
		if _, found := state.unifiedTokenInfos[unifiedTokenID]; !found {
			state.unifiedTokenInfos[unifiedTokenID] = make(map[common.Hash]*statedb.BridgeAggVaultState)
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
			v := statedb.NewBridgeAggVaultStateWithValue(0, 0, 0, vault.ExternalDecimal, false, vault.NetworkID, tokenID)
			state.unifiedTokenInfos[unifiedTokenID][tokenID] = v
		}
	}
	return state, nil
}

func (sp *stateProcessor) clearCache() {
	sp.UnshieldTxsCache = make(map[common.Hash]common.Hash)
}
