package toshardins

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
)

type TxSendInitDCBVoteTokenMetadataIns struct {
	Amount                 uint32
	ReceiverPaymentAddress privacy.PaymentAddress
}

func (txSendInitDCBVoteTokenMetadataIns *TxSendInitDCBVoteTokenMetadataIns) GetStringFormat() []string {
	panic("implement me")
}

func NewTxSendInitDCBVoteTokenMetadataIns(amount uint32, receiverPaymentAddress privacy.PaymentAddress) *TxSendInitDCBVoteTokenMetadataIns {
	return &TxSendInitDCBVoteTokenMetadataIns{Amount: amount, ReceiverPaymentAddress: receiverPaymentAddress}
}

func (txSendInitDCBVoteTokenMetadataIns *TxSendInitDCBVoteTokenMetadataIns) BuildTransaction(
	minerPrivateKey *privacy.SpendingKey,
	db database.DatabaseInterface,
) metadata.Transaction {
	meta := metadata.NewSendInitDCBVoteTokenMetadata(
		txSendInitDCBVoteTokenMetadataIns.Amount,
		txSendInitDCBVoteTokenMetadataIns.ReceiverPaymentAddress,
	)
	sendVoteTokenTransaction := transaction.NewEmptyTx(
		minerPrivateKey,
		db,
		meta,
	)
	return sendVoteTokenTransaction
}

type TxSendInitGOVVoteTokenMetadataIns struct {
	Amount                 uint32
	ReceiverPaymentAddress privacy.PaymentAddress
}

func (txSendInitGOVVoteTokenMetadataIns *TxSendInitGOVVoteTokenMetadataIns) GetStringFormat() []string {
	panic("implement me")
}

func NewTxSendInitGOVVoteTokenMetadataIns(amount uint32, receiverPaymentAddress privacy.PaymentAddress) *TxSendInitGOVVoteTokenMetadataIns {
	return &TxSendInitGOVVoteTokenMetadataIns{Amount: amount, ReceiverPaymentAddress: receiverPaymentAddress}
}

func (txSendInitGOVVoteTokenMetadataIns *TxSendInitGOVVoteTokenMetadataIns) BuildTransaction(
	minerPrivateKey *privacy.SpendingKey,
	db database.DatabaseInterface,
) metadata.Transaction {
	meta := metadata.NewSendInitGOVVoteTokenMetadata(
		txSendInitGOVVoteTokenMetadataIns.Amount,
		txSendInitGOVVoteTokenMetadataIns.ReceiverPaymentAddress,
	)
	sendVoteTokenTransaction := transaction.NewEmptyTx(
		minerPrivateKey,
		db,
		meta,
	)
	return sendVoteTokenTransaction
}

func NewTxSendInitVoteTokenMetadataIns(
	boardType byte,
	amount uint32,
	receiverPaymentAddress privacy.PaymentAddress,
) Instruction {
	if boardType == common.DCBBoard {
		return NewTxSendInitDCBVoteTokenMetadataIns(amount, receiverPaymentAddress)
	} else {
		return NewTxSendInitGOVVoteTokenMetadataIns(amount, receiverPaymentAddress)
	}
}
