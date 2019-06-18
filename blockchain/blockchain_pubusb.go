package blockchain

import (
	"math/rand"
)

func (blockchain *BlockChain) SubcribeNewShardBlock(ch chan *ShardBlock) int {
	blockchain.PubSub.mtx.Lock()
	defer blockchain.PubSub.mtx.Unlock()
	id := rand.Int()
	blockchain.PubSub.NewShardBlockEvent[id] = ch
	return id
}
func (blockchain *BlockChain) SubcribeNewBeaconBlock(ch chan *BeaconBlock) int {
	blockchain.PubSub.mtx.Lock()
	defer blockchain.PubSub.mtx.Unlock()
	id := rand.Int()
	blockchain.PubSub.NewBeaconBlockEvent[id] = ch
	return id
}
func (blockchain *BlockChain) UnsubcribeNewShardBlock(id int) {
	blockchain.PubSub.mtx.Lock()
	defer blockchain.PubSub.mtx.Unlock()
	if ch, ok := blockchain.PubSub.NewShardBlockEvent[id]; ok {
		close(ch)
		delete(blockchain.PubSub.NewShardBlockEvent, id)
	}
}
func (blockchain *BlockChain) UnsubcribeNewBeaconBlock(id int) {
	blockchain.PubSub.mtx.Lock()
	defer blockchain.PubSub.mtx.Unlock()
	if ch, ok := blockchain.PubSub.NewBeaconBlockEvent[id]; ok {
		close(ch)
		delete(blockchain.PubSub.NewBeaconBlockEvent, id)
	}
}
func (blockchain *BlockChain) NotifyNewShardBlockEvent(shardBlock *ShardBlock) {
	blockchain.PubSub.mtx.Lock()
	defer blockchain.PubSub.mtx.Unlock()
	for _, ch := range blockchain.PubSub.NewShardBlockEvent {
		go func(ch chan *ShardBlock) {
			ch <- shardBlock
		}(ch)
	}
}
func (blockchain *BlockChain) NotifyNewBeaconBlockEvent(beaconBlock *BeaconBlock) {
	blockchain.PubSub.mtx.Lock()
	defer blockchain.PubSub.mtx.Unlock()
	for _, ch := range blockchain.PubSub.NewBeaconBlockEvent {
		go func(ch chan *BeaconBlock) {
			ch <- beaconBlock
		}(ch)
	}
}
