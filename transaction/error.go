package transaction

import (
	"github.com/incognitochain/incognito-chain/transaction/utils"
)

const (
	UnexpectedError = iota
	WrongTokenTxTypeError
	CustomTokenExistedError
	WrongInputError
	WrongSigError
	DoubleSpendError
	TxNotExistError
	RandomCommitmentError
	InvalidSanityDataPRVError
	InvalidSanityDataPrivacyTokenError
	InvalidDoubleSpendPRVError
	InvalidDoubleSpendPrivacyTokenError
	InputCoinIsVeryLargeError
	PaymentInfoIsVeryLargeError
	SumInputCoinsAndOutputCoinsError
	InvalidInputCoinVersionErr
	TokenIDInvalidError
	TokenIDExistedError
	TokenIDExistedByCrossShardError
	PrivateKeySenderInvalidError
	SignTxError
	DecompressPaymentAddressError
	CanNotGetCommitmentFromIndexError
	CanNotDecompressCommitmentFromIndexError
	InitWithnessError
	WithnessProveError
	EncryptOutputError
	DecompressSigPubKeyError
	InitTxSignatureFromBytesError
	VerifyTxSigFailError
	DuplicatedOutputSndError
	SndExistedError
	InputCommitmentIsNotExistedError
	TxProofVerifyFailError
	VerifyOneOutOfManyProofFailedErr
	BatchTxProofVerifyFailError
	VerifyMinerCreatedTxBeforeGettingInBlockError
	CommitOutputCoinError
	GetShardIDByPublicKeyError

	NormalTokenPRVJsonError
	NormalTokenJsonError

	PrivacyTokenInitFeeParamsError
	PrivacyTokenInitPRVError
	PrivacyTokenInitTokenDataError
	PrivacyTokenPRVJsonError
	PrivacyTokenJsonError
	PrivacyTokenTxTypeNotHandleError

	ExceedSizeTx
	ExceedSizeInfoTxError
	ExceedSizeInfoOutCoinError

	RejectInvalidLockTime
	RejectTxSize
	RejectTxVersion
	RejectTxPublickeySigSize
	RejectTxType
	RejectTxInfoSize
	RejectTxMedataWithBlockChain

	GetCommitmentsInDatabaseError
	InvalidPaymentAddressError
	OnetimeAddressAlreadyExists
)

var ErrCodeMessage = utils.ErrCodeMessage

type TransactionError = utils.TransactionError

func NewTransactionErr(key int, err error, params ...interface{}) *TransactionError {
	return utils.NewTransactionErr(key, err, params)
}
