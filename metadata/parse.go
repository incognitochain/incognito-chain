package metadata

import (
	"encoding/json"

	metadataBridge "github.com/incognitochain/incognito-chain/metadata/bridge"
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
		md = &metadataBridge.IssuingEVMRequest{}
	case IssuingBSCRequestMeta:
		md = &metadataBridge.IssuingEVMRequest{}
	case IssuingPRVERC20RequestMeta:
		md = &metadataBridge.IssuingEVMRequest{}
	case IssuingPRVBEP20RequestMeta:
		md = &metadataBridge.IssuingEVMRequest{}
	case IssuingETHResponseMeta:
		md = &metadataBridge.IssuingEVMResponse{}
	case IssuingBSCResponseMeta:
		md = &metadataBridge.IssuingEVMResponse{}
	case IssuingPRVERC20ResponseMeta:
		md = &metadataBridge.IssuingEVMResponse{}
	case IssuingPRVBEP20ResponseMeta:
		md = &metadataBridge.IssuingEVMResponse{}
	case BeaconSalaryResponseMeta:
		md = &BeaconBlockSalaryRes{}
	case BurningRequestMeta:
		md = &metadataBridge.BurningRequest{}
	case BurningRequestMetaV2:
		md = &metadataBridge.BurningRequest{}
	case BurningPBSCRequestMeta:
		md = &metadataBridge.BurningRequest{}
	case BurningPRVBEP20RequestMeta:
		md = &metadataBridge.BurningRequest{}
	case BurningPRVERC20RequestMeta:
		md = &metadataBridge.BurningRequest{}
	case IssuingPLGRequestMeta:
		md = &metadataBridge.IssuingEVMRequest{}
	case IssuingPLGResponseMeta:
		md = &metadataBridge.IssuingEVMResponse{}
	case BurningPLGRequestMeta:
		md = &metadataBridge.BurningRequest{}
	case BurningPLGForDepositToSCRequestMeta:
		md = &metadataBridge.BurningRequest{}
	case IssuingFantomRequestMeta:
		md = &metadataBridge.IssuingEVMRequest{}
	case IssuingFantomResponseMeta:
		md = &metadataBridge.IssuingEVMResponse{}
	case IssuingAuroraRequestMeta:
		md = &metadataBridge.IssuingEVMAuroraRequest{}
	case IssuingAuroraResponseMeta:
		md = &metadataBridge.IssuingEVMResponse{}
	case IssuingAvaxRequestMeta:
		md = &metadataBridge.IssuingEVMRequest{}
	case IssuingAvaxResponseMeta:
		md = &metadataBridge.IssuingEVMResponse{}
	case BurningFantomRequestMeta:
		md = &metadataBridge.BurningRequest{}
	case BurningFantomForDepositToSCRequestMeta:
		md = &metadataBridge.BurningRequest{}
	case BurningAuroraForDepositToSCRequestMeta:
		md = &metadataBridge.BurningRequest{}
	case BurningAvaxForDepositToSCRequestMeta:
		md = &metadataBridge.BurningRequest{}
	case BurningAuroraRequestMeta:
		md = &metadataBridge.BurningRequest{}
	case BurningAvaxRequestMeta:
		md = &metadataBridge.BurningRequest{}
	case IssuingNearRequestMeta:
		md = &metadataBridge.IssuingWasmRequest{}
	case IssuingNearResponseMeta:
		md = &metadataBridge.IssuingEVMResponse{}
	case BurningNearRequestMeta:
		md = &metadataBridge.BurningRequest{}
	case ShardStakingMeta:
		md = &StakingMetadata{}
	case BeaconStakingMeta:
		md = &StakingMetadata{}
	case ReturnStakingMeta:
		md = &ReturnStakingMetadata{}
	case ReturnBeaconStakingMeta:
		md = &ReturnBeaconStakingMetadata{}
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
		md = &metadataBridge.BurningRequest{}
	case BurningPBSCForDepositToSCRequestMeta:
		md = &metadataBridge.BurningRequest{}
	case BurningForDepositToSCRequestMetaV2:
		md = &metadataBridge.BurningRequest{}
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
	case metadataCommon.PortalV4ShieldingRequestMeta:
		md = &PortalShieldingRequest{}
	case metadataCommon.PortalV4ShieldingResponseMeta:
		md = &PortalShieldingResponse{}
	case metadataCommon.PortalV4UnshieldingRequestMeta:
		md = &PortalUnshieldRequest{}
	case metadataCommon.PortalV4UnshieldingResponseMeta:
		md = &PortalUnshieldResponse{}
	case metadataCommon.PortalV4FeeReplacementRequestMeta:
		md = &PortalReplacementFeeRequest{}
	case metadataCommon.PortalV4SubmitConfirmedTxMeta:
		md = &PortalSubmitConfirmedTxRequest{}
	case metadataCommon.PortalV4ConvertVaultRequestMeta:
		md = &PortalConvertVaultRequest{}
	case metadataCommon.Pdexv3ModifyParamsMeta:
		md = &metadataPdexv3.ParamsModifyingRequest{}
	case metadataCommon.Pdexv3AddLiquidityRequestMeta:
		md = &metadataPdexv3.AddLiquidityRequest{}
	case metadataCommon.Pdexv3AddLiquidityResponseMeta:
		md = &metadataPdexv3.AddLiquidityResponse{}
	case metadataCommon.Pdexv3WithdrawLiquidityRequestMeta:
		md = &metadataPdexv3.WithdrawLiquidityRequest{}
	case metadataCommon.Pdexv3WithdrawLiquidityResponseMeta:
		md = &metadataPdexv3.WithdrawLiquidityResponse{}
	case metadataCommon.Pdexv3TradeRequestMeta:
		md = &metadataPdexv3.TradeRequest{}
	case metadataCommon.Pdexv3TradeResponseMeta:
		md = &metadataPdexv3.TradeResponse{}
	case metadataCommon.Pdexv3AddOrderRequestMeta:
		md = &metadataPdexv3.AddOrderRequest{}
	case metadataCommon.Pdexv3AddOrderResponseMeta:
		md = &metadataPdexv3.AddOrderResponse{}
	case metadataCommon.Pdexv3UserMintNftRequestMeta:
		md = &metadataPdexv3.UserMintNftRequest{}
	case metadataCommon.Pdexv3UserMintNftResponseMeta:
		md = &metadataPdexv3.UserMintNftResponse{}
	case metadataCommon.Pdexv3MintNftResponseMeta:
		md = &metadataPdexv3.MintNftResponse{}
	case metadataCommon.Pdexv3WithdrawOrderRequestMeta:
		md = &metadataPdexv3.WithdrawOrderRequest{}
	case metadataCommon.Pdexv3WithdrawOrderResponseMeta:
		md = &metadataPdexv3.WithdrawOrderResponse{}
	case metadataCommon.Pdexv3StakingRequestMeta:
		md = &metadataPdexv3.StakingRequest{}
	case metadataCommon.Pdexv3StakingResponseMeta:
		md = &metadataPdexv3.StakingResponse{}
	case metadataCommon.Pdexv3UnstakingRequestMeta:
		md = &metadataPdexv3.UnstakingRequest{}
	case metadataCommon.Pdexv3UnstakingResponseMeta:
		md = &metadataPdexv3.UnstakingResponse{}
	case metadataCommon.Pdexv3WithdrawLPFeeRequestMeta:
		md = &metadataPdexv3.WithdrawalLPFeeRequest{}
	case metadataCommon.Pdexv3WithdrawLPFeeResponseMeta:
		md = &metadataPdexv3.WithdrawalLPFeeResponse{}
	case metadataCommon.Pdexv3WithdrawProtocolFeeRequestMeta:
		md = &metadataPdexv3.WithdrawalProtocolFeeRequest{}
	case metadataCommon.Pdexv3WithdrawProtocolFeeResponseMeta:
		md = &metadataPdexv3.WithdrawalProtocolFeeResponse{}
	case metadataCommon.Pdexv3MintPDEXGenesisMeta:
		md = &metadataPdexv3.MintPDEXGenesisResponse{}
	case metadataCommon.Pdexv3WithdrawStakingRewardRequestMeta:
		md = &metadataPdexv3.WithdrawalStakingRewardRequest{}
	case metadataCommon.Pdexv3WithdrawStakingRewardResponseMeta:
		md = &metadataPdexv3.WithdrawalStakingRewardResponse{}
	case metadataCommon.BridgeAggModifyParamMeta:
		md = &metadataBridge.ModifyBridgeAggParamReq{}
	case metadataCommon.BridgeAggConvertTokenToUnifiedTokenRequestMeta:
		md = &metadataBridge.ConvertTokenToUnifiedTokenRequest{}
	case metadataCommon.BridgeAggConvertTokenToUnifiedTokenResponseMeta:
		md = &metadataBridge.ConvertTokenToUnifiedTokenResponse{}
	case metadataCommon.IssuingUnifiedTokenRequestMeta:
		md = &metadataBridge.ShieldRequest{}
	case metadataCommon.IssuingUnifiedTokenResponseMeta:
		md = &metadataBridge.ShieldResponse{}
	case metadataCommon.BurningUnifiedTokenRequestMeta:
		md = &metadataBridge.UnshieldRequest{}
	case metadataCommon.BurningUnifiedTokenResponseMeta:
		md = &metadataBridge.UnshieldResponse{}
	case metadataCommon.BurnForCallRequestMeta:
		md = &metadataBridge.BurnForCallRequest{}
	case metadataCommon.BurnForCallResponseMeta:
		md = &metadataBridge.BurnForCallResponse{}
	case metadataCommon.IssuingReshieldResponseMeta:
		md = &metadataBridge.IssuingReshieldResponse{}
	default:
		Logger.log.Debug("parse meta err: %+v\n", meta)
		return nil, errors.Errorf("Could not parse metadata with type: %d", theType)
	}

	err = json.Unmarshal(metaInBytes, &md)
	if err != nil {
		return nil, err
	}

	return md, nil
}
