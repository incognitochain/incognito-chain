package metadata

import (
	"encoding/json"
	"strconv"

	ec "github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/incognito-chain/common"

	"github.com/pkg/errors"
)

func calculateSize(meta Metadata) uint64 {
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return 0
	}
	return uint64(len(metaBytes))
}

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
	switch int(mtTemp["Type"].(float64)) {
	case IssuingRequestMeta:
		md = &IssuingRequest{}
	case IssuingResponseMeta:
		md = &IssuingResponse{}
	case ContractingRequestMeta:
		md = &ContractingRequest{}
	case IssuingETHRequestMeta:
		md = &IssuingETHRequest{}
	case IssuingETHResponseMeta:
		md = &IssuingETHResponse{}
	case BeaconSalaryResponseMeta:
		md = &BeaconBlockSalaryRes{}
	case BurningRequestMeta:
		md = &BurningRequest{}
	case BurningRequestMetaV2:
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
	strconv.Itoa(BeaconSwapConfirmMeta),
	strconv.Itoa(BridgeSwapConfirmMeta),
	strconv.Itoa(BurningConfirmMeta),
	strconv.Itoa(BurningConfirmForDepositToSCMeta),
	strconv.Itoa(BurningConfirmMetaV2),
	strconv.Itoa(BurningConfirmForDepositToSCMetaV2),
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
		IssuingETHResponseMeta,
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

func IsPortalRelayingMetaType(metaType int) bool {
	res, _ := common.SliceExists(portalRelayingMetaTypes, metaType)
	return res
}
