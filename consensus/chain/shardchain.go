package chain

import (
	"time"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/wire"
	peer "github.com/libp2p/go-libp2p-peer"
)

type ShardChain struct {
	ShardID         byte
	Node            Node
	BlockGen        *blockchain.BlockGenerator
	Blockchain      *blockchain.BlockChain
	ConsensusEngine ConsensusInterface
}

func (s *ShardChain) GetConsensusEngine() ConsensusInterface {
	return s.ConsensusEngine
}

func (s *ShardChain) PushMessageToValidator(msg wire.Message) error {
	return s.Node.PushMessageToShard(msg, s.ShardID, map[peer.ID]bool{})
}

func (s *ShardChain) GetLastBlockTimeStamp() uint64 {
	return uint64(s.Blockchain.BestState.Shard[s.ShardID].BestBlock.Header.Timestamp)
}

func (s *ShardChain) GetBlkMinTime() time.Duration {
	return time.Second * 5

}

func (s *ShardChain) IsReady() bool {
	return s.Blockchain.Synker.IsLatest(true, s.ShardID)
}

func (s *ShardChain) GetHeight() uint64 {
	return s.Blockchain.BestState.Shard[s.ShardID].BestBlock.Header.Height
}

func (s *ShardChain) GetCommitteeSize() int {
	return len(s.Blockchain.BestState.Shard[s.ShardID].ShardCommittee)
}

func (s *ShardChain) GetPubKeyCommitteeIndex(pubkey string) int {
	return common.IndexOfStr(pubkey, s.Blockchain.BestState.Shard[s.ShardID].ShardCommittee)
}

func (s *ShardChain) GetLastProposerIndex() int {
	return s.Blockchain.BestState.Shard[s.ShardID].ShardProposerIdx
}

func (s *ShardChain) CreateNewBlock(round int) BlockInterface {
	newBlock, err := s.BlockGen.NewBlockShard(s.ShardID, round, s.Blockchain.Synker.GetClosestShardToBeaconPoolState(), s.Blockchain.BestState.Beacon.BeaconHeight, time.Now())
	if err != nil {
		return nil
	}
	// err = s.BlockGen.FinalizeShardBlock(newBlock, userKeyset)
	// if err != nil {
	// 	return nil
	// }

	return newBlock
}

func (s *ShardChain) ValidateBlock(interface{}) error {
	return nil
}

func (s *ShardChain) ValidatePreSignBlock(block interface{}) error {
	_ = block.(*blockchain.BeaconBlock)
	return nil
}

func (s *ShardChain) InsertBlk(block interface{}, isValid bool) {
	if isValid {
		s.Blockchain.InsertShardBlock(block.(*blockchain.ShardBlock), true)
	}
}

func (s *ShardChain) GetActiveShardNumber() int {
	return s.Blockchain.GetActiveShardNumber()
}
