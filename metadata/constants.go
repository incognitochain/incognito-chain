package metadata

const (
	InvalidMeta = 1

	IssuingRequestMeta          = 24
	IssuingResponseMeta         = 25
	ContractingRequestMeta      = 26
	BurningRequestMeta          = 27
	ETHHeaderRelayingMeta       = 28
	ETHHeaderRelayingRewardMeta = 29
	IssuingETHRequestMeta       = 80
	IssuingETHResponseMeta      = 81

	ResponseBaseMeta             = 35
	ShardBlockReward             = 36
	AcceptedBlockRewardInfoMeta  = 37
	ShardBlockSalaryResponseMeta = 38
	BeaconRewardRequestMeta      = 39
	BeaconSalaryResponseMeta     = 40
	ReturnStakingMeta            = 41
	DevRewardRequestMeta         = 42
	ShardBlockRewardRequestMeta  = 43
	WithDrawRewardRequestMeta    = 44
	WithDrawRewardResponseMeta   = 45

	//statking
	ShardStakingMeta  = 63
	BeaconStakingMeta = 64

	// Incognito -> Ethereum bridge
	BeaconPubkeyRootMeta = 70
	BridgePubkeyRootMeta = 71
	BurningConfirmMeta   = 72
)

var minerCreatedMetaTypes = []int{
	ShardBlockReward,
	BeaconSalaryResponseMeta,
	IssuingResponseMeta,
	IssuingETHResponseMeta,
	ReturnStakingMeta,
	WithDrawRewardResponseMeta,
	ETHHeaderRelayingRewardMeta,
}

// Special rules for shardID: stored as 2nd param of instruction of BeaconBlock
const (
	AllShards  = -1
	BeaconOnly = -2
)
