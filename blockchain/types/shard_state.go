package types

import "github.com/incognitochain/incognito-chain/common"

type ShardState struct {
	ValidationData     string
	CommitteeFromBlock common.Hash
	Height             uint64
	Hash               common.Hash
	CrossShard         []byte //In this state, shard i send cross shard tx to which shard
	ProposerTime       int64
	Version            int
}

func NewShardState(validationData string,
	committeeFromBlock common.Hash,
	height uint64,
	hash common.Hash,
	crossShard []byte,
	proposerTime int64,
	version int,
) ShardState {
	newCrossShard := make([]byte, len(crossShard))
	copy(newCrossShard, crossShard)
	return ShardState{
		ValidationData:     validationData,
		CommitteeFromBlock: committeeFromBlock,
		Height:             height,
		Hash:               hash,
		CrossShard:         newCrossShard,
		ProposerTime:       proposerTime,
		Version:            version,
	}
}
