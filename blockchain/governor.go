package blockchain

import "github.com/ninjadotorg/constant/privacy"

type GovernorInfo struct {
	BoardIndex          uint32
	StartedBlock        uint32
	EndBlock            uint32 // = startedblock of decent governor
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

type Governor interface {
	GetBoardIndex() uint32
}
