package peerv2

import (
	"fmt"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/peerv2/mocks"
	"github.com/incognitochain/incognito-chain/peerv2/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestSubscribeNoChange(t *testing.T) {
	role := userRole{
		layer:   "",
		role:    "",
		shardID: -2,
	}
	consensusData := &mocks.ConsensusData{}
	consensusData.On("GetUserRole").Return(role.layer, role.role, role.shardID)
	sub := &SubManager{
		info: info{consensusData: consensusData},
		role: role,
	}
	forced := false
	err := sub.Subscribe(forced)
	assert.Nil(t, err)
}

func TestSubscribeRoleChanged(t *testing.T) {
	role := userRole{
		layer:   "",
		role:    "",
		shardID: -2,
	}
	registerer := &mocks.Registerer{}
	var pairs []*proto.MessageTopicPair
	err := fmt.Errorf("error preventing further advance")
	registerer.On("Register", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(pairs, &proto.UserRole{}, err)
	consensusData := &mocks.ConsensusData{}
	consensusData.On("GetUserRole").Once().Return(common.ShardRole, common.PendingRole, 1)
	sub := &SubManager{
		info: info{
			consensusData: consensusData,
			nodeMode:      common.NodeModeAuto,
			relayShard:    []byte{},
		},
		role:       role,
		registerer: registerer,
	}
	forced := false
	sub.Subscribe(forced)
	consensusData.AssertNumberOfCalls(t, "GetUserRole", 1)
}
