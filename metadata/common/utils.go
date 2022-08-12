package common

import (
	"encoding/json"
	"fmt"
	"strconv"

	ec "github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

func CalculateSize(meta Metadata) uint64 {
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return 0
	}
	return uint64(len(metaBytes))
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

type MetaInfo struct {
	HasInput   bool
	HasOutput  bool
	TxType     map[string]interface{}
	MetaAction int
}

const (
	NoAction = iota
	MetaRequestBeaconMintTxs
	MetaRequestShardMintTxs
)

var metaInfoMap map[int]*MetaInfo
var limitOfMetaAct map[int]int

func setLimitMetadataInBlock() {
	limitOfMetaAct = map[int]int{}
	limitOfMetaAct[MetaRequestBeaconMintTxs] = 400
	limitOfMetaAct[MetaRequestShardMintTxs] = 300
}

func buildMetaInfo() {
	type ListAndInfo struct {
		list []int
		info *MetaInfo
	}
	metaListNInfo := []ListAndInfo{}
	listTpNoInput := []int{
		PDETradeResponseMeta,
		PDEWithdrawalResponseMeta,
		PDEContributionResponseMeta,
		PDECrossPoolTradeResponseMeta,
		PortalRequestWithdrawRewardResponseMeta,
		PortalRedeemFromLiquidationPoolResponseMeta,
		PortalRedeemFromLiquidationPoolResponseMetaV3,
		PortalUserRequestPTokenResponseMeta,
		PortalRedeemRequestResponseMeta,

		WithDrawRewardResponseMeta,
		ReturnStakingMeta,

		IssuingETHResponseMeta,
		IssuingBSCResponseMeta,
		IssuingPRVERC20ResponseMeta,
		IssuingPRVBEP20ResponseMeta,
		IssuingPLGResponseMeta,
		IssuingFantomResponseMeta,
		IssuingAuroraResponseMeta,
		IssuingAvaxResponseMeta,
		IssuingResponseMeta,
		InitTokenResponseMeta,

		Pdexv3AddLiquidityResponseMeta,
		Pdexv3WithdrawLiquidityResponseMeta,
		Pdexv3TradeResponseMeta,
		Pdexv3AddOrderResponseMeta,
		Pdexv3WithdrawOrderResponseMeta,
		Pdexv3UserMintNftResponseMeta,
		Pdexv3MintNftResponseMeta,
		Pdexv3StakingResponseMeta,
		Pdexv3UnstakingResponseMeta,
		Pdexv3WithdrawLPFeeResponseMeta,
		Pdexv3WithdrawProtocolFeeResponseMeta,
		Pdexv3MintPDEXGenesisMeta,
		Pdexv3MintBlockRewardMeta,
		Pdexv3DistributeStakingRewardMeta,
		Pdexv3WithdrawStakingRewardResponseMeta,
		PortalV4ShieldingResponseMeta,
		PortalV4UnshieldingRequestMeta,
		PortalV4UnshieldingResponseMeta,
	}
	metaListNInfo = append(metaListNInfo, ListAndInfo{
		list: listTpNoInput,
		info: &MetaInfo{
			HasInput:  false,
			HasOutput: true,
			TxType: map[string]interface{}{
				common.TxCustomTokenPrivacyType: nil,
			},
		},
	})
	// listTpNoOutput := []int{}
	listTpNormal := []int{
		PDEContributionMeta,
		PDETradeRequestMeta,
		PDEPRVRequiredContributionRequestMeta,
		PDECrossPoolTradeRequestMeta,
		PortalRedeemRequestMeta,
		PortalRedeemFromLiquidationPoolMeta,
		PortalRedeemFromLiquidationPoolMetaV3,
		PortalRedeemRequestMetaV3,

		Pdexv3ModifyParamsMeta,
		Pdexv3AddLiquidityRequestMeta,
		Pdexv3WithdrawLiquidityRequestMeta,
		Pdexv3TradeRequestMeta,
		Pdexv3AddOrderRequestMeta,
		Pdexv3WithdrawOrderRequestMeta,
		Pdexv3UserMintNftRequestMeta,
		Pdexv3MintNftRequestMeta,
		Pdexv3StakingRequestMeta,
		Pdexv3UnstakingRequestMeta,
		Pdexv3WithdrawLPFeeRequestMeta,
		Pdexv3WithdrawProtocolFeeRequestMeta,
		Pdexv3WithdrawStakingRewardRequestMeta,

		PortalV4ShieldingResponseMeta,
		PortalV4UnshieldingRequestMeta,
		PortalV4UnshieldingResponseMeta,

		BurningRequestMeta,
		BurningRequestMetaV2,
		BurningPBSCRequestMeta,
		BurningForDepositToSCRequestMeta,
		BurningForDepositToSCRequestMetaV2,
		ContractingRequestMeta,
		BurningPBSCForDepositToSCRequestMeta,
		BurningPLGRequestMeta,
		BurningPLGForDepositToSCRequestMeta,
	}
	metaListNInfo = append(metaListNInfo, ListAndInfo{
		list: listTpNormal,
		info: &MetaInfo{
			HasInput:  true,
			HasOutput: true,
			TxType: map[string]interface{}{
				common.TxCustomTokenPrivacyType: nil,
			},
			MetaAction: NoAction,
		},
	})
	listNNoInput := []int{
		PDETradeResponseMeta,
		PDEWithdrawalResponseMeta,
		PDEContributionResponseMeta,
		PDECrossPoolTradeResponseMeta,
		PortalRequestWithdrawRewardResponseMeta,
		PortalRedeemFromLiquidationPoolResponseMeta,
		PortalRedeemFromLiquidationPoolResponseMetaV3,
		PDEFeeWithdrawalResponseMeta,
		PortalCustodianDepositResponseMeta,
		PortalCustodianWithdrawResponseMeta,
		PortalLiquidateCustodianResponseMeta,
		PortalCustodianTopupResponseMeta,
		PortalPortingResponseMeta,
		PortalCustodianTopupResponseMetaV2,
		PortalTopUpWaitingPortingResponseMeta,

		Pdexv3AddLiquidityResponseMeta,
		Pdexv3WithdrawLiquidityResponseMeta,
		Pdexv3TradeResponseMeta,
		Pdexv3AddOrderResponseMeta,
		Pdexv3WithdrawOrderResponseMeta,
		Pdexv3UserMintNftResponseMeta,
		Pdexv3MintNftResponseMeta,
		Pdexv3StakingResponseMeta,
		Pdexv3UnstakingResponseMeta,
		Pdexv3WithdrawLPFeeResponseMeta,
		Pdexv3WithdrawProtocolFeeResponseMeta,
		Pdexv3MintPDEXGenesisMeta,
		Pdexv3MintBlockRewardMeta,
		Pdexv3DistributeStakingRewardMeta,
		Pdexv3WithdrawStakingRewardResponseMeta,
		IssuingPRVERC20ResponseMeta,
	}
	metaListNInfo = append(metaListNInfo, ListAndInfo{
		list: listNNoInput,
		info: &MetaInfo{
			HasInput:  false,
			HasOutput: true,
			TxType: map[string]interface{}{
				common.TxNormalType: nil,
			},
			MetaAction: NoAction,
		},
	})
	// listNNoOutput := []int{}
	// listNNoInNoOut := []int{}
	listNNormal := []int{
		PDEContributionMeta,
		PDETradeRequestMeta,
		PDEPRVRequiredContributionRequestMeta,
		PDECrossPoolTradeRequestMeta,
		PDEWithdrawalRequestMeta,
		PDEFeeWithdrawalRequestMeta,
		PortalCustodianDepositMeta,
		PortalRequestPortingMeta,
		PortalUserRequestPTokenMeta,
		PortalExchangeRatesMeta,
		PortalRequestUnlockCollateralMeta,
		PortalCustodianWithdrawRequestMeta,
		PortalRequestWithdrawRewardMeta,
		PortalCustodianTopupMeta,
		PortalReqMatchingRedeemMeta,
		PortalCustodianTopupMetaV2,
		PortalCustodianDepositMetaV3,
		PortalCustodianWithdrawRequestMetaV3,
		PortalRequestUnlockCollateralMetaV3,
		PortalCustodianTopupMetaV3,
		PortalTopUpWaitingPortingRequestMetaV3,
		PortalRequestPortingMetaV3,
		PortalUnlockOverRateCollateralsMeta,
		RelayingBNBHeaderMeta,
		RelayingBTCHeaderMeta,
		PortalTopUpWaitingPortingRequestMeta,

		IssuingRequestMeta,
		IssuingETHRequestMeta,
		IssuingBSCRequestMeta,
		IssuingPRVERC20RequestMeta,
		IssuingPRVBEP20RequestMeta,
		ContractingRequestMeta,
		InitTokenRequestMeta,
		IssuingPLGRequestMeta,
		BurningPRVERC20RequestMeta,
		BurningPRVBEP20RequestMeta,

		ShardStakingMeta,
		BeaconStakingMeta,

		Pdexv3ModifyParamsMeta,
		Pdexv3AddLiquidityRequestMeta,
		Pdexv3WithdrawLiquidityRequestMeta,
		Pdexv3TradeRequestMeta,
		Pdexv3AddOrderRequestMeta,
		Pdexv3WithdrawOrderRequestMeta,
		Pdexv3UserMintNftRequestMeta,
		Pdexv3MintNftRequestMeta,
		Pdexv3StakingRequestMeta,
		Pdexv3UnstakingRequestMeta,
		Pdexv3WithdrawLPFeeRequestMeta,
		Pdexv3WithdrawProtocolFeeRequestMeta,
		Pdexv3WithdrawStakingRewardRequestMeta,

		PortalV4ShieldingRequestMeta,
		PortalV4FeeReplacementRequestMeta,
		PortalV4SubmitConfirmedTxMeta,
		PortalV4ConvertVaultRequestMeta,
	}
	metaListNInfo = append(metaListNInfo, ListAndInfo{
		list: listNNormal,
		info: &MetaInfo{
			HasInput:  true,
			HasOutput: true,
			TxType: map[string]interface{}{
				common.TxNormalType: nil,
			},
			MetaAction: NoAction,
		},
	})
	listNNoInNoOut := []int{
		WithDrawRewardRequestMeta,
		StopAutoStakingMeta,
		UnStakingMeta,
	}

	metaListNInfo = append(metaListNInfo, ListAndInfo{
		list: listNNoInNoOut,
		info: &MetaInfo{
			HasInput:  false,
			HasOutput: false,
			TxType: map[string]interface{}{
				common.TxNormalType: nil,
			},
			MetaAction: NoAction,
		},
	})

	listRSNoIn := []int{
		ReturnStakingMeta,
	}

	metaListNInfo = append(metaListNInfo, ListAndInfo{
		list: listRSNoIn,
		info: &MetaInfo{
			HasInput:  false,
			HasOutput: false,
			TxType: map[string]interface{}{
				common.TxReturnStakingType: nil,
			},
			MetaAction: NoAction,
		},
	})

	listSNoIn := []int{
		IssuingPRVERC20ResponseMeta,
		IssuingPRVBEP20ResponseMeta,
		PDETradeResponseMeta,
		PDEWithdrawalResponseMeta,
		PDEContributionResponseMeta,
		PDECrossPoolTradeResponseMeta,
		PDEFeeWithdrawalResponseMeta,
		PortalCustodianDepositResponseMeta,
		PortalCustodianWithdrawResponseMeta,
		PortalLiquidateCustodianResponseMeta,
		PortalRequestWithdrawRewardResponseMeta,
		PortalRedeemFromLiquidationPoolResponseMeta,
		PortalCustodianTopupResponseMeta,
		PortalPortingResponseMeta,
		PortalCustodianTopupResponseMetaV2,
		PortalRedeemFromLiquidationPoolResponseMetaV3,
		PortalTopUpWaitingPortingResponseMeta,

		WithDrawRewardResponseMeta,
		ReturnStakingMeta,
		Pdexv3AddLiquidityResponseMeta,
		Pdexv3WithdrawLiquidityResponseMeta,
		Pdexv3TradeResponseMeta,
		Pdexv3AddOrderResponseMeta,
		Pdexv3WithdrawOrderResponseMeta,
		Pdexv3StakingResponseMeta,
		Pdexv3UnstakingResponseMeta,
		Pdexv3WithdrawLPFeeResponseMeta,
		Pdexv3WithdrawProtocolFeeResponseMeta,
		Pdexv3MintBlockRewardMeta,
		Pdexv3DistributeStakingRewardMeta,
		Pdexv3WithdrawStakingRewardResponseMeta,
	}

	metaListNInfo = append(metaListNInfo, ListAndInfo{
		list: listSNoIn,
		info: &MetaInfo{
			HasInput:  false,
			HasOutput: false,
			TxType: map[string]interface{}{
				common.TxRewardType: nil,
			},
			MetaAction: NoAction,
		},
	})

	listRequestBeaconMintTxs := []int{
		PDETradeRequestMeta,
		// PDETradeResponseMeta,
		IssuingRequestMeta,
		IssuingResponseMeta,
		IssuingETHRequestMeta,
		IssuingPRVBEP20RequestMeta,
		IssuingPRVERC20RequestMeta,
		IssuingBSCRequestMeta,
		IssuingPLGRequestMeta,
		IssuingETHResponseMeta,
		IssuingBSCResponseMeta,
		IssuingPRVERC20ResponseMeta,
		IssuingPRVBEP20ResponseMeta,
		IssuingPLGResponseMeta,
		IssuingFantomResponseMeta,
		IssuingAuroraResponseMeta,
		IssuingAvaxResponseMeta,
		PDEWithdrawalRequestMeta,
		PDEWithdrawalResponseMeta,
		PDEPRVRequiredContributionRequestMeta,
		PDEContributionResponseMeta,
		PDECrossPoolTradeRequestMeta,
		PDECrossPoolTradeResponseMeta,
		PDEFeeWithdrawalRequestMeta,
		PDEFeeWithdrawalResponseMeta,
		PortalCustodianDepositMeta,
		PortalCustodianDepositResponseMeta,
		PortalRequestPortingMeta,
		PortalPortingResponseMeta,
		PortalUserRequestPTokenMeta,
		PortalUserRequestPTokenResponseMeta,
		PortalRedeemRequestMeta,
		PortalRedeemRequestResponseMeta,
		PortalCustodianWithdrawRequestMeta,
		PortalCustodianWithdrawResponseMeta,
		PortalLiquidateCustodianMeta,
		PortalLiquidateCustodianResponseMeta,
		PortalRequestWithdrawRewardMeta,
		PortalRequestWithdrawRewardResponseMeta,
		PortalRedeemFromLiquidationPoolMeta,
		PortalRedeemFromLiquidationPoolResponseMeta,
		PortalCustodianTopupMeta,
		PortalCustodianTopupResponseMeta,
		PortalCustodianTopupMetaV2,
		PortalCustodianTopupResponseMetaV2,
		PortalLiquidateCustodianMetaV3,
		PortalRedeemFromLiquidationPoolMetaV3,
		PortalRedeemFromLiquidationPoolResponseMetaV3,
		PortalRequestPortingMetaV3,
		PortalRedeemRequestMetaV3,
		PortalTopUpWaitingPortingRequestMeta,
		PortalTopUpWaitingPortingResponseMeta,
	}

	metaListNInfo = append(metaListNInfo, ListAndInfo{
		list: listRequestBeaconMintTxs,
		info: &MetaInfo{
			TxType:     map[string]interface{}{},
			MetaAction: MetaRequestBeaconMintTxs,
		},
	})

	listRequestShardMint := []int{
		WithDrawRewardRequestMeta,
	}

	metaListNInfo = append(metaListNInfo, ListAndInfo{
		list: listRequestShardMint,
		info: &MetaInfo{
			TxType:     map[string]interface{}{},
			MetaAction: MetaRequestShardMintTxs,
		},
	})
	metaInfoMap = map[int]*MetaInfo{}
	for _, value := range metaListNInfo {
		for _, metaType := range value.list {
			if info, ok := metaInfoMap[metaType]; ok {
				for k := range value.info.TxType {
					info.TxType[k] = nil
				}
				if (info.MetaAction == NoAction) && (value.info.MetaAction != NoAction) {
					info.MetaAction = value.info.MetaAction
				}
			} else {
				metaInfoMap[metaType] = &MetaInfo{
					HasInput:   value.info.HasInput,
					HasOutput:  value.info.HasOutput,
					MetaAction: value.info.MetaAction,
					TxType:     map[string]interface{}{},
				}
				for k := range value.info.TxType {
					metaInfoMap[metaType].TxType[k] = nil
				}
			}
		}
	}
}

func init() {
	buildMetaInfo()
	setLimitMetadataInBlock()
}

func NoInputNoOutput(metaType int) bool {
	if info, ok := metaInfoMap[metaType]; ok {
		return !(info.HasInput || info.HasOutput)
	}
	return false
}

func HasInputNoOutput(metaType int) bool {
	if info, ok := metaInfoMap[metaType]; ok {
		return info.HasInput && !info.HasOutput
	}
	return false
}

func NoInputHasOutput(metaType int) bool {
	if info, ok := metaInfoMap[metaType]; ok {
		return !info.HasInput && info.HasOutput
	}
	return false
}

func IsAvailableMetaInTxType(metaType int, txType string) bool {
	if info, ok := metaInfoMap[metaType]; ok {
		_, ok := info.TxType[txType]
		return ok
	}
	return false
}

func GetMetaAction(metaType int) int {
	if info, ok := metaInfoMap[metaType]; ok {
		return info.MetaAction
	}
	return NoAction
}

func GetLimitOfMeta(metaType int) int {
	if info, ok := metaInfoMap[metaType]; ok {
		if limit, ok := limitOfMetaAct[info.MetaAction]; ok {
			return limit
		}
	}
	return -1
}

// TODO: add more meta data types
var portalConfirmedMetas = []string{
	strconv.Itoa(PortalCustodianWithdrawConfirmMetaV3),
	strconv.Itoa(PortalRedeemFromLiquidationPoolConfirmMetaV3),
	strconv.Itoa(PortalLiquidateRunAwayCustodianConfirmMetaV3),
}

func HasPortalInstructions(instructions [][]string) bool {
	for _, inst := range instructions {
		for _, meta := range portalConfirmedMetas {
			if len(inst) > 0 && inst[0] == meta {
				return true
			}
		}
	}
	return false
}

// Validate portal external addresses for collateral tokens (ETH/ERC20)
func ValidatePortalExternalAddress(chainName string, tokenID string, address string) (bool, error) {
	switch chainName {
	case common.ETHChainName:
		return ec.IsHexAddress(address), nil
	}
	return true, nil
}

func IsPortalMetaTypeV3(metaType int) bool {
	res, _ := common.SliceExists(portalMetaTypesV3, metaType)
	return res
}

//Checks if a string payment address is supported by the underlying transaction.
//
//TODO: try another approach since the function itself is too complicated.
func AssertPaymentAddressAndTxVersion(paymentAddress interface{}, version int8) (privacy.PaymentAddress, error) {
	var addr privacy.PaymentAddress
	var ok bool
	//try to parse the payment address
	if addr, ok = paymentAddress.(privacy.PaymentAddress); !ok {
		//try the pointer
		if tmpAddr, ok := paymentAddress.(*privacy.PaymentAddress); !ok {
			//try the string one
			addrStr, ok := paymentAddress.(string)
			if !ok {
				return privacy.PaymentAddress{}, fmt.Errorf("cannot parse payment address - %v: Not a payment address or string address (txversion %v)", paymentAddress, version)
			}
			keyWallet, err := wallet.Base58CheckDeserialize(addrStr)
			if err != nil {
				return privacy.PaymentAddress{}, err
			}
			if len(keyWallet.KeySet.PrivateKey) > 0 {
				return privacy.PaymentAddress{}, fmt.Errorf("cannot parse payment address - %v: This is a private key", paymentAddress)
			}
			addr = keyWallet.KeySet.PaymentAddress
		} else {
			addr = *tmpAddr
		}
	}

	//Always check public spend and public view keys
	if addr.GetPublicSpend() == nil || addr.GetPublicView() == nil {
		return privacy.PaymentAddress{}, errors.New("PublicSpend or PublicView not found")
	}

	//If tx is in version 1, PublicOTAKey must be nil
	if version == 1 {
		if addr.GetOTAPublicKey() != nil {
			return privacy.PaymentAddress{}, errors.New("PublicOTAKey must be nil")
		}
	}

	//If tx is in version 2, PublicOTAKey must not be nil
	if version == 2 {
		if addr.GetOTAPublicKey() == nil {
			return privacy.PaymentAddress{}, errors.New("PublicOTAKey not found")
		}
	}

	return addr, nil
}

func IsPortalRelayingMetaType(metaType int) bool {
	res, _ := common.SliceExists(portalRelayingMetaTypes, metaType)
	return res
}

func IsPortalMetaTypeV4(metaType int) bool {
	res, _ := common.SliceExists(portalV4MetaTypes, metaType)
	return res
}

//genTokenID generates a (deterministically) random tokenID for the request transaction.
//From now on, users cannot generate their own tokenID.
//The generated tokenID is calculated as the hash of the following components:
//	- The Tx hash
//	- The shardID at which the request is sent
func GenTokenIDFromRequest(txHash string, shardID byte) *common.Hash {
	record := txHash + strconv.FormatUint(uint64(shardID), 10)

	tokenID := common.HashH([]byte(record))
	return &tokenID
}

type OTADeclaration struct {
	PublicKey [32]byte
	TokenID   common.Hash
}

func CheckIncognitoAddress(address, txRandom string) (bool, error, int) {
	version := 0
	if len(txRandom) > 0 {
		version = 2
		_, _, err := coin.ParseOTAInfoFromString(address, txRandom)
		if err != nil {
			return false, err, version
		}
	} else {
		version = 1
		_, err := AssertPaymentAddressAndTxVersion(address, 1)
		return err == nil, err, version
	}
	return true, nil, version
}

func IsPdexv3Type(metadataType int) bool {
	switch metadataType {
	case Pdexv3ModifyParamsMeta:
		return true
	case Pdexv3UserMintNftRequestMeta:
		return true
	case Pdexv3UserMintNftResponseMeta:
		return true
	case Pdexv3MintNftRequestMeta:
		return true
	case Pdexv3MintNftResponseMeta:
		return true
	case Pdexv3AddLiquidityRequestMeta:
		return true
	case Pdexv3AddLiquidityResponseMeta:
		return true
	case Pdexv3TradeRequestMeta:
		return true
	case Pdexv3TradeResponseMeta:
		return true
	case Pdexv3AddOrderRequestMeta:
		return true
	case Pdexv3AddOrderResponseMeta:
		return true
	case Pdexv3WithdrawOrderRequestMeta:
		return true
	case Pdexv3WithdrawOrderResponseMeta:
		return true
	case Pdexv3WithdrawLiquidityRequestMeta:
		return true
	case Pdexv3WithdrawLiquidityResponseMeta:
		return true
	case Pdexv3WithdrawLPFeeRequestMeta:
		return true
	case Pdexv3WithdrawLPFeeResponseMeta:
		return true
	case Pdexv3WithdrawProtocolFeeRequestMeta:
		return true
	case Pdexv3WithdrawProtocolFeeResponseMeta:
		return true
	case Pdexv3MintPDEXGenesisMeta:
		return true
	case Pdexv3MintBlockRewardMeta:
		return true
	case Pdexv3StakingRequestMeta:
		return true
	case Pdexv3StakingResponseMeta:
		return true
	case Pdexv3UnstakingRequestMeta:
		return true
	case Pdexv3UnstakingResponseMeta:
		return true
	case Pdexv3DistributeStakingRewardMeta:
		return true
	case Pdexv3WithdrawStakingRewardRequestMeta:
		return true
	case Pdexv3WithdrawStakingRewardResponseMeta:
		return true
	case Pdexv3DistributeMiningOrderRewardMeta:
		return true
	default:
		return false
	}
}

func IsPDETx(metadata Metadata) bool {
	if metadata != nil {
		return IsPDEType(metadata.GetType())
	}
	return false
}

func IsPDEType(metadataType int) bool {
	switch metadataType {
	case PDEContributionMeta:
		return true
	case PDETradeRequestMeta:
		return true
	case PDETradeResponseMeta:
		return true
	case PDEWithdrawalRequestMeta:
		return true
	case PDEWithdrawalResponseMeta:
		return true
	case PDEContributionResponseMeta:
		return true
	case PDEPRVRequiredContributionRequestMeta:
		return true
	case PDECrossPoolTradeRequestMeta:
		return true
	case PDECrossPoolTradeResponseMeta:
		return true
	case PDEFeeWithdrawalRequestMeta:
		return true
	case PDEFeeWithdrawalResponseMeta:
		return true
	case PDETradingFeesDistributionMeta:
		return true
	default:
		return false
	}
}

func IsPdexv3Tx(metadata Metadata) bool {
	if metadata != nil {
		return IsPdexv3Type(metadata.GetType())
	}
	return false
}

// NOTE: append new bridge unshield metadata type
func IsBridgeUnshieldMetaType(metadataType int) bool {
	switch metadataType {
	case ContractingRequestMeta:
		return true
	case BurningRequestMeta:
		return true
	case BurningRequestMetaV2:
		return true
	case BurningPBSCRequestMeta:
		return true
	case BurningPBSCForDepositToSCRequestMeta:
		return true
	case BurningForDepositToSCRequestMeta:
		return true
	case BurningForDepositToSCRequestMetaV2:
		return true
	case BurningPRVERC20RequestMeta:
		return true
	case BurningPRVBEP20RequestMeta:
		return true
	case BurningPLGRequestMeta:
		return true
	case BurningPLGForDepositToSCRequestMeta:
		return true
	case BurningFantomRequestMeta:
		return true
	case BurningFantomForDepositToSCRequestMeta:
		return true
	case BurningAuroraRequestMeta:
		return true
	case BurningAvaxRequestMeta:
		return true
	default:
		return false
	}
}

// NOTE: append new bridge agg unshield metadata type
func IsBridgeAggUnshieldMetaType(metadataType int) bool {
	switch metadataType {
	case BurningUnifiedTokenRequestMeta:
		return true
	default:
		return false
	}
}
