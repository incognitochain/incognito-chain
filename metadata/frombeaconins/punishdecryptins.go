package frombeaconins

import (
	"github.com/big0t/constant-chain/common"
	"github.com/big0t/constant-chain/database"
	"github.com/big0t/constant-chain/metadata"
	"github.com/big0t/constant-chain/privacy"
	"github.com/big0t/constant-chain/transaction"
)

type PunishDecryptIns struct {
	boardType      common.BoardType
	paymentAddress privacy.PaymentAddress
}

func (punishDecryptIns PunishDecryptIns) GetStringFormat() ([]string, error) {
	panic("implement me")
}

func (punishDecryptIns PunishDecryptIns) BuildTransaction(minerPrivateKey *privacy.SpendingKey, db database.DatabaseInterface) (metadata.Transaction, error) {
	paymentAddress := punishDecryptIns.paymentAddress
	var meta metadata.Metadata
	if punishDecryptIns.boardType == common.DCBBoard {
		meta = metadata.NewPunishDCBDecryptMetadata(paymentAddress)
	} else {
		meta = metadata.NewPunishGOVDecryptMetadata(paymentAddress)
	}
	newTx := transaction.NewEmptyTx(minerPrivateKey, db, meta)
	return newTx, nil
}

func NewPunishDecryptIns(boardType common.BoardType, paymentAddress privacy.PaymentAddress) *PunishDecryptIns {
	return &PunishDecryptIns{boardType: boardType, paymentAddress: paymentAddress}
}
