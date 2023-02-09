package common

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	UnexpectedError = iota

	IssuingEvmRequestDecodeInstructionError
	IssuingEvmRequestUnmarshalJsonError
	IssuingEvmRequestNewIssuingEVMRequestFromMapError
	IssuingEvmRequestValidateTxWithBlockChainError
	IssuingEvmRequestValidateSanityDataError
	IssuingEvmRequestBuildReqActionsError
	IssuingEvmRequestVerifyProofAndParseReceipt

	IssuingRequestDecodeInstructionError
	IssuingRequestUnmarshalJsonError
	IssuingRequestNewIssuingRequestFromMapEror
	IssuingRequestValidateTxWithBlockChainError
	IssuingRequestValidateSanityDataError
	IssuingRequestBuildReqActionsError
	IssuingRequestVerifyProofAndParseReceipt

	BeaconBlockRewardNewBeaconBlockRewardInfoFromStrError
	BeaconBlockRewardBuildInstructionForBeaconBlockRewardError

	StopAutoStakingRequestNotInCommitteeListError
	StopAutoStakingRequestGetStakingTransactionError
	StopAutoStakingRequestStakingTransactionNotFoundError
	StopAutoStakingRequestInvalidTransactionSenderError
	StopAutoStakingRequestNoAutoStakingAvaiableError
	StopAutoStakingRequestTypeAssertionError
	StopAutoStakingRequestAlreadyStopError

	WrongIncognitoDAOPaymentAddressError

	// pde
	PDEWithdrawalRequestFromMapError
	CouldNotGetExchangeRateError
	RejectInvalidFee
	PDEFeeWithdrawalRequestFromMapError
	PDECouldNotGenerateHashFromStringError
	PDEInvalidMetadataValueError
	PDENotBurningTxError
	PDEInvalidTxTypeError
	PDETxWrongVersionError
	//

	// pDex v3
	Pdexv3ModifyParamsValidateSanityDataError
	Pdexv3WithdrawLPFeeValidateSanityDataError
	Pdexv3WithdrawProtocolFeeValidateSanityDataError
	Pdexv3WithdrawStakingRewardValidateSanityDataError

	// portal
	PortalRequestPTokenParamError
	PortalRedeemRequestParamError
	PortalRedeemLiquidateExchangeRatesParamError

	// Unstake
	UnStakingRequestNotInCommitteeListError
	UnStakingRequestGetStakerInfoError
	UnStakingRequestNotFoundStakerInfoError
	UnStakingRequestStakingTransactionNotFoundError
	UnStakingRequestInvalidTransactionSenderError
	UnStakingRequestNoAutoStakingAvaiableError
	UnStakingRequestTypeAssertionError
	UnStakingRequestAlreadyStopError
	UnStakingRequestInvalidFormatRequestKey
	UnstakingRequestAlreadyUnstake

	// eth utils
	VerifyProofAndParseReceiptError

	// init privacy custom token
	InitTokenRequestDecodeInstructionError
	InitTokenRequestUnmarshalJsonError
	InitTokenRequestNewInitPTokenRequestFromMapError
	InitTokenRequestValidateTxWithBlockChainError
	InitTokenRequestValidateSanityDataError
	InitTokenRequestBuildReqActionsError

	InitTokenResponseValidateSanityDataError

	// portal v3
	PortalCustodianDepositV3ValidateWithBCError
	PortalCustodianDepositV3ValidateSanityDataError
	NewPortalCustodianDepositV3MetaFromMapError
	PortalUnlockOverRateCollateralsError

	// portal v4
	PortalV4ShieldRequestValidateSanityDataError
	PortalV4UnshieldRequestValidateSanityDataError
	PortalV4FeeReplacementRequestMetaError
	PortalV4SubmitConfirmedTxRequestMetaError
	PortalV4ConvertVaultRequestMetaError

	// relaying header
	RelayingHeaderMetaError

	// bridge agg
	BridgeAggModifyParamValidateSanityDataError
	BridgeAggConvertRequestValidateSanityDataError
	BridgeAggShieldValidateSanityDataError
	BridgeAggUnshieldValidateSanityDataError

	// near bridge
	IssuingWasmRequestDecodeInstructionError
	IssuingWasmRequestUnmarshalJsonError
	IssuingWasmRequestNewIssuingWasmRequestFromMapError
	IssuingWasmRequestValidateTxWithBlockChainError
	IssuingWasmRequestValidateSanityDataError
	IssuingWasmRequestBuildReqActionsError
	IssuingWasmRequestVerifyProofAndParseReceipt

	ConsensusMetadataTypeAssertionError
	ConsensusMetadataInvalidTransactionSenderError
	AddStakingNotInCommitteeListError
	AddStakingCommitteeNotFoundError

	ReDelegateNotInCommitteeListError
	ReDelegateCommitteeNotFoundError
)

