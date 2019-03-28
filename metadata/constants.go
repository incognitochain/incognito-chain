package metadata

const (
	LoanKeyDigestLength = 32
)

const (
	InvalidMeta = 1

	LoanRequestMeta  = 2
	LoanResponseMeta = 3
	LoanWithdrawMeta = 4
	LoanUnlockMeta   = 5
	LoanPaymentMeta  = 6

	DividendSubmitMeta  = 7
	DividendPaymentMeta = 8

	CrowdsaleRequestMeta = 10
	CrowdsalePaymentMeta = 11

	// CMB
	CMBInitRequestMeta      = 12
	CMBInitResponseMeta     = 13 // offchain multisig
	CMBInitRefundMeta       = 14 // miner
	CMBDepositContractMeta  = 15
	CMBDepositSendMeta      = 16
	CMBWithdrawRequestMeta  = 17
	CMBWithdrawResponseMeta = 18 // offchain multisig
	CMBLoanContractMeta     = 19

	BuyFromGOVRequestMeta        = 20
	BuyFromGOVResponseMeta       = 21
	BuyBackRequestMeta           = 22
	BuyBackResponseMeta          = 23
	IssuingRequestMeta           = 24
	IssuingResponseMeta          = 25
	ContractingRequestMeta       = 26
	ContractingReponseMeta       = 27
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

	//Voting
	NewDCBConstitutionIns    = 39
	NewGOVConstitutionIns    = 40
	UpdateDCBConstitutionIns = 41
	UpdateGOVConstitutionIns = 42

	SubmitDCBProposalMeta          = 43
	VoteDCBBoardMeta               = 44
	SubmitGOVProposalMeta          = 45
	VoteGOVBoardMeta               = 46
	RewardProposalWinnerMeta       = 47
	RewardDCBProposalSubmitterMeta = 48
	RewardGOVProposalSubmitterMeta = 49
	ShareRewardOldDCBBoardMeta     = 50
	ShareRewardOldGOVBoardMeta     = 51
	PunishDCBDecryptMeta           = 52
	PunishGOVDecryptMeta           = 53
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

// Special rules for shardID: stored as 2nd param of instruction of BeaconBlock
