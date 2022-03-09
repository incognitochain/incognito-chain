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
	shardID byte,
	unifiedTokenInfos map[common.Hash]map[common.Hash]*Vault,
	sDBs map[int]*statedb.StateDB,
) ([][]string, map[common.Hash]map[common.Hash]*Vault, error) {
	action := metadataCommon.NewAction()
	err := action.FromString(contentStr)
	if err != nil {
		return [][]string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(OtherError, err)
	}
	inst := metadataCommon.NewInstructionWithValue(
		metadataCommon.BridgeAggModifyListTokenMeta,
		common.AcceptedStatusStr,
		shardID,
		utils.EmptyString,
	)
	rejectContent := metadataCommon.NewRejectContentWithValue(action.TxReqID, 0)
	md, _ := action.Meta.(*metadataBridgeAgg.ModifyListToken)
	for k, v := range md.NewListTokens {
		if err := CheckTokenIDExisted(sDBs, k); err != nil {
			rejectContent.ErrorCode = ErrCodeMessage[NotFoundTokenIDInNetwork].Code
			temp, err := inst.StringSliceWithRejectContent(rejectContent)
			if err != nil {
				return [][]string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(NotFoundTokenIDInNetwork, err)
			}
			return [][]string{temp}, unifiedTokenInfos, nil
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
					return [][]string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(NotFoundTokenIDInNetwork, err)
				}
				return [][]string{temp}, unifiedTokenInfos, nil
			}
			if _, found := unifiedTokenInfos[k][tokenID]; !found {
				unifiedTokenInfos[k][tokenID] = NewVault()
			}
		}
	}
	acceptedInst := metadataBridgeAgg.AcceptedModifyListToken{
		NewListTokens: md.NewListTokens,
	}
	contentBytes, err := json.Marshal(acceptedInst)
	if err != nil {
		return [][]string{}, unifiedTokenInfos, err
	}
	inst.Content = base64.StdEncoding.EncodeToString(contentBytes)
	return [][]string{inst.StringSlice()}, unifiedTokenInfos, nil
}
