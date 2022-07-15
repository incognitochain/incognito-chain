package pruner

import (
	"sync"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/incdb"
)

const (
	IDLE    = 0
	RUNNING = 1
	ERROR   = -1
)

type TraverseHelper struct {
	db          incdb.Database
	shardID     byte
	finalHeight uint64
	wg          *sync.WaitGroup
	heightCh    chan uint64
	rootHashCh  chan blockchain.ShardRootHash
}

type UpdateStatus struct {
	ShardID byte
	Status  byte
}

type Config struct {
	ShouldPruneByHash bool `json:"ShouldPruneByHash"`
}

type ExtendedConfig struct {
	Config
	ShardID byte `json:"ShardID"`
}
