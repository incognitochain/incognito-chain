package blockchain

import (
	"encoding/json"
	"time"

	"github.com/incognitochain/incognito-chain/common"
)

type ShardChain struct {
	BestState  *ShardBestState
	BlockGen   *BlockGenerator
	Blockchain *BlockChain
	ChainName  string
	// ChainConsensus  ConsensusInterface
	// ConsensusEngine ConsensusEngineInterface
}

func (chain *ShardChain) GetLastBlockTimeStamp() int64 {
	// return uint64(s.Blockchain.BestState.Beacon.BestBlock.Header.Timestamp)
	return chain.BestState.BestBlock.Header.Timestamp
}

func (chain *ShardChain) GetMinBlkInterval() time.Duration {
	return chain.BestState.BlockInterval
}

func (chain *ShardChain) GetMaxBlkCreateTime() time.Duration {
	return chain.BestState.BlockMaxCreateTime
}

func (chain *ShardChain) IsReady() bool {
	return chain.Blockchain.Synker.IsLatest(false, 0)
}

func (chain *ShardChain) CurrentHeight() uint64 {
	return chain.BestState.BestBlock.Header.Height
}

func (chain *ShardChain) GetCommittee() []string {
	return chain.BestState.ShardCommittee
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

func (chain *ShardChain) CreateNewBlock(round int) common.BlockInterface {
	newBlock, err := chain.BlockGen.NewBlockBeacon(round, chain.Blockchain.Synker.GetClosestShardToBeaconPoolState())
	if err != nil {
		return nil
	}
	return newBlock
}

func (chain *ShardChain) ValidateAndInsertBlock(block common.BlockInterface) error {
	_ = block
	return nil
}

func (chain *ShardChain) ValidateBlockWithBlockChain(common.BlockInterface) error {
	return nil
}

func (chain *ShardChain) InsertBlk(block common.BlockInterface, isValid bool) {
	chain.Blockchain.InsertShardBlock(block.(*ShardBlock), isValid)
}

func (chain *ShardChain) GetActiveShardNumber() int {
	return 0
}

func (chain *ShardChain) GetChainName() string {
	return chain.ChainName
}

func (chain *ShardChain) GetConsensusType() string {
	return chain.BestState.ConsensusAlgorithm
}

func (chain *ShardChain) GetShardID() int {
	return int(chain.BestState.ShardID)
}

func (chain *ShardChain) GetPubkeyRole(pubkey string, round int) (string, byte) {
	chain.GetCommittee()
	return "", 0
}

func (chain *ShardChain) UnmarshalBlock(blockString []byte) (common.BlockInterface, error) {
	var shardBlk ShardBlock
	err := json.Unmarshal(blockString, &shardBlk)
	if err != nil {
		return nil, err
	}
	return shardBlk, nil
}
