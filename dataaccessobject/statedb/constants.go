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

	ShardStakerObjectType   = 35
	BeaconStakerObjectType  = 79
	CommitteeDataObjectType = 82
	// Portal v3
	PortalExternalTxObjectType      = 36
	PortalConfirmProofObjectType    = 37
	PortalUnlockOverRateCollaterals = 38

	SlashingCommitteeObjectType = 39

	// bsc bridge
	BridgeBSCTxObjectType = 40

	// Portal v4
	PortalV4StatusObjectType                      = 41
	PortalV4UTXOObjectType                        = 42
	PortalV4ShieldRequestObjectType               = 43
	PortalWaitingUnshieldObjectType               = 44
	PortalProcessedUnshieldRequestBatchObjectType = 45

	RewardRequestV3ObjectType = 46

	// PRV EVM bridge
	BridgePRVEVMObjectType = 47

	// Polygon bridge
	BridgePLGTxObjectType = 66

	// Fantom bridge
	BridgeFTMTxObjectType = 70

	// Fantom bridge
	BridgeAURORATxObjectType = 80

	// Fantom bridge
	BridgeAVAXTxObjectType = 81

	// pDex v3
	Pdexv3StatusObjectType                    = 48
	Pdexv3ParamsObjectType                    = 49
	Pdexv3ContributionObjectType              = 50
	Pdexv3PoolPairObjectType                  = 51
	Pdexv3ShareObjectType                     = 52
	Pdexv3NftObjectType                       = 53
	Pdexv3OrderObjectType                     = 54
	Pdexv3StakerObjectType                    = 55
	Pdexv3PoolPairLpFeePerShareObjectType     = 56
	Pdexv3PoolPairProtocolFeeObjectType       = 57
	Pdexv3PoolPairStakingPoolFeeObjectType    = 58
	Pdexv3ShareTradingFeeObjectType           = 59
	Pdexv3ShareLastLPFeesPerShareObjectType   = 60
	Pdexv3StakingPoolRewardPerShareObjectType = 61
	Pdexv3StakerRewardObjectType              = 62
	Pdexv3StakerLastRewardPerShareObjectType  = 63
	Pdexv3PoolPairMakingVolumeObjectType      = 64
	Pdexv3PoolPairOrderRewardObjectType       = 65
	Pdexv3ShareLastLmRewardPerShareObjectType = 67
	Pdexv3PoolPairLmRewardPerShareObjectType  = 68
	Pdexv3PoolPairLmLockedShareObjectType     = 69

	// bridge agg
	BridgeAggStatusObjectType             = 71
	BridgeAggUnifiedTokenObjectType       = 72
	BridgeAggConvertedTokenObjectType     = 73
	BridgeAggVaultObjectType              = 74
	BridgeAggWaitingUnshieldReqObjectType = 75
	BridgeAggParamObjectType              = 76

	// Near bridge
	BridgeNEARTxObjectType = 77

	// Decentralized Bridge Hub
	BridgeHubParamObjectType             = 83
	BridgeHubStakerInfoObjectType        = 84
	BridgeHubBridgeInfoObjectType        = 85
	BridgeHubBridgeInfoNetworkObjectType = 86
	BridgeHubStatusObjectType            = 87
	BridgeHubTxObjectType                = 88
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
	BeaconWaitingPool
	CommonShardPool
	BeaconLockingPool
	ShardPool
	BeaconCommittee
	ShardCommittee
	SyncingValidators
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
	BridgeMinusOperator = "-"
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
