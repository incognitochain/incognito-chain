package blockchain

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/multiview"
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
	Ready      bool
}

func (chain *BeaconChain) GetBestView() multiview.View {
	return chain.multiView.GetBestView()
}

func (chain *BeaconChain) GetFinalView() multiview.View {
	return chain.multiView.GetFinalView()
}

func (chain *BeaconChain) GetViewByHash(hash common.Hash) multiview.View {
	if chain.multiView.GetViewByHash(hash) == nil {
		return nil
	}
	return chain.multiView.GetViewByHash(hash)
}

func (chain *BeaconChain) InsertBatchBlock(blocks []common.BlockInterface) (int, error) {

	curEpoch := chain.GetBestView().(*BeaconBestState).Epoch
	sameCommitteeBlock := blocks
	for i, v := range blocks {
		if v.GetCurrentEpoch() == curEpoch+1 {
			sameCommitteeBlock = blocks[:i+1]
			break
		}
	}

	for i, blk := range sameCommitteeBlock {
		if i == len(sameCommitteeBlock)-1 {
			break
		}
		if blk.GetHeight() != sameCommitteeBlock[i+1].GetHeight()-1 {
			sameCommitteeBlock = blocks[:i+1]
			break
		}
	}

	for i := len(sameCommitteeBlock) - 1; i >= 0; i-- {
		if err := chain.ValidateBlockSignatures(sameCommitteeBlock[i], chain.GetCommittee()); err != nil {
			sameCommitteeBlock = sameCommitteeBlock[:i]
		} else {
			break
		}
	}

	if len(sameCommitteeBlock) > 0 {
		if sameCommitteeBlock[0].GetHeight()-1 != chain.CurrentHeight() {
			return 0, errors.New(fmt.Sprintf("Not expected height: %d %d", sameCommitteeBlock[0].GetHeight()-1, chain.CurrentHeight()))
		}
	}

	for _, v := range sameCommitteeBlock {
		err := chain.InsertBlk(v)
		if err != nil {
			return 0, err
		}
	}
	return len(sameCommitteeBlock), nil
}

func (s *BeaconChain) GetShardBestViewHash() map[byte]common.Hash {
	return s.GetBestView().(*BeaconBestState).GetBestShardHash()
}

func (s *BeaconChain) GetShardBestViewHeight() map[byte]uint64 {
	return s.GetBestView().(*BeaconBestState).GetBestShardHeight()
}

func (s *BeaconChain) GetCurrentCrossShardHeightToShard(sid byte) map[byte]uint64 {

	res := make(map[byte]uint64)
	for fromShard, toShardStatus := range s.GetBestView().(*BeaconBestState).LastCrossShardState {
		for toShard, currentHeight := range toShardStatus {
			if toShard == sid {
				res[fromShard] = currentHeight
			}
		}
	}
	return res
}

func (s *BeaconChain) GetEpoch() uint64 {
	return s.GetBestView().(*BeaconBestState).Epoch
}

func (s *BeaconChain) GetBestViewHeight() uint64 {
	return s.GetBestView().(*BeaconBestState).BeaconHeight
}

func (s *BeaconChain) GetFinalViewHeight() uint64 {
	return s.GetBestView().(*BeaconBestState).BeaconHeight
}

func (s *BeaconChain) GetBestViewHash() string {
	return s.GetBestView().(*BeaconBestState).BestBlockHash.String()
}

func (s *BeaconChain) GetFinalViewHash() string {
	return s.GetBestView().(*BeaconBestState).BestBlockHash.String()
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
	return chain.Ready
}
func (chain *BeaconChain) SetReady(ready bool) {
	chain.Ready = ready
}

func (chain *BeaconChain) CurrentHeight() uint64 {
	return chain.multiView.GetBestView().(*BeaconBestState).BestBlock.Header.Height
}

func (chain *BeaconChain) GetCommittee() []incognitokey.CommitteePublicKey {
	return chain.multiView.GetBestView().(*BeaconBestState).GetBeaconCommittee()
}

func (chain *BeaconChain) GetPendingCommittee() []incognitokey.CommitteePublicKey {
	return chain.GetBestView().(*BeaconBestState).GetBeaconPendingValidator()
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

func (chain *BeaconChain) CreateNewBlock(version int, proposer string, round int) (common.BlockInterface, error) {
	newBlock, err := chain.Blockchain.NewBlockBeacon_V2(chain.GetBestView().(*BeaconBestState), version, proposer, round, nil)
	if err != nil {
		return nil, err
	}
	if version == 2 {
		newBlock.Header.Proposer = proposer
		newBlock.Header.ProposeTime = time.Now().Unix()
	}

	return newBlock, nil
}

//this function for version 2
func (chain *BeaconChain) CreateNewBlockFromOldBlock(oldBlock common.BlockInterface, proposer string) (common.BlockInterface, error) {
	b, _ := json.Marshal(oldBlock)
	newBlock := new(BeaconBlock)
	json.Unmarshal(b, &newBlock)
	newBlock.Header.Proposer = proposer
	newBlock.Header.ProposeTime = time.Now().Unix()
	return oldBlock, nil
}

func (chain *BeaconChain) InsertBlk(block common.BlockInterface) error {
	//if chain.Blockchain.config.ConsensusEngine.IsOngoing(common.BeaconChainKey) {
	//	return NewBlockChainError(ConsensusIsOngoingError, errors.New(fmt.Sprint(common.BeaconChainKey, block.Hash())))
	//}
	return chain.Blockchain.InsertBeaconBlock(block.(*BeaconBlock), true)
	if err := chain.Blockchain.InsertBeaconBlock_V2(block.(*BeaconBlock), true); err != nil {
		Logger.log.Info(err)
		return err
	}
	return nil
}

func (chain *BeaconChain) InsertAndBroadcastBlock(block common.BlockInterface) error {
	go chain.Blockchain.config.Server.PushBlockToAll(block, true)
	err := chain.Blockchain.InsertBeaconBlock(block.(*BeaconBlock), true)
	if err := chain.Blockchain.InsertBeaconBlock_V2(block.(*BeaconBlock), true); err != nil {
		Logger.log.Info(err)
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
	return chain.Blockchain.ValidateProposedBeaconBlock_V2(block.(*BeaconBlock))
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
	if err := chain.Blockchain.config.ConsensusEngine.ValidateProducerPosition(block, committee); err != nil {
		return err
	}

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
