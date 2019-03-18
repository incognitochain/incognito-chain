package frombeaconins

import (
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/privacy"
)

type PunishDecryptIns struct {
	boardType      common.BoardType
	paymentAddress privacy.PaymentAddress
}

func (punishDecryptIns PunishDecryptIns) GetStringFormat() ([]string, error) {
	panic("implement me")
}

func NewPunishDecryptIns(boardType common.BoardType, paymentAddress privacy.PaymentAddress) *PunishDecryptIns {
	return &PunishDecryptIns{boardType: boardType, paymentAddress: paymentAddress}
}
