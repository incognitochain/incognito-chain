package pubsub

const ChanWorkLoad = 100

// TOPIC
const (
	NewShardblockTopic            = "newshardblocktopic"
	NewBeaconBlockTopic           = "newbeaconblocktopic"
	TransactionHashEnterNodeTopic = "transactionhashenternodetopic"
	ShardRoleTopic                = "shardroletopic"
	BeaconRoleTopic               = "beaconroletopic"
	MempoolInfoTopic              = "mempoolinfotopic"
	BeaconBeststateTopic          = "beaconbeststatetopic"
	ShardBeststateTopic           = "shardbeststatetopic"
	TestTopic                     = "testtopic"
)

var Topics = []string{
	NewShardblockTopic,
	NewBeaconBlockTopic,
	MempoolInfoTopic,
	TestTopic,
	TransactionHashEnterNodeTopic,
	ShardRoleTopic,
	BeaconRoleTopic,
	BeaconBeststateTopic,
	ShardBeststateTopic,
}
