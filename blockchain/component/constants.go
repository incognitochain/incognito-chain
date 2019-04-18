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
	ShareRewardOldDCBBoardSupportterIns
	ShareRewardOldGOVBoardSupportterIns
	SendBackTokenVoteBoardFailIns

	ConfirmBuySellRequestMeta
	ConfirmBuyBackRequestMeta
	RewardDCBProposalVoterIns
	RewardGOVProposalVoterIns
)

const (
	AllShards  = -1
	BeaconOnly = -2
)
