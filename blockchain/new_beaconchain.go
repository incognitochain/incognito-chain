package blockchain

import (
	"time"

	"github.com/incognitochain/incognito-chain/common"
)

type BeaconChain struct {
	BestState       *BeaconBestState
	BlockGen        *BlockGenerator
	Blockchain      *BlockChain
	ChainConsensus  ConsensusInterface
	ConsensusEngine ConsensusEngineInterface
}

func (chain *BeaconChain) GetLastBlockTimeStamp() int64 {
	// return uint64(s.Blockchain.BestState.Beacon.BestBlock.Header.Timestamp)
	return chain.BestState.BestBlock.Header.Timestamp
}

func (chain *BeaconChain) GetBlkInterval() time.Duration {
	return chain.BestState.BlockInterval
}

func (chain *BeaconChain) GetBlkMaxCreateTime() time.Duration {
	return chain.BestState.BlockMaxCreateTime
}

func (chain *BeaconChain) IsReady() bool {
	return chain.Blockchain.Synker.IsLatest(false, 0)
}

func (chain *BeaconChain) CurrentHeight() uint64 {
	return chain.BestState.BestBlock.Header.Height
}

func (chain *BeaconChain) GetCommitteeSize() int {
	return len(chain.BestState.GetBeaconCommittee())
}

func (chain *BeaconChain) GetPubKeyCommitteeIndex(pubkey string) int {
	return common.IndexOfStr(pubkey, chain.BestState.GetBeaconCommittee())
}

func (chain *BeaconChain) GetLastProposerIndex() int {
	return chain.BestState.BeaconProposerIdx
}

func (chain *BeaconChain) CreateNewBlock(round int) BlockInterface {
	newBlock, err := chain.BlockGen.NewBlockBeacon(round, chain.Blockchain.Synker.GetClosestShardToBeaconPoolState())
	if err != nil {
		return nil
	}
	return newBlock
}

func (chain *BeaconChain) ValidateBlock(block BeaconBlock) error {
	_ = block
	return nil
}

// func (s *BeaconChain) ValidatePreSignBlock(block interface{}) error {
// 	_ = block.(*blockchain.BeaconBlock)
// 	return nil
// }

func (chain *BeaconChain) InsertBlk(block *BeaconBlock, isValid bool) {
	chain.Blockchain.InsertBeaconBlock(block, isValid)
}

// func (chain *BeaconChain) GetActiveShardNumber() int {
// 	return chain.Blockchain.GetActiveShardNumber()
// }
