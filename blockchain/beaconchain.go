package blockchain

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/multiview"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/pkg/errors"
)

type BeaconChain struct {
	multiView *multiview.MultiView

	BlockGen   *BlockGenerator
	Blockchain *BlockChain
	ChainName  string
	lock       sync.RWMutex
}

func (chain *BeaconChain) GetBestState() *BeaconBestState {
	view := chain.multiView.GetBestView()
	return view.(*BeaconBestState)
}

func (chain *BeaconChain) GetLastBlockTimeStamp() int64 {
	return chain.multiView.GetBestView().(*BeaconBestState).BestBlock.Header.Timestamp
}

func (chain *BeaconChain) GetMinBlkInterval() time.Duration {
	return chain.multiView.GetBestView().(*BeaconBestState).BlockInterval
}

func (chain *BeaconChain) GetMaxBlkCreateTime() time.Duration {
	return chain.multiView.GetBestView().(*BeaconBestState).BlockMaxCreateTime
}

func (chain *BeaconChain) IsReady() bool {
	return chain.Blockchain.Synker.IsLatest(false, 0)
}

func (chain *BeaconChain) CurrentHeight() uint64 {
	return chain.multiView.GetBestView().(*BeaconBestState).BestBlock.Header.Height
}

func (chain *BeaconChain) GetCommittee() []incognitokey.CommitteePublicKey {
	return chain.multiView.GetBestView().(*BeaconBestState).GetBeaconCommittee()
}

func (chain *BeaconChain) GetCommitteeSize() int {
	return len(chain.multiView.GetBestView().(*BeaconBestState).BeaconCommittee)
}

func (chain *BeaconChain) GetPubKeyCommitteeIndex(pubkey string) int {
	for index, key := range chain.multiView.GetBestView().(*BeaconBestState).GetBeaconCommittee() {
		if key.GetMiningKeyBase58(chain.multiView.GetBestView().(*BeaconBestState).ConsensusAlgorithm) == pubkey {
			return index
		}
	}
	return -1
}

func (chain *BeaconChain) GetLastProposerIndex() int {
	return chain.multiView.GetBestView().(*BeaconBestState).BeaconProposerIndex
}

func (chain *BeaconChain) CreateNewBlock(round int) (common.BlockInterface, error) {
	// chain.lock.Lock()
	// defer chain.lock.Unlock()
	newBlock, err := chain.BlockGen.NewBlockBeacon(round, chain.Blockchain.Synker.GetClosestShardToBeaconPoolState())
	if err != nil {
		return nil, err
	}
	return newBlock, nil
}

func (chain *BeaconChain) InsertBlk(block common.BlockInterface) error {
	if chain.Blockchain.config.ConsensusEngine.IsOngoing(common.BeaconChainKey) {
		return NewBlockChainError(ConsensusIsOngoingError, errors.New(fmt.Sprint(common.BeaconChainKey, block.Hash())))
	}
	return chain.Blockchain.InsertBeaconBlock(block.(*BeaconBlock), true)
}

func (chain *BeaconChain) InsertAndBroadcastBlock(block common.BlockInterface) error {
	go chain.Blockchain.config.Server.PushBlockToAll(block, true)
	err := chain.Blockchain.InsertBeaconBlock(block.(*BeaconBlock), true)
	if err != nil {
		return err
	}
	return nil
}

func (chain *BeaconChain) GetActiveShardNumber() int {
	return chain.multiView.GetBestView().(*BeaconBestState).ActiveShards
}

func (chain *BeaconChain) GetChainName() string {
	return chain.ChainName
}

func (chain *BeaconChain) GetPubkeyRole(pubkey string, round int) (string, byte) {
	return chain.multiView.GetBestView().(*BeaconBestState).GetPubkeyRole(pubkey, round)
}

func (chain *BeaconChain) ValidatePreSignBlock(block common.BlockInterface) error {
	return chain.Blockchain.VerifyPreSignBeaconBlock(block.(*BeaconBlock), true)
}

