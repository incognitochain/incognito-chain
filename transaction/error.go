package transaction

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/pkg/errors"
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

var ErrCodeMessage = map[int]struct {
	Code    int
	Message string
}{
	// for common
	UnexpectedError:                               {-1000, "Unexpected error"},
	WrongTokenTxTypeError:                         {-1001, "Can't handle this TokenTxType"},
	CustomTokenExistedError:                       {-1002, "This token is existed in network"},
	WrongInputError:                               {-1003, "Wrong input transaction"},
	WrongSigError:                                 {-1004, "Wrong signature"},
	DoubleSpendError:                              {-1005, "Double spend"},
	TxNotExistError:                               {-1006, "Not exist tx for this"},
	RandomCommitmentError:                         {-1007, "Number of list commitments indices must be corresponding with number of input coins"},
	InputCoinIsVeryLargeError:                     {-1008, "Input coins in tx are very large: %d"},
	PaymentInfoIsVeryLargeError:                   {-1009, "Input coins in tx are very large: %d"},
	TokenIDInvalidError:                           {-1010, "Invalid TokenID: %+v"},
	PrivateKeySenderInvalidError:                  {-1011, "Invalid private key"},
	SignTxError:                                   {-1012, "Can not sign tx"},
	DecompressPaymentAddressError:                 {-1013, "Can not decompress public key from payment address %+v"},
	CanNotGetCommitmentFromIndexError:             {-1014, "Can not get commitment from index=%d shardID=%+v"},
	CanNotDecompressCommitmentFromIndexError:      {-1015, "Can not get commitment from index=%d shardID=%+v value=%+v"},
	InitWithnessError:                             {-1016, "Can not init witness for privacy with param: %s"},
	WithnessProveError:                            {-1017, "Can not prove with witness hashPrivacy=%+v param: %+s"},
	EncryptOutputError:                            {-1018, "Can not encrypt output"},
	DecompressSigPubKeyError:                      {-1019, "Can not decompress sig pubkey of tx"},
	InitTxSignatureFromBytesError:                 {-1020, "Can not init signature for tx from bytes"},
	VerifyTxSigFailError:                          {-1021, "Verify signature of tx is fail"},
	DuplicatedOutputSndError:                      {-1022, "Duplicate output"},
	SndExistedError:                               {-1023, "Snd existed: %s"},
	InputCommitmentIsNotExistedError:              {-1024, "Input's commitment is not existed"},
	TxProofVerifyFailError:                        {-1025, "Can not verify proof of tx %s"},
	VerifyMinerCreatedTxBeforeGettingInBlockError: {-1026, "Verify Miner Created Tx Before Getting In Block error"},
	CommitOutputCoinError:                         {-1027, "Commit all output error"},
	TokenIDExistedError:                           {-1028, "This token is existed in network"},
	TokenIDExistedByCrossShardError:               {-1029, "This token is existed in network by cross shard"},
	ExceedSizeInfoTxError:                         {-1030, "Size of tx info exceed max size info"},
	ExceedSizeInfoOutCoinError:                    {-1031, "Size of output coin's info exceed max size info"},
	ExceedSizeTx:                                  {-1032, "Size of tx info exceed max size of tx"},
	RejectInvalidLockTime:                         {-1033, "Wrong tx locktime"},
	RejectTxSize:                                  {-1034, "Tx size is too large"},
	RejectTxVersion:                               {-1035, "Wrong version tx"},
	RejectTxPublickeySigSize:                      {-1036, "Wrong tx Sig PK"},
	RejectTxType:                                  {-1037, "Wrong tx type"},
	RejectTxInfoSize:                              {-1038, "Wrong tx info length"},
	RejectTxMedataWithBlockChain:                  {-1039, "Reject invalid metadata with blockchain"},
	BatchTxProofVerifyFailError:                   {-1040, "Can not verify proof of batch txs %s"},
	VerifyOneOutOfManyProofFailedErr:              {-1041, "Verify one out of many proof failed"},

	// for PRV
	InvalidSanityDataPRVError:  {-2000, "Invalid sanity data for PRV"},
	InvalidDoubleSpendPRVError: {-2001, "Double spend PRV in blockchain"},

	// for privacy token
	InvalidSanityDataPrivacyTokenError:  {-3000, "Invalid sanity data for privacy Token"},
	InvalidDoubleSpendPrivacyTokenError: {-3001, "Double spend privacy Token in blockchain"},
	PrivacyTokenJsonError:               {-3002, "Json data error"},
	PrivacyTokenPRVJsonError:            {-3003, "Json data error"},
	PrivacyTokenInitPRVError:            {-3004, "Init tx for PRV error"},
	PrivacyTokenTxTypeNotHandleError:    {-3005, "Can not handle this tx type for privacy token"},
	PrivacyTokenInitTokenDataError:      {-3006, "Can not init data for privacy token tx"},
	PrivacyTokenInitFeeParamsError: 		 {-3007, "Init tx token fee params error"},

	// for normal token
	NormalTokenPRVJsonError: {-4000, "Json data error"},
	NormalTokenJsonError:    {-4001, "Json data error"},

	GetCommitmentsInDatabaseError: {-5000, "Get commitments from database got error"},
	InvalidPaymentAddressError: {-5001, "Payment address is null"},
	OnetimeAddressAlreadyExists: { -5002, "OnetimeAddress is already exists in database"},

	// For tx conversion
	SumInputCoinsAndOutputCoinsError:			   {-6000, "Sum input coins and output coins are not the same"},
	InvalidInputCoinVersionErr:					   {-6001, "Input coins version is invalid"},
}

type TransactionError struct {
	Code    int
	Message string
	err     error
}

func (e TransactionError) Error() string {
	return fmt.Sprintf("%+v: %+v %+v", e.Code, e.Message, e.err)
}

func NewTransactionErr(key int, err error, params ...interface{}) *TransactionError {
	e := &TransactionError{
		err:  errors.Wrap(err, common.EmptyString),
		Code: ErrCodeMessage[key].Code,
	}
	e.Message = ErrCodeMessage[key].Message
	if len(params) > 0 {
		e.Message = fmt.Sprintf(ErrCodeMessage[key].Message, params)
	}
	return e
}
