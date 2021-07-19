package pdex

import (
	"encoding/base64"
	"encoding/json"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataPDexV3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
)

type stateProducerV2 struct {
	stateProducerBase
}

func buildModifyParamsInst(
	params metadataPDexV3.PDexV3Params,
	shardID byte,
	reqTxID common.Hash,
	status string,
) []string {
	modifyingParamsReqContent := metadataPDexV3.PDexV3ParamsModifyingRequestContent{
		Content: params,
		TxReqID: reqTxID,
		ShardID: shardID,
	}
	modifyingParamsReqContentBytes, _ := json.Marshal(modifyingParamsReqContent)
	return []string{
		strconv.Itoa(metadataCommon.PDexV3ModifyParamsMeta),
		strconv.Itoa(int(shardID)),
		status,
		string(modifyingParamsReqContentBytes),
	}
}

func isValidPDexV3Params(params Params) bool {
	if params.DefaultFeeRateBPS > MaxFeeRateBPS {
		return false
	}
	for _, feeRate := range params.FeeRateBPS {
		if feeRate > MaxFeeRateBPS {
			return false
		}
	}
	if params.PRVDiscountPercent > MaxPRVDiscountPercent {
		return false
	}
	if params.TradingStakingPoolRewardPercent+params.TradingProtocolFeePercent > 100 {
		return false
	}
	if params.LimitProtocolFeePercent+params.LimitStakingPoolRewardPercent > 100 {
		return false
	}
	return true
}

func (sp *stateProducerV2) modifyParams(
	actions [][]string,
	beaconHeight uint64,
	params Params,
) ([][]string, Params, error) {
	instructions := [][]string{}

	for _, action := range actions {
		contentStr := action[1]
		contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while decoding content string of pdex v3 modify params action: %+v", err)
			return [][]string{}, params, err
		}
		var modifyParamsRequestAction metadataPDexV3.PDexV3ParamsModifyingRequestAction
		err = json.Unmarshal(contentBytes, &modifyParamsRequestAction)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while unmarshaling pdex v3 modify params action: %+v", err)
			return [][]string{}, params, err
		}

		// check conditions
		metadataParams := modifyParamsRequestAction.Meta.PDexV3Params
		newParams := Params(metadataParams)
		isValidParams := isValidPDexV3Params(newParams)

		status := ""
		if isValidParams {
			status = RequestAcceptedChainStatus
			params = newParams
		} else {
			status = RequestRejectedChainStatus
		}

		inst := buildModifyParamsInst(
			metadataParams,
			modifyParamsRequestAction.ShardID,
			modifyParamsRequestAction.TxReqID,
			status,
		)
		if err != nil {
			return [][]string{}, params, nil
		}
		instructions = append(instructions, inst)
	}

	return instructions, params, nil
}
