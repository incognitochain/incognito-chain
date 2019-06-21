package pubsub

const ChanWorkLoad = 100

// TOPIC
const (
	NewshardblockTopic            = "newshardblocktopic"
	NewBeaconBlockTopic           = "newbeaconblocktopic"
	TransactionHashEnterNodeTopic = "transactionhashenternodetopic"
	ShardRoleTopic                = "shardroletopic"
	BeaconRoleTopic               = "beaconroletopic"
	TestTopic                     = "testtopic"
)

var Topics = []string{
	NewshardblockTopic,
	NewBeaconBlockTopic,
	TestTopic,
	TransactionHashEnterNodeTopic,
}
