package blockchain

import (
	"fmt"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/pkg/errors"
)

func (blockchain *BlockChain) ValidateProposedShardBlock_V2(shardProposedBlock *ShardBlock) (err error) {
	//check prev view exit
	preViewHash := shardProposedBlock.GetPrevHash()
	preView := blockchain.ShardChain[shardProposedBlock.Header.ShardID].GetViewByHash(preViewHash)
	if preView == nil {
		return fmt.Errorf("Cannot find view %+v", preViewHash.String())
	}

	//create block and compare content
	shardFlow := &ShardProcessState{
		curView:             preView.(*ShardBestState),
		newView:             nil,
		blockchain:          blockchain,
		version:             shardProposedBlock.Header.Version,
		producer:            shardProposedBlock.Header.Proposer,
		round:               1,
		newBlock:            NewShardBlock(),
		startTime:           time.Unix(shardProposedBlock.Header.Timestamp, 0),
		crossShardBlocks:    make(map[byte][]*CrossShardBlock),
		confirmBeaconHeight: shardProposedBlock.Header.BeaconHeight,
		confirmBeaconHash:   shardProposedBlock.Header.BeaconHash,
	}

	if err = shardFlow.PreValidateProcess(shardProposedBlock); err != nil {
		return err
	}
	if err := shardFlow.BuildBody(); err != nil {
		return err
	}

	shardFlow.newView, err = shardFlow.curView.updateShardBestState(blockchain, shardFlow.newBlock, shardFlow.beaconBlocks, newCommitteeChange())
	if err != nil {
		return err
	}

	if err := shardFlow.BuildHeader(); err != nil {
		return err
	}

	//TODO: postProcessing (optional strictInsert) -> check header with new view
	if err = isSameShardContent(shardProposedBlock, shardFlow.newBlock); err != nil {
		return err
	}

	return
}

func (shardFlowState *ShardProcessState) PreValidateProcess(proposeBlock *ShardBlock) (err error) {
	//TODO: basic validation

	//check block version
	if proposeBlock.GetVersion() != shardFlowState.blockchain.config.ConsensusEngine.GetCurrentConsensusVersion() {
		return fmt.Errorf("Block created with wrong version %v. Expect %v", proposeBlock.GetVersion(), shardFlowState.blockchain.config.ConsensusEngine.GetCurrentConsensusVersion())
	}

	//validate block signature
	if err := shardFlowState.blockchain.config.ConsensusEngine.ValidateProducerPosition(proposeBlock, shardFlowState.curView.ShardCommittee); err != nil {
		return err
	}

	if err := shardFlowState.blockchain.config.ConsensusEngine.ValidateProducerSig(proposeBlock, shardFlowState.blockchain.BeaconChain.GetConsensusType()); err != nil {
		return err
	}

	//get enough beacon blocks
	// fetch beacon blocks
	previousBeaconHeight := shardFlowState.blockchain.GetBeaconBestState().GetHeight()
	if proposeBlock.Header.BeaconHeight > previousBeaconHeight {
		err := shardFlowState.blockchain.config.Server.PushMessageGetBlockBeaconByHeight(previousBeaconHeight, proposeBlock.Header.BeaconHeight)
		if err != nil {
			return errors.New(fmt.Sprintf("Beacon %d not ready, latest is %d", proposeBlock.Header.BeaconHeight, previousBeaconHeight))
		}
		ticker := time.NewTicker(5 * time.Second)
		<-ticker.C
		previousBeaconHeight = shardFlowState.blockchain.GetBeaconBestState().GetHeight()
		if proposeBlock.Header.BeaconHeight > previousBeaconHeight {
			return errors.New(fmt.Sprintf("Beacon %d not ready, latest is %d", proposeBlock.Header.BeaconHeight, previousBeaconHeight))
		}
	}

	//Fetch beacon block from height to confirm beacon block
	shardFlowState.beaconBlocks, err = FetchBeaconBlockFromHeight(shardFlowState.blockchain.GetDatabase(), shardFlowState.curView.BeaconHeight+1, shardFlowState.confirmBeaconHeight)
	if err != nil {
		return err
	}

	// this  beacon height is already seen by shard best state
	if shardFlowState.confirmBeaconHeight == shardFlowState.curView.BeaconHeight {
		shardFlowState.isOldBeaconHeight = true
	}

	shardFlowState.txs = shardFlowState.newBlock.Body.Transactions

	toShard := shardFlowState.curView.ShardID
	var toShardAllCrossShardBlock = make(map[byte][]*CrossShardBlock)
	crossShardRequired := make(map[byte][]common.Hash)
	for fromShard, crossTransactions := range shardFlowState.newBlock.Body.CrossTransactions {
		for _, crossTransaction := range crossTransactions {
			crossShardRequired[fromShard] = append(crossShardRequired[fromShard], crossTransaction.BlockHash)
		}
	}
	crossShardBlksFromPool, err := shardFlowState.blockchain.config.Syncker.GetCrossShardBlocksForShardValidator(toShard, crossShardRequired)
	if err != nil {
		return NewBlockChainError(CrossShardBlockError, fmt.Errorf("Unable to get required crossShard blocks from pool in time"))
	}
	for sid, v := range crossShardBlksFromPool {
		for _, b := range v {
			toShardAllCrossShardBlock[sid] = append(toShardAllCrossShardBlock[sid], b.(*CrossShardBlock))
		}
	}
	shardFlowState.crossShardBlocks = toShardAllCrossShardBlock
	return nil
}

func isSameShardContent(src, target *ShardBlock) (err error) {
	//compare header
	//Version
	//Height
	//Epoch
	//Round
	//Timestamp
	//PreviousBlockHash
	//InstructionHash
	//ShardStateHash
	//InstructionMerkleRoot
	//BeaconCommitteeAndValidatorRoot
	//BeaconCandidateRoot
	//ShardCandidateRoot
	//ShardCommitteeAndValidatorRoot
	//AutoStakingRoot
	//ConsensusType
	//Producer
	//ProducerPubKeyStr
	//Proposer    string `json:"Proposer"`
	//ProposeTime int64  `json:"ProposeTime"`

	//compare body
	//ShardState   map[byte][]ShardState
	//Instructions [][]string

	return nil
}
