package metadata

const (
	LoanKeyDigestLength = 32
	LoanKeyLength       = 32
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
	SealedLv1DCBBallotMeta
	SealedLv2DCBBallotMeta
	SealedLv3DCBBallotMeta
	NormalDCBBallotMetaFromSealerMeta
	NormalDCBBallotMetaFromOwnerMeta
	SealedLv1GOVBallotMeta
	SealedLv2GOVBallotMeta
	SealedLv3GOVBallotMeta
	NormalGOVBallotMetaFromSealerMeta
	NormalGOVBallotMetaFromOwnerMeta
	RewardProposalWinnerMeta
	RewardDCBProposalSubmitterMeta
	RewardGOVProposalSubmitterMeta
	PunishDCBDecryptMeta
	PunishGOVDecryptMeta
)

const (
	MaxDivTxsPerBlock = 1000
	PayoutFrequency   = 1000 // Payout dividend every 1000 blocks
)

// update oracle board actions
const (
	Add    = 1
	Remove = 2
)
