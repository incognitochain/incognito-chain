package metadata

const (
	InvalidMeta = 1

	IssuingRequestMeta          = 24
	IssuingResponseMeta         = 25
	ContractingRequestMeta      = 26
	ETHHeaderRelayingMeta       = 27
	ETHHeaderRelayingRewardMeta = 28
	IssuingETHRequestMeta       = 29
	IssuingETHResponseMeta      = 30

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
