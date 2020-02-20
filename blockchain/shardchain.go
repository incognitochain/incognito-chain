package blockchain

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/pkg/errors"
)

type ShardChain struct {
	BestState  *ShardBestState
	BlockGen   *BlockGenerator
	Blockchain *BlockChain
	ChainName  string
	lock       sync.RWMutex
}

func (s *ShardChain) InsertBatchBlock([]common.BlockInterface) error {
	panic("implement me")
}

func (s *ShardChain) GetBestViewHeight() uint64 {
	return s.CurrentHeight()
}

func (s *ShardChain) GetFinalViewHeight() uint64 {
	return s.CurrentHeight()
}

func (s *ShardChain) GetBestViewHash() string {
	return s.BestState.Hash().String()
}

func (s *ShardChain) GetFinalViewHash() string {
	return s.BestState.Hash().String()
}
func (chain *ShardChain) GetLastBlockTimeStamp() int64 {
	return chain.BestState.BestBlock.Header.Timestamp
}

func (chain *ShardChain) GetMinBlkInterval() time.Duration {
	return chain.BestState.BlockInterval
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
	result := []incognitokey.CommitteePublicKey{}
	return append(result, chain.BestState.ShardCommittee...)
}

func (chain *ShardChain) GetPendingCommittee() []incognitokey.CommitteePublicKey {
	result := []incognitokey.CommitteePublicKey{}
	return append(result, chain.BestState.ShardPendingValidator...)
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
	chain.lock.Lock()
	defer chain.lock.Unlock()
	start := time.Now()
	Logger.log.Infof("Begin Create New Block %+v", start)
	beaconHeight := chain.Blockchain.Synker.States.ClosestState.ClosestBeaconState
	if chain.Blockchain.BestState.Beacon.BeaconHeight < beaconHeight {
		beaconHeight = chain.Blockchain.BestState.Beacon.BeaconHeight
	} else {
		if beaconHeight < GetBestStateShard(byte(chain.GetShardID())).BeaconHeight {
			beaconHeight = GetBestStateShard(byte(chain.GetShardID())).BeaconHeight
		}
	}
	Logger.log.Infof("Begin Enter New Block Shard %+v", time.Now())
	newBlock, err := chain.BlockGen.NewBlockShard(byte(chain.GetShardID()), round, chain.Blockchain.Synker.GetClosestCrossShardPoolState(), beaconHeight, start)
	Logger.log.Infof("Begin Finish New Block Shard %+v", time.Now())
	if err != nil {
		return nil, err
	}
	Logger.log.Infof("Finish Create New Block %+v", start)
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

func (chain *ShardChain) ValidateBlockSignatures(block common.BlockInterface, committee []incognitokey.CommitteePublicKey) error {
	if err := chain.Blockchain.config.ConsensusEngine.ValidateProducerSig(block, chain.GetConsensusType()); err != nil {
		return err
	}
	if err := chain.Blockchain.config.ConsensusEngine.ValidateBlockCommitteSig(block, committee, chain.GetConsensusType()); err != nil {
		return nil
	}
	return nil
}

func (chain *ShardChain) InsertBlk(block common.BlockInterface) error {
	if chain.Blockchain.config.ConsensusEngine.IsOngoing(chain.ChainName) {
		return NewBlockChainError(ConsensusIsOngoingError, errors.New(fmt.Sprint(chain.ChainName, block.Hash())))
	}
	return chain.Blockchain.InsertShardBlock(block.(*ShardBlock), false)
}

func (chain *ShardChain) InsertAndBroadcastBlock(block common.BlockInterface) error {
	go chain.Blockchain.config.Server.PushBlockToAll(block, false)
	err := chain.Blockchain.InsertShardBlock(block.(*ShardBlock), true)
	if err != nil {
		return err
	}
	return nil
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