var ErrCodeMessage = map[int]struct {
	Code    int
	Message string
}{
	UnexpectedError: {-1, "Unexpected error"},

	// -1xxx issuing eth request
	IssuingEvmRequestDecodeInstructionError:           {-1001, "Can not decode instruction"},
	IssuingEvmRequestUnmarshalJsonError:               {-1002, "Can not unmarshall json"},
	IssuingEvmRequestNewIssuingEVMRequestFromMapError: {-1003, "Can no new issuing evm request from map"},
	IssuingEvmRequestValidateTxWithBlockChainError:    {-1004, "Validate tx with block chain error"},
	IssuingEvmRequestValidateSanityDataError:          {-1005, "Validate sanity data error"},
	IssuingEvmRequestBuildReqActionsError:             {-1006, "Build request action error"},
	IssuingEvmRequestVerifyProofAndParseReceipt:       {-1007, "Verify proof and parse receipt"},

	// -2xxx issuing eth request
	IssuingRequestDecodeInstructionError:        {-2001, "Can not decode instruction"},
	IssuingRequestUnmarshalJsonError:            {-2002, "Can not unmarshall json"},
	IssuingRequestNewIssuingRequestFromMapEror:  {-2003, "Can no new issuing eth request from map"},
	IssuingRequestValidateTxWithBlockChainError: {-2004, "Validate tx with block chain error"},
	IssuingRequestValidateSanityDataError:       {-2005, "Validate sanity data error"},
	IssuingRequestBuildReqActionsError:          {-2006, "Build request action error"},
	IssuingRequestVerifyProofAndParseReceipt:    {-2007, "Verify proof and parse receipt"},

	// -3xxx beacon block reward
	BeaconBlockRewardNewBeaconBlockRewardInfoFromStrError:      {-3000, "Can not new beacon block reward from string"},
	BeaconBlockRewardBuildInstructionForBeaconBlockRewardError: {-3001, "Can not build instruction for beacon block reward"},

	// -4xxx staking error
	StopAutoStakingRequestNotInCommitteeListError:         {-4000, "Stop Auto-Staking Request Not In Committee List Error"},
	StopAutoStakingRequestStakingTransactionNotFoundError: {-4001, "Stop Auto-Staking Request Staking Transaction Not Found Error"},
	StopAutoStakingRequestInvalidTransactionSenderError:   {-4002, "Stop Auto-Staking Request Invalid Transaction Sender Error"},
	StopAutoStakingRequestNoAutoStakingAvaiableError:      {-4003, "Stop Auto-Staking Request No Auto Staking Avaliable Error"},
	StopAutoStakingRequestTypeAssertionError:              {-4004, "Stop Auto-Staking Request Type Assertion Error"},
	StopAutoStakingRequestAlreadyStopError:                {-4005, "Stop Auto Staking Request Already Stop Error"},
	StopAutoStakingRequestGetStakingTransactionError:      {-4006, "Stop Auto Staking Request Get Staking Transaction Error"},
	UnStakingRequestNotInCommitteeListError:               {-4100, "Unstaking Request Not In Committee List Error"},
	UnStakingRequestGetStakerInfoError:                    {-4101, "Unstaking Request Get Staker Info Error"},
	UnStakingRequestNotFoundStakerInfoError:               {-4102, "Unstaking Request Not Found Staker Info Error"},
	UnStakingRequestStakingTransactionNotFoundError:       {-4103, "Unstaking Request Staking Transaction Not Found Error"},
	UnStakingRequestInvalidTransactionSenderError:         {-4104, "Unstaking Request Invalid Transaction Sender Error"},
	UnStakingRequestNoAutoStakingAvaiableError:            {-4105, "UnStaking Request No Auto Staking Available Error"},
	UnStakingRequestTypeAssertionError:                    {-4106, "UnStaking Request Type Assertion Error"},
	UnStakingRequestAlreadyStopError:                      {-4107, "UnStaking Request Already Stop Error"},
	UnStakingRequestInvalidFormatRequestKey:               {-4108, "Unstaking Request Key Is Invalid Format"},
	UnstakingRequestAlreadyUnstake:                        {-4109, "Public Key Has Been Already Unstaked"},
	// -5xxx dev reward error
	WrongIncognitoDAOPaymentAddressError: {-5001, "Invalid dev account"},

	// pde
	PDEWithdrawalRequestFromMapError:       {-6001, "PDE withdrawal request Error"},
	CouldNotGetExchangeRateError:           {-6002, "Could not get the exchange rate error"},
	RejectInvalidFee:                       {-6003, "Reject invalid fee"},
	PDECouldNotGenerateHashFromStringError: {-6003, "Could not generate hash from string"},
	PDEInvalidMetadataValueError:           {-6004, "Invalid metadata value"},
	PDENotBurningTxError:                   {-6004, "Tx is not a burning tx"},
	PDEInvalidTxTypeError:                  {-6005, "Invalid tx type"},
	PDETxWrongVersionError:                 {-6006, "Invalid tx version"},

	// portal
	PortalRequestPTokenParamError:                {-7001, "Portal request ptoken param error"},
	PortalRedeemRequestParamError:                {-7002, "Portal redeem request param error"},
	PortalRedeemLiquidateExchangeRatesParamError: {-7003, "Portal redeem liquidate exchange rates param error"},

	// eth utils
	VerifyProofAndParseReceiptError: {-8001, "Verify proof and parse receipt eth error"},

	// init privacy custom token
	InitTokenRequestDecodeInstructionError:           {-8002, "Cannot decode instruction"},
	InitTokenRequestUnmarshalJsonError:               {-8003, "Cannot unmarshall json"},
	InitTokenRequestNewInitPTokenRequestFromMapError: {-8004, "Cannot new InitPToken eth request from map"},
	InitTokenRequestValidateTxWithBlockChainError:    {-8005, "Validate tx with block chain error"},
	InitTokenRequestValidateSanityDataError:          {-8006, "Validate sanity data error"},
	InitTokenRequestBuildReqActionsError:             {-8007, "Build request action error"},

	InitTokenResponseValidateSanityDataError: {-8008, "Validate sanity data error"},

	// portal v3
	PortalCustodianDepositV3ValidateWithBCError:     {-9001, "Validate with blockchain tx portal custodian deposit v3 error"},
	PortalCustodianDepositV3ValidateSanityDataError: {-9002, "Validate sanity data tx portal custodian deposit v3 error"},
	NewPortalCustodianDepositV3MetaFromMapError:     {-9003, "New portal custodian deposit v3 metadata from map error"},
	PortalUnlockOverRateCollateralsError:            {-9004, "Validate with blockchain tx portal custodian unlock over rate v3 error"},

	// portal v4
	PortalV4ShieldRequestValidateSanityDataError:   {-10001, "Validate sanity data portal v4 shielding request error"},
	PortalV4UnshieldRequestValidateSanityDataError: {-10002, "Validate sanity data portal v4 unshielding request error"},
	PortalV4FeeReplacementRequestMetaError:         {-10003, "Portal batch unshield request metadata error"},
	PortalV4SubmitConfirmedTxRequestMetaError:      {-10004, "Portal submit external confirmed tx metadata error"},
	PortalV4ConvertVaultRequestMetaError:           {-10005, "Portal convert vault tx metadata error"},

	// relaying header
	RelayingHeaderMetaError: {-11005, " relaying header metadata error"},

	// bridge agg
	BridgeAggModifyParamValidateSanityDataError:    {-12000, "Modify list token validate sanity error"},
	BridgeAggConvertRequestValidateSanityDataError: {-12001, "Convert request sanity error"},
	BridgeAggShieldValidateSanityDataError:         {-12002, "Shield request sanity error"},
	BridgeAggUnshieldValidateSanityDataError:       {-12003, "Unshield request sanity error"},

	// near bridge error definition
	IssuingWasmRequestDecodeInstructionError:            {-1301, "Can not decode instruction"},
	IssuingWasmRequestUnmarshalJsonError:                {-1302, "Can not unmarshall json"},
	IssuingWasmRequestNewIssuingWasmRequestFromMapError: {-1303, "Can no new issuing wasm request from map"},
	IssuingWasmRequestValidateTxWithBlockChainError:     {-1304, "Validate tx with block chain error"},
	IssuingWasmRequestValidateSanityDataError:           {-1305, "Validate sanity data error"},
	IssuingWasmRequestBuildReqActionsError:              {-1306, "Build request action error"},
	IssuingWasmRequestVerifyProofAndParseReceipt:        {-1307, "Verify proof and parse receipt"},

	ConsensusMetadataTypeAssertionError:            {-4120, "ConsensusMetadata Type Assertion Error"},
	ConsensusMetadataInvalidTransactionSenderError: {-4121, "ConsensusMetadata Invalid Transaction Sender Error"},
	AddStakingNotInCommitteeListError:              {-4122, "AddStaking Not In Committee List Error"},
	AddStakingCommitteeNotFoundError:               {-4123, "AddStaking Committee Not Found Error"},
}

type MetadataTxError struct {
	Code    int    // The code to send with reject messages
	Message string // Human readable message of the issue
	Err     error
}

// Error satisfies the error interface and prints human-readable errors.
func (e MetadataTxError) Error() string {
	return fmt.Sprintf("%d: %s %+v", e.Code, e.Message, e.Err)
}

func NewMetadataTxError(key int, err error, params ...interface{}) *MetadataTxError {
	return &MetadataTxError{
		Code:    ErrCodeMessage[key].Code,
		Message: fmt.Sprintf(ErrCodeMessage[key].Message, params),
		Err:     errors.Wrap(err, ErrCodeMessage[key].Message),
	}
}
