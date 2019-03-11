package blockchain

import "github.com/big0t/constant-chain/privacy"

type GovernorInfo struct {
	BoardIndex          uint32
	StartedBlock        uint64
	EndBlock            uint64 // = startedblock of decent governor
	BoardPaymentAddress []privacy.PaymentAddress
	StartAmountToken    uint64 //Sum of DCB token stack to all member of this board
}

func (governorInfo GovernorInfo) GetBoardIndex() uint32 {
	return governorInfo.BoardIndex
}

type DCBGovernor struct {
	GovernorInfo
}

type GOVGovernor struct {
	GovernorInfo
}
