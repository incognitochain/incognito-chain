package blockchain

type DCBGovernor struct {
	StartedBlock    uint32
	EndBlock        uint32
	DCBBoardPubKeys []string
}

type GOVGovernor struct {
	StartedBlock    uint32
	EndBlock        uint32
	GOVBoardPubKeys []string
}

type CMBGovernor struct {
	StartedBlock    uint32
	EndBlock        uint32
	CMBBoardPubKeys []string
}
