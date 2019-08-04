// Copyright (c) 2014-2016 The thaibaoautonomous developers
// Use of this source Code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	UnExpectedError = iota
	UpdateMerkleTreeForBlockError
	UnmashallJsonShardBlockError
	MashallJsonShardBlockError
	UnmashallJsonShardBestStateError
	MashallJsonShardBestStateError
	MashallJsonError
	CanNotCheckDoubleSpendError
	HashError
	WrongVersionError
	WrongBlockHeightError
	DatabaseError
	EpochError
	WrongTimestampError
	InstructionHashError
	ShardStateHashError
	RandomError
	VerificationError
	ShardError
	BeaconError
	SignatureError
	CrossShardBlockError
	CandidateError
	ShardIDError
	ProducerError
	ShardStateError
	TransactionFromNewBlockError
	GenerateInstructionError
	SwapError
	DuplicateShardBlockError
	CommitteeOrValidatorError
	ShardBlockSanityError
	StoreIncomingCrossShardError
	DeleteIncomingCrossShardError
	WrongShardIDError
	CloneShardBestStateError
	ShardBestStateNotCompatibleError
	RegisterEstimatorFeeError
	FetchPreviousBlockError
	TransactionRootHashError
	ShardTransactionRootHashError
	CrossShardTransactionRootHashError
	FetchBeaconBlocksError
	WrongBlockTotalFeeError
	ShardIntructionFromTransactionAndInsError
	InstructionsHashError
	FlattenAndConvertStringInstError
	InstructionMerkleRootError
	FetchBeaconBlockHashError
	FetchBeaconBlockError
	BeaconBlockNotCompatibleError
	SwapInstructionError
	TransactionCreatedByMinerError
	ResponsedTransactionWithMetadataError
	UnmashallJsonShardCommitteesError
	MashallJsonShardCommitteesError
	VerifyCrossShardBlockError
	NextCrossShardBlockError
	FetchShardCommitteeError
	CrossTransactionHashError
	VerifyCrossShardCustomTokenError
	ShardCommitteeRootHashError
	ShardPendingValidatorRootHashError
	StoreShardBlockError
	StoreBestStateError
	FetchAndStoreTransactionError
	FetchAndStoreCrossTransactionError
	RemoveCommitteeRewardError
	StoreBurningConfirmError
	SwapValidatorError
)

