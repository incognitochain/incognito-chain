package pdexv3

import (
	"github.com/incognitochain/incognito-chain/common"
)

type FeeReceiverAddress struct {
	Token0ReceiverAddress string `json:"Token0ReceiverAddress"`
	Token1ReceiverAddress string `json:"Token1ReceiverAddress"`
	PRVReceiverAddress    string `json:"PRVReceiverAddress"`
	PDEXReceiverAddress   string `json:"PDEXReceiverAddress"`
}

type ReceiverInfo struct {
	TokenID    common.Hash `json:"TokenID"`
	AddressStr string      `json:"AddressStr"`
	Amount     uint64      `json:"Amount"`
}

func (feeReceiverAddress *FeeReceiverAddress) ToString() string {
	return feeReceiverAddress.Token0ReceiverAddress + "," +
		feeReceiverAddress.Token1ReceiverAddress + "," +
		feeReceiverAddress.PRVReceiverAddress + "," +
		feeReceiverAddress.PDEXReceiverAddress
}
