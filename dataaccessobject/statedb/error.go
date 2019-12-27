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
	ErrInvalidBurningConfirmStateType         = "invald burning confirm state type"
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
)

var ErrCodeMessage = map[int]struct {
	Code    int
	message string
}{
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

	// -2xxx store, get, update, delete at usecase level
	StoreSerialNumberError: {-2000, "Store Serial Number Error"},
	GetSerialNumberError:   {-2001, "Get Serial Number Error"},
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
