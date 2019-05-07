package component

import (
	"bytes"

	"github.com/constant-money/constant-chain/common"
)

type TradeData struct {
	TradeID   []byte
	BondID    *common.Hash
	Buy       bool
	Activated bool
	Amount    uint64
	ReqAmount uint64
}

func (td *TradeData) Compare(td2 *TradeData) bool {
	return bytes.Equal(td.TradeID, td2.TradeID) &&
		td.BondID.IsEqual(td2.BondID) &&
		td.Buy == td2.Buy &&
		td.Activated == td2.Activated &&
		td.Amount == td2.Amount &&
		td.ReqAmount == td2.ReqAmount
}
