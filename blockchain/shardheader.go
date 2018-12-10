package blockchain

import (
	"github.com/ninjadotorg/constant/common"
)

type BlockHeaderShard struct {
	BlockHeaderGeneric
	ShardID byte
}

func (f BlockHeaderShard) Hash() common.Hash {
	return common.Hash{}
}

func (f BlockHeaderShard) UnmarshalJSON([]byte) error {
	return nil
}
