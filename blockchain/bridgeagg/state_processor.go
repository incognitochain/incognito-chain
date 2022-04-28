package bridgeagg

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataBridge "github.com/incognitochain/incognito-chain/metadata/bridge"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type stateProcessor struct {
	UnshieldTxsCache map[common.Hash]common.Hash
}

func (sp *stateProcessor) modifyRewardReserve(
	inst metadataCommon.Instruction,
	unifiedTokenInfos map[common.Hash]map[uint]*Vault,
	sDB *statedb.StateDB,
) (map[common.Hash]map[uint]*Vault, error) {
	var status byte
	var txReqID common.Hash
	var errorCode uint
	switch inst.Status {
	case common.AcceptedStatusStr:
		contentBytes, err := base64.StdEncoding.DecodeString(inst.Content)
		if err != nil {
			return unifiedTokenInfos, err
		}
		Logger.log.Info("Processing inst content:", string(contentBytes))
		acceptedInst := metadataBridge.AcceptedModifyRewardReserve{}
		err = json.Unmarshal(contentBytes, &acceptedInst)
		if err != nil {
			return unifiedTokenInfos, err
		}
		for unifiedTokenID, vaults := range acceptedInst.Vaults {
			_, found := unifiedTokenInfos[unifiedTokenID]
			if !found {
				return unifiedTokenInfos, fmt.Errorf("Cannot find unifiedTokenID %s", unifiedTokenID.String())
			}
			for _, vault := range vaults {
				if _, found := unifiedTokenInfos[unifiedTokenID][vault.NetworkID()]; !found {
					return unifiedTokenInfos, fmt.Errorf("Cannot find vault tokenID %s", vault.TokenID().String())
				}
				v := unifiedTokenInfos[unifiedTokenID][vault.NetworkID()]
				err = v.updateRewardReserve(vault.RewardReserve, vault.IsPaused)
				if err != nil {
					return unifiedTokenInfos, err
				}
				unifiedTokenInfos[unifiedTokenID][vault.NetworkID()] = v
			}
		}
		txReqID = acceptedInst.TxReqID
		status = common.AcceptedStatusByte
	case common.RejectedStatusStr:
		rejectContent := metadataCommon.NewRejectContent()
		if err := rejectContent.FromString(inst.Content); err != nil {
			return unifiedTokenInfos, err
		}
		errorCode = rejectContent.ErrorCode
		txReqID = rejectContent.TxReqID
		status = common.RejectedStatusByte
	default:
		return unifiedTokenInfos, errors.New("Can not recognize status")
	}
	modifyRewardReserveStatus := ModifyRewardReserveStatus{
		Status:    status,
		ErrorCode: errorCode,
	}
	contentBytes, _ := json.Marshal(modifyRewardReserveStatus)
	return unifiedTokenInfos, statedb.TrackBridgeAggStatus(
		sDB,
		statedb.BridgeAggRewardReserveModifyingStatusPrefix(),
		txReqID.Bytes(),
		contentBytes,
	)
}

