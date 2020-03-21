package blockchain

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/pubsub"
)

func (blockchain *BlockChain) InsertShardBlock_V2(shardBlock *ShardBlock, strictInsert bool) (err error) {

	//check already insert
	if view := blockchain.ShardChain[shardBlock.Header.ShardID].GetViewByHash(*shardBlock.Hash()); view != nil {
		return fmt.Errorf("View already insert %+v", *shardBlock.Hash())
	}

	//check prev view exit
	preViewHash := shardBlock.GetPrevHash()
	preView := blockchain.ShardChain[shardBlock.Header.ShardID].GetViewByHash(preViewHash)
	if preView == nil {
		return fmt.Errorf("Cannot find view %+v", preViewHash.String())
	}

	//create block and compare content
	processState := &ShardProcessState{
		curView:          preView.(*ShardBestState),
		newView:          nil,
		blockchain:       blockchain,
		version:          shardBlock.Header.Version,
		producer:         shardBlock.Header.Producer,
		round:            1,
		newBlock:         shardBlock,
		crossShardBlocks: make(map[byte][]*CrossShardBlock),
	}

	if err = processState.PreInsertProcess(shardBlock); err != nil {
		return err
	}

	// Backup beststate
	processState.newView, err = processState.curView.updateShardBestState(blockchain, processState.newBlock, processState.beaconBlocks, newCommitteeChange())
	if err != nil {
		return err
	}
	//TODO: postProcessing (optional strictInsert) -> check header root hash with new view

	//store process block
	if err := blockchain.processStoreShardBlock(processState.newView, processState.newBlock, newCommitteeChange()); err != nil {
		return err
	}

	go blockchain.config.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.NewShardblockTopic, shardBlock))
	go blockchain.config.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.BeaconBeststateTopic, processState.newView))

	// For masternode: broadcast new committee to highways
	// go blockchain.config.Highway.BroadcastCommittee(
	// 	blockchain.config.ChainParams.Epoch,
	// 	processState.newView.BeaconCommittee,
	// 	processState.newView.ShardCommittee,
	// 	processState.newView.ShardPendingValidator,
	// )

	return nil
}

func (shardFlowState *ShardProcessState) PreInsertProcess(proposeBlock *ShardBlock) error {
	//TODO: basic validation (pre processing)

	//validate block signature
	if err := shardFlowState.blockchain.BeaconChain.ValidateBlockSignatures(proposeBlock, shardFlowState.curView.GetCommittee()); err != nil {
		return err
	}

	return nil
}
