package frombeaconins

import (
	"encoding/json"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
	"strconv"
)

type InstructionFromBeacon interface {
	GetStringFormat() ([]string, error)
	BuildTransaction(minerPrivateKey *privacy.SpendingKey, db database.DatabaseInterface) (metadata.Transaction, error)
}

type TxAcceptDCBBoardIns struct {
	BoardPaymentAddress []privacy.PaymentAddress
	StartAmountToken    uint64
}

func (txAcceptDCBBoardIns *TxAcceptDCBBoardIns) GetStringFormat() ([]string, error) {
	content, err := json.Marshal(txAcceptDCBBoardIns)
	if err != nil {
		return nil, err
	}
	return []string{
		strconv.Itoa(metadata.AcceptDCBBoardMeta),
		strconv.Itoa(-1),
		string(content),
	}, nil
}

func (txAcceptDCBBoardIns *TxAcceptDCBBoardIns) BuildTransaction(
	minerPrivateKey *privacy.SpendingKey,
	db database.DatabaseInterface,
) (metadata.Transaction, error) {
	meta := metadata.NewAcceptDCBBoardMetadata(txAcceptDCBBoardIns.BoardPaymentAddress,
		txAcceptDCBBoardIns.StartAmountToken)
	tx := transaction.NewEmptyTx(minerPrivateKey, db, meta)
	return tx, nil
}

type TxAcceptGOVBoardIns struct {
	BoardPaymentAddress []privacy.PaymentAddress
	StartAmountToken    uint64
}

func (txAcceptGOVBoardIns *TxAcceptGOVBoardIns) GetStringFormat() ([]string, error) {
	content, err := json.Marshal(txAcceptGOVBoardIns)
	if err != nil {
		return nil, err
	}
	return []string{
		strconv.Itoa(metadata.AcceptGOVBoardMeta),
		strconv.Itoa(-1),
		string(content),
	}, nil
}

func (txAcceptGOVBoardIns *TxAcceptGOVBoardIns) BuildTransaction(
	minerPrivateKey *privacy.SpendingKey,
	db database.DatabaseInterface,
) (metadata.Transaction, error) {
	meta := metadata.NewAcceptGOVBoardMetadata(txAcceptGOVBoardIns.BoardPaymentAddress,
		txAcceptGOVBoardIns.StartAmountToken)
	tx := transaction.NewEmptyTx(minerPrivateKey, db, meta)
	return tx, nil
}

//used in 2 cases:
//1. In Beacon chain
//2. In shard
func NewTxAcceptBoardIns(
	boardType byte,
	boardPaymentAddress []privacy.PaymentAddress,
	startAmountToken uint64,
) InstructionFromBeacon {
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
