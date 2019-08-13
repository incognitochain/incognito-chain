package database

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	DriverExistErr = iota
	DriverNotRegisterErr

	// LevelDB
	OpenDbErr
	NotExistValue
	LvDbNotFound

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
)

var ErrCodeMessage = map[int]struct {
	Code    int
	message string
}{
	// -1xxx driver
	DriverExistErr:       {-1000, "Driver is already registered"},
	DriverNotRegisterErr: {-1001, "Driver is not registered"},

	// -2xxx levelDb
	OpenDbErr:     {-2000, "Open database error"},
	NotExistValue: {-2001, "H is not existed"},
	LvDbNotFound:  {-2002, "lvdb not found"},

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
}

type DatabaseError struct {
	err     error
	Code    int
	Message string
}

func (e DatabaseError) GetErrorCode() int {
	return e.Code
}

func (e DatabaseError) Error() string {
	return fmt.Sprintf("%d: %+v", e.Code, e.err)
}

func NewDatabaseError(key int, err error, params ...interface{}) *DatabaseError {
	return &DatabaseError{
		err:     errors.Wrap(err, ErrCodeMessage[key].message),
		Code:    ErrCodeMessage[key].Code,
		Message: fmt.Sprintf(ErrCodeMessage[key].message, params),
	}
}
