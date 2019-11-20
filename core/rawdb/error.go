package rawdb

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	JsonMarshalError = iota
	JsonUnMarshalError

	// LevelDB
	OpenDbErr
	NotExistValue
	LvdbNotFound
	LvdbDeleteError
	LvdbPutError
	LvdbGetError
	LvdbHasError

	// BlockChain err
	NotImplHashMethod
	BlockExisted
	UnexpectedError
	KeyExisted

	// Serial Number Error
	StoreSerialNumbersError
	GetSerialNumbersLengthError
	HasSerialNumberError
	CleanSerialNumbersError

	// Output coin
	StoreOutputCoinsError
	GetOutputCoinByPublicKeyError

	// Commitment
	StoreCommitmentsError
	HasCommitmentError
	HasCommitmentInexError
	GetCommitmentByIndexError
	GetCommitmentIndexError
	GetCommitmentLengthError
	CleanCommitmentError

	// snderivator
	StoreSNDerivatorsError
	HasSNDerivatorError
	CleanSNDerivatorError

	// transaction
	StoreTransactionIndexError
	GetTransactionIndexByIdError
	StoreTxByPublicKeyError
	GetTxByPublicKeyError

	// Beacon
	StoreCrossShardNextHeightError
	HasCrossShardNextHeightError
	FetchCrossShardNextHeightError
	StoreBeaconBlockError
	HasBeaconBlockError
	FetchBeaconBlockError
	StoreBeaconBlockIndexError
	GetIndexOfBeaconBlockError
	DeleteBeaconBlockError
	StoreBeaconBestStateError
	FetchBeaconBestStateError
	CleanBeaconBestStateError
	GetBeaconBlockHashByIndexError
	StoreAcceptedShardToBeaconError
	GetAcceptedShardToBeaconError
	StoreBeaconCommitteeByHeightError
	StoreShardCommitteeByHeightError
	FetchShardCommitteeByHeightError
	FetchBeaconCommitteeByHeightError
	HasShardCommitteeByHeightError
	StoreAutoStakingByHeightError
	FetchAutoStakingByHeightError

	// Bridge
	BridgeUnexpectedError
	InsertETHTxHashIssuedError
	IsETHTxHashIssuedError

	// reward
	GetCommitteeRewardError
	RemoveCommitteeRewardError

	// slash
	GetProducersBlackListError
	StoreProducersBlackListError

	// pde
	GetWaitingPDEContributionByPairIDError
	GetPDEPoolForPairKeyError
	StoreWaitingPDEContributionError
	DeleteWaitingPDEContributionError
	StorePDEPoolForPairError
	GetPDEShareError
	AddShareAmountUpError
	GetPDETradeFeeError
	AddTradeFeeUpError
	DeduceTradeFeeError
	GetAllRecordsByPrefixError
	DeduceShareError
	TrackPDEStatusError
	GetPDEStatusError
)

