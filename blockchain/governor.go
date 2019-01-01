package blockchain

type GovernorInfo struct {
	boardIndex       uint32
	StartedBlock     uint32
	EndBlock         uint32 // = startedblock of decent governor
	BoardPubKeys     [][]byte
	StartAmountToken uint64 //Sum of DCB token stack to all member of this board
}

func (governorInfo GovernorInfo) BoardIndex() uint32 {
	return governorInfo.boardIndex
}

type DCBGovernor struct {
	GovernorInfo
}

type GOVGovernor struct {
	GovernorInfo
}

type CMBGovernor struct {
	StartedBlock    uint32
	EndBlock        uint32
	CMBBoardPubKeys [][]byte
}

type Governor interface {
	BoardIndex() uint32
}
