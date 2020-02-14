package rawdbv2

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	// Beacon
	StoreBeaconBlockError = iota
	StoreBeaconBlockIndexError
	GetIndexOfBeaconBlockError
	HasBeaconBlockError
	GetBeaconBlockByHashError
	GetBeaconBlockByIndexError
	DeleteBeaconBlockError
	StoreBeaconBestStateError
	GetBeaconBestStateError
	StoreBeaconConsensusRootHashError
	GetBeaconConsensusRootHashError
	StoreBeaconRewardRootHashError
	GetBeaconRewardRootHashError
	StoreBeaconFeatureRootHashError
	GetBeaconFeatureRootHashError
	StoreBeaconSlashRootHashError
	GetBeaconSlashRootHashError
	StoreShardCommitteeRewardRootHashError
	GetShardCommitteeRewardRootHashError
	DeleteShardCommitteeRewardRootHashError
	StoreShardConsensusRootHashError
	GetShardConsensusRootHashError
	DeleteShardConsensusRootHashError
	StoreShardTransactionRootHashError
	GetShardTransactionRootHashError
	DeleteShardTransactionRootHashError
	StoreShardFeatureRootHashError
	GetShardFeatureRootHashError
	DeleteShardFeatureRootHashError
	StoreShardSlashRootHashError
	GetShardSlashRootHashError
	DeleteShardSlashRootHashError
	StorePreviousBeaconBestStateError
	GetPreviousBeaconBestStateError
	CleanUpPreviousBeaconBestStateError
	// Shard
	StoreShardBlockError
	StoreShardBlockIndexError
	HasShardBlockError
	GetShardBlockByHashError
	GetShardBlockByIndexError
	DeleteShardBlockError
	StoreCrossShardNextHeightError
	FetchCrossShardNextHeightError
	GetIndexOfBlockError
	StoreShardBestStateError
	StoreFeeEstimatorError
	GetFeeEstimatorError
	StorePreviousShardBestStateError
	GetPreviousShardBestStateError
	CleanUpPreviousShardBestStateError
	RestoreCrossShardNextHeightsError
	// tx
	StoreTransactionIndexError
	GetTransactionByHashError
	DeleteTransactionByHashError
	StoreTxByPublicKeyError
	GetTxByPublicKeyError
)

