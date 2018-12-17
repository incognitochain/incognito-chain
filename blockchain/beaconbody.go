package blockchain

import (
	"encoding/json"

	"github.com/ninjadotorg/constant/common"
)

type BlockBodyBeacon struct {
	ShardState   [][]common.Hash
	Instructions [][]string
}

func (self *BlockBodyBeacon) toString() string {
	res := ""

	for _, l := range self.ShardState {
		for _, r := range l {
			res += r.String()
		}
	}

	for _, l := range self.Instructions {
		for _, r := range l {
			res += r
		}
	}

	return res
}

func (self *BlockBodyBeacon) Hash() common.Hash {
	return common.DoubleHashH([]byte(self.toString()))
}

func (self *BlockBodyBeacon) UnmarshalJSON(data []byte) error {
	blkBody := &BlockBodyBeacon{}

	err := json.Unmarshal(data, blkBody)
	if err != nil {
		return NewBlockChainError(UnmashallJsonBlockError, err)
	}
	self = blkBody
	return nil
}
