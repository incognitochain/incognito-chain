package blockchain

import (
	"encoding/json"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
)

type BeaconBody struct {
	// Shard State extract from shard to beacon block
	// Store all shard state == store content of all shard to beacon block
	ShardState   map[byte][]ShardState
	Instructions [][]string
}

type ShardState struct {
	Height     uint64
	Hash       common.Hash
	CrossShard []byte //In this state, shard i send cross shard tx to which shard
}

func (beaconBlock *BeaconBody) toString() string {
	res := ""
	for _, l := range beaconBlock.ShardState {
		for _, r := range l {
			res += strconv.Itoa(int(r.Height))
			res += r.Hash.String()
			crossShard, _ := json.Marshal(r.CrossShard)
			res += string(crossShard)

		}
	}
	for _, l := range beaconBlock.Instructions {
		for _, r := range l {
			res += r
		}
	}
	return res
}

func (beaconBody BeaconBody) Hash() common.Hash {
	return common.HashH([]byte(beaconBody.toString()))
}
