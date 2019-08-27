package blockchain

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type BeaconChain struct {
	BestState  *BeaconBestState
	BlockGen   *BlockGenerator
	Blockchain *BlockChain
	ChainName  string
	lock       sync.RWMutex
}

func (chain *BeaconChain) GetLastBlockTimeStamp() int64 {
	chain.BestState.lock.RLock()
	defer chain.BestState.lock.RUnlock()
	return chain.BestState.BestBlock.Header.Timestamp
}

func (chain *BeaconChain) GetMinBlkInterval() time.Duration {
	chain.BestState.lock.RLock()
	defer chain.BestState.lock.RUnlock()
	// return chain.BestState.BlockInterval
	return common.MinBeaconBlkInterval
}

func (chain *BeaconChain) GetMaxBlkCreateTime() time.Duration {
	chain.BestState.lock.RLock()
	defer chain.BestState.lock.RUnlock()
	return chain.BestState.BlockMaxCreateTime
}

func (chain *BeaconChain) IsReady() bool {
	return chain.Blockchain.Synker.IsLatest(false, 0)
}

func (chain *BeaconChain) CurrentHeight() uint64 {
	chain.BestState.lock.RLock()
	defer chain.BestState.lock.RUnlock()
	return chain.BestState.BestBlock.Header.Height
}

func (chain *BeaconChain) GetCommittee() []incognitokey.CommitteePublicKey {
	chain.BestState.lock.RLock()
	defer chain.BestState.lock.RUnlock()
	return chain.BestState.GetBeaconCommittee()
}

func (chain *BeaconChain) GetCommitteeSize() int {
	chain.BestState.lock.RLock()
	defer chain.BestState.lock.RUnlock()
	return len(chain.BestState.GetBeaconCommittee())
}

func (chain *BeaconChain) GetPubKeyCommitteeIndex(pubkey string) int {
	for index, key := range chain.BestState.GetBeaconCommittee() {
		if key.GetMiningKeyBase58(chain.BestState.ConsensusAlgorithm) == pubkey {
			return index
		}
	}
	return -1
}

func (chain *BeaconChain) GetLastProposerIndex() int {
	chain.BestState.lock.RLock()
	defer chain.BestState.lock.RUnlock()
	return chain.BestState.BeaconProposerIndex
}

func (chain *BeaconChain) CreateNewBlock(round int) (common.BlockInterface, error) {
	chain.lock.Lock()
	defer chain.lock.Unlock()
	newBlock, err := chain.BlockGen.NewBlockBeacon(round, chain.Blockchain.Synker.GetClosestShardToBeaconPoolState())
	if err != nil {
		return nil, err
	}
	return newBlock, nil
}

func (chain *BeaconChain) InsertBlk(block common.BlockInterface) error {
	chain.lock.Lock()
	defer chain.lock.Unlock()
	return chain.Blockchain.InsertBeaconBlock(block.(*BeaconBlock), true)
}

func (chain *BeaconChain) GetActiveShardNumber() int {
	chain.BestState.lock.RLock()
	defer chain.BestState.lock.RUnlock()
	return chain.BestState.ActiveShards
}

func (chain *BeaconChain) GetChainName() string {
	return chain.ChainName
}

func (chain *BeaconChain) GetPubkeyRole(pubkey string, round int) (string, byte) {
	chain.BestState.lock.RLock()
	defer chain.BestState.lock.RUnlock()
	return chain.BestState.GetPubkeyRole(pubkey, round)
}

func (chain *BeaconChain) ValidatePreSignBlock(block common.BlockInterface) error {
	return chain.Blockchain.VerifyPreSignBeaconBlock(block.(*BeaconBlock), true)
}

func (chain *BeaconChain) ValidateAndInsertBlock(block common.BlockInterface) error {
	chain.lock.Lock()
	defer chain.lock.Unlock()
	var beaconBestState BeaconBestState
	beaconBlock := block.(*BeaconBlock)
	chain.BestState.cloneBeaconBestState(&beaconBestState)
	producerPublicKey := beaconBlock.Header.Producer
	producerPosition := (beaconBestState.BeaconProposerIndex + beaconBlock.Header.Round) % len(beaconBestState.BeaconCommittee)
	tempProducer := beaconBestState.BeaconCommittee[producerPosition].GetMiningKeyBase58(beaconBestState.ConsensusAlgorithm)
	if strings.Compare(tempProducer, producerPublicKey) != 0 {
		return NewBlockChainError(BeaconBlockProducerError, fmt.Errorf("Expect Producer Public Key to be equal but get %+v From Index, %+v From Header", tempProducer, producerPublicKey))
	}
	if err := chain.ValidateBlockSignatures(block, beaconBestState.BeaconCommittee); err != nil {
		return err
	}
	return chain.Blockchain.InsertBeaconBlock(beaconBlock, true)
}

func (chain *BeaconChain) ValidateBlockSignatures(block common.BlockInterface, committee []incognitokey.CommitteePublicKey) error {
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
	chain.BestState.lock.RLock()
	defer chain.BestState.lock.RUnlock()
	return chain.BestState.ConsensusAlgorithm
}

func (chain *BeaconChain) GetShardID() int {
	return -1
}

func (chain *BeaconChain) GetAllCommittees() map[string]map[string][]incognitokey.CommitteePublicKey {
	chain.BestState.lock.RLock()
	defer chain.BestState.lock.RUnlock()
	var result map[string]map[string][]incognitokey.CommitteePublicKey
	result = make(map[string]map[string][]incognitokey.CommitteePublicKey)
	result[chain.BestState.ConsensusAlgorithm] = make(map[string][]incognitokey.CommitteePublicKey)
	result[chain.BestState.ConsensusAlgorithm][common.BEACON_CHAINKEY] = append([]incognitokey.CommitteePublicKey{}, chain.BestState.BeaconCommittee...)
	for shardID, consensusType := range chain.BestState.ShardConsensusAlgorithm {
		if _, ok := result[consensusType]; !ok {
			result[consensusType] = make(map[string][]incognitokey.CommitteePublicKey)
		}
		result[consensusType][common.GetShardChainKey(shardID)] = append([]incognitokey.CommitteePublicKey{}, chain.BestState.ShardCommittee[shardID]...)
	}
	return result
}

func (chain *BeaconChain) UnmarshalBlock(blockString []byte) (common.BlockInterface, error) {
	var beaconBlk BeaconBlock
	err := json.Unmarshal(blockString, &beaconBlk)
	if err != nil {
		return nil, err
	}
	return &beaconBlk, nil
}
