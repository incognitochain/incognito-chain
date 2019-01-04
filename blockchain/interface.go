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
	PushBlock(ShardBlock) error
	GetBlocks(byte, uint64) ([]ShardBlock, error)
	RemoveBlocks(byte, uint64) error
}

type NodeBeaconPool interface {
	PushBlock(BeaconBlock) error
	GetBlocks(uint64) ([]BeaconBlock, error)
	RemoveBlocks(uint64) error
}
