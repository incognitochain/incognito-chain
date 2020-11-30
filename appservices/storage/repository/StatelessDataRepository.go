package repository

import (
	"github.com/incognitochain/incognito-chain/appservices/storage/model"
	"github.com/incognitochain/incognito-chain/common"
)

type TransactionStorer interface {
	StoreBeaconState (beaconState model.Transaction) error
}

type TransactionRetriver interface {
	GetTransactionByBlockHash (hash common.Hash) model.Transaction
	GetTransactionByBlockHeight(height uint64) model.Transaction
	GetTransactionByHash(hash common.Hash) model.Transaction
	GetAllTransaction (offset uint, limit uint) []model.Transaction

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
