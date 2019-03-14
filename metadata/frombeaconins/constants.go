package frombeaconins

import (
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/transaction"
)

var (
	mintDCBTokenParam = transaction.CustomTokenParamTx{
		PropertyID:     common.DCBTokenID.String(),
		PropertyName:   "",
		PropertySymbol: "",
		Amount:         0,
		TokenTxType:    transaction.CustomTokenMint,
		Receiver:       nil,
	}
	mintGOVTokenParam = transaction.CustomTokenParamTx{
		PropertyID:     common.GOVTokenID.String(),
		PropertyName:   "",
		PropertySymbol: "",
		Amount:         0,
		TokenTxType:    transaction.CustomTokenMint,
		Receiver:       nil,
	}
)