// func (chain *BeaconChain) ValidateAndInsertBlock(block common.BlockInterface) error {
// 	var beaconBestState BeaconBestState
// 	beaconBlock := block.(*BeaconBlock)
// 	beaconBestState.cloneBeaconBestStateFrom(chain.multiView.GetBestView().(*BeaconBestState))
// 	producerPublicKey := beaconBlock.Header.Producer
// 	producerPosition := (beaconBestState.BeaconProposerIndex + beaconBlock.Header.Round) % len(beaconBestState.BeaconCommittee)
// 	tempProducer := beaconBestState.BeaconCommittee[producerPosition].GetMiningKeyBase58(beaconBestState.ConsensusAlgorithm)
// 	if strings.Compare(tempProducer, producerPublicKey) != 0 {
// 		return NewBlockChainError(BeaconBlockProducerError, fmt.Errorf("Expect Producer Public Key to be equal but get %+v From Index, %+v From Header", tempProducer, producerPublicKey))
// 	}
// 	if err := chain.ValidateBlockSignatures(block, beaconBestState.BeaconCommittee); err != nil {
// 		return err
// 	}
// 	return chain.Blockchain.InsertBeaconBlock(beaconBlock, false)
// }

func (chain *BeaconChain) ValidateBlockSignatures(block common.BlockInterface, committee []incognitokey.CommitteePublicKey) error {
	if err := chain.Blockchain.config.ConsensusEngine.ValidateProducerSig(block, chain.GetConsensusType()); err != nil {
		return err
	}
	if err := chain.Blockchain.config.ConsensusEngine.ValidateBlockCommitteSig(block, committee, chain.GetConsensusType()); err != nil {
		return nil
	}
	return nil
}

func (chain *BeaconChain) GetConsensusType() string {
	return chain.multiView.GetBestView().(*BeaconBestState).ConsensusAlgorithm
}

func (chain *BeaconChain) GetShardID() int {
	return -1
}

func (chain *BeaconChain) GetAllCommittees() map[string]map[string][]incognitokey.CommitteePublicKey {
	var result map[string]map[string][]incognitokey.CommitteePublicKey
	result = make(map[string]map[string][]incognitokey.CommitteePublicKey)
	result[chain.multiView.GetBestView().(*BeaconBestState).ConsensusAlgorithm] = make(map[string][]incognitokey.CommitteePublicKey)
	result[chain.multiView.GetBestView().(*BeaconBestState).ConsensusAlgorithm][common.BeaconChainKey] = append([]incognitokey.CommitteePublicKey{}, chain.multiView.GetBestView().(*BeaconBestState).BeaconCommittee...)
	for shardID, consensusType := range chain.multiView.GetBestView().(*BeaconBestState).GetShardConsensusAlgorithm() {
		if _, ok := result[consensusType]; !ok {
			result[consensusType] = make(map[string][]incognitokey.CommitteePublicKey)
		}
		result[consensusType][common.GetShardChainKey(shardID)] = append([]incognitokey.CommitteePublicKey{}, chain.multiView.GetBestView().(*BeaconBestState).ShardCommittee[shardID]...)
	}
	return result
}

func (chain *BeaconChain) GetBeaconPendingList() []incognitokey.CommitteePublicKey {
	var result []incognitokey.CommitteePublicKey
	result = append(result, chain.multiView.GetBestView().(*BeaconBestState).BeaconPendingValidator...)
	return result
}

func (chain *BeaconChain) GetShardsPendingList() map[string]map[string][]incognitokey.CommitteePublicKey {
	var result map[string]map[string][]incognitokey.CommitteePublicKey
	result = make(map[string]map[string][]incognitokey.CommitteePublicKey)
	for shardID, consensusType := range chain.multiView.GetBestView().(*BeaconBestState).GetShardConsensusAlgorithm() {
		if _, ok := result[consensusType]; !ok {
			result[consensusType] = make(map[string][]incognitokey.CommitteePublicKey)
		}
		result[consensusType][common.GetShardChainKey(shardID)] = append([]incognitokey.CommitteePublicKey{}, chain.multiView.GetBestView().(*BeaconBestState).ShardPendingValidator[shardID]...)
	}
	return result
}

func (chain *BeaconChain) GetShardsWaitingList() []incognitokey.CommitteePublicKey {
	var result []incognitokey.CommitteePublicKey
	result = append(result, chain.multiView.GetBestView().(*BeaconBestState).CandidateShardWaitingForNextRandom...)
	result = append(result, chain.multiView.GetBestView().(*BeaconBestState).CandidateShardWaitingForCurrentRandom...)
	return result
}

func (chain *BeaconChain) GetBeaconWaitingList() []incognitokey.CommitteePublicKey {
	var result []incognitokey.CommitteePublicKey
	result = append(result, chain.multiView.GetBestView().(*BeaconBestState).CandidateBeaconWaitingForNextRandom...)
	result = append(result, chain.multiView.GetBestView().(*BeaconBestState).CandidateBeaconWaitingForCurrentRandom...)
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
