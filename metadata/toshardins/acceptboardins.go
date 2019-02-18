package toshardins

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
)

type Instruction interface {
	GetStringFormat() []string
}

type TxAcceptDCBBoardIns struct {
	BoardPaymentAddress []privacy.PaymentAddress
	StartAmountToken    uint64
}

func (txAcceptDCBBoardIns *TxAcceptDCBBoardIns) GetStringFormat() []string {
	panic("implement me")
}

func (txAcceptDCBBoardIns *TxAcceptDCBBoardIns) BuildTransaction(
	minerPrivateKey *privacy.SpendingKey,
	db database.DatabaseInterface,
) metadata.Transaction {
	meta := metadata.NewAcceptDCBBoardMetadata(txAcceptDCBBoardIns.BoardPaymentAddress,
		txAcceptDCBBoardIns.StartAmountToken)
	tx := transaction.NewEmptyTx(minerPrivateKey, db, meta)
	return tx
}

type TxAcceptGOVBoardIns struct {
	BoardPaymentAddress []privacy.PaymentAddress
	StartAmountToken    uint64
}

func (txAcceptGOVBoardIns *TxAcceptGOVBoardIns) GetStringFormat() []string {
	panic("implement me")
}

func (txAcceptGOVBoardIns *TxAcceptGOVBoardIns) BuildTransaction(
	minerPrivateKey *privacy.SpendingKey,
	db database.DatabaseInterface,
) metadata.Transaction {
	meta := metadata.NewAcceptGOVBoardMetadata(txAcceptGOVBoardIns.BoardPaymentAddress,
		txAcceptGOVBoardIns.StartAmountToken)
	tx := transaction.NewEmptyTx(minerPrivateKey, db, meta)
	return tx
}

//used in 2 cases:
//1. In Beacon chain
//2. In shard
func NewTxAcceptBoardIns(
	boardType byte,
	boardPaymentAddress []privacy.PaymentAddress,
	startAmountToken uint64,
) Instruction {
	if boardType == common.DCBBoard {
		return NewTxAcceptDCBBoardIns(
			boardPaymentAddress,
			startAmountToken,
		)
	} else {
		return NewTxAcceptGOVBoardIns(
			boardPaymentAddress,
			startAmountToken,
		)
	}
}

func NewTxAcceptDCBBoardIns(
	boardPaymentAddress []privacy.PaymentAddress,
	startAmountToken uint64,
) *TxAcceptDCBBoardIns {
	return &TxAcceptDCBBoardIns{
		BoardPaymentAddress: boardPaymentAddress,
		StartAmountToken:    startAmountToken,
	}
}

func NewTxAcceptGOVBoardIns(
	boardPaymentAddress []privacy.PaymentAddress,
	startAmountToken uint64,
) *TxAcceptGOVBoardIns {
	return &TxAcceptGOVBoardIns{
		BoardPaymentAddress: boardPaymentAddress,
		StartAmountToken:    startAmountToken,
	}
}
