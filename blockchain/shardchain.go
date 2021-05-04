package blockchain

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/incognitochain/incognito-chain/blockchain/committeestate"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/multiview"
)

type ShardChain struct {
	shardID   int
	multiView *multiview.MultiView

	BlockGen    *BlockGenerator
	Blockchain  *BlockChain
	hashHistory *lru.Cache
	ChainName   string
	Ready       bool

	insertLock sync.Mutex
}

func NewShardChain(shardID int, multiView *multiview.MultiView, blockGen *BlockGenerator, blockchain *BlockChain, chainName string) *ShardChain {
	return &ShardChain{shardID: shardID, multiView: multiView, BlockGen: blockGen, Blockchain: blockchain, ChainName: chainName}
}

func (chain *ShardChain) GetDatabase() incdb.Database {
	return chain.Blockchain.GetShardChainDatabase(byte(chain.shardID))
}

func (chain *ShardChain) GetFinalView() multiview.View {
	return chain.multiView.GetFinalView()
}

func (chain *ShardChain) GetBestView() multiview.View {
	return chain.multiView.GetBestView()
}

func (chain *ShardChain) GetViewByHash(hash common.Hash) multiview.View {
	return chain.multiView.GetViewByHash(hash)
}

func (chain *ShardChain) GetBestState() *ShardBestState {
	return chain.multiView.GetBestView().(*ShardBestState)
}

func (s *ShardChain) GetEpoch() uint64 {
	return s.GetBestState().Epoch
}

func (s *ShardChain) InsertBatchBlock([]types.BlockInterface) (int, error) {
	panic("implement me")
}

func (s *ShardChain) GetCrossShardState() map[byte]uint64 {

	res := make(map[byte]uint64)
	for index, key := range s.GetBestState().BestCrossShard {
		res[index] = key
	}
	return res
}

func (s *ShardChain) GetAllViewHash() (res []common.Hash) {
	for _, v := range s.multiView.GetAllViewsWithBFS() {
		res = append(res, *v.GetHash())
	}
	return
}

func (s *ShardChain) GetBestViewHeight() uint64 {
	return s.CurrentHeight()
}

func (s *ShardChain) GetFinalViewHeight() uint64 {
	return s.GetFinalView().GetHeight()
}

func (s *ShardChain) GetBestViewHash() string {
	return s.GetBestState().BestBlockHash.String()
}

func (s *ShardChain) GetFinalViewHash() string {
	return s.GetBestState().Hash().String()
}
func (chain *ShardChain) GetLastBlockTimeStamp() int64 {
	return chain.GetBestState().BestBlock.Header.Timestamp
}

func (chain *ShardChain) GetMinBlkInterval() time.Duration {
	return chain.GetBestState().BlockInterval
}

func (chain *ShardChain) GetMaxBlkCreateTime() time.Duration {
	return chain.GetBestState().BlockMaxCreateTime
}

func (chain *ShardChain) IsReady() bool {
	return chain.Ready
}

func (chain *ShardChain) SetReady(ready bool) {
	chain.Ready = ready
}

func (chain *ShardChain) CurrentHeight() uint64 {
	return chain.GetBestState().BestBlock.Header.Height
}

func (chain *ShardChain) GetCommittee() []incognitokey.CommitteePublicKey {
	result := []incognitokey.CommitteePublicKey{}
	return append(result, chain.GetBestState().shardCommitteeState.GetShardCommittee()...)
}

func (chain *ShardChain) GetLastCommittee() []incognitokey.CommitteePublicKey {
	v := chain.multiView.GetViewByHash(*chain.GetBestView().GetPreviousHash())
	if v == nil {
		return nil
	}
	result := []incognitokey.CommitteePublicKey{}
	return append(result, v.GetCommittee()...)
}

func (chain *ShardChain) GetCommitteeByHeight(h uint64) ([]incognitokey.CommitteePublicKey, error) {
	bcStateRootHash := chain.Blockchain.GetBeaconBestState().ConsensusStateDBRootHash
	bcDB := chain.Blockchain.GetBeaconChainDatabase()
	bcStateDB, err := statedb.NewWithPrefixTrie(bcStateRootHash, statedb.NewDatabaseAccessWarper(bcDB))
	if err != nil {
		return nil, err
	}
	return statedb.GetOneShardCommittee(bcStateDB, byte(chain.shardID)), nil
}

func (chain *ShardChain) GetPendingCommittee() []incognitokey.CommitteePublicKey {
	result := []incognitokey.CommitteePublicKey{}
	return append(result, chain.GetBestState().shardCommitteeState.GetShardSubstitute()...)
}

func (chain *ShardChain) GetCommitteeSize() int {
	return len(chain.GetBestState().shardCommitteeState.GetShardCommittee())
}

func (chain *ShardChain) GetPubKeyCommitteeIndex(pubkey string) int {
	for index, key := range chain.GetBestState().shardCommitteeState.GetShardCommittee() {
		if key.GetMiningKeyBase58(chain.GetBestState().ConsensusAlgorithm) == pubkey {
			return index
		}
	}
	return -1
}

