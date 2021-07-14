package common

import (
	"fmt"
	"strconv"

	ec "github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

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
		IssuingResponseMeta,
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

		BurningRequestMeta,
		BurningRequestMetaV2,
		BurningPBSCRequestMeta,
		BurningForDepositToSCRequestMeta,
		BurningForDepositToSCRequestMetaV2,
		ContractingRequestMeta,
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
		ContractingRequestMeta,

		ShardStakingMeta,
		BeaconStakingMeta,
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
		IssuingBSCRequestMeta,
		IssuingETHResponseMeta,
		IssuingBSCResponseMeta,
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
