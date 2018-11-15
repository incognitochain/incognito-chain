package transaction

import (
	"fmt"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/pkg/errors"
)

type TxDivTokenVout struct {
	TxTokenVout
	LastPayout uint64 // seconds since unix epoch
}

func (tx TxDivTokenVout) Hash() *common.Hash {
	record := common.EmptyString
	record += fmt.Sprintf("%d", tx.Value)
	record += base58.Base58Check{}.Encode(tx.PaymentAddress.Apk[:], 0)
	record += fmt.Sprintf("%d", tx.LastPayout)
	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

type TxDivTokenData struct {
	PropertyID     common.Hash // = hash of TxTokenData data
	PropertyName   string
	PropertySymbol string

	Type   int // action type
	Amount uint64
	Vins   []TxTokenVin
	Vouts  []TxDivTokenVout
}

func (tx TxDivTokenData) Hash() (*common.Hash, error) {
	if tx.Vouts == nil {
		return nil, errors.New("Vout is empty")
	}
	record := tx.PropertyName + tx.PropertySymbol + fmt.Sprintf("%d", tx.Amount)
	for _, out := range tx.Vouts {
		record += string(out.PaymentAddress.Apk[:])
	}
	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash, nil
}
