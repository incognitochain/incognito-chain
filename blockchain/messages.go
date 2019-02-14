package blockchain

import (
	"fmt"

	libp2p "github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/cashec"
)

func (self *BlockChain) OnBlockShardReceived(block *ShardBlock) {
	if self.newShardBlkCh[block.Header.ShardID] != nil {
		*self.newShardBlkCh[block.Header.ShardID] <- block
	}
}
func (self *BlockChain) OnBlockBeaconReceived(block *BeaconBlock) {
	if self.syncStatus.Beacon {
		self.newBeaconBlkCh <- block
	}
}

func (self *BlockChain) GetBeaconState() (*BeaconChainState, error) {
	state := &BeaconChainState{
		Height:          self.BestState.Beacon.BeaconHeight,
		BlockHash:       self.BestState.Beacon.BestBlockHash,
		ShardsPoolState: self.config.ShardToBeaconPool.GetValidPendingBlockHash(),
	}
	return state, nil
}

func (self *BlockChain) OnBeaconStateReceived(state *BeaconChainState, peerID libp2p.ID) {
	if self.syncStatus.Beacon {
		self.BeaconStateCh <- &PeerBeaconChainState{
			state, peerID,
		}
	}
}

func (self *BlockChain) GetShardState(shardID byte) (*ShardChainState, error) {
	//TODO: check node mode -> node role
	state := &ShardChainState{
		Height:    self.BestState.Shard[shardID].ShardHeight,
		ShardID:   shardID,
		BlockHash: self.BestState.Shard[shardID].BestShardBlockHash,
	}
	return state, nil
}

func (self *BlockChain) OnShardStateReceived(state *ShardChainState, peerID libp2p.ID) {
	if self.newShardBlkCh[state.ShardID] != nil {
		self.ShardStateCh[state.ShardID] <- &PeerShardChainState{
			state, peerID,
		}
	}
}

func (self *BlockChain) OnShardToBeaconBlockReceived(block ShardToBeaconBlock) {
	//TODO: check node mode -> node mode & role before add block to pool
	fmt.Println("Blockchain Message/OnShardToBeaconBlockReceived: Block Height", block.Header.Height)
	blkHash := block.Header.Hash()
	err := cashec.ValidateDataB58(block.Header.Producer, block.ProducerSig, blkHash.GetBytes())

	if err != nil {
		Logger.log.Debugf("Invalid Producer Signature of block height %+v in Shard %+v", block.Header.Height, block.Header.ShardID)
		return
	}
	if block.Header.Version != VERSION {
		Logger.log.Debugf("Invalid Verion of block height %+v in Shard %+v", block.Header.Height, block.Header.ShardID)
		return
	}

	//TODO: what if shard to beacon from old committee
	if err = ValidateAggSignature(block.ValidatorsIdx, self.BestState.Beacon.ShardCommittee[block.Header.ShardID], block.AggregatedSig, block.R, block.Hash()); err != nil {
		Logger.log.Error(err)
		return
	}

	if err = self.config.ShardToBeaconPool.AddShardToBeaconBlock(block); err != nil {
		Logger.log.Error(err)
		return
	}

	//TODO review: synblock already find?
	//if self.BestState.Beacon.BestShardHeight[block.Header.ShardID] < block.Header.Height-1 {
	//	self.config.Server.PushMessageGetShardToBeacons(block.Header.ShardID, self.BestState.Beacon.BestShardHeight[block.Header.ShardID]+1, block.Header.Height)
	//}
}

func (self *BlockChain) OnCrossShardBlockReceived(block CrossShardBlock) {
	//TODO: check node mode -> node role before add block to pool
	err := self.config.CrossShardPool.AddCrossShardBlock(block)
	if err != nil {
		Logger.log.Error(err)
	}
}
