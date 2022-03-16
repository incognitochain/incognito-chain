package bridgeagg

import (
	"encoding/base64"
	"encoding/json"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataBridgeAgg "github.com/incognitochain/incognito-chain/metadata/bridgeagg"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/utils"
)

type stateProducer struct {
}

func (sp *stateProducer) modifyListTokens(
	contentStr string,
	unifiedTokenInfos map[common.Hash]map[uint]*Vault,
	sDBs map[int]*statedb.StateDB,
) ([]string, map[common.Hash]map[uint]*Vault, error) {
	action := metadataCommon.NewAction()
	md := &metadataBridgeAgg.ModifyListToken{}
	action.Meta = md
	err := action.FromString(contentStr)
	if err != nil {
		return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(OtherError, err)
	}
	inst := metadataCommon.NewInstructionWithValue(
		metadataCommon.BridgeAggModifyListTokenMeta,
		common.AcceptedStatusStr,
		action.ShardID,
		utils.EmptyString,
	)
	rejectContent := metadataCommon.NewRejectContentWithValue(action.TxReqID, 0, nil)

	for unifiedTokenID, vaults := range md.NewListTokens {
		if err := CheckTokenIDExisted(sDBs, unifiedTokenID); err != nil {
			rejectContent.ErrorCode = ErrCodeMessage[NotFoundTokenIDInNetworkError].Code
			temp, err := inst.StringSliceWithRejectContent(rejectContent)
			if err != nil {
				return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(NotFoundTokenIDInNetworkError, err)
			}
			return temp, unifiedTokenInfos, nil
		}
		_, found := unifiedTokenInfos[unifiedTokenID]
		if !found {
			unifiedTokenInfos[unifiedTokenID] = make(map[uint]*Vault)
		}
		for _, vault := range vaults {
			if err := CheckTokenIDExisted(sDBs, vault.TokenID()); err != nil {
				rejectContent.ErrorCode = ErrCodeMessage[NotFoundTokenIDInNetworkError].Code
				temp, err := inst.StringSliceWithRejectContent(rejectContent)
				if err != nil {
					return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(NotFoundTokenIDInNetworkError, err)
				}
				return temp, unifiedTokenInfos, nil
			}
			if _, found := unifiedTokenInfos[unifiedTokenID][vault.NetworkID()]; !found {
				unifiedTokenInfos[unifiedTokenID][vault.NetworkID()] = NewVaultWithValue(
					*statedb.NewBridgeAggVaultState(), []byte{}, vault.TokenID(),
				)
			} else {
				newRewardReserve := vault.RewardReserve
				v := unifiedTokenInfos[unifiedTokenID][vault.NetworkID()]
				lastUpdatedRewardReserve := v.LastUpdatedRewardReserve()
				currentRewardReserve := v.CurrentRewardReserve()
				if newRewardReserve < lastUpdatedRewardReserve-currentRewardReserve {
					rejectContent.ErrorCode = ErrCodeMessage[InvalidRewardReserveError].Code
					temp, err := inst.StringSliceWithRejectContent(rejectContent)
					if err != nil {
						return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(InvalidRewardReserveError, err)
					}
					return temp, unifiedTokenInfos, nil
				}
				v.SetLastUpdatedRewardReserve(vault.RewardReserve)
				v.SetCurrentRewardReserve(vault.RewardReserve)
				v.tokenID = vault.TokenID()
				unifiedTokenInfos[unifiedTokenID][vault.NetworkID()] = v
			}
		}
	}
	acceptedContent := metadataBridgeAgg.AcceptedModifyListToken{
		ModifyListToken: *md,
		TxReqID:         action.TxReqID,
	}
	contentBytes, err := json.Marshal(acceptedContent)
	if err != nil {
		return []string{}, unifiedTokenInfos, err
	}
	inst.Content = base64.StdEncoding.EncodeToString(contentBytes)
	return inst.StringSlice(), unifiedTokenInfos, nil
}

