package metadata

const (
	LoanKeyDigestLength = 32
)

const (
	InvalidMeta = iota

	LoanRequestMeta
	LoanResponseMeta
	LoanWithdrawMeta
	LoanUnlockMeta
	LoanPaymentMeta

	DividendSubmitMeta
	DividendPaymentMeta

	CrowdsaleRequestMeta
	CrowdsalePaymentMeta

	// CMB
	CMBInitRequestMeta
	CMBInitResponseMeta // offchain multisig
	CMBInitRefundMeta   // miner
	CMBDepositContractMeta
	CMBDepositSendMeta
	CMBWithdrawRequestMeta
	CMBWithdrawResponseMeta // offchain multisig
	CMBLoanContractMeta

	BuyFromGOVRequestMeta
	BuyFromGOVResponseMeta
	BuyBackRequestMeta
	BuyBackResponseMeta
	IssuingRequestMeta
	IssuingResponseMeta
	ContractingRequestMeta
	ContractingReponseMeta
	OracleFeedMeta
	OracleRewardMeta
	RefundMeta
	UpdatingOracleBoardMeta
	MultiSigsRegistrationMeta
	MultiSigsSpendingMeta
	WithSenderAddressMeta
	ResponseBaseMeta
	BuyGOVTokenRequestMeta
	ShardBlockSalaryRequestMeta
	ShardBlockSalaryResponseMeta

	//Voting
	NewDCBConstitutionIns
	NewGOVConstitutionIns
	UpdateDCBConstitutionIns
	UpdateGOVConstitutionIns

	SubmitDCBProposalMeta
	VoteDCBBoardMeta
	SubmitGOVProposalMeta
	VoteGOVBoardMeta
	RewardProposalWinnerMeta
	RewardDCBProposalSubmitterMeta
	RewardGOVProposalSubmitterMeta
	ShareRewardOldDCBBoardMeta
	ShareRewardOldGOVBoardMeta
	PunishDCBDecryptMeta
	PunishGOVDecryptMeta
	SendBackTokenVoteBoardFailMeta
	DCBVoteProposalMeta
	GOVVoteProposalMeta
)

const (
	// STAKING
	ShardStakingMeta  = 1
	BeaconStakingMeta = 2
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
const (
	AllShards  = -1
	BeaconOnly = -2
)
