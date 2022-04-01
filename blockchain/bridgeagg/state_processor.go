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

func (sp *stateProcessor) modifyListTokens(
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
		acceptedInst := metadataBridge.AcceptedModifyListToken{}
		err = json.Unmarshal(contentBytes, &acceptedInst)
		if err != nil {
			return unifiedTokenInfos, err
		}
		for unifiedTokenID, vaults := range acceptedInst.NewListTokens {
			_, found := unifiedTokenInfos[unifiedTokenID]
			if !found {
				unifiedTokenInfos[unifiedTokenID] = make(map[uint]*Vault)
			}
			for _, vault := range vaults {
				if _, found := unifiedTokenInfos[unifiedTokenID][vault.NetworkID()]; !found {
					return unifiedTokenInfos, errors.New("Cannot find vault")
				} else {
					v := unifiedTokenInfos[unifiedTokenID][vault.NetworkID()]
					v.SetLastUpdatedRewardReserve(vault.RewardReserve)
					v.SetCurrentRewardReserve(vault.RewardReserve)
					unifiedTokenInfos[unifiedTokenID][vault.NetworkID()] = v
				}
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
	modifyListTokenStatus := ModifyListTokenStatus{
		Status:    status,
		ErrorCode: errorCode,
	}
	contentBytes, _ := json.Marshal(modifyListTokenStatus)
	return unifiedTokenInfos, statedb.TrackBridgeAggStatus(
		sDB,
		statedb.BridgeAggListTokenModifyingStatusPrefix(),
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
		acceptedContent := metadataBridge.AcceptedConvertTokenToUnifiedToken{}
		err = json.Unmarshal(contentBytes, &acceptedContent)
		if err != nil {
			return unifiedTokenInfos, err
		}
		if vaults, found := unifiedTokenInfos[acceptedContent.UnifiedTokenID]; found {
			if vault, found := vaults[acceptedContent.NetworkID]; found {
				err := vault.convert(acceptedContent.Amount)
				if err != nil {
					return unifiedTokenInfos, err
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
	switch inst.Status {
	case common.AcceptedStatusStr:
		contentBytes, err := base64.StdEncoding.DecodeString(inst.Content)
		if err != nil {
			return unifiedTokenInfos, err
		}
		acceptedInst := metadataBridge.AcceptedShieldRequest{}
		err = json.Unmarshal(contentBytes, &acceptedInst)
		if err != nil {
			return unifiedTokenInfos, err
		}
		for _, data := range acceptedInst.Data {
			vault := unifiedTokenInfos[acceptedInst.IncTokenID][data.NetworkID] // check available before
			if acceptedInst.IsReward {
				err = vault.decreaseCurrentRewardReserve(data.IssuingAmount)
				if err != nil {
					return unifiedTokenInfos, err
				}
			} else {
				err = vault.increaseReserve(data.IssuingAmount)
				if err != nil {
					return unifiedTokenInfos, err
				}
			}
			unifiedTokenInfos[acceptedInst.IncTokenID][data.NetworkID] = vault
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
	shieldStatus := ShieldStatus{
		Status:    status,
		ErrorCode: errorCode,
	}
	contentBytes, _ := json.Marshal(shieldStatus)
	return unifiedTokenInfos, statedb.TrackBridgeAggStatus(
		sDB,
		statedb.BridgeAggListShieldStatusPrefix(),
		txReqID.Bytes(),
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
	switch inst.Status {
	case common.AcceptedStatusStr:
		contentBytes, err := base64.StdEncoding.DecodeString(inst.Content)
		if err != nil {
			return unifiedTokenInfos, err
		}
		acceptedContent := metadataBridge.AcceptedUnshieldRequest{}
		err = json.Unmarshal(contentBytes, &acceptedContent)
		if err != nil {
			return unifiedTokenInfos, err
		}
		for _, data := range acceptedContent.Data {
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
			txReqID = acceptedContent.TxReqID
			status = common.AcceptedStatusByte
		}
		sp.UnshieldTxsCache[txReqID] = acceptedContent.TxReqID
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
	shieldStatus := ShieldStatus{
		Status:    status,
		ErrorCode: errorCode,
	}
	contentBytes, _ := json.Marshal(shieldStatus)
	return unifiedTokenInfos, statedb.TrackBridgeAggStatus(
		sDB,
		statedb.BridgeAggListUnshieldStatusPrefix(),
		txReqID.Bytes(),
		contentBytes,
	)
}

func (sp *stateProcessor) clearCache() {
	sp.UnshieldTxsCache = make(map[common.Hash]common.Hash)
}
