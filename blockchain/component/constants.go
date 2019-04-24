package component

const (
	VoteProposalIns          = 100 + iota
	NewDCBConstitutionIns    //1
	NewGOVConstitutionIns    //2
	UpdateDCBConstitutionIns //3
	UpdateGOVConstitutionIns //4
	VoteBoardIns             //5
	SubmitProposalIns        //6

	AcceptDCBProposalIns //7
	AcceptDCBBoardIns    //8
	AcceptGOVProposalIns //9
	AcceptGOVBoardIns    //10

	RewardDCBProposalSubmitterIns       //11
	RewardGOVProposalSubmitterIns       //12
	ShareRewardOldDCBBoardSupportterIns //13
	ShareRewardOldGOVBoardSupportterIns //14
	SendBackTokenVoteBoardFailIns       //15

	ConfirmBuySellRequestMeta //16
	ConfirmBuyBackRequestMeta //17
	RewardDCBProposalVoterIns //18
	RewardGOVProposalVoterIns //19
	KeepOldDCBProposalIns     //20
	KeepOldGOVProposalIns     //21
)

const (
	AllShards  = -1
	BeaconOnly = -2
)
