package pdexv3

import (
	"testing"

	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	. "github.com/stretchr/testify/assert"
)

func TestActionMarshalling(t *testing.T) {
	inst := []string{"285", "1", "0", "a9abecb293ad3d3253967e1a5bcaed45853404fb641784947051d0a1005cba71", "{\"Content\":{\"Receiver\":\"15kAkMVD3FE446h7nw1uujpQ31rf35XDTdctLuSynPD4jwr6ZS8Yoq8wu7yAwifUWq5maokoREYaUpz5LR4D1MNHGKSzFbvwmn4S4PweEqoa7N3A18yme3sYouggdBnebB46TfrdfgeSyWNn\",\"Amount\":137,\"TradePath\":[\"1124bb0628650378036312f5f75a94b6fe8b530d8230b4dde318ea9f2fef6cb1-3e888ea85b42c29057c05d52590793a7f0581cc69525e7b7f5837f51eeea4db5-9e13a8bf557202180ccb8c31d585f8cc35f564eb6b69ee5b5533b6c336ec8ce1\"],\"TokenToBuy\":\"3e888ea85b42c29057c05d52590793a7f0581cc69525e7b7f5837f51eeea4db5\",\"PairChanges\":[[160,-137]],\"RewardEarned\":[{\"1124bb0628650378036312f5f75a94b6fe8b530d8230b4dde318ea9f2fef6cb1\":25}],\"OrderChanges\":[{}]}}"}
	currentTrade := &Action{Content: &metadataPdexv3.AcceptedTrade{}}
	err := currentTrade.FromStringSlice(inst)
	NoError(t, err)
	result := currentTrade.StringSlice()
	for i := 0; i < len(result); i++ {
		Equal(t, result[i], inst[i])
	}
}
