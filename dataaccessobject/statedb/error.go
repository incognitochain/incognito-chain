package statedb

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	ErrInvalidByteArrayType                   = "invalid byte array type"
	ErrInvalidHashType                        = "invalid hash type"
	ErrInvalidBigIntType                      = "invalid big int type"
	ErrInvalidCommitteeStateType              = "invalid committee state type"
	ErrInvalidPaymentAddressType              = "invalid payment address type"
	ErrInvalidIncognitoPublicKeyType          = "invalid incognito public key type"
	ErrInvalidCommitteeRewardStateType        = "invalid reward receiver state type "
	ErrInvalidRewardRequestStateType          = "invalid reward request state type"
	ErrInvalidBlackListProducerStateType      = "invalid black list producer state type"
	ErrInvalidSerialNumberStateType           = "invalid serial number state type"
	ErrInvalidCommitmentStateType             = "invalid commitment state type"
	ErrInvalidSNDerivatorStateType            = "invalid snderivator state type"
	ErrInvalidOutputCoinStateType             = "invalid output coin state type"
	ErrInvalidTokenStateType                  = "invalid token state type"
	ErrInvalidWaitingPDEContributionStateType = "invalid waiting pde contribution state type"
	ErrInvalidPDEPoolPairStateType            = "invalid pde pool pair state type"
	ErrInvalidPDEShareStateType               = "invalid pde shard state type"
	ErrInvalidPDEStatusStateType              = "invalid pde status state type"
	ErrInvalidBridgeEthTxStateType            = "invalid bridge eth tx state type"
	ErrInvalidBridgeTokenInfoStateType        = "invalid bridge token info state type"
	ErrInvalidBridgeStatusStateType           = "invalid bridge status state type"
	ErrInvalidBurningConfirmStateType         = "invalid burning confirm state type"
)
const (
	InvalidByteArrayTypeError = iota
	InvalidHashTypeError
	InvalidBigIntTypeError
	InvalidCommitteeStateTypeError
	InvalidPaymentAddressTypeError
	InvalidIncognitoPublicKeyTypeError
	InvalidCommitteeRewardStateTypeError
	InvalidRewardRequestStateTypeError
	InvalidBlackListProducerStateTypeError
	InvalidSerialNumberStateTypeError
	InvalidCommitmentStateTypeError
	InvalidSNDerivatorStateTypeError
	InvalidOutputCoinStateTypeError
	// general error
	MethodNotSupportError
	// transaction related error
	StoreSerialNumberError
	GetSerialNumberError
	StoreCommitmentError
	GetCommitmentError
	StoreCommitmentIndexError
	GetCommitmentIndexError
	StoreCommitmentLengthError
	GetCommitmentLengthError
	StoreOutputCoinError
	GetOutputCoinError
	StoreSNDerivatorError
	GetSNDerivatorError
	StorePrivacyTokenError
	GetPrivacyTokenError
	GetPrivacyTokenTxsError
	PrivacyTokenIDExistedError
	// Consensus Related Error
	StoreBeaconCommitteeError
	GetBeaconCommitteeError
	StoreShardCommitteeError
	GetShardCommitteeError
	StoreAllShardCommitteeError
	StoreNextEpochCandidateError
	StoreCurrentEpochCandidateError
	StoreRewardRequestError
	GetRewardRequestError
	StoreCommitteeRewardError
	GetCommitteeRewardError
	ListCommitteeRewardError
	RemoveCommitteeRewardError
	StoreBlackListProducersError

	DeleteBeaconCommitteeError
	DeleteOneShardCommitteeError
	DeleteAllShardCommitteeError
	DeleteNextEpochShardCandidateError
	DeleteCurrentEpochShardCandidateError
	DeleteNextEpochBeaconCandidateError
	DeleteCurrentEpochBeaconCandidateError
	DeleteAllShardSubstitutesValidatorError
	DeleteBeaconSubstituteValidatorError
	// pdex error
	StoreWaitingPDEContributionError
	StorePDEPoolPairError
	StorePDEShareError
	GetPDEPoolForPairError
	TrackPDEStatusError
	GetPDEStatusError
	// bridge error
	BridgeInsertETHTxHashIssuedError
	IsETHTxHashIssuedError
	IsBridgeTokenExistedByTypeError
	CanProcessCIncTokenError
	CanProcessTokenPairError
	UpdateBridgeTokenInfoError
	GetAllBridgeTokensError
	TrackBridgeReqWithStatusError
	GetBridgeReqWithStatusError
	// burning confirm
	StoreBurningConfirmError
	GetBurningConfirmError
)

