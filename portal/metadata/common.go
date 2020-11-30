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
	//case PortalRequestPortingMeta, PortalRequestPortingMetaV3:
	//	md = &PortalUserRegister{}
	//case PortalUserRequestPTokenMeta:
	//	md = &PortalRequestPTokens{}
	//case PortalCustodianDepositResponseMeta:
	//	md = &PortalCustodianDepositResponse{}
	//case PortalUserRequestPTokenResponseMeta:
	//	md = &PortalRequestPTokensResponse{}
	//case PortalRedeemRequestMeta, PortalRedeemRequestMetaV3:
	//	md = &PortalRedeemRequest{}
	//case PortalRedeemRequestResponseMeta:
	//	md = &PortalRedeemRequestResponse{}
	//case PortalRequestUnlockCollateralMeta, PortalRequestUnlockCollateralMetaV3:
	//	md = &PortalRequestUnlockCollateral{}
	//case PortalExchangeRatesMeta:
	//	md = &PortalExchangeRates{}
	//case RelayingBNBHeaderMeta:
	//	md = &RelayingHeader{}
	//case RelayingBTCHeaderMeta:
	//	md = &RelayingHeader{}
	//case PortalCustodianWithdrawRequestMeta:
	//	md = &PortalCustodianWithdrawRequest{}
	//case PortalCustodianWithdrawResponseMeta:
	//	md = &PortalCustodianWithdrawResponse{}
	//case PortalLiquidateCustodianMeta, PortalLiquidateCustodianMetaV3:
	//	md = &PortalLiquidateCustodian{}
	//case PortalLiquidateCustodianResponseMeta:
	//	md = &PortalLiquidateCustodianResponse{}
	//case PortalRequestWithdrawRewardMeta:
	//	md = &PortalRequestWithdrawReward{}
	//case PortalRequestWithdrawRewardResponseMeta:
	//	md = &PortalWithdrawRewardResponse{}
	//case PortalRedeemFromLiquidationPoolMeta:
	//	md = &PortalRedeemLiquidateExchangeRates{}
	//case PortalRedeemFromLiquidationPoolResponseMeta:
	//	md = &PortalRedeemLiquidateExchangeRatesResponse{}
	//case PortalCustodianTopupMetaV2:
	//	md = &PortalLiquidationCustodianDepositV2{}
	//case PortalCustodianTopupResponseMetaV2:
	//	md = &PortalLiquidationCustodianDepositResponseV2{}
	//case PortalCustodianTopupMeta:
	//	md = &PortalLiquidationCustodianDeposit{}
	//case PortalCustodianTopupResponseMeta:
	//	md = &PortalLiquidationCustodianDepositResponse{}
	//case BurningForDepositToSCRequestMeta:
	//	md = &BurningRequest{}
	//case BurningForDepositToSCRequestMetaV2:
	//	md = &BurningRequest{}
	//case PortalPortingResponseMeta:
	//	md = &PortalFeeRefundResponse{}
	//case PortalReqMatchingRedeemMeta:
	//	md = &PortalReqMatchingRedeem{}
	//case PortalTopUpWaitingPortingRequestMeta:
	//	md = &PortalTopUpWaitingPortingRequest{}
	//case PortalTopUpWaitingPortingResponseMeta:
	//	md = &PortalTopUpWaitingPortingResponse{}
	//case PortalCustodianDepositMetaV3:
	//	md = &PortalCustodianDepositV3{}
	//case PortalCustodianWithdrawRequestMetaV3:
	//	md = &PortalCustodianWithdrawRequestV3{}
	//case PortalRedeemFromLiquidationPoolMetaV3:
	//	md = &PortalRedeemFromLiquidationPoolV3{}
	//case PortalRedeemFromLiquidationPoolResponseMetaV3:
	//	md = &PortalRedeemFromLiquidationPoolResponseV3{}
	//case PortalCustodianTopupMetaV3:
	//	md = &PortalLiquidationCustodianDepositV3{}
	//case PortalTopUpWaitingPortingRequestMetaV3:
	//	md = &PortalTopUpWaitingPortingRequestV3{}
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


//todo: should move to portal common
func IsValidPortalRemoteAddress(
	bcr basemeta.ChainRetriever,
	remoteAddress string,
	tokenID string,
) bool {
	if tokenID == common.PortalBNBIDStr {
		return bnb.IsValidBNBAddress(remoteAddress, bcr.GetBNBChainID())
	} else if tokenID == common.PortalBTCIDStr {
		btcHeaderChain := bcr.GetBTCHeaderChain()
		if btcHeaderChain == nil {
			return false
		}
		return btcHeaderChain.IsBTCAddressValid(remoteAddress)
	}
	return false
}

func IsPortalToken(tokenIDStr string) bool {
	isExisted, _ := common.SliceExists(common.PortalSupportedIncTokenIDs, tokenIDStr)
	return isExisted
}

func IsSupportedTokenCollateralV3(bcr basemeta.ChainRetriever, beaconHeight uint64, externalTokenID string) bool {
	isSupported, _ := common.SliceExists(bcr.GetSupportedCollateralTokenIDs(beaconHeight), externalTokenID)
	return isSupported
}

func IsPortalExchangeRateToken(tokenIDStr string, bcr basemeta.ChainRetriever, beaconHeight uint64) bool {
	return IsPortalToken(tokenIDStr) || tokenIDStr == common.PRVIDStr || IsSupportedTokenCollateralV3(bcr, beaconHeight, tokenIDStr)
}


// Validate portal remote addresses for portal tokens (BTC, BNB)
func ValidatePortalRemoteAddresses(remoteAddresses map[string]string, chainRetriever basemeta.ChainRetriever) (bool, error){
	if len(remoteAddresses) == 0 {
		return false, errors.New("remote addresses should be at least one address")
	}
	for tokenID, remoteAddr := range remoteAddresses {
		if !IsPortalToken(tokenID) {
			return false, errors.New("TokenID in remote address is invalid")
		}
		if len(remoteAddr) == 0 {
			return false, errors.New("Remote address is invalid")
		}
		if !IsValidPortalRemoteAddress(chainRetriever, remoteAddr, tokenID) {
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