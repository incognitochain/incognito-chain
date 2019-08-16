package blockchain

import (
	"time"

	"github.com/incognitochain/incognito-chain/common"
)

type BeaconChain struct {
	BestState  *BeaconBestState
	BlockGen   *BlockGenerator
	Blockchain *BlockChain
	ChainName  string
	// ChainConsensus  ConsensusInterface
	// ConsensusEngine ConsensusEngineInterface
}

func (chain *BeaconChain) GetLastBlockTimeStamp() int64 {
	// return uint64(s.Blockchain.BestState.Beacon.BestBlock.Header.Timestamp)
	return chain.BestState.BestBlock.Header.Timestamp
}

func (chain *BeaconChain) GetMinBlkInterval() time.Duration {
	return chain.BestState.BlockInterval
}

func (chain *BeaconChain) GetMaxBlkCreateTime() time.Duration {
	return chain.BestState.BlockMaxCreateTime
}

func (chain *BeaconChain) IsReady() bool {
	return chain.Blockchain.Synker.IsLatest(false, 0)
}

func (chain *BeaconChain) CurrentHeight() uint64 {
	return chain.BestState.BestBlock.Header.Height
}

func (chain *BeaconChain) GetCommittee() []string {
	return chain.BestState.GetBeaconCommittee()
}

func (chain *BeaconChain) GetCommitteeSize() int {
	return len(chain.BestState.GetBeaconCommittee())
}

func (chain *BeaconChain) GetPubKeyCommitteeIndex(pubkey string) int {
	return common.IndexOfStr(pubkey, chain.BestState.GetBeaconCommittee())
}

func (chain *BeaconChain) GetLastProposerIndex() int {
	return chain.BestState.BeaconProposerIndex
}

func (chain *BeaconChain) CreateNewBlock(round int) common.BlockInterface {
	newBlock, err := chain.BlockGen.NewBlockBeacon(round, chain.Blockchain.Synker.GetClosestShardToBeaconPoolState())
	if err != nil {
		return nil
	}
	return newBlock
}

func (chain *BeaconChain) InsertBlk(block common.BlockInterface, isValid bool) {
	chain.Blockchain.InsertBeaconBlock(block.(*BeaconBlock), isValid)
}

func (chain *BeaconChain) GetActiveShardNumber() int {
	return chain.BestState.ActiveShards
}

func (chain *BeaconChain) GetChainName() string {
	return chain.ChainName
}

func (chain *BeaconChain) GetPubkeyRole(pubkey string, round int) (string, byte) {
	return "", 0
}

func (chain *BeaconChain) ValidateBlock(block common.BlockInterface) error {
	_ = block
	return nil
}

func (chain *BeaconChain) ValidateBlockSanity(common.BlockInterface) error {
	return nil
}

func (chain *BeaconChain) ValidateBlockWithBlockChain(common.BlockInterface) error {
	return nil
}

func (chain *BeaconChain) GetConsensusType() string {
	return chain.BestState.ConsensusAlgorithm
}

func (chain *BeaconChain) GetShardID() int {
	return -1
}