var ErrCodeMessage = map[int]struct {
	Code    int
	message string
}{
	HasBeaconBlockError:        {-1000, "Has Beacon Block Error"},
	GetBeaconBlockByHashError:  {-1001, "Get Beacon Block By Hash Error"},
	GetBeaconBlockByIndexError: {-1002, "Get Beacon Block By Index Error"},
	DeleteBeaconBlockError:     {-1003, "Delete Beacon Block Error"},
	StoreBeaconBlockIndexError: {-1004, "Store Beacon Block Index Error"},
	GetIndexOfBeaconBlockError: {-1005, "Get Index Of Beacon Block Error"},
	StoreBeaconBestStateError:  {-1006, "Store Beacon Best State Error"},
	GetBeaconBestStateError:    {-1007, "Fetch Beacon Best State Error"},

	StoreShardBlockError:           {-2000, "Store Shard Block Error"},
	HasShardBlockError:             {-2001, "Has Shard Block Error"},
	GetShardBlockByHashError:       {-2002, "Get Shard Block By Hash Error"},
	GetShardBlockByIndexError:      {-2003, "Get Shard Block By Index Error"},
	DeleteShardBlockError:          {-2004, "Delete Shard Block Error"},
	StoreCrossShardNextHeightError: {-2005, "Store Cross Shard Next Height Error"},
	FetchCrossShardNextHeightError: {-2006, "Fetch Cross Shard Next Height Error"},
	StoreShardBlockIndexError:      {-2007, "Store Shard Block Index Error"},
	GetIndexOfBlockError:           {-2008, "Get Index Of Shard Block Error"},
	StoreShardBestStateError:       {-2009, "Store Shard Best State Error"},
	StoreFeeEstimatorError:         {-2010, "Store Fee Estimator Error"},
	GetFeeEstimatorError:           {-2011, "Get Fee Estimator Error"},

	StoreTransactionIndexError:   {-3000, "Store Transaction Index Error"},
	GetTransactionByHashError:    {-3001, "Get Transaction By Hash Error"},
	StoreTxByPublicKeyError:      {-3002, "Store Tx By PublicKey Error"},
	GetTxByPublicKeyError:        {-3003, "Get Tx By Public Key Error"},
	DeleteTransactionByHashError: {-3004, "Delete Transaction By Hash Error"},

	StoreBeaconConsensusRootHashError:       {-4000, "Store Beacon Consensus Root Hash Error"},
	GetBeaconConsensusRootHashError:         {-4001, "Get Beacon Consensus Root Hash Error"},
	StoreBeaconRewardRootHashError:          {-4002, "Store Beacon Reward Root Hash Error"},
	GetBeaconRewardRootHashError:            {-4003, "Get Beacon Reward Root Hash Error"},
	StoreBeaconFeatureRootHashError:         {-4004, "Store Beacon Feature Root Hash Error"},
	GetBeaconFeatureRootHashError:           {-4005, "Get Beacon Feature Root Hash Error"},
	StoreBeaconSlashRootHashError:           {-4006, "Store Beacon Slash Root Hash Error"},
	GetBeaconSlashRootHashError:             {-4007, "Get Beacon Slash Root Hash Error"},
	StoreShardCommitteeRewardRootHashError:  {-4008, "Store Shard Committee Reward Root Hash Error"},
	GetShardCommitteeRewardRootHashError:    {-4009, "Get Shard Committee Reward Root Hash Error"},
	StoreShardConsensusRootHashError:        {-4010, "Store Shard Consensus Root Hash Error"},
	GetShardConsensusRootHashError:          {-4011, "Get Shard Consensus Root Hash Error"},
	StoreShardTransactionRootHashError:      {-4012, "Store Shard Transaction Root Hash Error"},
	GetShardTransactionRootHashError:        {-4013, "Get Shard Transaction Root Hash Error"},
	StoreShardFeatureRootHashError:          {-4014, "Store Shard Feature Root Hash Error"},
	GetShardFeatureRootHashError:            {-4015, "Get Shard Feature Root Hash Error"},
	StoreShardSlashRootHashError:            {-4016, "Store Shard Slash Root Hash Error"},
	GetShardSlashRootHashError:              {-4017, "Get Shard Slash Root Hash Error"},
	StorePreviousBeaconBestStateError:       {-4018, "Store Previous Beacon Best State Error"},
	GetPreviousBeaconBestStateError:         {-4019, "Get Previous Beacon Best State Error"},
	CleanUpPreviousBeaconBestStateError:     {-4020, "Clean Previous Beacon Best State Error"},
	StorePreviousShardBestStateError:        {-4021, "Store Previous Shard Best State Error"},
	GetPreviousShardBestStateError:          {-4022, "Get Previous Shard Best State Error"},
	CleanUpPreviousShardBestStateError:      {-4023, "Clean Previous Shard Best State Error"},
	DeleteShardCommitteeRewardRootHashError: {-4024, "Delete Shard Committee Reward Root Hash Error"},
	DeleteShardConsensusRootHashError:       {-4025, "Delete Shard Consensus Root Hash Error"},
	DeleteShardTransactionRootHashError:     {-4026, "Delete Shard Transaction Root Hash Error"},
	DeleteShardFeatureRootHashError:         {-4027, "Delete Shard Feature Root Hash Error"},
	DeleteShardSlashRootHashError:           {-4028, "Delete Shard Slash Root Hash Error"},
	RestoreCrossShardNextHeightsError:       {-4029, "Restore Cross Shard Next Heights Error"},
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
