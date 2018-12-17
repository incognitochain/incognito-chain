package blockchain

import (
	"encoding/json"
	"fmt"

	"github.com/ninjadotorg/constant/common"
)

type BlockHeaderBeacon struct {
	BlockHeaderGeneric
	DataHash common.Hash `json:"DataHash"`
}

func (self *BlockHeaderBeacon) toString() string {
	res := ""
	res += fmt.Sprintf("%v", self.BlockHeaderGeneric.Version)
	res += fmt.Sprintf("%v", self.BlockHeaderGeneric.Height)
	res += fmt.Sprintf("%v", self.BlockHeaderGeneric.Timestamp)
	res += self.BlockHeaderGeneric.PrevBlockHash.String()
	res += self.DataHash.String()
	return res
}

func (self *BlockHeaderBeacon) Hash() common.Hash {
	return common.DoubleHashH([]byte(self.toString()))
}

func (self *BlockHeaderBeacon) UnmarshalJSON(data []byte) error {
	blkHeader := &BlockHeaderBeacon{}
	err := json.Unmarshal(data, blkHeader)
	if err != nil {
		return NewBlockChainError(UnmashallJsonBlockError, err)
	}
	self = blkHeader
	return nil
}

func (self *BlockHeaderBeacon) GetHeight() uint64 {
	return self.Height
}
