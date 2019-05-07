package metadata

const (
	InvalidMeta = 1

	LoanRequestMeta  = 2
	LoanResponseMeta = 3
	LoanWithdrawMeta = 4
	LoanUnlockMeta   = 5
	LoanPaymentMeta  = 6

	// Dividend: removed 7-8

	CrowdsaleRequestMeta = 10
	CrowdsalePaymentMeta = 11

	// CMB: removed 12-19

	BuyFromGOVRequestMeta        = 20
	BuyFromGOVResponseMeta       = 21
	BuyBackRequestMeta           = 22
	BuyBackResponseMeta          = 23
	IssuingRequestMeta           = 24
	IssuingResponseMeta          = 25
	ContractingRequestMeta       = 26
	ContractingResponseMeta      = 27
	OracleFeedMeta               = 28
	OracleRewardMeta             = 29
	RefundMeta                   = 30
	UpdatingOracleBoardMeta      = 31
	MultiSigsRegistrationMeta    = 32
	MultiSigsSpendingMeta        = 33
	WithSenderAddressMeta        = 34
	ResponseBaseMeta             = 35
	BuyGOVTokenRequestMeta       = 36
	ShardBlockSalaryRequestMeta  = 37
	ShardBlockSalaryResponseMeta = 38

	SubmitDCBProposalMeta          = 43
	VoteDCBBoardMeta               = 44
	SubmitGOVProposalMeta          = 45
	VoteGOVBoardMeta               = 46
	RewardProposalWinnerMeta       = 47
	RewardDCBProposalSubmitterMeta = 48
	RewardGOVProposalSubmitterMeta = 49
	ShareRewardOldDCBBoardMeta     = 50
	ShareRewardOldGOVBoardMeta     = 51
	RewardDCBProposalVoterMeta     = 52
	RewardGOVProposalVoterMeta     = 53
	SendBackTokenVoteBoardFailMeta = 54
	DCBVoteProposalMeta            = 55
	GOVVoteProposalMeta            = 56

	SendBackTokenToOldSupporterMeta = 59

	//statking
	ShardStakingMeta  = 63
	BeaconStakingMeta = 64

	TradeActivationMeta = 65
)

const (
	MaxDivTxsPerBlock = 1000
)

// update oracle board actions
const (
	Add = iota + 1
	Remove
)

var minerCreatedMetaTypes = []int{
	BuyFromGOVRequestMeta,
	BuyBackRequestMeta,
	ShardBlockSalaryResponseMeta,
	CrowdsalePaymentMeta,
	IssuingResponseMeta,
	ContractingResponseMeta,
}

// Special rules for shardID: stored as 2nd param of instruction of BeaconBlock
