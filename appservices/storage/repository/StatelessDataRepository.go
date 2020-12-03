package repository

import (
	"context"
	"github.com/incognitochain/incognito-chain/appservices/storage/model"
	"github.com/incognitochain/incognito-chain/common"
)

type TransactionStorer interface {
	StoreTransaction (ctx context.Context, transaction model.Transaction) error
}

type TransactionRetriver interface {
	GetTransactionByBlockHash (hash common.Hash) model.Transaction
	GetTransactionByBlockHeight(height uint64) model.Transaction
	GetTransactionByHash(hash common.Hash) model.Transaction
	GetAllTransaction (offset uint, limit uint) []model.Transaction

}

type InputCoinStorer interface {
	StoreInputCoin (ctx context.Context, inputCoin model.InputCoin) error
}

type ShardCommitmentIndexStorer interface {
	StoreCommitment (ctx context.Context, commitment model.Commitment) error
}

type ShardOutputCoinStorer interface {
	StoreOutputCoin (ctx context.Context, outputCoin model.OutputCoin) error
}

type CrossShardOutputCoinStorer interface {
	StoreCrossShardOutputCoin (ctx context.Context, outputCoin model.OutputCoin) error
}


type TransactionRepository interface {
	TransactionStorer
	TransactionRetriver
}

type InstructionStorer interface {
	StoreBeaconState (beaconState model.Instruction) error
}

type InstructionRetriver interface {
	GetInstructionByBlockHash (hash common.Hash) model.Instruction
	GetInstructionByBlockHeight(height uint64) model.Instruction
	GetAllInstruction (offset uint, limit uint) []model.Instruction
}

type InstructionRepository interface {
	InstructionStorer
	InstructionRetriver
}
