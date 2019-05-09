package frombeaconins

import (
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/transaction"
)

var (
	mintDCBTokenParam = transaction.CustomTokenParamTx{
		PropertyID:     common.DCBTokenID.String(),
		PropertyName:   common.DCBTokenName,
		PropertySymbol: common.DCBTokenSymbol,
		Amount:         0,
		TokenTxType:    transaction.CustomTokenInit,
		Receiver:       nil,
		Mintable:       true,
	}
	mintGOVTokenParam = transaction.CustomTokenParamTx{
		PropertyID:     common.GOVTokenID.String(),
		PropertyName:   common.GOVTokenName,
		PropertySymbol: common.GOVTokenSymbol,
		Amount:         0,
		TokenTxType:    transaction.CustomTokenInit,
		Receiver:       nil,
		Mintable:       true,
	}
)
