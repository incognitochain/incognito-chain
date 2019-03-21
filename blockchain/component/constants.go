package component

const (
	VoteProposalIns = 100 + iota
	NewDCBConstitutionIns
	NewGOVConstitutionIns
	UpdateDCBConstitutionIns
	UpdateGOVConstitutionIns
	VoteBoardIns
	SubmitProposalIns

	AcceptDCBProposalIns
	AcceptDCBBoardIns
	AcceptGOVProposalIns
	AcceptGOVBoardIns

	RewardDCBProposalSubmitterIns
	RewardGOVProposalSubmitterIns
	ShareRewardOldDCBBoardIns
	ShareRewardOldGOVBoardIns
)

const (
	AllShards  = -1
	BeaconOnly = -2
)
