package blockchain

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/pubsub"
)

func (blockchain *BlockChain) InsertBeaconBlock_V2(beaconBlock *BeaconBlock, strictInsert bool) (err error) {
	//check prev view exit
	preViewHash := beaconBlock.GetPrevHash()
	preView := blockchain.BeaconChain.GetViewByHash(preViewHash)
	if preView == nil {
		return fmt.Errorf("Cannot find view %+v", preViewHash.String())
	}

	//create block and compare content
	processState := &BeaconProcessState{
		curView:            preView.(*BeaconBestState),
		newView:            nil,
		blockchain:         blockchain,
		version:            beaconBlock.Header.Version,
		proposer:           beaconBlock.Header.Proposer,
		round:              1,
		newBlock:           beaconBlock,
		shardToBeaconBlock: make(map[byte][]*ShardToBeaconBlock),
	}

	if err = processState.PreInsertProcess(beaconBlock); err != nil {
		Logger.log.Error(err)
		return err
	}

	// Backup beststate
	committeeChange := newCommitteeChange()
	processState.newView, err = processState.curView.updateBeaconBestState(beaconBlock, blockchain.config.ChainParams.Epoch, blockchain.config.ChainParams.AssignOffset, blockchain.config.ChainParams.RandomTime, committeeChange)
	if err != nil {
		return err
	}
	//TODO: postProcessing (optional strictInsert) -> check header root hash with new view

	//store process block
	if err := blockchain.processStoreBeaconBlock(processState.newView, beaconBlock, processState.curView.BeaconCommittee, processState.curView.ShardCommittee, processState.curView.RewardReceiver, committeeChange); err != nil {
		return err
	}

	go blockchain.config.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.NewBeaconBlockTopic, beaconBlock))
	go blockchain.config.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.BeaconBeststateTopic, processState.newView))

	// For masternode: broadcast new committee to highways
	go blockchain.config.Highway.BroadcastCommittee(
		blockchain.config.ChainParams.Epoch,
		processState.newView.BeaconCommittee,
		processState.newView.ShardCommittee,
		processState.newView.ShardPendingValidator,
	)

	return nil
}

func (s *BeaconProcessState) PreInsertProcess(proposeBlock *BeaconBlock) error {
	//TODO: basic validation (pre processing)

	//validate block signature
	if err := s.blockchain.BeaconChain.ValidateProducerPosition(proposeBlock, s.curView.GetCommittee()); err != nil {
		return err
	}
	if err := s.blockchain.BeaconChain.ValidateBlockSignatures(proposeBlock, s.curView.GetCommittee()); err != nil {
		return err
	}
	return nil
}
