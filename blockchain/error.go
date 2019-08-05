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
	UnmashallJsonBeaconBlockError
	MashallJsonBeaconBlockError
	UnmashallJsonBeaconBestStateError
	MashallJsonBeaconBestStateError
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
	CloneBeaconBestStateError
	ShardBestStateNotCompatibleError
	RegisterEstimatorFeeError
	FetchPreviousBlockError
	TransactionRootHashError
	ShardTransactionRootHashError
	CrossShardTransactionRootHashError
	FetchBeaconBlocksError
	WrongBlockTotalFeeError
	ShardIntructionFromTransactionAndInstructionError
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
	CrossShardBitMapError
	ShardCommitteeLengthAndCommitteeIndexError
)

var ErrCodeMessage = map[int]struct {
	Code    int
	message string
}{
	UnExpectedError:                                   {-1000, "Unexpected error"},
	UpdateMerkleTreeForBlockError:                     {-1001, "updateShardBestState Merkle Commitments Tree For Block is failed"},
	UnmashallJsonShardBlockError:                      {-1002, "Unmarshall Json Shard Block Is Failed"},
	MashallJsonShardBlockError:                        {-1003, "Marshall Json Shard Block Is Failed"},
	UnmashallJsonShardBestStateError:                  {-1004, "Unmarshall Json Shard Best State Is Failed"},
	MashallJsonShardBestStateError:                    {-1005, "Marshall Json Shard Best State Is Failed"},
	UnmashallJsonBeaconBlockError:                     {-1006, "Unmarshall Json Beacon Block Is Failed"},
	MashallJsonBeaconBlockError:                       {-1007, "Marshall Json Beacon Block Is Failed"},
	UnmashallJsonBeaconBestStateError:                 {-1008, "Unmarshall Json Beacon Best State Is Failed"},
	MashallJsonBeaconBestStateError:                   {-1009, "Marshall Json Beacon Best State Is Failed"},
	CanNotCheckDoubleSpendError:                       {-1010, "CanNotCheckDoubleSpend Error"},
	HashError:                                         {-1011, "Hash error"},
	WrongVersionError:                                 {-1012, "Version error"},
	WrongBlockHeightError:                             {-1013, "Block height error"},
	DatabaseError:                                     {-1014, "Database Error"},
	EpochError:                                        {-1015, "Epoch Error"},
	WrongTimestampError:                               {-1016, "Timestamp Error"},
	InstructionHashError:                              {-1017, "Instruction Hash Error"},
	ShardStateHashError:                               {-1018, "ShardState Hash Error"},
	RandomError:                                       {-1019, "Random Number Error"},
	VerificationError:                                 {-1020, "Verify Block Error"},
	BeaconError:                                       {-1021, "Beacon Error"},
	CrossShardBlockError:                              {-1022, "CrossShardBlockError"},
	SignatureError:                                    {-1023, "Signature Error"},
	CandidateError:                                    {-1024, "Candidate Error"},
	ShardIDError:                                      {-1025, "ShardID Error"},
	ProducerError:                                     {-1026, "Producer Error"},
	ShardStateError:                                   {-1027, "Shard State Error"},
	TransactionFromNewBlockError:                      {-1028, "Transaction invalid"},
	GenerateInstructionError:                          {-1029, "Instruction Error"},
	SwapError:                                         {-1030, "Swap Error"},
	MashallJsonError:                                  {-1031, "MashallJson Error"},
	DuplicateShardBlockError:                          {-1032, "Duplicate Shard Block Error"},
	CommitteeOrValidatorError:                         {-1033, "Committee or Validator Error"},
	ShardBlockSanityError:                             {-1034, "Shard Block Sanity Data Error"},
	StoreIncomingCrossShardError:                      {-1035, "Store Incoming Cross Shard Block Error"},
	DeleteIncomingCrossShardError:                     {-1036, "Delete Incoming Cross Shard Block Error"},
	WrongShardIDError:                                 {-1037, "Wrong Shard ID Error"},
	CloneShardBestStateError:                          {-1038, "Clone Shard Best State Error"},
	CloneBeaconBestStateError:                         {-1039, "Clone Beacon Best State Error"},
	ShardBestStateNotCompatibleError:                  {-1075, "New Block and Shard Best State Is NOT Compatible"},
	RegisterEstimatorFeeError:                         {-1040, "Register Fee Estimator Error"},
	FetchPreviousBlockError:                           {-1041, "Failed To Fetch Previous Block Error"},
	TransactionRootHashError:                          {-1042, "Transaction Root Hash Error"},
	ShardTransactionRootHashError:                     {-1043, "Shard Transaction Root Hash Error"},
	CrossShardTransactionRootHashError:                {-1044, "Cross Shard Transaction Root Hash Error"},
	FetchBeaconBlocksError:                            {-1045, "Fetch Beacon Blocks Error"},
	FetchBeaconBlockHashError:                         {-1046, "Fetch Beacon Block Hash Error"},
	FetchBeaconBlockError:                             {-1047, "Fetch Beacon Block Error"},
	WrongBlockTotalFeeError:                           {-1048, "Wrong Block Total Fee Error"},
	ShardIntructionFromTransactionAndInstructionError: {-1049, "Shard Instruction From Transaction And Instruction Error"},
	InstructionsHashError:                             {-1050, "Instruction Hash Error"},
	FlattenAndConvertStringInstError:                  {-1051, "Flatten And Convert String Instruction Error"},
	InstructionMerkleRootError:                        {-1052, "Instruction Merkle Root Error"},
	BeaconBlockNotCompatibleError:                     {-1053, "Beacon Block Not Compatible Error"},
	SwapInstructionError:                              {-1054, "Swap Instruction Error"},
	TransactionCreatedByMinerError:                    {-1055, "Transaction Created By Miner Error"},
	ResponsedTransactionWithMetadataError:             {-1056, "Responsed Transaction With Metadata Error"},
	UnmashallJsonShardCommitteesError:                 {-1057, "Unmashall Json Shard Committees Error"},
	MashallJsonShardCommitteesError:                   {-1058, "Mashall Json Shard Committees Error"},
	VerifyCrossShardBlockError:                        {-1059, "Verify Cross Shard Block Error"},
	NextCrossShardBlockError:                          {-1060, "Next Cross Shard Block Error"},
	FetchShardCommitteeError:                          {-1061, "Fetch Shard Committee Error"},
	CrossTransactionHashError:                         {-1062, "Cross Transaction Hash Error"},
	VerifyCrossShardCustomTokenError:                  {-1063, "Verify Cross Shard Custom Token Error"},
	ShardCommitteeRootHashError:                       {-1064, "Shard Committee Root Hash Error"},
	ShardPendingValidatorRootHashError:                {-1065, "Shard Pending Validator Root Hash Error"},
	StoreShardBlockError:                              {-1066, "Store Shard Block Error"},
	StoreBestStateError:                               {-1067, "Store Shard Shard Best State Error"},
	FetchAndStoreTransactionError:                     {-1068, "Fetch And Store Transaction Error"},
	FetchAndStoreCrossTransactionError:                {-1069, "Fetch And Store Cross Transaction Error"},
	RemoveCommitteeRewardError:                        {-1070, "Remove Committee Reward Error"},
	StoreBurningConfirmError:                          {-1071, "Store Burning Confirm Error"},
	SwapValidatorError:                                {-1072, "Swap Validator Error"},
	CrossShardBitMapError:                             {-1073, "Cross Shard Bitmap Error"},
	ShardCommitteeLengthAndCommitteeIndexError:        {-1074, "Shard Committee Length And Committee Index Error"},
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