var ErrCodeMessage = map[int]struct {
	Code    int
	message string
}{
	UnExpectedError:                           {-1000, "Unexpected error"},
	UpdateMerkleTreeForBlockError:             {-1001, "updateShardBestState Merkle Commitments Tree For Block is failed"},
	UnmashallJsonShardBlockError:              {-1002, "Unmarshall Json Shard Block Is Failed"},
	MashallJsonShardBlockError:                {-1003, "Marshall Json Shard Block Is Failed"},
	UnmashallJsonShardBestStateError:          {-1004, "Unmarshall Json Shard Best State Is Failed"},
	MashallJsonShardBestStateError:            {-1005, "Marshall Json Shard Best State Is Failed"},
	CanNotCheckDoubleSpendError:               {-1006, "CanNotCheckDoubleSpend Error"},
	HashError:                                 {-1007, "Hash error"},
	WrongVersionError:                         {-1008, "Version error"},
	WrongBlockHeightError:                     {-1009, "Block height error"},
	DatabaseError:                             {-1010, "Database Error"},
	EpochError:                                {-1011, "Epoch Error"},
	WrongTimestampError:                       {-1012, "Timestamp Error"},
	InstructionHashError:                      {-1013, "Instruction Hash Error"},
	ShardStateHashError:                       {-1014, "ShardState Hash Error"},
	RandomError:                               {-1015, "Random Number Error"},
	VerificationError:                         {-1016, "Verify Block Error"},
	BeaconError:                               {-1017, "Beacon Error"},
	CrossShardBlockError:                      {-1018, "CrossShardBlockError"},
	SignatureError:                            {-1019, "Signature Error"},
	CandidateError:                            {-1020, "Candidate Error"},
	ShardIDError:                              {-1021, "ShardID Error"},
	ProducerError:                             {-1022, "Producer Error"},
	ShardStateError:                           {-1023, "Shard State Error"},
	TransactionFromNewBlockError:              {-1024, "Transaction invalid"},
	GenerateInstructionError:                  {-1025, "Instruction Error"},
	SwapError:                                 {-1026, "Swap Error"},
	MashallJsonError:                          {-1027, "MashallJson Error"},
	DuplicateShardBlockError:                  {-1028, "Duplicate Shard Block Error"},
	CommitteeOrValidatorError:                 {-1029, "Committee or Validator Error"},
	ShardBlockSanityError:                     {-1030, "Shard Block Sanity Data Error"},
	StoreIncomingCrossShardError:              {-1031, "Store Incoming Cross Shard Block Error"},
	DeleteIncomingCrossShardError:             {-1032, "Delete Incoming Cross Shard Block Error"},
	WrongShardIDError:                         {-1033, "Wrong Shard ID Error"},
	CloneShardBestStateError:                  {-1034, "Clone Shard Best State Error"},
	ShardBestStateNotCompatibleError:          {-1035, "New Block and Shard Best State Is NOT Compatible"},
	RegisterEstimatorFeeError:                 {-1036, "Register Fee Estimator Error"},
	FetchPreviousBlockError:                   {-1037, "Failed To Fetch Previous Block Error"},
	TransactionRootHashError:                  {-1038, "Transaction Root Hash Error"},
	ShardTransactionRootHashError:             {-1039, "Shard Transaction Root Hash Error"},
	CrossShardTransactionRootHashError:        {-1040, "Cross Shard Transaction Root Hash Error"},
	FetchBeaconBlocksError:                    {-1041, "Fetch Beacon Blocks Error"},
	FetchBeaconBlockHashError:                 {-1047, "Fetch Beacon Block Hash Error"},
	FetchBeaconBlockError:                     {-1068, "Fetch Beacon Block Error"},
	WrongBlockTotalFeeError:                   {-1042, "Wrong Block Total Fee Error"},
	ShardIntructionFromTransactionAndInsError: {-1043, "Shard Instruction From Transaction And Instruction Error"},
	InstructionsHashError:                     {-1044, "Instruction Hash Error"},
	FlattenAndConvertStringInstError:          {-1045, "Flatten And Convert String Instruction Error"},
	InstructionMerkleRootError:                {-1046, "Instruction Merkle Root Error"},
	BeaconBlockNotCompatibleError:             {-1048, "Beacon Block Not Compatible Error"},
	SwapInstructionError:                      {-1049, "Swap Instruction Error"},
	TransactionCreatedByMinerError:            {-1050, "Transaction Created By Miner Error"},
	ResponsedTransactionWithMetadataError:     {-1051, "Responsed Transaction With Metadata Error"},
	UnmashallJsonShardCommitteesError:         {-1052, "Unmashall Json Shard Committees Error"},
	MashallJsonShardCommitteesError:           {-1053, "Mashall Json Shard Committees Error"},
	VerifyCrossShardBlockError:                {-1054, "Verify Cross Shard Block Error"},
	NextCrossShardBlockError:                  {-1055, "Next Cross Shard Block Error"},
	FetchShardCommitteeError:                  {-1056, "Fetch Shard Committee Error"},
	CrossTransactionHashError:                 {-1057, "Cross Transaction Hash Error"},
	VerifyCrossShardCustomTokenError:          {-1058, "Verify Cross Shard Custom Token Error"},
	ShardCommitteeRootHashError:               {-1059, "Shard Committee Root Hash Error"},
	ShardPendingValidatorRootHashError:        {-1060, "Shard Pending Validator Root Hash Error"},
	StoreShardBlockError:                      {-1061, "Store Shard Block Error"},
	StoreBestStateError:                       {-1062, "Store Shard Shard Best State Error"},
	FetchAndStoreTransactionError:             {-1063, "Fetch And Store Transaction Error"},
	FetchAndStoreCrossTransactionError:        {-1064, "Fetch And Store Cross Transaction Error"},
	RemoveCommitteeRewardError:                {-1065, "Remove Committee Reward Error"},
	StoreBurningConfirmError:                  {-1066, "Store Burning Confirm Error"},
	SwapValidatorError:                        {-1067, "Swap Validator Error"},
}

type BlockChainError struct {
	Code    int
	Message string
	err     error
}

func (e BlockChainError) Error() string {
	return fmt.Sprintf("%d: %s \n %+v", e.Code, e.Message, e.err)
}

func NewBlockChainError(key int, err error) *BlockChainError {
	return &BlockChainError{
		Code:    ErrCodeMessage[key].Code,
		Message: ErrCodeMessage[key].message,
		err:     errors.Wrap(err, ErrCodeMessage[key].message),
	}
}