var ErrCodeMessage = map[int]struct {
	Code    int
	message string
}{
	// -1xxx marshal and unmarshal
	JsonMarshalError:   {-1000, "Json Marshal Error"},
	JsonUnMarshalError: {-1001, "Json UnMarshal Error"},

	// -2xxx levelDb
	OpenDbErr:       {-2000, "Open database error"},
	NotExistValue:   {-2001, "H is not existed"},
	LvdbNotFound:    {-2002, "lvdb not found"},
	LvdbDeleteError: {-2003, "lvdb delete error"},
	LvdbPutError:    {-2004, "lvdb put error"},
	LvdbGetError:    {-2005, "lvdb get error"},
	LvdbHasError:    {-2006, "lvdb has error"},

	// -3xxx blockchain
	NotImplHashMethod: {-3000, "Data does not implement Hash() method"},
	BlockExisted:      {-3001, "Block already existed"},
	UnexpectedError:   {-3002, "Unexpected error"},
	KeyExisted:        {-3003, "PubKey already existed in database"},

	// -4xxx serial number
	StoreSerialNumbersError:     {-4000, "Store serial number error"},
	GetSerialNumbersLengthError: {-4001, "Get serial numbers length error"},
	HasSerialNumberError:        {-4002, "Has serial number error data=%+v shard=%+v token=%+v"},
	CleanSerialNumbersError:     {-4003, "Clean serial numbers"},

	// -5xxx output coin
	StoreOutputCoinsError:         {-5000, "Store output coin error"},
	GetOutputCoinByPublicKeyError: {-5001, "Get output coin by public key error"},

	// -6xxx commitment
	StoreCommitmentsError:     {-6000, "Store commitment error"},
	HasCommitmentError:        {-6001, "Has commitment error commitment=%+v shard=%+v token=%+v"},
	HasCommitmentInexError:    {-6002, "Has commitment error commitmentIndex=%+v shard=%+v token=%+v"},
	GetCommitmentByIndexError: {-6003, "Get commitment error commitmentIndex=%+v shard=%+v token=%+v"},
	GetCommitmentIndexError:   {-6004, "Get commitment index error commitment=%+v shard=%+v token=%+v"},
	GetCommitmentLengthError:  {-6005, "Get commitment length error"},
	CleanCommitmentError:      {-6006, "Clean commitment error"},

	// -7xxx snderivator
	StoreSNDerivatorsError: {-7000, "Store snd error"},
	HasSNDerivatorError:    {-7001, "Has snd error data=%+v shard=%+v token=%+v"},
	CleanSNDerivatorError:  {-7002, "Clean snd error"},

	// -8xxx transaction
	StoreTransactionIndexError:   {-8000, "Store transaction index error tx=%+v block=%+v index=%+v"},
	GetTransactionIndexByIdError: {-8001, "Get transaction index by id error id=%+v"},
	StoreTxByPublicKeyError:      {-8002, "Store tx by public key error tx=%+v pubkey=%+v shardID=%+v"},
	GetTxByPublicKeyError:        {-8003, "Get tx by public key error publlic key = %+v"},

	// -9xxx beacon
	StoreCrossShardNextHeightError:    {-9000, "Cannot store cross shard next height"},
	HasCrossShardNextHeightError:      {-9001, "Has cross shard next height"},
	FetchCrossShardNextHeightError:    {-9002, "Fetch cross shard next height error"},
	StoreBeaconBlockError:             {-9003, "Store beacon block error"},
	HasBeaconBlockError:               {-9004, "Has beacon block error"},
	FetchBeaconBlockError:             {-9005, "Fetch beacon block error"},
	StoreBeaconBlockIndexError:        {-9006, "Store beacon block index"},
	GetIndexOfBeaconBlockError:        {-9007, "Get index of beacon block error hash=%+v"},
	DeleteBeaconBlockError:            {-9008, "Delete beacon block error hash=%+v index=%+v"},
	StoreBeaconBestStateError:         {-9009, "Store beacon best state error"},
	FetchBeaconBestStateError:         {-9010, "Fetch beacon beststate error"},
	CleanBeaconBestStateError:         {-9011, "Clean beacon beststate error"},
	GetBeaconBlockHashByIndexError:    {-9012, "Get beacon block hash by index error index=%+v"},
	StoreAcceptedShardToBeaconError:   {-9013, "Store accepted shard to beacon error"},
	GetAcceptedShardToBeaconError:     {-9014, "Get accepted shard to beacon error"},
	StoreBeaconCommitteeByHeightError: {-9015, "Store beacon committee by height error"},
	StoreShardCommitteeByHeightError:  {-9016, "Store shard committee by height error"},
	FetchShardCommitteeByHeightError:  {-9017, "Fetch committee by height=%+v error"},
	FetchBeaconCommitteeByHeightError: {-9018, "Fetch beacon committee by height=%+v error"},
	HasShardCommitteeByHeightError:    {-9019, "Has committee shard by height error"},
	StoreAutoStakingByHeightError:     {-9020, "Store Auto Staking By Height Error"},
	FetchAutoStakingByHeightError:     {-9021, "Fetch Auto Staking By Height Error"},

	// -10xxx bridge
	BridgeUnexpectedError:      {-10000, "Insert ETH tx hash issued error"},
	InsertETHTxHashIssuedError: {-10001, "Insert ETH tx hash issued error"},
	IsETHTxHashIssuedError:     {-10002, "Is eth tx hash issued error"},

	// -11xxx reward
	GetCommitteeRewardError:    {-11000, "Get committee reward error"},
	RemoveCommitteeRewardError: {-11001, "Remove committee reward error"},

	// -12xxx Slash
	GetProducersBlackListError:   {-12000, "Get producers black list error"},
	StoreProducersBlackListError: {-12001, "Store producers black list error"},

	// -13xxx PDE
	GetWaitingPDEContributionByPairIDError: {-13001, "Get waiting pde contribution by pair id error"},
	GetPDEPoolForPairKeyError:              {-13002, "Get pde pool for pair key error"},
	StoreWaitingPDEContributionError:       {-13003, "Store waiting pde contribution error"},
	DeleteWaitingPDEContributionError:      {-13004, "Delete waiting pde contribution error"},
	StorePDEPoolForPairError:               {-13005, "Store pde pool for pair error"},
	GetPDEShareError:                       {-13006, "Get pde share error"},
	AddShareAmountUpError:                  {-13007, "Add share amount up error"},
	GetPDETradeFeeError:                    {-13008, "Get pde trade fee error"},
	AddTradeFeeUpError:                     {-13009, "Add trade fee up error"},
	DeduceTradeFeeError:                    {-13010, "Deduce trade fee error"},
	GetAllRecordsByPrefixError:             {-13011, "Get all records by prefix error"},
	DeduceShareError:                       {-13012, "Deduce share error"},
	TrackPDEStatusError:                    {-13013, "Track pde status error"},
	GetPDEStatusError:                      {-13014, "Get pde status error"},
}

type RawdbError struct {
	err     error
	Code    int
	Message string
}

func (e RawdbError) GetErrorCode() int {
	return e.Code
}
func (e RawdbError) GetError() error {
	return e.err
}
func (e RawdbError) GetMessage() string {
	return e.Message
}

func (e RawdbError) Error() string {
	return fmt.Sprintf("%d: %+v", e.Code, e.err)
}

func NewRawdbError(key int, err error, params ...interface{}) *RawdbError {
	return &RawdbError{
		err:     errors.Wrap(err, ErrCodeMessage[key].message),
		Code:    ErrCodeMessage[key].Code,
		Message: fmt.Sprintf(ErrCodeMessage[key].message, params),
	}
}
