package metadata

import (
	"github.com/incognitochain/incognito-chain/common"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

// export interfaces
type Metadata = metadataCommon.Metadata
type MetadataBase = metadataCommon.MetadataBase
type MetadataBaseWithSignature = metadataCommon.MetadataBaseWithSignature
type Transaction = metadataCommon.Transaction
type ChainRetriever = metadataCommon.ChainRetriever
type ShardViewRetriever = metadataCommon.ShardViewRetriever
type BeaconViewRetriever = metadataCommon.BeaconViewRetriever
type MempoolRetriever = metadataCommon.MempoolRetriever
type ValidationEnviroment = metadataCommon.ValidationEnviroment
type TxDesc = metadataCommon.TxDesc

// export structs
type OTADeclaration = metadataCommon.OTADeclaration
type MintData = metadataCommon.MintData
type AccumulatedValues = metadataCommon.AccumulatedValues

// wrap logger
type MetadataLogger struct {
	log common.Logger
}

func (ml *MetadataLogger) Init(inst common.Logger) {
	ml.log = inst
	metadataCommon.Logger.Init(inst)
}

var Logger = MetadataLogger{}

var AcceptedWithdrawRewardRequestVersion = metadataCommon.AcceptedWithdrawRewardRequestVersion

// export functions
var AssertPaymentAddressAndTxVersion = metadataCommon.AssertPaymentAddressAndTxVersion
var GenTokenIDFromRequest = metadataCommon.GenTokenIDFromRequest
var NewMetadataBase = metadataCommon.NewMetadataBase
var NewMetadataBaseWithSignature = metadataCommon.NewMetadataBaseWithSignature
var ValidatePortalExternalAddress = metadataCommon.ValidatePortalExternalAddress
var NewMetadataTxError = metadataCommon.NewMetadataTxError
var IsAvailableMetaInTxType = metadataCommon.IsAvailableMetaInTxType
var NoInputNoOutput = metadataCommon.NoInputNoOutput
var NoInputHasOutput = metadataCommon.NoInputHasOutput
var IsPortalRelayingMetaType = metadataCommon.IsPortalRelayingMetaType
var IsPortalMetaTypeV3 = metadataCommon.IsPortalMetaTypeV3
var GetMetaAction = metadataCommon.GetMetaAction
var IsPDEType = metadataCommon.IsPDEType
var IsPdexv3Type = metadataCommon.IsPdexv3Type
var GetLimitOfMeta = metadataCommon.GetLimitOfMeta
var IsPDETx = metadataCommon.IsPDETx
var IsPdexv3Tx = metadataCommon.IsPdexv3Tx
var ConvertPrivacyTokenToNativeToken = metadataCommon.ConvertPrivacyTokenToNativeToken
var ConvertNativeTokenToPrivacyToken = metadataCommon.ConvertNativeTokenToPrivacyToken
var HasBridgeInstructions = metadataCommon.HasBridgeInstructions
var HasPortalInstructions = metadataCommon.HasPortalInstructions

var calculateSize = metadataCommon.CalculateSize

// export package constants
const (
	InvalidMeta                  = metadataCommon.InvalidMeta
	IssuingRequestMeta           = metadataCommon.IssuingRequestMeta
	IssuingResponseMeta          = metadataCommon.IssuingResponseMeta
	ContractingRequestMeta       = metadataCommon.ContractingRequestMeta
	IssuingETHRequestMeta        = metadataCommon.IssuingETHRequestMeta
	IssuingETHResponseMeta       = metadataCommon.IssuingETHResponseMeta
	ShardBlockReward             = metadataCommon.ShardBlockReward
	ShardBlockSalaryResponseMeta = metadataCommon.ShardBlockSalaryResponseMeta
	BeaconRewardRequestMeta      = metadataCommon.BeaconRewardRequestMeta
	BeaconSalaryResponseMeta     = metadataCommon.BeaconSalaryResponseMeta
	ReturnStakingMeta            = metadataCommon.ReturnStakingMeta
	MintDelegationRewardMeta     = metadataCommon.MintDelegationRewardMeta
	ReturnBeaconStakingMeta      = metadataCommon.ReturnBeaconStakingMeta
	IncDAORewardRequestMeta      = metadataCommon.IncDAORewardRequestMeta
	WithDrawRewardRequestMeta    = metadataCommon.WithDrawRewardRequestMeta
	WithDrawRewardResponseMeta   = metadataCommon.WithDrawRewardResponseMeta
	// staking
	ShardStakingMeta    = metadataCommon.ShardStakingMeta
	StopAutoStakingMeta = metadataCommon.StopAutoStakingMeta
	BeaconStakingMeta   = metadataCommon.BeaconStakingMeta
	UnStakingMeta       = metadataCommon.UnStakingMeta
	AddStakingMeta      = metadataCommon.AddStakingMeta
	ReDelegateMeta      = metadataCommon.ReDelegateMeta
	// Incognito -> Ethereum bridge
	BeaconSwapConfirmMeta = metadataCommon.BeaconSwapConfirmMeta
	BridgeSwapConfirmMeta = metadataCommon.BridgeSwapConfirmMeta
	BurningRequestMeta    = metadataCommon.BurningRequestMeta
	BurningRequestMetaV2  = metadataCommon.BurningRequestMetaV2
	BurningConfirmMeta    = metadataCommon.BurningConfirmMeta
	BurningConfirmMetaV2  = metadataCommon.BurningConfirmMetaV2
	// pde
	PDEContributionMeta                   = metadataCommon.PDEContributionMeta
	PDETradeRequestMeta                   = metadataCommon.PDETradeRequestMeta
	PDETradeResponseMeta                  = metadataCommon.PDETradeResponseMeta
	PDEWithdrawalRequestMeta              = metadataCommon.PDEWithdrawalRequestMeta
	PDEWithdrawalResponseMeta             = metadataCommon.PDEWithdrawalResponseMeta
	PDEContributionResponseMeta           = metadataCommon.PDEContributionResponseMeta
	PDEPRVRequiredContributionRequestMeta = metadataCommon.PDEPRVRequiredContributionRequestMeta
	PDECrossPoolTradeRequestMeta          = metadataCommon.PDECrossPoolTradeRequestMeta
	PDECrossPoolTradeResponseMeta         = metadataCommon.PDECrossPoolTradeResponseMeta
	PDEFeeWithdrawalRequestMeta           = metadataCommon.PDEFeeWithdrawalRequestMeta
	PDEFeeWithdrawalResponseMeta          = metadataCommon.PDEFeeWithdrawalResponseMeta
	PDETradingFeesDistributionMeta        = metadataCommon.PDETradingFeesDistributionMeta
	// pDEX v3
	Pdexv3TradeRequestMeta          = metadataCommon.Pdexv3TradeRequestMeta
	Pdexv3TradeResponseMeta         = metadataCommon.Pdexv3TradeResponseMeta
	Pdexv3AddOrderRequestMeta       = metadataCommon.Pdexv3AddOrderRequestMeta
	Pdexv3AddOrderResponseMeta      = metadataCommon.Pdexv3AddOrderResponseMeta
	Pdexv3WithdrawOrderRequestMeta  = metadataCommon.Pdexv3WithdrawOrderRequestMeta
	Pdexv3WithdrawOrderResponseMeta = metadataCommon.Pdexv3WithdrawOrderResponseMeta
	// portal
	PortalCustodianDepositMeta                  = metadataCommon.PortalCustodianDepositMeta
	PortalRequestPortingMeta                    = metadataCommon.PortalRequestPortingMeta
	PortalUserRequestPTokenMeta                 = metadataCommon.PortalUserRequestPTokenMeta
	PortalCustodianDepositResponseMeta          = metadataCommon.PortalCustodianDepositResponseMeta
	PortalUserRequestPTokenResponseMeta         = metadataCommon.PortalUserRequestPTokenResponseMeta
	PortalExchangeRatesMeta                     = metadataCommon.PortalExchangeRatesMeta
	PortalRedeemRequestMeta                     = metadataCommon.PortalRedeemRequestMeta
	PortalRedeemRequestResponseMeta             = metadataCommon.PortalRedeemRequestResponseMeta
	PortalRequestUnlockCollateralMeta           = metadataCommon.PortalRequestUnlockCollateralMeta
	PortalCustodianWithdrawRequestMeta          = metadataCommon.PortalCustodianWithdrawRequestMeta
	PortalCustodianWithdrawResponseMeta         = metadataCommon.PortalCustodianWithdrawResponseMeta
	PortalLiquidateCustodianMeta                = metadataCommon.PortalLiquidateCustodianMeta
	PortalLiquidateCustodianResponseMeta        = metadataCommon.PortalLiquidateCustodianResponseMeta
	PortalLiquidateTPExchangeRatesMeta          = metadataCommon.PortalLiquidateTPExchangeRatesMeta
	PortalExpiredWaitingPortingReqMeta          = metadataCommon.PortalExpiredWaitingPortingReqMeta
	PortalRewardMeta                            = metadataCommon.PortalRewardMeta
	PortalRequestWithdrawRewardMeta             = metadataCommon.PortalRequestWithdrawRewardMeta
	PortalRequestWithdrawRewardResponseMeta     = metadataCommon.PortalRequestWithdrawRewardResponseMeta
	PortalRedeemFromLiquidationPoolMeta         = metadataCommon.PortalRedeemFromLiquidationPoolMeta
	PortalRedeemFromLiquidationPoolResponseMeta = metadataCommon.PortalRedeemFromLiquidationPoolResponseMeta
	PortalCustodianTopupMeta                    = metadataCommon.PortalCustodianTopupMeta
	PortalCustodianTopupResponseMeta            = metadataCommon.PortalCustodianTopupResponseMeta
	PortalTotalRewardCustodianMeta              = metadataCommon.PortalTotalRewardCustodianMeta
	PortalPortingResponseMeta                   = metadataCommon.PortalPortingResponseMeta
	PortalReqMatchingRedeemMeta                 = metadataCommon.PortalReqMatchingRedeemMeta
	PortalPickMoreCustodianForRedeemMeta        = metadataCommon.PortalPickMoreCustodianForRedeemMeta
	PortalCustodianTopupMetaV2                  = metadataCommon.PortalCustodianTopupMetaV2
	PortalCustodianTopupResponseMetaV2          = metadataCommon.PortalCustodianTopupResponseMetaV2
	// Portal v3
	PortalCustodianDepositMetaV3                  = metadataCommon.PortalCustodianDepositMetaV3
	PortalCustodianWithdrawRequestMetaV3          = metadataCommon.PortalCustodianWithdrawRequestMetaV3
	PortalRewardMetaV3                            = metadataCommon.PortalRewardMetaV3
	PortalRequestUnlockCollateralMetaV3           = metadataCommon.PortalRequestUnlockCollateralMetaV3
	PortalLiquidateCustodianMetaV3                = metadataCommon.PortalLiquidateCustodianMetaV3
	PortalLiquidateByRatesMetaV3                  = metadataCommon.PortalLiquidateByRatesMetaV3
	PortalRedeemFromLiquidationPoolMetaV3         = metadataCommon.PortalRedeemFromLiquidationPoolMetaV3
	PortalRedeemFromLiquidationPoolResponseMetaV3 = metadataCommon.PortalRedeemFromLiquidationPoolResponseMetaV3
	PortalCustodianTopupMetaV3                    = metadataCommon.PortalCustodianTopupMetaV3
	PortalTopUpWaitingPortingRequestMetaV3        = metadataCommon.PortalTopUpWaitingPortingRequestMetaV3
	PortalRequestPortingMetaV3                    = metadataCommon.PortalRequestPortingMetaV3
	PortalRedeemRequestMetaV3                     = metadataCommon.PortalRedeemRequestMetaV3
	PortalUnlockOverRateCollateralsMeta           = metadataCommon.PortalUnlockOverRateCollateralsMeta
	// Incognito => Ethereum's SC for portal
	PortalCustodianWithdrawConfirmMetaV3         = metadataCommon.PortalCustodianWithdrawConfirmMetaV3
	PortalRedeemFromLiquidationPoolConfirmMetaV3 = metadataCommon.PortalRedeemFromLiquidationPoolConfirmMetaV3
	PortalLiquidateRunAwayCustodianConfirmMetaV3 = metadataCommon.PortalLiquidateRunAwayCustodianConfirmMetaV3
	// Note: don't use this metadata type for others
	PortalResetPortalDBMeta = metadataCommon.PortalResetPortalDBMeta
	// relaying
	RelayingBNBHeaderMeta                 = metadataCommon.RelayingBNBHeaderMeta
	RelayingBTCHeaderMeta                 = metadataCommon.RelayingBTCHeaderMeta
	PortalTopUpWaitingPortingRequestMeta  = metadataCommon.PortalTopUpWaitingPortingRequestMeta
	PortalTopUpWaitingPortingResponseMeta = metadataCommon.PortalTopUpWaitingPortingResponseMeta
	// incognito mode for smart contract
	BurningForDepositToSCRequestMeta   = metadataCommon.BurningForDepositToSCRequestMeta
	BurningForDepositToSCRequestMetaV2 = metadataCommon.BurningForDepositToSCRequestMetaV2
	BurningConfirmForDepositToSCMeta   = metadataCommon.BurningConfirmForDepositToSCMeta
	BurningConfirmForDepositToSCMetaV2 = metadataCommon.BurningConfirmForDepositToSCMetaV2
	InitTokenRequestMeta               = metadataCommon.InitTokenRequestMeta
	InitTokenResponseMeta              = metadataCommon.InitTokenResponseMeta
	// incognito mode for bsc
	IssuingBSCRequestMeta                = metadataCommon.IssuingBSCRequestMeta
	IssuingBSCResponseMeta               = metadataCommon.IssuingBSCResponseMeta
	BurningPBSCRequestMeta               = metadataCommon.BurningPBSCRequestMeta
	BurningBSCConfirmMeta                = metadataCommon.BurningBSCConfirmMeta
	AllShards                            = metadataCommon.AllShards
	BeaconOnly                           = metadataCommon.BeaconOnly
	StopAutoStakingAmount                = metadataCommon.StopAutoStakingAmount
	ReDelegateFee                        = metadataCommon.ReDelegateFee
	EVMConfirmationBlocks                = metadataCommon.EVMConfirmationBlocks
	PLGConfirmationBlocks                = metadataCommon.PLGConfirmationBlocks
	FantomConfirmationBlocks             = metadataCommon.FantomConfirmationBlocks
	NoAction                             = metadataCommon.NoAction
	MetaRequestBeaconMintTxs             = metadataCommon.MetaRequestBeaconMintTxs
	MetaRequestShardMintTxs              = metadataCommon.MetaRequestShardMintTxs
	BurningPBSCForDepositToSCRequestMeta = metadataCommon.BurningPBSCForDepositToSCRequestMeta
	BurningPBSCConfirmForDepositToSCMeta = metadataCommon.BurningPBSCConfirmForDepositToSCMeta

	IssuingPRVERC20RequestMeta  = metadataCommon.IssuingPRVERC20RequestMeta
	IssuingPRVERC20ResponseMeta = metadataCommon.IssuingPRVERC20ResponseMeta
	IssuingPRVBEP20RequestMeta  = metadataCommon.IssuingPRVBEP20RequestMeta
	IssuingPRVBEP20ResponseMeta = metadataCommon.IssuingPRVBEP20ResponseMeta
	BurningPRVERC20RequestMeta  = metadataCommon.BurningPRVERC20RequestMeta
	BurningPRVERC20ConfirmMeta  = metadataCommon.BurningPRVERC20ConfirmMeta
	BurningPRVBEP20RequestMeta  = metadataCommon.BurningPRVBEP20RequestMeta
	BurningPRVBEP20ConfirmMeta  = metadataCommon.BurningPRVBEP20ConfirmMeta

	IssuingPLGRequestMeta  = metadataCommon.IssuingPLGRequestMeta
	IssuingPLGResponseMeta = metadataCommon.IssuingPLGResponseMeta
	BurningPLGRequestMeta  = metadataCommon.BurningPLGRequestMeta
	BurningPLGConfirmMeta  = metadataCommon.BurningPLGConfirmMeta

	BurningPLGForDepositToSCRequestMeta = metadataCommon.BurningPLGForDepositToSCRequestMeta
	BurningPLGConfirmForDepositToSCMeta = metadataCommon.BurningPLGConfirmForDepositToSCMeta

	IssuingFantomRequestMeta  = metadataCommon.IssuingFantomRequestMeta
	IssuingFantomResponseMeta = metadataCommon.IssuingFantomResponseMeta
	BurningFantomRequestMeta  = metadataCommon.BurningFantomRequestMeta
	BurningFantomConfirmMeta  = metadataCommon.BurningFantomConfirmMeta

	BurningFantomForDepositToSCRequestMeta = metadataCommon.BurningFantomForDepositToSCRequestMeta
	BurningFantomConfirmForDepositToSCMeta = metadataCommon.BurningFantomConfirmForDepositToSCMeta

	BurnForCallRequestMeta      = metadataCommon.BurnForCallRequestMeta
	BurnForCallResponseMeta     = metadataCommon.BurnForCallResponseMeta
	BurnForCallConfirmMeta      = metadataCommon.BurnForCallConfirmMeta
	IssuingReshieldResponseMeta = metadataCommon.IssuingReshieldResponseMeta

	IssuingAuroraRequestMeta  = metadataCommon.IssuingAuroraRequestMeta
	IssuingAuroraResponseMeta = metadataCommon.IssuingAuroraResponseMeta
	BurningAuroraRequestMeta  = metadataCommon.BurningAuroraRequestMeta
	BurningAuroraConfirmMeta  = metadataCommon.BurningAuroraConfirmMeta

	IssuingAvaxRequestMeta  = metadataCommon.IssuingAvaxRequestMeta
	IssuingAvaxResponseMeta = metadataCommon.IssuingAvaxResponseMeta
	BurningAvaxRequestMeta  = metadataCommon.BurningAvaxRequestMeta
	BurningAvaxConfirmMeta  = metadataCommon.BurningAvaxConfirmMeta

	BurningAuroraForDepositToSCRequestMeta = metadataCommon.BurningAuroraForDepositToSCRequestMeta
	BurningAuroraConfirmForDepositToSCMeta = metadataCommon.BurningAuroraConfirmForDepositToSCMeta

	BurningAvaxForDepositToSCRequestMeta = metadataCommon.BurningAvaxForDepositToSCRequestMeta
	BurningAvaxConfirmForDepositToSCMeta = metadataCommon.BurningAvaxConfirmForDepositToSCMeta

	IssuingNearRequestMeta  = metadataCommon.IssuingNearRequestMeta
	IssuingNearResponseMeta = metadataCommon.IssuingNearResponseMeta
	BurningNearRequestMeta  = metadataCommon.BurningNearRequestMeta
	BurningNearConfirmMeta  = metadataCommon.BurningNearConfirmMeta

	BurningPRVRequestMeta        = metadataCommon.BurningPRVRequestMeta
	BurningPRVRequestConfirmMeta = metadataCommon.BurningPRVRequestConfirmMeta
)

// export error codes
const (
	UnexpectedError                                            = metadataCommon.UnexpectedError
	IssuingEvmRequestDecodeInstructionError                    = metadataCommon.IssuingEvmRequestDecodeInstructionError
	IssuingEvmRequestUnmarshalJsonError                        = metadataCommon.IssuingEvmRequestUnmarshalJsonError
	IssuingEvmRequestNewIssuingEVMRequestFromMapError          = metadataCommon.IssuingEvmRequestNewIssuingEVMRequestFromMapError
	IssuingEvmRequestValidateTxWithBlockChainError             = metadataCommon.IssuingEvmRequestValidateTxWithBlockChainError
	IssuingEvmRequestValidateSanityDataError                   = metadataCommon.IssuingEvmRequestValidateSanityDataError
	IssuingEvmRequestBuildReqActionsError                      = metadataCommon.IssuingEvmRequestBuildReqActionsError
	IssuingEvmRequestVerifyProofAndParseReceipt                = metadataCommon.IssuingEvmRequestVerifyProofAndParseReceipt
	IssuingRequestDecodeInstructionError                       = metadataCommon.IssuingRequestDecodeInstructionError
	IssuingRequestUnmarshalJsonError                           = metadataCommon.IssuingRequestUnmarshalJsonError
	IssuingRequestNewIssuingRequestFromMapEror                 = metadataCommon.IssuingRequestNewIssuingRequestFromMapEror
	IssuingRequestValidateTxWithBlockChainError                = metadataCommon.IssuingRequestValidateTxWithBlockChainError
	IssuingRequestValidateSanityDataError                      = metadataCommon.IssuingRequestValidateSanityDataError
	IssuingRequestBuildReqActionsError                         = metadataCommon.IssuingRequestBuildReqActionsError
	IssuingRequestVerifyProofAndParseReceipt                   = metadataCommon.IssuingRequestVerifyProofAndParseReceipt
	BeaconBlockRewardNewBeaconBlockRewardInfoFromStrError      = metadataCommon.BeaconBlockRewardNewBeaconBlockRewardInfoFromStrError
	BeaconBlockRewardBuildInstructionForBeaconBlockRewardError = metadataCommon.BeaconBlockRewardBuildInstructionForBeaconBlockRewardError
	StopAutoStakingRequestNotInCommitteeListError              = metadataCommon.StopAutoStakingRequestNotInCommitteeListError
	StopAutoStakingRequestGetStakingTransactionError           = metadataCommon.StopAutoStakingRequestGetStakingTransactionError
	StopAutoStakingRequestStakingTransactionNotFoundError      = metadataCommon.StopAutoStakingRequestStakingTransactionNotFoundError
	StopAutoStakingRequestInvalidTransactionSenderError        = metadataCommon.StopAutoStakingRequestInvalidTransactionSenderError

	StopAutoStakingRequestNoAutoStakingAvaiableError = metadataCommon.StopAutoStakingRequestNoAutoStakingAvaiableError
	StopAutoStakingRequestTypeAssertionError         = metadataCommon.StopAutoStakingRequestTypeAssertionError
	StopAutoStakingRequestAlreadyStopError           = metadataCommon.StopAutoStakingRequestAlreadyStopError
	WrongIncognitoDAOPaymentAddressError             = metadataCommon.WrongIncognitoDAOPaymentAddressError
	ConsensusMetadataTypeAssertionError              = metadataCommon.ConsensusMetadataTypeAssertionError
	ConsensusMetadataInvalidTransactionSenderError   = metadataCommon.StopAutoStakingRequestInvalidTransactionSenderError
	AddStakingRequestNotInCommitteeListError         = metadataCommon.AddStakingNotInCommitteeListError
	AddStakingCommitteeNotFoundError                 = metadataCommon.AddStakingCommitteeNotFoundError
	ReDelegateRequestNotInCommitteeListError         = metadataCommon.ReDelegateNotInCommitteeListError
	ReDelegateCommitteeNotFoundError                 = metadataCommon.ReDelegateCommitteeNotFoundError
	ReDelegateInvalidTxError                         = metadataCommon.ReDelegateInvalidTxError
	// pde
	PDEWithdrawalRequestFromMapError    = metadataCommon.PDEWithdrawalRequestFromMapError
	CouldNotGetExchangeRateError        = metadataCommon.CouldNotGetExchangeRateError
	RejectInvalidFee                    = metadataCommon.RejectInvalidFee
	PDEFeeWithdrawalRequestFromMapError = metadataCommon.PDEFeeWithdrawalRequestFromMapError
	// portal
	PortalRequestPTokenParamError                = metadataCommon.PortalRequestPTokenParamError
	PortalRedeemRequestParamError                = metadataCommon.PortalRedeemRequestParamError
	PortalRedeemLiquidateExchangeRatesParamError = metadataCommon.PortalRedeemLiquidateExchangeRatesParamError
	// Unstake
	UnStakingRequestNotInCommitteeListError         = metadataCommon.UnStakingRequestNotInCommitteeListError
	UnStakingRequestGetStakerInfoError              = metadataCommon.UnStakingRequestGetStakerInfoError
	UnStakingRequestNotFoundStakerInfoError         = metadataCommon.UnStakingRequestNotFoundStakerInfoError
	UnStakingRequestStakingTransactionNotFoundError = metadataCommon.UnStakingRequestStakingTransactionNotFoundError
	UnStakingRequestInvalidTransactionSenderError   = metadataCommon.UnStakingRequestInvalidTransactionSenderError
	UnStakingRequestNoAutoStakingAvaiableError      = metadataCommon.UnStakingRequestNoAutoStakingAvaiableError
	UnStakingRequestTypeAssertionError              = metadataCommon.UnStakingRequestTypeAssertionError
	UnStakingRequestAlreadyStopError                = metadataCommon.UnStakingRequestAlreadyStopError
	UnStakingRequestInvalidFormatRequestKey         = metadataCommon.UnStakingRequestInvalidFormatRequestKey
	UnstakingRequestAlreadyUnstake                  = metadataCommon.UnstakingRequestAlreadyUnstake
	// eth utils
	VerifyProofAndParseReceiptError = metadataCommon.VerifyProofAndParseReceiptError
	// init privacy custom token
	InitTokenRequestDecodeInstructionError           = metadataCommon.InitTokenRequestDecodeInstructionError
	InitTokenRequestUnmarshalJsonError               = metadataCommon.InitTokenRequestUnmarshalJsonError
	InitTokenRequestNewInitPTokenRequestFromMapError = metadataCommon.InitTokenRequestNewInitPTokenRequestFromMapError
	InitTokenRequestValidateTxWithBlockChainError    = metadataCommon.InitTokenRequestValidateTxWithBlockChainError
	InitTokenRequestValidateSanityDataError          = metadataCommon.InitTokenRequestValidateSanityDataError
	InitTokenRequestBuildReqActionsError             = metadataCommon.InitTokenRequestBuildReqActionsError
	InitTokenResponseValidateSanityDataError         = metadataCommon.InitTokenResponseValidateSanityDataError
	// portal v3
	PortalCustodianDepositV3ValidateWithBCError     = metadataCommon.PortalCustodianDepositV3ValidateWithBCError
	PortalCustodianDepositV3ValidateSanityDataError = metadataCommon.PortalCustodianDepositV3ValidateSanityDataError
	NewPortalCustodianDepositV3MetaFromMapError     = metadataCommon.NewPortalCustodianDepositV3MetaFromMapError
	PortalUnlockOverRateCollateralsError            = metadataCommon.PortalUnlockOverRateCollateralsError
)
