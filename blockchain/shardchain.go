package blockchain

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdb_consensus"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/txpool"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/multiview"
	"github.com/incognitochain/incognito-chain/portal/portalv4"
)

type ShardChain struct {
	shardID   int
	multiView multiview.MultiView

	BlockGen    *BlockGenerator
	Blockchain  *BlockChain
	hashHistory *lru.Cache
	ChainName   string
	Ready       bool

	TxPool      txpool.TxPool
	TxsVerifier txpool.TxVerifier

	insertLock sync.Mutex
}

func NewShardChain(
	shardID int,
	multiView multiview.MultiView,
	blockGen *BlockGenerator,
	blockchain *BlockChain,
	chainName string,
	tp txpool.TxPool,
	tv txpool.TxVerifier,
) *ShardChain {
	return &ShardChain{
		shardID:     shardID,
		multiView:   multiView,
		BlockGen:    blockGen,
		Blockchain:  blockchain,
		ChainName:   chainName,
		TxPool:      tp,
		TxsVerifier: tv,
	}
}

func (chain *ShardChain) GetDatabase() incdb.Database {
	return chain.Blockchain.GetShardChainDatabase(byte(chain.shardID))
}

func (chain *ShardChain) GetMultiView() multiview.MultiView {
	return chain.multiView
}

func (chain *ShardChain) CloneMultiView() multiview.MultiView {
	return chain.multiView.Clone()
}

