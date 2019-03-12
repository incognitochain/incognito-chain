package component

const (
	NormalVoteProposalFromSealerIns = 100 + iota
	NormalVoteProposalFromOwnerIns
	PunishDecryptIns
	NewDCBConstitutionIns
	NewGOVConstitutionIns
	UpdateDCBConstitutionIns
	UpdateGOVConstitutionIns
	VoteBoardIns
	VoteGOVBoardIns

	AcceptDCBProposalIns
	AcceptDCBBoardIns

	AcceptGOVProposalIns
	AcceptGOVBoardIns
)
