package pubsub

const ChanWorkLoad = 100

// TOPIC
const (
	NewshardblockTopic = "newshardblocktopic"
	NewBeaconBlockTopic = "newbeaconblocktopic"
	TestTopic = "testtopic"
)

var Topics = []string{
	NewshardblockTopic,
	NewBeaconBlockTopic,
	TestTopic,
}