func (chain *ShardChain) SetMultiView(multiView multiview.MultiView) {
	chain.multiView = multiView
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

func (chain *ShardChain) AddView(view multiview.View) bool {
	curBestView := chain.multiView.GetBestView()
	added, err := chain.multiView.(*multiview.ShardMultiView).AddView(view)
	if (curBestView != nil) && (added == 1) {
		go func(chain *ShardChain, curBestView multiview.View) {
			sBestView := chain.GetBestState()
			if (time.Now().Unix() - sBestView.GetBlockTime()) > (int64(15 * common.TIMESLOT)) {
				return
			}
			if (curBestView.GetHash().String() != sBestView.GetHash().String()) && (chain.TxPool != nil) {
				bcHash := sBestView.GetBeaconHash()
				includePdexv3Tx := false
				if sBestView.BeaconHeight >= config.Param().PDexParams.Pdexv3BreakPointHeight {
					includePdexv3Tx = true
				}
				bcView, err := chain.Blockchain.GetBeaconViewStateDataFromBlockHash(bcHash, true, includePdexv3Tx)
				if err != nil {
					Logger.log.Errorf("Can not get beacon view from hash %, sView Hash %v, err %v", bcHash.String(), sBestView.GetHash().String(), err)
				} else {
					chain.TxPool.FilterWithNewView(chain.Blockchain, sBestView, bcView)
				}
			}
		}(chain, curBestView)
	}
	return err == nil
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
	if version >= types.MULTI_VIEW_VERSION {
		newBlock.Header.Proposer = proposer
		newBlock.Header.ProposeTime = startTime
	}

	if version >= types.LEMMA2_VERSION {
		previousBlock, err := chain.GetBlockByHash(newBlock.Header.PreviousBlockHash)
		if err != nil {
			return nil, err
		}
		previousProposeTimeSlot := common.CalculateTimeSlot(previousBlock.GetProposeTime())
		currentTimeSlot := common.CalculateTimeSlot(newBlock.Header.ProposeTime)

		if newBlock.Header.Timestamp == newBlock.Header.ProposeTime &&
			newBlock.Header.Producer == newBlock.Header.Proposer &&
			previousProposeTimeSlot+1 == currentTimeSlot {
			newBlock.Header.FinalityHeight = newBlock.Header.Height - 1
		} else {
			newBlock.Header.FinalityHeight = 0
		}
	}

	Logger.log.Infof("[dcs] new block header proposer %v proposerTime %v", newBlock.Header.Proposer, newBlock.Header.ProposeTime)

	Logger.log.Infof("Finish Create New Block")
	return newBlock, nil
}

func (chain *ShardChain) CreateNewBlockFromOldBlock(oldBlock types.BlockInterface, proposer string, startTime int64, isValidRePropose bool) (types.BlockInterface, error) {
	b, _ := json.Marshal(oldBlock)
	newBlock := new(types.ShardBlock)
	json.Unmarshal(b, &newBlock)

	newBlock.Header.Proposer = proposer
	newBlock.Header.ProposeTime = startTime
	if newBlock.Header.Version >= types.LEMMA2_VERSION {

		if isValidRePropose {
			newBlock.Header.FinalityHeight = newBlock.Header.Height - 1
		} else {
			newBlock.Header.FinalityHeight = 0
		}
	}

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

func (chain *ShardChain) InsertAndBroadcastBlock(block types.BlockInterface) error {

	go chain.Blockchain.config.Server.PushBlockToAll(block, "", false)

	if err := chain.InsertBlock(block, false); err != nil {
		return err
	}

	return nil
}

func (chain *ShardChain) InsertWithPrevValidationData(block types.BlockInterface, newValidationData string) error {

	if err := chain.InsertBlock(block, false); err != nil {
		return err
	}

	if err := chain.ReplacePreviousValidationData(block.GetPrevHash(), newValidationData); err != nil {
		return err
	}

	return nil
}

func (chain *ShardChain) InsertAndBroadcastBlockWithPrevValidationData(block types.BlockInterface, newValidationData string) error {

	go chain.Blockchain.config.Server.PushBlockToAll(block, newValidationData, false)

	return chain.InsertWithPrevValidationData(block, newValidationData)
}

func (chain *ShardChain) ReplacePreviousValidationData(previousBlockHash common.Hash, newValidationData string) error {
	if err := chain.Blockchain.ReplacePreviousValidationData(previousBlockHash, newValidationData); err != nil {
		Logger.log.Error(err)
		return err
	}

	return nil
}

func (chain *ShardChain) GetBlockByHash(hash common.Hash) (types.BlockInterface, error) {
	block, _, err := chain.Blockchain.GetShardBlockByHash(hash)
	return block, err
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

func (chain *ShardChain) GetPortalParamsV4(beaconHeight uint64) portalv4.PortalParams {
	return chain.Blockchain.GetPortalParamsV4(beaconHeight)
}

//CommitteesV2 get committees by block for shardChain
// Input block must be ShardBlock
func (chain *ShardChain) GetCommitteeV2(block types.BlockInterface) ([]incognitokey.CommitteePublicKey, error) {
	var isShardView bool
	var shardView *ShardBestState
	shardView, isShardView = chain.GetViewByHash(block.GetPrevHash()).(*ShardBestState)
	if !isShardView {
		shardView = chain.GetBestState()
	}
	shardBlock, isShardBlock := block.(*types.ShardBlock)
	if !isShardBlock {
		return []incognitokey.CommitteePublicKey{}, fmt.Errorf("Shard Chain NOT insert Shard Block Types")
	}
	_, signingCommittees, err := shardView.getSigningCommittees(shardBlock, chain.Blockchain)
	if err != nil {
		return []incognitokey.CommitteePublicKey{}, err
	}
	return signingCommittees, nil
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
func (chain *ShardChain) GetProposerByTimeSlotFromCommitteeList(ts int64, committees []incognitokey.CommitteePublicKey) (incognitokey.CommitteePublicKey, int) {
	proposer, proposerIndex := GetProposer(
		ts,
		committees,
		chain.GetBestState().GetProposerLength(),
	)
	return proposer, proposerIndex
}

func (chain *ShardChain) GetSigningCommittees(
	proposerIndex int, committees []incognitokey.CommitteePublicKey, blockVersion int,
) []incognitokey.CommitteePublicKey {
	res := []incognitokey.CommitteePublicKey{}
	if blockVersion >= types.BLOCK_PRODUCINGV3_VERSION {
		res = FilterSigningCommitteeV3(committees, proposerIndex)
	} else {
		res = append(res, committees...)
	}
	return res
}

func (chain *ShardChain) StoreFinalityProof(block types.BlockInterface, finalityProof interface{}, reProposeSig interface{}) error {
	err := rawdb_consensus.StoreShardFinalityProof(
		rawdb_consensus.GetConsensusDatabase(),
		byte(chain.shardID),
		*block.Hash(),
		block.GetPrevHash(),
		finalityProof,
		reProposeSig,
		block.GetAggregateRootHash(),
		block.GetProducer(),
		common.CalculateTimeSlot(block.GetProduceTime()),
		block.GetProposer(),
		common.CalculateTimeSlot(block.GetProposeTime()),
	)
	if err != nil {
		return err
	}

	return nil
}

func (chain *ShardChain) GetFinalityProof(hash common.Hash) (*types.ShardBlock, map[string]interface{}, error) {

	shardBlock, err := chain.GetBlockByHash(hash)
	if err != nil {
		return nil, nil, err
	}

	m, err := rawdb_consensus.GetShardFinalityProof(rawdb_consensus.GetConsensusDatabase(), byte(chain.shardID), hash)
	if err != nil {
		return nil, nil, err
	}

	return shardBlock.(*types.ShardBlock), m, nil
}
