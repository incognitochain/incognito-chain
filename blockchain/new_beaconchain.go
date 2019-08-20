package blockchain

import (
	"encoding/json"
	"fmt"
	"strings"
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

func (chain *BeaconChain) ValidateAndInsertBlock(block common.BlockInterface) error {
	var beaconBestState BeaconBestState
	beaconBlock := block.(*BeaconBlock)
	chain.BestState.cloneBeaconBestState(&beaconBestState)
	producerPublicKey := beaconBlock.Header.Producer
	producerPosition := (beaconBestState.BeaconProposerIndex + beaconBlock.Header.Round) % len(beaconBestState.BeaconCommittee)
	tempProducer := beaconBestState.BeaconCommittee[producerPosition]
	if strings.Compare(tempProducer, producerPublicKey) != 0 {
		return NewBlockChainError(BeaconBlockProducerError, fmt.Errorf("Expect Producer Public Key to be equal but get %+v From Index, %+v From Header", tempProducer, producerPublicKey))
	}

	return nil
}

func (chain *BeaconChain) ValidateBlockSignatures(block common.BlockInterface, committee []string) error {
	if err := chain.Blockchain.config.ConsensusEngine.ValidateProducerSig(block, chain.GetConsensusType()); err != nil {
		return err
	}
	if err := chain.Blockchain.config.ConsensusEngine.ValidateBlockCommitteSig(block, committee, chain.GetConsensusType()); err != nil {
		return nil
	}

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

func (chain *BeaconChain) GetAllCommittees() map[string]map[string][]string {
	var result map[string]map[string][]string
	result = make(map[string]map[string][]string)

	result[chain.BestState.ConsensusAlgorithm] = make(map[string][]string)
	result[chain.BestState.ConsensusAlgorithm][common.BEACON_CHAINKEY] = append([]string{}, chain.BestState.BeaconCommittee...)
	for shardID, consensusType := range chain.BestState.ShardConsensusAlgorithm {
		if _, ok := result[consensusType]; !ok {
			result[consensusType] = make(map[string][]string)
		}
		result[consensusType][common.GetShardChainKey(shardID)] = append([]string{}, chain.BestState.ShardCommittee[shardID]...)
	}
	return result
}

func (chain *BeaconChain) UnmarshalBlock(blockString []byte) (common.BlockInterface, error) {
	var beaconBlk BeaconBlock
	err := json.Unmarshal(blockString, &beaconBlk)
	if err != nil {
		return nil, err
	}
	return beaconBlk, nil
}
