package blockchain

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
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
	// return chain.BestState.BlockInterval
	return common.MinShardBlkInterval
}

func (chain *ShardChain) GetMaxBlkCreateTime() time.Duration {
	return chain.BestState.BlockMaxCreateTime
}

func (chain *ShardChain) IsReady() bool {
	return chain.Blockchain.Synker.IsLatest(true, chain.BestState.ShardID)
}

func (chain *ShardChain) CurrentHeight() uint64 {
	return chain.BestState.BestBlock.Header.Height
}

func (chain *ShardChain) GetCommittee() []incognitokey.CommitteePublicKey {
	return chain.BestState.ShardCommittee
}

func (chain *ShardChain) GetCommitteeSize() int {
	return len(chain.BestState.ShardCommittee)
}

func (chain *ShardChain) GetPubKeyCommitteeIndex(pubkey string) int {
	for index, key := range chain.BestState.ShardCommittee {
		if key.GetMiningKeyBase58(chain.BestState.ConsensusAlgorithm) == pubkey {
			return index
		}
	}
	return -1
}

func (chain *ShardChain) GetLastProposerIndex() int {
	return chain.BestState.ShardProposerIdx
}

func (chain *ShardChain) CreateNewBlock(round int) (common.BlockInterface, error) {
	start := time.Now()
	beaconHeight := chain.Blockchain.Synker.States.ClosestState.ClosestBeaconState
	if chain.Blockchain.BestState.Beacon.BeaconHeight < beaconHeight {
		beaconHeight = chain.Blockchain.BestState.Beacon.BeaconHeight
	}
	newBlock, err := chain.BlockGen.NewBlockShard(byte(chain.GetShardID()), round, chain.Blockchain.Synker.GetClosestCrossShardPoolState(), beaconHeight, start)
	if err != nil {
		return nil, err
	}
	return newBlock, nil
}

func (chain *ShardChain) ValidateAndInsertBlock(block common.BlockInterface) error {
	//@Bahamoot review later
	var shardBestState ShardBestState
	shardBlock := block.(*ShardBlock)
	chain.BestState.cloneShardBestState(&shardBestState)
	producerPublicKey := shardBlock.Header.Producer
	producerPosition := (shardBestState.ShardProposerIdx + shardBlock.Header.Round) % len(shardBestState.ShardCommittee)
	tempProducer := beaconBestState.BeaconCommittee[producerPosition].GetMiningKeyBase58(shardBestState.ConsensusAlgorithm)
	if strings.Compare(tempProducer, producerPublicKey) != 0 {
		return NewBlockChainError(BeaconBlockProducerError, fmt.Errorf("Expect Producer Public Key to be equal but get %+v From Index, %+v From Header", tempProducer, producerPublicKey))
	}
	if err := chain.ValidateBlockSignatures(block, beaconBestState.BeaconCommittee); err != nil {
		return err
	}
	return chain.InsertBlk(block)
}

func (chain *ShardChain) ValidateBlockSignatures(block common.BlockInterface, committee []incognitokey.CommitteePublicKey) error {
	if err := chain.Blockchain.config.ConsensusEngine.ValidateProducerSig(block, chain.GetConsensusType()); err != nil {
		return err
	}
	if err := chain.Blockchain.config.ConsensusEngine.ValidateBlockCommitteSig(block, committee, chain.GetConsensusType()); err != nil {
		return nil
	}
	return nil
}

func (chain *ShardChain) ValidateBlockWithBlockChain(common.BlockInterface) error {
	return nil
}

func (chain *ShardChain) InsertBlk(block common.BlockInterface) error {
	return chain.Blockchain.InsertShardBlock(block.(*ShardBlock), false)
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
	return chain.BestState.GetPubkeyRole(pubkey, round), chain.BestState.ShardID
}

func (chain *ShardChain) UnmarshalBlock(blockString []byte) (common.BlockInterface, error) {
	var shardBlk ShardBlock
	err := json.Unmarshal(blockString, &shardBlk)
	if err != nil {
		return nil, err
	}
	return &shardBlk, nil
}

func (chain *ShardChain) ValidatePreSignBlock(block common.BlockInterface) error {
	return chain.Blockchain.VerifyPreSignShardBlock(block.(*ShardBlock), chain.BestState.ShardID)
}
