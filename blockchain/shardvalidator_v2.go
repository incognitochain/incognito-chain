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
	processState := &ShardProcessState{
		curView:          preView.(*ShardBestState),
		newView:          nil,
		blockchain:       blockchain,
		version:          shardProposedBlock.Header.Version,
		proposer:         shardProposedBlock.Header.Proposer,
		round:            1,
		newBlock:         shardProposedBlock,
		crossShardBlocks: make(map[byte][]*CrossShardBlock),
		maxBeaconHeight:  shardProposedBlock.Header.BeaconHeight,
	}

	if err = processState.PreValidateProcess(shardProposedBlock); err != nil {
		return err
	}
	if err := processState.BuildBody(); err != nil {
		return err
	}

	processState.newView, err = processState.curView.updateShardBestState(blockchain, processState.newBlock, processState.beaconBlocks, newCommitteeChange())
	if err != nil {
		return err
	}

	if err := processState.BuildHeader(); err != nil {
		return err
	}

	//TODO: postProcessing (optional strictInsert) -> check header with new view
	if err = isSameShardContent(shardProposedBlock, processState.newBlock); err != nil {
		return err
	}

	return
}

func (createState *ShardProcessState) PreValidateProcess(proposeBlock *ShardBlock) error {
	//TODO: basic validation

	//validate block signature
	if err := createState.blockchain.config.ConsensusEngine.ValidateProducerPosition(proposeBlock, createState.curView.ShardCommittee); err != nil {
		return err
	}

	if err := createState.blockchain.config.ConsensusEngine.ValidateProducerSig(proposeBlock, createState.blockchain.BeaconChain.GetConsensusType()); err != nil {
		return err
	}
	if createState.maxBeaconHeight-createState.curView.BeaconHeight > MAX_BEACON_BLOCK {
		createState.maxBeaconHeight = createState.curView.BeaconHeight + MAX_BEACON_BLOCK
	}
	//get enough beacon blocks
	// fetch beacon blocks
	previousBeaconHeight := createState.curView.BeaconHeight
	if proposeBlock.Header.BeaconHeight > previousBeaconHeight {
		err := createState.blockchain.config.Server.PushMessageGetBlockBeaconByHeight(previousBeaconHeight, proposeBlock.Header.BeaconHeight)
		if err != nil {
			return errors.New(fmt.Sprintf("Beacon %d not ready, latest is %d", proposeBlock.Header.BeaconHeight, previousBeaconHeight))
		}
		ticker := time.NewTicker(5 * time.Second)
		<-ticker.C
		if proposeBlock.Header.BeaconHeight > previousBeaconHeight {
			return errors.New(fmt.Sprintf("Beacon %d not ready, latest is %d", proposeBlock.Header.BeaconHeight, previousBeaconHeight))
		}
	}

	createState.txs = createState.newBlock.Body.Transactions

	toShard := createState.curView.ShardID
	var toShardAllCrossShardBlock = make(map[byte][]*CrossShardBlock)
	crossShardRequired := make(map[byte][]common.Hash)
	for fromShard, crossTransactions := range createState.newBlock.Body.CrossTransactions {
		for _, crossTransaction := range crossTransactions {
			crossShardRequired[fromShard] = append(crossShardRequired[fromShard], crossTransaction.BlockHash)
		}
	}
	crossShardBlksFromPool, err := createState.blockchain.config.Syncker.GetCrossShardBlocksForShardValidator(toShard, crossShardRequired)
	if err != nil {
		return NewBlockChainError(CrossShardBlockError, fmt.Errorf("Unable to get required crossShard blocks from pool in time"))
	}
	for sid, v := range crossShardBlksFromPool {
		for _, b := range v {
			toShardAllCrossShardBlock[sid] = append(toShardAllCrossShardBlock[sid], b.(*CrossShardBlock))
		}
	}
	createState.crossShardBlocks = toShardAllCrossShardBlock
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
