package pruner

import (
	"sync"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/incdb"
)

type TraverseHelper struct {
	db          incdb.Database
	shardID     byte
	finalHeight uint64
	wg          *sync.WaitGroup
	heightCh    chan uint64
	rootHashCh  chan blockchain.ShardRootHash
}