func (chain *ShardChain) GetLastProposerIndex() int {
	return chain.GetBestState().ShardProposerIdx
}

func (chain *ShardChain) CreateNewBlock(
	version int, proposer string, round int, startTime int64,
	committees []incognitokey.CommitteePublicKey,
	committeeViewHash common.Hash) (types.BlockInterface, error) {
	Logger.log.Infof("Begin Start New Block Shard %+v", time.Now())
	newBlock, err := chain.Blockchain.NewBlockShard(
		chain.GetBestState(),
		version, proposer, round,
		startTime, committees, committeeViewHash)
	Logger.log.Infof("Finish New Block Shard %+v", time.Now())
	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}
	if version >= 2 {
		newBlock.Header.Proposer = proposer
		newBlock.Header.ProposeTime = startTime
	}
	Logger.log.Infof("[dcs] new block header proposer %v proposerTime %v", newBlock.Header.Proposer, newBlock.Header.ProposeTime)

	Logger.log.Infof("Finish Create New Block")
	return newBlock, nil
}

func (chain *ShardChain) CreateNewBlockFromOldBlock(
	oldBlock types.BlockInterface,
	proposer string, startTime int64,
	committees []incognitokey.CommitteePublicKey,
	committeeViewHash common.Hash,
) (types.BlockInterface, error) {
	b, _ := json.Marshal(oldBlock)
	newBlock := new(types.ShardBlock)
	json.Unmarshal(b, &newBlock)

	newBlock.Header.Proposer = proposer
	newBlock.Header.ProposeTime = startTime

	return newBlock, nil
}

// func (chain *ShardChain) ValidateAndInsertBlock(block common.BlockInterface) error {
// 	//@Bahamoot review later
// 	chain.lock.Lock()
// 	defer chain.lock.Unlock()
// 	var shardBestState ShardBestState
// 	shardBlock := block.(*ShardBlock)
// 	shardBestState.cloneShardBestStateFrom(chain.BestState)
// 	producerPublicKey := shardBlock.Header.Producer
// 	producerPosition := (shardBestState.ShardProposerIdx + shardBlock.Header.Round) % len(shardBestState.ShardCommittee)
// 	tempProducer := shardBestState.ShardCommittee[producerPosition].GetMiningKeyBase58(shardBestState.ConsensusAlgorithm)
// 	if strings.Compare(tempProducer, producerPublicKey) != 0 {
// 		return NewBlockChainError(BeaconBlockProducerError, fmt.Errorf("Expect Producer Public Key to be equal but get %+v From Index, %+v From Header", tempProducer, producerPublicKey))
// 	}
// 	if err := chain.ValidateBlockSignatures(block, shardBestState.ShardCommittee); err != nil {
// 		return err
// 	}
// 	return chain.Blockchain.InsertShardBlock(shardBlock, false)
// }

func (chain *ShardChain) ValidateBlockSignatures(block types.BlockInterface, committees []incognitokey.CommitteePublicKey) error {
	if err := chain.Blockchain.config.ConsensusEngine.ValidateProducerSig(block, chain.GetConsensusType()); err != nil {
		return err
	}
	if err := chain.Blockchain.config.ConsensusEngine.ValidateBlockCommitteSig(block, committees); err != nil {
		return err
	}
	return nil
}

func (chain *ShardChain) InsertBlock(block types.BlockInterface, shouldValidate bool) error {
	err := chain.Blockchain.InsertShardBlock(block.(*types.ShardBlock), shouldValidate)
	if err != nil {
		Logger.log.Error(err)
		return err
	}

	return nil
}

func (chain *ShardChain) ReplacePreviousValidationData(previousBlockHash common.Hash, newValidationData string) error {
	if err := chain.Blockchain.ReplacePreviousValidationData(previousBlockHash, newValidationData); err != nil {
		Logger.log.Error(err)
		return err
	}

	return nil
}

func (chain *ShardChain) CheckExistedBlk(block types.BlockInterface) bool {
	blkHash := block.Hash()
	_, err := rawdbv2.GetShardBlockByHash(chain.Blockchain.GetShardChainDatabase(byte(chain.shardID)), *blkHash)
	return err == nil
}

func (chain *ShardChain) GetActiveShardNumber() int {
	return 0
}

func (chain *ShardChain) GetChainName() string {
	return chain.ChainName
}

func (chain *ShardChain) GetConsensusType() string {
	return chain.GetBestState().ConsensusAlgorithm
}

func (chain *ShardChain) GetShardID() int {
	return chain.shardID
}

func (chain *ShardChain) IsBeaconChain() bool {
	return false
}

func (chain *ShardChain) UnmarshalBlock(blockString []byte) (types.BlockInterface, error) {
	var shardBlk types.ShardBlock
	err := json.Unmarshal(blockString, &shardBlk)
	if err != nil {
		return nil, err
	}
	return &shardBlk, nil
}

