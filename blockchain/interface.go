package blockchain

import "github.com/ninjadotorg/constant/common"

type BFTBlockInterface interface {
	// UnmarshalJSON(data []byte) error
}

type ShardToBeaconPool interface {
	RemoveBlock(map[byte]uint64) error
	GetFinalBlock() map[byte][]ShardToBeaconBlock
}
type CrossShardPool interface {
	RemoveBlock([]common.Hash) error
	GetBlock() map[byte][]CrossShardBlock
}

type NodeShardPool interface {
}

type NodeBeaconPool interface {
}
