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
	DividendMeta
	CrowdsaleRequestMeta
	CrowdsaleResponseMeta
	CrowdsalePaymentMeta

	// CMB
	CMBInitRequestMeta
	CMBInitResponseMeta  // offchain multisig
	CMBInitRefundMeta    // miner
	CMBDepositContractMeta
	CMBDepositSendMeta
	CMBWithdrawRequestMeta
	CMBWithdrawResponseMeta  // offchain multisig

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

	//Voting
	SubmitDCBProposalMeta
	VoteDCBProposalMeta
	VoteDCBBoardMeta
	AcceptDCBProposalMeta
	AcceptDCBBoardMeta

	SubmitGOVProposalMeta
	VoteGOVProposalMeta
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
)

const (
	MaxDivTxsPerBlock = 1000
	PayoutFrequency   = 1000 // Payout dividend every 1000 blocks
)

// update oracle board actions
const (
	Add = iota + 1
	Remove
)
