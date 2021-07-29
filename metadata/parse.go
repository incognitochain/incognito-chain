package metadata

import (
	"encoding/json"

	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/pkg/errors"
)

func ParseMetadata(meta interface{}) (Metadata, error) {

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

	var md Metadata
	typeFloat, ok := mtTemp["Type"].(float64)
	if !ok {
		return nil, errors.Errorf("Could not parse metadata with type: %v", mtTemp["Type"])
	}
	theType := int(typeFloat)
	switch theType {
	case InitTokenRequestMeta:
		md = &InitTokenRequest{}
	case InitTokenResponseMeta:
		md = &InitTokenResponse{}
	case IssuingRequestMeta:
		md = &IssuingRequest{}
	case IssuingResponseMeta:
		md = &IssuingResponse{}
	case ContractingRequestMeta:
		md = &ContractingRequest{}
	case IssuingETHRequestMeta:
		md = &IssuingEVMRequest{}
	case IssuingBSCRequestMeta:
		md = &IssuingEVMRequest{}
	case IssuingETHResponseMeta:
		md = &IssuingEVMResponse{}
	case IssuingBSCResponseMeta:
		md = &IssuingEVMResponse{}
	case BeaconSalaryResponseMeta:
		md = &BeaconBlockSalaryRes{}
	case BurningRequestMeta:
		md = &BurningRequest{}
	case BurningRequestMetaV2:
		md = &BurningRequest{}
	case BurningPBSCRequestMeta:
		md = &BurningRequest{}
	case ShardStakingMeta:
		md = &StakingMetadata{}
	case BeaconStakingMeta:
		md = &StakingMetadata{}
	case ReturnStakingMeta:
		md = &ReturnStakingMetadata{}
	case WithDrawRewardRequestMeta:
		md = &WithDrawRewardRequest{}
	case WithDrawRewardResponseMeta:
		md = &WithDrawRewardResponse{}
	case UnStakingMeta:
		md = &UnStakingMetadata{}
	case StopAutoStakingMeta:
		md = &StopAutoStakingMetadata{}
	case PDEContributionMeta:
		md = &PDEContribution{}
	case PDEPRVRequiredContributionRequestMeta:
		md = &PDEContribution{}
	case PDETradeRequestMeta:
		md = &PDETradeRequest{}
	case PDETradeResponseMeta:
		md = &PDETradeResponse{}
	case PDECrossPoolTradeRequestMeta:
		md = &PDECrossPoolTradeRequest{}
	case PDECrossPoolTradeResponseMeta:
		md = &PDECrossPoolTradeResponse{}
	case PDEWithdrawalRequestMeta:
		md = &PDEWithdrawalRequest{}
	case PDEWithdrawalResponseMeta:
		md = &PDEWithdrawalResponse{}
	case PDEFeeWithdrawalRequestMeta:
		md = &PDEFeeWithdrawalRequest{}
	case PDEFeeWithdrawalResponseMeta:
		md = &PDEFeeWithdrawalResponse{}
	case PDEContributionResponseMeta:
		md = &PDEContributionResponse{}
	case PortalCustodianDepositMeta:
		md = &PortalCustodianDeposit{}
	case PortalRequestPortingMeta, PortalRequestPortingMetaV3:
		md = &PortalUserRegister{}
	case PortalUserRequestPTokenMeta:
		md = &PortalRequestPTokens{}
	case PortalCustodianDepositResponseMeta:
		md = &PortalCustodianDepositResponse{}
	case PortalUserRequestPTokenResponseMeta:
		md = &PortalRequestPTokensResponse{}
	case PortalRedeemRequestMeta, PortalRedeemRequestMetaV3:
		md = &PortalRedeemRequestV3{}
	case PortalRedeemRequestResponseMeta:
		md = &PortalRedeemRequestResponse{}
	case PortalRequestUnlockCollateralMeta, PortalRequestUnlockCollateralMetaV3:
		md = &PortalRequestUnlockCollateral{}
	case PortalExchangeRatesMeta:
		md = &PortalExchangeRates{}
	case PortalUnlockOverRateCollateralsMeta:
		md = &PortalUnlockOverRateCollaterals{}
	case RelayingBNBHeaderMeta:
		md = &RelayingHeader{}
	case RelayingBTCHeaderMeta:
		md = &RelayingHeader{}
	case PortalCustodianWithdrawRequestMeta:
		md = &PortalCustodianWithdrawRequest{}
	case PortalCustodianWithdrawResponseMeta:
		md = &PortalCustodianWithdrawResponse{}
	case PortalLiquidateCustodianMeta, PortalLiquidateCustodianMetaV3:
		md = &PortalLiquidateCustodian{}
	case PortalLiquidateCustodianResponseMeta:
		md = &PortalLiquidateCustodianResponse{}
	case PortalRequestWithdrawRewardMeta:
		md = &PortalRequestWithdrawReward{}
	case PortalRequestWithdrawRewardResponseMeta:
		md = &PortalWithdrawRewardResponse{}
	case PortalRedeemFromLiquidationPoolMeta:
		md = &PortalRedeemLiquidateExchangeRates{}
	case PortalRedeemFromLiquidationPoolResponseMeta:
		md = &PortalRedeemLiquidateExchangeRatesResponse{}
	case PortalCustodianTopupMetaV2:
		md = &PortalLiquidationCustodianDepositV2{}
	case PortalCustodianTopupResponseMetaV2:
		md = &PortalLiquidationCustodianDepositResponseV2{}
	case PortalCustodianTopupMeta:
		md = &PortalLiquidationCustodianDeposit{}
	case PortalCustodianTopupResponseMeta:
		md = &PortalLiquidationCustodianDepositResponse{}
	case BurningForDepositToSCRequestMeta:
		md = &BurningRequest{}
	case BurningForDepositToSCRequestMetaV2:
		md = &BurningRequest{}
	case PortalPortingResponseMeta:
		md = &PortalFeeRefundResponse{}
	case PortalReqMatchingRedeemMeta:
		md = &PortalReqMatchingRedeem{}
	case PortalTopUpWaitingPortingRequestMeta:
		md = &PortalTopUpWaitingPortingRequest{}
	case PortalTopUpWaitingPortingResponseMeta:
		md = &PortalTopUpWaitingPortingResponse{}
	case PortalCustodianDepositMetaV3:
		md = &PortalCustodianDepositV3{}
	case PortalCustodianWithdrawRequestMetaV3:
		md = &PortalCustodianWithdrawRequestV3{}
	case PortalRedeemFromLiquidationPoolMetaV3:
		md = &PortalRedeemFromLiquidationPoolV3{}
	case PortalRedeemFromLiquidationPoolResponseMetaV3:
		md = &PortalRedeemFromLiquidationPoolResponseV3{}
	case PortalCustodianTopupMetaV3:
		md = &PortalLiquidationCustodianDepositV3{}
	case PortalTopUpWaitingPortingRequestMetaV3:
		md = &PortalTopUpWaitingPortingRequestV3{}
	case metadataCommon.Pdexv3ModifyParamsMeta:
		md = &metadataPdexv3.ParamsModifyingRequest{}
	case metadataCommon.Pdexv3AddLiquidityRequestMeta:
		md = &metadataPdexv3.AddLiquidity{}
	case metadataCommon.Pdexv3WithdrawLPFeeRequestMeta:
		md = &metadataPdexv3.WithdrawalLPFeeRequest{}
	case metadataCommon.Pdexv3WithdrawProtocolFeeRequestMeta:
		md = &metadataPdexv3.WithdrawalProtocolFeeRequest{}
	default:
		Logger.Log.Debug("[db] parse meta err: %+v\n", meta)
		return nil, errors.Errorf("Could not parse metadata with type: %d", theType)
	}

	err = json.Unmarshal(metaInBytes, &md)
	if err != nil {
		return nil, err
	}

	return md, nil
}
