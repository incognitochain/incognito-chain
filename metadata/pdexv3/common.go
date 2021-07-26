package pdexv3

import "strconv"

type FeeReceiverAddress struct {
	Token0ReceiverAddress string `json:"Token0ReceiverAddress"`
	Token1ReceiverAddress string `json:"Token1ReceiverAddress"`
	PRVReceiverAddress    string `json:"PRVReceiverAddress"`
	PDEXReceiverAddress   string `json:"PDEXReceiverAddress"`
}

type FeeReceiverAmount struct {
	Token0ReceiverAmount uint64 `json:"Token0ReceiverAmount"`
	Token1ReceiverAmount uint64 `json:"Token1ReceiverAmount"`
	PRVReceiverAmount    uint64 `json:"PRVReceiverAmount"`
	PDEXReceiverAmount   uint64 `json:"PDEXReceiverAmount"`
}

func (feeReceiverAddress *FeeReceiverAddress) ToString() string {
	return feeReceiverAddress.Token0ReceiverAddress + "," +
		feeReceiverAddress.Token1ReceiverAddress + "," +
		feeReceiverAddress.PRVReceiverAddress + "," +
		feeReceiverAddress.PDEXReceiverAddress
}

func (feeReceiverAmount *FeeReceiverAmount) ToString() string {
	return strconv.FormatUint(feeReceiverAmount.Token0ReceiverAmount, 10) + "," +
		strconv.FormatUint(feeReceiverAmount.Token1ReceiverAmount, 10) + "," +
		strconv.FormatUint(feeReceiverAmount.PRVReceiverAmount, 10) + "," +
		strconv.FormatUint(feeReceiverAmount.PDEXReceiverAmount, 10)
}
