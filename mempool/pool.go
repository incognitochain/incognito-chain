package mempool

import (
	"github.com/constant-money/constant-chain/blockchain"
	"sort"
)

type PendingShardBlock struct {
	Queue map[uint64]*blockchain.ShardBlock
	Priority []uint64 // priority for block height, min height has first priority
	MaxLength int
}
// if queue does not reach full capacity then just enqueue new shard block
func(pending *PendingShardBlock) Enqueue(block *blockchain.ShardBlock){
	pending.Queue[block.Header.Height] = block
	pending.Priority = append(pending.Priority, block.Header.Height)
	if len(pending.Queue) == pending.MaxLength {
		sort.Slice(pending.Priority, func(i, j int) bool {
			return pending.Priority[i] < pending.Priority[j]
		})
	}
}
func (pending *PendingShardBlock) Dequeue(){
	toBeRemoved := len(pending.Priority) - 1
	delete(pending.Queue,pending.Priority[toBeRemoved])
	pending.Priority = pending.Priority[:toBeRemoved]
}

