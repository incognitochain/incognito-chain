package blockchain

type DCBGovernor struct {
	StartedBlock        uint32
	EndBlock            uint32 // = startedblock of decent governor
	DCBBoardPubKeys     [][]byte
	StartAmountDCBToken uint64 //Sum of DCB token stack to all member of this board
}

type GOVGovernor struct {
	StartedBlock        uint32
	EndBlock            uint32 // = startedblock of decent governor
	GOVBoardPubKeys     [][]byte
	StartAmountGOVToken uint64 //Sum of GOV token stack to all member of this board
}

type CMBGovernor struct {
	StartedBlock    uint32
	EndBlock        uint32
	CMBBoardPubKeys [][]byte
}
