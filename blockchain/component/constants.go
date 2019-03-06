package component

const (
	SealedLv1Or2VoteProposalIns = 100 + iota
	SealedLv3VoteProposalIns
	NormalVoteProposalFromSealerIns
	NormalVoteProposalFromOwnerIns
	PunishDecryptIns
	NewDCBConstitutionIns
	NewGOVConstitutionIns
	UpdateDCBConstitutionIns
	UpdateGOVConstitutionIns
	VoteDCBBoardIns
	VoteGOVBoardIns

	AcceptDCBProposalIns
	AcceptDCBBoardIns

	AcceptGOVProposalIns
	AcceptGOVBoardIns
)
