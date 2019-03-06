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
	ShareRewardOldDCBBoardMeta
	ShareRewardOldGOVBoardMeta
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

//Stake amount
// count in miliconstant
const (
	STAKE_SHARD_AMOUNT  = 1
	STAKE_BEACON_AMOUNT = 2
)

const ()

// Special rules for shardID: stored as 2nd param of instruction of BeaconBlock
const (
	AllShards  = -1
	BeaconOnly = -2
)
