package peerv2

import (
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/stretchr/testify/assert"
)

func TestGetShardIDsNormalAutoMode(t *testing.T) {
	role := userRole{shardID: -2}
	nodeMode := common.NodeModeAuto
	relayShard := []byte{}
	shardIDs := getWantedShardIDs(role, nodeMode, relayShard)

	// Must return shardID 255 to be able to get beacon's topics
	assert.Equal(t, []byte{255}, shardIDs, "incorrect shardIDs")
}

func TestGetShardIDsValidatorAutoMode(t *testing.T) {
	role := userRole{shardID: 3}
	nodeMode := common.NodeModeAuto
	relayShard := []byte{}
	shardIDs := getWantedShardIDs(role, nodeMode, relayShard)

	// shardID = 3 is enough when we want msg blockbeacon
	assert.Equal(t, []byte{3}, shardIDs, "incorrect shardIDs")
}

func TestGetShardIDsRelayBeacon(t *testing.T) {
	role := userRole{shardID: -2}
	nodeMode := common.NodeModeRelay
	relayShard := []byte{}
	shardIDs := getWantedShardIDs(role, nodeMode, relayShard)

	assert.Equal(t, []byte{255}, shardIDs, "incorrect shardIDs")
}

func TestGetShardIDsRelayShards(t *testing.T) {
	role := userRole{shardID: -2}
	nodeMode := common.NodeModeRelay
	relayShard := []byte{1, 2, 5, 7}
	shardIDs := getWantedShardIDs(role, nodeMode, relayShard)

	assert.Equal(t, []byte{1, 2, 5, 7, 255}, shardIDs, "incorrect shardIDs")
}
