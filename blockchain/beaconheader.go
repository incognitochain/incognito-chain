package blockchain

import (
	"encoding/json"
	"fmt"
	"github.com/ninjadotorg/constant/common"
)

type BeaconBlockHeader struct {
	BlockHeaderGeneric
	DataHash common.Hash `json:"DataHash"`
}

func (self *BeaconBlockHeader) toString() string {
	res := ""
	res += fmt.Sprintf("%v", self.BlockHeaderGeneric.Version)
	res += fmt.Sprintf("%v", self.BlockHeaderGeneric.Height)
	res += fmt.Sprintf("%v", self.BlockHeaderGeneric.Timestamp)
	res += self.BlockHeaderGeneric.PrevBlockHash.String()
	res += self.DataHash.String()
	return res
}

func (self *BeaconBlockHeader) Hash() common.Hash {
	return common.DoubleHashH([]byte(self.toString()))
}

func (self *BeaconBlockHeader) UnmarshalJSON(data []byte) error {
	blkHeader := &BeaconBlockHeader{}
	err := json.Unmarshal(data, blkHeader)
	if err != nil {
		return NewBlockChainError(UnmashallJsonBlockError, err)
	}
	self = blkHeader
	return nil
}
