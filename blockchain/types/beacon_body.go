package types

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

func NewBeaconBody(shardState map[byte][]ShardState, instructions [][]string) BeaconBody {
	return BeaconBody{ShardState: shardState, Instructions: instructions}
}

func (b *BeaconBody) SetInstructions(inst [][]string) {
	b.Instructions = inst
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
