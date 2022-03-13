package bridgeagg

import (
	"encoding/base64"
	"encoding/json"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataBridgeAgg "github.com/incognitochain/incognito-chain/metadata/bridgeagg"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/utils"
)

type stateProducer struct {
}

func (sp *stateProducer) modifyListTokens(
	contentStr string,
	unifiedTokenInfos map[common.Hash]map[common.Hash]*Vault,
	sDBs map[int]*statedb.StateDB,
) ([]string, map[common.Hash]map[common.Hash]*Vault, error) {
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
	for k, v := range md.NewListTokens {
		if err := CheckTokenIDExisted(sDBs, k); err != nil {
			rejectContent.ErrorCode = ErrCodeMessage[NotFoundTokenIDInNetwork].Code
			temp, err := inst.StringSliceWithRejectContent(rejectContent)
			if err != nil {
				return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(NotFoundTokenIDInNetwork, err)
			}
			return temp, unifiedTokenInfos, nil
		}
		_, found := unifiedTokenInfos[k]
		if !found {
			unifiedTokenInfos[k] = make(map[common.Hash]*Vault)
		}
		for _, tokenID := range v {
			if err := CheckTokenIDExisted(sDBs, tokenID); err != nil {
				rejectContent.ErrorCode = ErrCodeMessage[NotFoundTokenIDInNetwork].Code
				temp, err := inst.StringSliceWithRejectContent(rejectContent)
				if err != nil {
					return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(NotFoundTokenIDInNetwork, err)
				}
				return temp, unifiedTokenInfos, nil
			}
			if _, found := unifiedTokenInfos[k][tokenID]; !found {
				unifiedTokenInfos[k][tokenID] = NewVault()
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
	contentStr string, unifiedTokenInfos map[common.Hash]map[common.Hash]*Vault, sDBs map[int]*statedb.StateDB,
) ([]string, map[common.Hash]map[common.Hash]*Vault, error) {
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
		rejectContent.ErrorCode = ErrCodeMessage[NotFoundTokenIDInNetwork].Code
		temp, err := inst.StringSliceWithRejectContent(rejectContent)
		if err != nil {
			return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(NotFoundTokenIDInNetwork, err)
		}
		return temp, unifiedTokenInfos, nil
	}
	if vault, found := unifiedTokenInfos[md.UnifiedTokenID][md.TokenID]; !found {
		rejectContent.ErrorCode = ErrCodeMessage[NotFoundTokenIDInNetwork].Code
		temp, err := inst.StringSliceWithRejectContent(rejectContent)
		if err != nil {
			return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(NotFoundTokenIDInNetwork, err)
		}
		return temp, unifiedTokenInfos, nil
	} else {
		vault.Convert(md.Amount)
		unifiedTokenInfos[md.UnifiedTokenID][md.TokenID] = vault
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