func (sp *stateProcessor) convert(
	inst metadataCommon.Instruction,
	unifiedTokenInfos map[common.Hash]map[uint]*Vault,
	sDB *statedb.StateDB,
) (map[common.Hash]map[uint]*Vault, error) {
	var status byte
	var txReqID common.Hash
	var errorCode uint
	switch inst.Status {
	case common.AcceptedStatusStr:
		contentBytes, err := base64.StdEncoding.DecodeString(inst.Content)
		if err != nil {
			return unifiedTokenInfos, err
		}
		Logger.log.Info("Processing inst content:", string(contentBytes))
		acceptedContent := metadataBridge.AcceptedConvertTokenToUnifiedToken{}
		err = json.Unmarshal(contentBytes, &acceptedContent)
		if err != nil {
			return unifiedTokenInfos, err
		}
		if vaults, found := unifiedTokenInfos[acceptedContent.UnifiedTokenID]; found {
			if vault, found := vaults[acceptedContent.NetworkID]; found {
				err := vault.increaseReserve(acceptedContent.MintAmount)
				if err != nil {
					return unifiedTokenInfos, NewBridgeAggErrorWithValue(InvalidConvertAmountError, err)
				}
				unifiedTokenInfos[acceptedContent.UnifiedTokenID][acceptedContent.NetworkID] = vault
			} else {
				return unifiedTokenInfos, NewBridgeAggErrorWithValue(NotFoundNetworkIDError, err)
			}
		} else {
			return unifiedTokenInfos, NewBridgeAggErrorWithValue(NotFoundTokenIDInNetworkError, err)
		}
		txReqID = acceptedContent.TxReqID
		status = common.AcceptedStatusByte
	case common.RejectedStatusStr:
		rejectContent := metadataCommon.NewRejectContent()
		if err := rejectContent.FromString(inst.Content); err != nil {
			return unifiedTokenInfos, err
		}
		errorCode = rejectContent.ErrorCode
		txReqID = rejectContent.TxReqID
		status = common.RejectedStatusByte
	default:
		return unifiedTokenInfos, errors.New("Can not recognize status")
	}
	convertStatus := ConvertStatus{
		Status:    status,
		ErrorCode: errorCode,
	}
	contentBytes, _ := json.Marshal(convertStatus)
	return unifiedTokenInfos, statedb.TrackBridgeAggStatus(
		sDB,
		statedb.BridgeAggConvertStatusPrefix(),
		txReqID.Bytes(),
		contentBytes,
	)
}

func (sp *stateProcessor) shield(
	inst metadataCommon.Instruction,
	unifiedTokenInfos map[common.Hash]map[uint]*Vault,
	sDB *statedb.StateDB,
) (map[common.Hash]map[uint]*Vault, error) {
	var status byte
	var txReqID common.Hash
	var errorCode uint
	var suffix []byte
	var shieldStatusData []ShieldStatusData
	switch inst.Status {
	case common.AcceptedStatusStr:
		contentBytes, err := base64.StdEncoding.DecodeString(inst.Content)
		if err != nil {
			return unifiedTokenInfos, err
		}
		Logger.log.Info("Processing inst content:", string(contentBytes))
		acceptedInst := metadataBridge.AcceptedShieldRequest{}
		err = json.Unmarshal(contentBytes, &acceptedInst)
		if err != nil {
			return unifiedTokenInfos, err
		}
		for _, data := range acceptedInst.Data {
			vault := unifiedTokenInfos[acceptedInst.TokenID][data.NetworkID] // check available before
			statusData := ShieldStatusData{}
			if acceptedInst.IsReward {
				err = vault.decreaseCurrentRewardReserve(data.IssuingAmount)
				if err != nil {
					return unifiedTokenInfos, err
				}
				statusData.Reward = data.IssuingAmount
			} else {
				err = vault.increaseReserve(data.IssuingAmount)
				if err != nil {
					return unifiedTokenInfos, err
				}
				statusData.Amount = data.IssuingAmount
			}
			shieldStatusData = append(shieldStatusData, statusData)
			unifiedTokenInfos[acceptedInst.TokenID][data.NetworkID] = vault
		}
		txReqID = acceptedInst.TxReqID
		status = common.AcceptedStatusByte
		suffix = append(txReqID.Bytes(), common.BoolToByte(acceptedInst.IsReward))
	case common.RejectedStatusStr:
		rejectContent := metadataCommon.NewRejectContent()
		if err := rejectContent.FromString(inst.Content); err != nil {
			return unifiedTokenInfos, err
		}
		errorCode = rejectContent.ErrorCode
		txReqID = rejectContent.TxReqID
		status = common.RejectedStatusByte
		suffix = txReqID.Bytes()
	default:
		return unifiedTokenInfos, errors.New("Can not recognize status")
	}
	shieldStatus := ShieldStatus{
		Status:    status,
		ErrorCode: errorCode,
		Data:      shieldStatusData,
	}
	contentBytes, _ := json.Marshal(shieldStatus)
	return unifiedTokenInfos, statedb.TrackBridgeAggStatus(
		sDB,
		statedb.BridgeAggShieldStatusPrefix(),
		suffix,
		contentBytes,
	)
}

