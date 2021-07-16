package statedb

// version
const (
	defaultVersion = 0
)

// Object type
const (
	TestObjectType                   = 0
	CommitteeObjectType              = 1
	CommitteeRewardObjectType        = 2
	RewardRequestObjectType          = 3
	BlackListProducerObjectType      = 4
	SerialNumberObjectType           = 5
	CommitmentObjectType             = 6
	CommitmentIndexObjectType        = 7
	CommitmentLengthObjectType       = 8
	SNDerivatorObjectType            = 9
	OutputCoinObjectType             = 10
	OTACoinObjectType                = 11
	OTACoinIndexObjectType           = 12
	OTACoinLengthObjectType          = 13
	OnetimeAddressObjectType         = 14
	TokenObjectType                  = 15
	WaitingPDEContributionObjectType = 16
	PDEPoolPairObjectType            = 17
	PDEShareObjectType               = 18
	PDEStatusObjectType              = 19
	BridgeEthTxObjectType            = 20
	BridgeTokenInfoObjectType        = 21
	BridgeStatusObjectType           = 22
	BurningConfirmObjectType         = 23
	TokenTransactionObjectType       = 24

	// portal
	//final exchange rates
	PortalFinalExchangeRatesStateObjectType = 25
	//waiting porting request
	PortalWaitingPortingRequestObjectType = 26
	//liquidation
	PortalLiquidationPoolObjectType = 27
	PortalStatusObjectType          = 28
	CustodianStateObjectType        = 29
	WaitingRedeemRequestObjectType  = 30
	PortalRewardInfoObjectType      = 31
	LockedCollateralStateObjectType = 32
	RewardFeatureStateObjectType    = 33

	// PDEX v2
	PDETradingFeeObjectType = 34

	StakerObjectType = 35

	// Portal v3
	PortalExternalTxObjectType      = 36
	PortalConfirmProofObjectType    = 37
	PortalUnlockOverRateCollaterals = 38

	SlashingCommitteeObjectType = 39

	// bsc bridge
	BridgeBSCTxObjectType = 40
)

// Prefix length
const (
	prefixHashKeyLength = 12
	prefixKeyLength     = 20
)

// Committee Role
const (
	NextEpochShardCandidate = iota
	NextEpochBeaconCandidate
	CurrentEpochShardCandidate
	CurrentEpochBeaconCandidate
	SubstituteValidator
	CurrentValidator
	CommonBeaconPool
	CommonShardPool
	BeaconPool
	ShardPool
	BeaconCommittee
	ShardCommittee
)
const (
	BeaconChainID    = -1
	CandidateChainID = -2
)

// PDE Track Status type
const (
	WaitingContributionStatus = iota
	TradeStatus
	WithdrawStatus
)

// bridge
const (
	BridgeMinorOperator = "-"
	BridgePlusOperator  = "+"
)

// commitment
var (
	zeroBigInt = []byte("zero")
)

// token type
const (
	InitToken = iota
	CrossShardToken
	BridgeToken
	UnknownToken
)
