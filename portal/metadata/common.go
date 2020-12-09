package metadata

import (
	"encoding/json"
	"fmt"
	ec "github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/relaying/bnb"
	"github.com/incognitochain/incognito-chain/basemeta"
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
	switch int(mtTemp["Type"].(float64)) {
	case basemeta.PortalCustodianDepositMeta:
		md = &PortalCustodianDeposit{}
	case basemeta.PortalRequestPortingMeta, basemeta.PortalRequestPortingMetaV3:
		md = &PortalUserRegister{}
	case basemeta.PortalUserRequestPTokenMeta:
		md = &PortalRequestPTokens{}
	case basemeta.PortalCustodianDepositResponseMeta:
		md = &PortalCustodianDepositResponse{}
	case basemeta.PortalUserRequestPTokenResponseMeta:
		md = &PortalRequestPTokensResponse{}
	case basemeta.PortalRedeemRequestMeta, basemeta.PortalRedeemRequestMetaV3:
		md = &PortalRedeemRequest{}
	case basemeta.PortalRedeemRequestResponseMeta:
		md = &PortalRedeemRequestResponse{}
	case basemeta.PortalRequestUnlockCollateralMeta, basemeta.PortalRequestUnlockCollateralMetaV3:
		md = &PortalRequestUnlockCollateral{}
	case basemeta.PortalExchangeRatesMeta:
		md = &PortalExchangeRates{}
	case basemeta.RelayingBNBHeaderMeta:
		md = &RelayingHeader{}
	case basemeta.RelayingBTCHeaderMeta:
		md = &RelayingHeader{}
	case basemeta.PortalCustodianWithdrawRequestMeta:
		md = &PortalCustodianWithdrawRequest{}
	case basemeta.PortalCustodianWithdrawResponseMeta:
		md = &PortalCustodianWithdrawResponse{}
	case basemeta.PortalLiquidateCustodianMeta, basemeta.PortalLiquidateCustodianMetaV3:
		md = &PortalLiquidateCustodian{}
	case basemeta.PortalLiquidateCustodianResponseMeta:
		md = &PortalLiquidateCustodianResponse{}
	case basemeta.PortalRequestWithdrawRewardMeta:
		md = &PortalRequestWithdrawReward{}
	case basemeta.PortalRequestWithdrawRewardResponseMeta:
		md = &PortalWithdrawRewardResponse{}
	case basemeta.PortalRedeemFromLiquidationPoolMeta:
		md = &PortalRedeemLiquidateExchangeRates{}
	case basemeta.PortalRedeemFromLiquidationPoolResponseMeta:
		md = &PortalRedeemLiquidateExchangeRatesResponse{}
	case basemeta.PortalCustodianTopupMetaV2:
		md = &PortalLiquidationCustodianDepositV2{}
	case basemeta.PortalCustodianTopupResponseMetaV2:
		md = &PortalLiquidationCustodianDepositResponseV2{}
	case basemeta.PortalCustodianTopupMeta:
		md = &PortalLiquidationCustodianDeposit{}
	case basemeta.PortalCustodianTopupResponseMeta:
		md = &PortalLiquidationCustodianDepositResponse{}
	case basemeta.PortalPortingResponseMeta:
		md = &PortalFeeRefundResponse{}
	case basemeta.PortalReqMatchingRedeemMeta:
		md = &PortalReqMatchingRedeem{}
	case basemeta.PortalTopUpWaitingPortingRequestMeta:
		md = &PortalTopUpWaitingPortingRequest{}
	case basemeta.PortalTopUpWaitingPortingResponseMeta:
		md = &PortalTopUpWaitingPortingResponse{}
	case basemeta.PortalCustodianDepositMetaV3:
		md = &PortalCustodianDepositV3{}
	case basemeta.PortalCustodianWithdrawRequestMetaV3:
		md = &PortalCustodianWithdrawRequestV3{}
	case basemeta.PortalRedeemFromLiquidationPoolMetaV3:
		md = &PortalRedeemFromLiquidationPoolV3{}
	case basemeta.PortalRedeemFromLiquidationPoolResponseMetaV3:
		md = &PortalRedeemFromLiquidationPoolResponseV3{}
	case basemeta.PortalCustodianTopupMetaV3:
		md = &PortalLiquidationCustodianDepositV3{}
	case basemeta.PortalTopUpWaitingPortingRequestMetaV3:
		md = &PortalTopUpWaitingPortingRequestV3{}
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

func IsValidPortalRemoteAddress(
	bcr basemeta.ChainRetriever,
	remoteAddress string,
	tokenID string,
	beaconHeight uint64,
) bool {
	if tokenID == common.PortalBNBIDStr {
		return bnb.IsValidBNBAddress(remoteAddress, bcr.GetBNBChainID(beaconHeight))
	} else if tokenID == common.PortalBTCIDStr {
		btcHeaderChain := bcr.GetBTCHeaderChain()
		if btcHeaderChain == nil {
			return false
		}
		return btcHeaderChain.IsBTCAddressValid(remoteAddress)
	}
	return false
}

// Validate portal remote addresses for portal tokens (BTC, BNB)
func ValidatePortalRemoteAddresses(remoteAddresses map[string]string, chainRetriever basemeta.ChainRetriever, beaconHeight uint64) (bool, error){
	if len(remoteAddresses) == 0 {
		return false, errors.New("remote addresses should be at least one address")
	}
	for tokenID, remoteAddr := range remoteAddresses {
		if !chainRetriever.IsPortalToken(beaconHeight, tokenID) {
			return false, errors.New("TokenID in remote address is invalid")
		}
		if len(remoteAddr) == 0 {
			return false, errors.New("Remote address is invalid")
		}
		if !IsValidPortalRemoteAddress(chainRetriever, remoteAddr, tokenID, beaconHeight) {
			return false, fmt.Errorf("Remote address %v is not a valid address of tokenID %v", remoteAddr, tokenID)
		}
	}

	return true, nil
}

// Validate portal external addresses for collateral tokens (ETH/ERC20)
func ValidatePortalExternalAddress(chainName string, tokenID string, address string) (bool, error) {
	switch chainName {
	case common.ETHChainName:
		return ec.IsHexAddress(address), nil
	}
	return true, nil
}