func (sp *stateProcessor) unshield(
	inst metadataCommon.Instruction,
	unifiedTokenInfos map[common.Hash]map[uint]*Vault,
	sDB *statedb.StateDB,
) (map[common.Hash]map[uint]*Vault, error) {
	var status byte
	var txReqID common.Hash
	var errorCode uint
	var unshieldStatusData []UnshieldStatusData
	switch inst.Status {
	case common.AcceptedStatusStr:
		contentBytes, err := base64.StdEncoding.DecodeString(inst.Content)
		if err != nil {
			return unifiedTokenInfos, err
		}
		Logger.log.Info("Processing inst content:", string(contentBytes))
		acceptedContent := metadataBridge.AcceptedUnshieldRequest{}
		err = json.Unmarshal(contentBytes, &acceptedContent)
		if err != nil {
			return unifiedTokenInfos, err
		}
		txReqID = acceptedContent.TxReqID
		for index, data := range acceptedContent.Data {
			vault := unifiedTokenInfos[acceptedContent.TokenID][data.NetworkID] // check available before
			err = vault.increaseCurrentRewardReserve(data.Fee)
			if err != nil {
				return unifiedTokenInfos, err
			}
			err = vault.decreaseReserve(data.Amount)
			if err != nil {
				return unifiedTokenInfos, err
			}
			unifiedTokenInfos[acceptedContent.TokenID][data.NetworkID] = vault
			status = common.AcceptedStatusByte
			newTxReqID := common.HashH(append(txReqID.Bytes(), common.IntToBytes(index)...))
			sp.UnshieldTxsCache[newTxReqID] = acceptedContent.TokenID
			unshieldStatusData = append(unshieldStatusData, UnshieldStatusData{
				ReceivedAmount: data.Amount,
				Fee:            data.Fee,
			})
		}
	case common.RejectedStatusStr:
		rejectContent := metadataCommon.NewRejectContent()
		if err := rejectContent.FromString(inst.Content); err != nil {
			return unifiedTokenInfos, err
		}
		errorCode = rejectContent.ErrorCode
		txReqID = rejectContent.TxReqID
		status = common.RejectedStatusByte
	default:
		return unifiedTokenInfos, errors.New("Can not recognize status")
	}
	unshieldStatus := UnshieldStatus{
		Status:    status,
		Data:      unshieldStatusData,
		ErrorCode: errorCode,
	}
	contentBytes, _ := json.Marshal(unshieldStatus)
	return unifiedTokenInfos, statedb.TrackBridgeAggStatus(
		sDB,
		statedb.BridgeAggUnshieldStatusPrefix(),
		txReqID.Bytes(),
		contentBytes,
	)
}

func (sp *stateProcessor) clearCache() {
	sp.UnshieldTxsCache = make(map[common.Hash]common.Hash)
}

func (sp *stateProcessor) addToken(
	inst []string,
	unifiedTokenInfos map[common.Hash]map[uint]*Vault,
	sDB *statedb.StateDB,
) (map[common.Hash]map[uint]*Vault, error) {
	content := metadataBridge.AddToken{}
	err := content.FromStringSlice(inst)
	if err != nil {
		return unifiedTokenInfos, err
	}
	for unifiedTokenID, vaults := range content.NewListTokens {
		if _, found := unifiedTokenInfos[unifiedTokenID]; !found {
			unifiedTokenInfos[unifiedTokenID] = make(map[uint]*Vault)
		}
		err = statedb.UpdateBridgeTokenInfo(sDB, unifiedTokenID, GetExternalTokenIDForUnifiedToken(), false, 0, "+")
		if err != nil {
			return unifiedTokenInfos, err
		}
		for networkID, vault := range vaults {
			tokenID, err := common.Hash{}.NewHashFromStr(vault.IncTokenID)
			if err != nil {
				return unifiedTokenInfos, err
			}
			externalTokenID, err := getExternalTokenIDByNetworkID(vault.ExternalTokenID, networkID)
			if err != nil {
				return unifiedTokenInfos, err
			}
			err = statedb.UpdateBridgeTokenInfo(sDB, *tokenID, externalTokenID, false, 0, "+")
			if err != nil {
				return unifiedTokenInfos, err
			}
			state := statedb.NewBridgeAggVaultStateWithValue(0, 0, 0, vault.ExternalDecimal, false)
			v := NewVaultWithValue(*state, *tokenID)
			unifiedTokenInfos[unifiedTokenID][networkID] = v
		}
	}
	return unifiedTokenInfos, nil
}