var ErrCodeMessage = map[int]struct {
	Code    int
	message string
}{
	MethodNotSupportError: {-1, "Method Not Support"},
	// -1xxx reposistory level
	InvalidByteArrayTypeError:              {-1000, "invalid byte array type"},
	InvalidHashTypeError:                   {-1001, "invalid hash type"},
	InvalidBigIntTypeError:                 {-1002, "invalid big int type"},
	InvalidCommitteeStateTypeError:         {-1003, "invalid committee state type"},
	InvalidPaymentAddressTypeError:         {-1004, "invalid payment address type"},
	InvalidIncognitoPublicKeyTypeError:     {-1005, "invalid incognito public key type"},
	InvalidCommitteeRewardStateTypeError:   {-1006, "invalid reward receiver state type "},
	InvalidRewardRequestStateTypeError:     {-1007, "invalid reward request state type"},
	InvalidBlackListProducerStateTypeError: {-1008, "invalid black list producer state type"},
	InvalidSerialNumberStateTypeError:      {-1009, "invalid serial number state type"},
	InvalidCommitmentStateTypeError:        {-1010, "invalid commitment state type"},
	InvalidSNDerivatorStateTypeError:       {-1011, "invalid snderivator state type"},
	InvalidOutputCoinStateTypeError:        {-1011, "invalid output coin state type"},
	// -2xxx transaction error
	StoreSerialNumberError:     {-2000, "Store Serial Number Error"},
	GetSerialNumberError:       {-2001, "Get Serial Number Error"},
	StoreCommitmentError:       {-2002, "Store Commitment Error"},
	GetCommitmentError:         {-2003, "Get Commitment Error"},
	StoreCommitmentIndexError:  {-2004, "Store Commitment Index Error"},
	GetCommitmentIndexError:    {-2005, "Get Commitment Index Error"},
	StoreCommitmentLengthError: {-2006, "Store Commitment Length Error"},
	GetCommitmentLengthError:   {-2007, "Get Commitment Length Error"},
	StoreOutputCoinError:       {-2008, "Store Output Coin Error"},
	GetOutputCoinError:         {-2009, "Get Output Coin Error"},
	StoreSNDerivatorError:      {-2010, "Store SNDeriavator Error"},
	GetSNDerivatorError:        {-2011, "Get SNDeriavator Error"},
	StorePrivacyTokenError:     {-2012, "Store Privacy Token Error"},
	GetPrivacyTokenError:       {-2013, "Get Privacy Token Error"},
	GetPrivacyTokenTxsError:    {-2014, "Get Privacy Token Txs Error"},
	PrivacyTokenIDExistedError: {-2015, "Privacy Token ID Existed Error"},
	// -3xxx: consensus error
	StoreBeaconCommitteeError:       {-3000, "Store Beacon Committee Error"},
	GetBeaconCommitteeError:         {-3001, "Get Beacon Committee Error"},
	StoreShardCommitteeError:        {-3002, "Store Shard Committee Error"},
	GetShardCommitteeError:          {-3003, "Get Shard Committee Error"},
	StoreAllShardCommitteeError:     {-3004, "Store All Shard Committee Error"},
	StoreNextEpochCandidateError:    {-3005, "Store Next Epoch Candidate Error"},
	StoreCurrentEpochCandidateError: {-3006, "Store Next Current Candidate Error"},
	StoreRewardRequestError:         {-3007, "Store Reward Request Error"},
	GetRewardRequestError:           {-3008, "Get Reward Request Error"},
	StoreCommitteeRewardError:       {-3009, "Store Committee Reward Error"},
	GetCommitteeRewardError:         {-3010, "Get Committee Reward Error"},
	ListCommitteeRewardError:        {-3011, "List Committee Reward Error"},
	RemoveCommitteeRewardError:      {-3012, "Remove Committee Reward Error"},
	StoreBlackListProducersError:    {-3013, "Store Black List Producers Error"},
	// -4xxx: pdex error
	StoreWaitingPDEContributionError: {-4000, "Store Waiting PDEX Contribution Error"},
	StorePDEPoolPairError:            {-4001, "Store PDEX Pool Pair Error"},
	StorePDEShareError:               {-4002, "Store PDEX Share Error"},
	GetPDEPoolForPairError:           {-4003, "Get PDEX Pool Pair Error"},
	TrackPDEStatusError:              {-4004, "Track PDEX Status Error"},
	GetPDEStatusError:                {-4005, "Get PDEX Status Error"},
	// -5xxx: bridge error
	BridgeInsertETHTxHashIssuedError: {-5000, "Bridge Insert ETH Tx Hash Issued Error"},
	IsETHTxHashIssuedError:           {-5001, "Is ETH Tx Hash Issued Error"},
	IsBridgeTokenExistedByTypeError:  {-5002, "Is Bridge Token Existed By Type Error"},
	CanProcessCIncTokenError:         {-5003, "Can Process Centralized Inc Token Error"},
	CanProcessTokenPairError:         {-5004, "Can Process Token Pair Error"},
	UpdateBridgeTokenInfoError:       {-5005, "Update Bridge Token Info Error"},
	GetAllBridgeTokensError:          {-5006, "Get All Bridge Tokens Error"},
	TrackBridgeReqWithStatusError:    {-5007, "Track Bridge Request With Status Error"},
	GetBridgeReqWithStatusError:      {-5008, "Get Bridge Request With Status Error"},
	// -6xxx: burning confirm
	StoreBurningConfirmError: {-6000, "Store Burning Confirm Error"},
	GetBurningConfirmError:   {-6001, "Get Burning Confirm Error"},
}

type StatedbError struct {
	err     error
	Code    int
	Message string
}

func (e StatedbError) GetErrorCode() int {
	return e.Code
}
func (e StatedbError) GetError() error {
	return e.err
}
func (e StatedbError) GetMessage() string {
	return e.Message
}

func (e StatedbError) Error() string {
	return fmt.Sprintf("%d: %+v", e.Code, e.err)
}

func NewStatedbError(key int, err error, params ...interface{}) *StatedbError {
	return &StatedbError{
		err:     errors.Wrap(err, ErrCodeMessage[key].message),
		Code:    ErrCodeMessage[key].Code,
		Message: fmt.Sprintf(ErrCodeMessage[key].message, params),
	}
}