func (sp *stateProducer) convert(
	contentStr string, unifiedTokenInfos map[common.Hash]map[uint]*Vault, sDBs map[int]*statedb.StateDB,
) ([]string, map[common.Hash]map[uint]*Vault, error) {
	action := metadataCommon.NewAction()
	md := &metadataBridgeAgg.ConvertTokenToUnifiedTokenRequest{}
	action.Meta = md
	err := action.FromString(contentStr)
	if err != nil {
		return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(OtherError, err)
	}
	inst := metadataCommon.NewInstructionWithValue(
		metadataCommon.BridgeAggConvertTokenToUnifiedTokenRequestMeta,
		common.AcceptedStatusStr,
		action.ShardID,
		utils.EmptyString,
	)
	rejectContent := metadataCommon.NewRejectContentWithValue(action.TxReqID, 0, md)
	if _, found := unifiedTokenInfos[md.UnifiedTokenID]; !found {
		rejectContent.ErrorCode = ErrCodeMessage[NotFoundTokenIDInNetworkError].Code
		temp, err := inst.StringSliceWithRejectContent(rejectContent)
		if err != nil {
			return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(NotFoundTokenIDInNetworkError, err)
		}
		return temp, unifiedTokenInfos, nil
	}
	if vault, found := unifiedTokenInfos[md.UnifiedTokenID][md.NetworkID]; !found {
		rejectContent.ErrorCode = ErrCodeMessage[NotFoundNetworkIDError].Code
		temp, err := inst.StringSliceWithRejectContent(rejectContent)
		if err != nil {
			return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(NotFoundTokenIDInNetworkError, err)
		}
		return temp, unifiedTokenInfos, nil
	} else {
		if vault.tokenID.String() != md.TokenID.String() {
			rejectContent.ErrorCode = ErrCodeMessage[NotFoundTokenIDInNetworkError].Code
			temp, err := inst.StringSliceWithRejectContent(rejectContent)
			if err != nil {
				return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(NotFoundTokenIDInNetworkError, err)
			}
			return temp, unifiedTokenInfos, nil

		}
		err = vault.convert(md.Amount)
		if err != nil {
			Logger.log.Warnf("Invalid shield amount error: %v tx %s", err, action.TxReqID.String())
			rejectContent.ErrorCode = ErrCodeMessage[InvalidConvertAmountError].Code
			temp, e := inst.StringSliceWithRejectContent(rejectContent)
			if e != nil {
				return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(OtherError, e)
			}
			return temp, unifiedTokenInfos, nil

		}
		unifiedTokenInfos[md.UnifiedTokenID][md.NetworkID] = vault
		acceptedContent := metadataBridgeAgg.AcceptedConvertTokenToUnifiedToken{
			ConvertTokenToUnifiedTokenRequest: *md,
			TxReqID:                           action.TxReqID,
		}
		contentBytes, err := json.Marshal(acceptedContent)
		if err != nil {
			return []string{}, unifiedTokenInfos, err
		}
		inst.Content = base64.StdEncoding.EncodeToString(contentBytes)

	}
	return inst.StringSlice(), unifiedTokenInfos, nil
}

func (sp *stateProducer) shield(
	contentStr []string, unifiedTokenInfos map[common.Hash]map[uint]*Vault,
) ([]string, map[common.Hash]map[uint]*Vault, error) {
	res := []string{}
	tempInst := metadataCommon.NewInstruction()
	if err := tempInst.FromStringSlice(contentStr); err != nil {
		return res, unifiedTokenInfos, NewBridgeAggErrorWithValue(OtherError, err)
	}
	contentBytes, err := base64.StdEncoding.DecodeString(tempInst.Content)
	if err != nil {
		return res, unifiedTokenInfos, NewBridgeAggErrorWithValue(OtherError, err)
	}
	content := metadata.IssuingEVMAcceptedInst{}
	err = json.Unmarshal(contentBytes, &content)
	if err != nil {
		return res, unifiedTokenInfos, NewBridgeAggErrorWithValue(OtherError, err)
	}
	inst := metadataCommon.NewInstructionWithValue(
		metadataCommon.IssuingUnifiedTokenRequestMeta,
		common.AcceptedStatusStr,
		content.ShardID,
		utils.EmptyString,
	)
	rejectContent := metadataCommon.NewRejectContentWithValue(content.TxReqID, 0, nil)
	vault := unifiedTokenInfos[content.IncTokenID][content.NetworkID] // check available before
	actualAmount, err := vault.shield(content.IssuingAmount)
	if err != nil {
		Logger.log.Warnf("Calculate shield amount error: %v tx %s", err, content.TxReqID.String())
		rejectContent.ErrorCode = ErrCodeMessage[CalculateShieldAmountError].Code
		temp, e := inst.StringSliceWithRejectContent(rejectContent)
		if e != nil {
			return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(OtherError, err)
		}
		return temp, unifiedTokenInfos, nil
	}
	unifiedTokenInfos[content.IncTokenID][content.NetworkID] = vault
	// build instruction content
	content.Reward = actualAmount - content.IssuingAmount
	content.IssuingAmount = actualAmount

	contentBytes, err = json.Marshal(content)
	if err != nil {
		return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(OtherError, err)
	}
	inst.Content = base64.StdEncoding.EncodeToString(contentBytes)
	return inst.StringSlice(), unifiedTokenInfos, nil
}
