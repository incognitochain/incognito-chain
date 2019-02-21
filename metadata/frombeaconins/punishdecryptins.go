package frombeaconins

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
)

type PunishDecryptIns struct {
	boardType      byte
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

func NewPunishDecryptIns(boardType byte, paymentAddress privacy.PaymentAddress) *PunishDecryptIns {
	return &PunishDecryptIns{boardType: boardType, paymentAddress: paymentAddress}
}
