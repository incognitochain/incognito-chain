package metadata

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/basemeta"

	"strconv"

	"github.com/pkg/errors"
)

func ParseMetadata(meta interface{}) (basemeta.Metadata, error) {
	if meta == nil {
		return nil, nil
	}

	mtTemp := map[string]interface{}{}
	metaInBytes, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(metaInBytes, &mtTemp)
	if err != nil {
		return nil, err
	}
	var md basemeta.Metadata
	metaType := int(mtTemp["Type"].(float64))

	switch metaType {
	case basemeta.IssuingRequestMeta:
		md = &IssuingRequest{}
	case basemeta.IssuingResponseMeta:
		md = &IssuingResponse{}
	case basemeta.ContractingRequestMeta:
		md = &ContractingRequest{}
	case basemeta.IssuingETHRequestMeta:
		md = &IssuingETHRequest{}
	case basemeta.IssuingETHResponseMeta:
		md = &IssuingETHResponse{}
	case basemeta.BeaconSalaryResponseMeta:
		md = &BeaconBlockSalaryRes{}
	case basemeta.BurningRequestMeta:
		md = &BurningRequest{}
	case basemeta.BurningRequestMetaV2:
		md = &BurningRequest{}
	case basemeta.ShardStakingMeta:
		md = &StakingMetadata{}
	case basemeta.BeaconStakingMeta:
		md = &StakingMetadata{}
	case basemeta.ReturnStakingMeta:
		md = &ReturnStakingMetadata{}
	case basemeta.WithDrawRewardRequestMeta:
		md = &WithDrawRewardRequest{}
	case basemeta.WithDrawRewardResponseMeta:
		md = &WithDrawRewardResponse{}
	case basemeta.StopAutoStakingMeta:
		md = &StopAutoStakingMetadata{}
	case basemeta.PDEContributionMeta:
		md = &PDEContribution{}
	case basemeta.PDEPRVRequiredContributionRequestMeta:
		md = &PDEContribution{}
	case basemeta.PDETradeRequestMeta:
		md = &PDETradeRequest{}
	case basemeta.PDETradeResponseMeta:
		md = &PDETradeResponse{}
	case basemeta.PDECrossPoolTradeRequestMeta:
		md = &PDECrossPoolTradeRequest{}
	case basemeta.PDECrossPoolTradeResponseMeta:
		md = &PDECrossPoolTradeResponse{}
	case basemeta.PDEWithdrawalRequestMeta:
		md = &PDEWithdrawalRequest{}
	case basemeta.PDEWithdrawalResponseMeta:
		md = &PDEWithdrawalResponse{}
	case basemeta.PDEFeeWithdrawalRequestMeta:
		md = &PDEFeeWithdrawalRequest{}
	case basemeta.PDEFeeWithdrawalResponseMeta:
		md = &PDEFeeWithdrawalResponse{}
	case basemeta.PDEContributionResponseMeta:
		md = &PDEContributionResponse{}
	case basemeta.BurningForDepositToSCRequestMeta:
		md = &BurningRequest{}
	case basemeta.BurningForDepositToSCRequestMetaV2:
		md = &BurningRequest{}
	default:
		Logger.log.Debug("[db] parse meta err: %+v\n", meta)
		return nil, errors.Errorf("Could not parse metadata with type: %d", int(mtTemp["Type"].(float64)))
	}

	err = json.Unmarshal(metaInBytes, &md)
	if err != nil {
		return nil, err
	}
	return md, nil
}


var bridgeMetas = []string{
	strconv.Itoa(basemeta.BeaconSwapConfirmMeta),
	strconv.Itoa(basemeta.BridgeSwapConfirmMeta),
	strconv.Itoa(basemeta.BurningConfirmMeta),
	strconv.Itoa(basemeta.BurningConfirmForDepositToSCMeta),
	strconv.Itoa(basemeta.BurningConfirmMetaV2),
	strconv.Itoa(basemeta.BurningConfirmForDepositToSCMetaV2),
}

func HasBridgeInstructions(instructions [][]string) bool {
	for _, inst := range instructions {
		for _, meta := range bridgeMetas {
			if len(inst) > 0 && inst[0] == meta {
				return true
			}
		}
	}
	return false
}

// TODO: add more meta data types
var portalMetas = []string{
	strconv.Itoa(basemeta.PortalCustodianWithdrawConfirmMetaV3),
	strconv.Itoa(basemeta.PortalRedeemFromLiquidationPoolConfirmMetaV3),
	strconv.Itoa(basemeta.PortalLiquidateRunAwayCustodianConfirmMetaV3),
}

func HasPortalInstructions(instructions [][]string) bool {
	for _, inst := range instructions {
		for _, meta := range portalMetas {
			if len(inst) > 0 && inst[0] == meta {
				return true
			}
		}
	}
	return false
}
