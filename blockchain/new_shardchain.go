package blockchain

import (
	"time"

	"github.com/incognitochain/incognito-chain/common"
)

type ShardChain struct {
	BestState       *ShardBestState
	BlockGen        *BlockGenerator
	Blockchain      *BlockChain
	ChainConsensus  ConsensusInterface
	ConsensusEngine ConsensusEngineInterface
}

func (chain *ShardChain) GetLastBlockTimeStamp() int64 {
	// return uint64(s.Blockchain.BestState.Beacon.BestBlock.Header.Timestamp)
	return chain.BestState.BestBlock.Header.Timestamp
}

func (chain *ShardChain) GetBlkInterval() time.Duration {
	return chain.BestState.BlockInterval
}

func (chain *ShardChain) GetBlkMaxCreateTime() time.Duration {
	return chain.BestState.BlockMaxCreateTime
}

func (chain *ShardChain) IsReady() bool {
	return chain.Blockchain.Synker.IsLatest(false, 0)
}

func (chain *ShardChain) CurrentHeight() uint64 {
	return chain.BestState.BestBlock.Header.Height
}

func (chain *ShardChain) GetCommitteeSize() int {
	return len(chain.BestState.ShardCommittee)
}

func (chain *ShardChain) GetPubKeyCommitteeIndex(pubkey string) int {
	return common.IndexOfStr(pubkey, chain.BestState.ShardCommittee)
}

func (chain *ShardChain) GetLastProposerIndex() int {
	return chain.BestState.ShardProposerIdx
}

func (chain *ShardChain) CreateNewBlock(round int) BlockInterface {
	newBlock, err := chain.BlockGen.NewBlockBeacon(round, chain.Blockchain.Synker.GetClosestShardToBeaconPoolState())
	if err != nil {
		return nil
	}
	return newBlock
}

func (chain *ShardChain) ValidateBlock(block BeaconBlock) error {
	_ = block
	return nil
}

// func (s *ShardChain) ValidatePreSignBlock(block interface{}) error {
// 	_ = block.(*blockchain.BeaconBlock)
// 	return nil
// }

func (chain *ShardChain) InsertBlk(block *BeaconBlock, isValid bool) {
	chain.Blockchain.InsertBeaconBlock(block, isValid)
}

// func (chain *ShardChain) GetActiveShardNumber() int {
// 	return chain.Blockchain.GetActiveShardNumber()
// }
