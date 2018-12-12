package blockchain

import (
	"encoding/json"
	"github.com/ninjadotorg/constant/common"
)

type BeaconBlockBody struct {
	ShardState   [][]common.Hash
	Instructions [][]string
}

func (self *BeaconBlockBody) toString() string {
	res := ""
	//if self.ShardState != nil {
	//	for s, l := range self.ShardState {
	//		res += string(s)
	//		for _, h := range l {
	//			res += h.String()
	//		}
	//	}
	//}
	//if self.StateInstruction != nil {
	//	for s, i := range self.StateInstruction {
	//		res += s + i
	//	}
	//}
	return res
}

func (self *BeaconBlockBody) Hash() common.Hash {
	return common.DoubleHashH([]byte(self.toString()))
}

func (self *BeaconBlockBody) UnmarshalJSON(data []byte) error {
	blkBody := &BeaconBlockBody{}

	err := json.Unmarshal(data, blkBody)
	if err != nil {
		return NewBlockChainError(UnmashallJsonBlockError, err)
	}
	self = blkBody
	return nil
}
