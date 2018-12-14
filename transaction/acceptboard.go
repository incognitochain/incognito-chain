package transaction

type TxAcceptDCBBoard struct {
	*Tx
	DCBBoardPubKeys     [][]byte
	StartAmountDCBToken uint64
}

type TxAcceptGOVBoard struct {
	*Tx
	GOVBoardPubKeys     [][]byte
	StartAmountGOVToken uint64
}