func (chain *ShardChain) ValidatePreSignBlock(block types.BlockInterface, signingCommittees, committees []incognitokey.CommitteePublicKey) error {
	return chain.Blockchain.VerifyPreSignShardBlock(block.(*types.ShardBlock), signingCommittees, committees, byte(block.(*types.ShardBlock).GetShardID()))
}

func (chain *ShardChain) GetAllView() []multiview.View {
	return chain.multiView.GetAllViewsWithBFS()
}

//CommitteesV2 get committees by block for shardChain
// Input block must be ShardBlock
func (chain *ShardChain) GetCommitteeV2(block types.BlockInterface) ([]incognitokey.CommitteePublicKey, error) {
	var err error
	var isShardView bool
	var shardView *ShardBestState
	shardView, isShardView = chain.GetViewByHash(block.GetPrevHash()).(*ShardBestState)
	if !isShardView {
		shardView = chain.GetBestState()
	}
	committees := []incognitokey.CommitteePublicKey{}

	shardBlock, isShardBlock := block.(*types.ShardBlock)
	if !isShardBlock {
		return committees, fmt.Errorf("Shard Chain NOT insert Shard Block Types")
	}
	if shardView.shardCommitteeState.Version() == committeestate.SELF_SWAP_SHARD_VERSION || shardBlock.Header.CommitteeFromBlock.IsZeroValue() {
		committees = append(committees, chain.GetBestState().shardCommitteeState.GetShardCommittee()...)
	} else {
		committees, err = chain.Blockchain.getShardCommitteeFromBeaconHash(block.CommitteeFromBlock(), byte(chain.shardID))
		if err != nil {
			return committees, err
		}
	}

	return committees, nil
}

func (chain *ShardChain) CommitteeStateVersion() int {
	return chain.GetBestState().shardCommitteeState.Version()
}

//BestViewCommitteeFromBlock ...
func (chain *ShardChain) BestViewCommitteeFromBlock() common.Hash {
	return chain.GetBestState().CommitteeFromBlock()
}

func (chain *ShardChain) GetChainDatabase() incdb.Database {
	return chain.Blockchain.GetShardChainDatabase(byte(chain.shardID))
}

func (chain *ShardChain) CommitteeEngineVersion() int {
	return chain.multiView.GetBestView().CommitteeStateVersion()
}

//ProposerByTimeSlot ...
func (chain *ShardChain) GetProposerByTimeSlot(
	committeeViewHash common.Hash, shardID byte, ts int64,
	committees []incognitokey.CommitteePublicKey) (incognitokey.CommitteePublicKey, int, error) {
	lenProposers := 0
	beaconHeight := chain.Blockchain.BeaconChain.GetFinalView().GetHeight()
	if beaconHeight >= chain.Blockchain.config.ChainParams.StakingFlowV2Height {
		tempCommitteeInfo, err := chain.Blockchain.getTempCommitteeInfoByHash(committeeViewHash, shardID)
		if err != nil {
			return incognitokey.CommitteePublicKey{}, -1, err
		}
		beaconHeight = tempCommitteeInfo.BeaconHeight()
	}
	if beaconHeight >= chain.Blockchain.config.ChainParams.BlockProducingV3Height {
		lenProposers = chain.Blockchain.config.ChainParams.NumberOfShardFixedBlockValidators
	} else {
		lenProposers = chain.Blockchain.GetBestStateShard(shardID).MinShardCommitteeSize
	}
	id := GetProposerByTimeSlot(ts, lenProposers)
	return committees[id], id, nil
}

func (chain *ShardChain) SigningCommittees(
	committeeViewHash common.Hash, proposerIndex int, committees []incognitokey.CommitteePublicKey, shardID byte,
) []incognitokey.CommitteePublicKey {
	res := []incognitokey.CommitteePublicKey{}
	beaconHeight := chain.Blockchain.BeaconChain.GetFinalView().GetHeight()
	if beaconHeight >= chain.Blockchain.config.ChainParams.StakingFlowV2Height {
		tempCommitteeInfo, err := chain.Blockchain.getTempCommitteeInfoByHash(committeeViewHash, shardID)
		if err != nil {
			return []incognitokey.CommitteePublicKey{}
		}
		beaconHeight = tempCommitteeInfo.BeaconHeight()
	}

	if beaconHeight >= chain.Blockchain.config.ChainParams.BlockProducingV3Height {
		for i, v := range committees {
			if (i % MaxSubsetCommittees) == (proposerIndex % MaxSubsetCommittees) {
				res = append(res, v)
			}
		}
	} else {
		res = committees
	}
	return res
}

func (chain *ShardChain) CommitteesFromViewHashForShard(
	committeeHash common.Hash, shardID byte,
) ([]incognitokey.CommitteePublicKey, error) {
	return []incognitokey.CommitteePublicKey{}, nil
}

func (chain *ShardChain) GetSigningCommittees(
	committees []incognitokey.CommitteePublicKey,
	block types.BlockInterface,
) ([]incognitokey.CommitteePublicKey, error) {
	return chain.Blockchain.getCommitteesForSigning(committees, block)
}
