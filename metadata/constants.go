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

	DividendPaymentMeta
	DividendSubmitMeta

	CrowdsaleRequestMeta
	CrowdsalePaymentMeta

	// Reserve
	ReserveRequestMeta
	ReserveResponseMeta
	ReservePaymentMeta

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
	OracleFeedMeta
	OracleRewardMeta
	RefundMeta
	UpdatingOracleBoardMeta
	MultiSigsRegistrationMeta
	MultiSigsSpendingMeta
	WithSenderAddressMeta
	ResponseBaseMeta
	BuyGOVTokenRequestMeta

	//Voting
	SubmitDCBProposalMeta
	VoteDCBBoardMeta
	AcceptDCBProposalMeta
	AcceptDCBBoardMeta

	SubmitGOVProposalMeta
	VoteGOVBoardMeta
	AcceptGOVProposalMeta
	AcceptGOVBoardMeta

	SendInitDCBVoteTokenMeta
	SendInitGOVVoteTokenMeta
	SealedLv1DCBVoteProposalMeta
	SealedLv2DCBVoteProposalMeta
	SealedLv3DCBVoteProposalMeta
	NormalDCBVoteProposalFromSealerMeta
	NormalDCBVoteProposalFromOwnerMeta
	SealedLv1GOVVoteProposalMeta
	SealedLv2GOVVoteProposalMeta
	SealedLv3GOVVoteProposalMeta
	NormalGOVVoteProposalFromSealerMeta
	NormalGOVVoteProposalFromOwnerMeta
	RewardProposalWinnerMeta
	RewardDCBProposalSubmitterMeta
	RewardGOVProposalSubmitterMeta
	RewardShareOldDCBBoardMeta
	RewardShareOldGOVBoardMeta
	PunishDCBDecryptMeta
	PunishGOVDecryptMeta
	SendBackTokenVoteFailMeta

	// STAKING
	ShardStakingMeta
	BeaconStakingMeta
)

const (
	MaxDivTxsPerBlock = 1000
)

// update oracle board actions
const (
	Add = iota + 1
	Remove
)
