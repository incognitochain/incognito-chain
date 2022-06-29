package blsbft

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metrics/monitor"
	"time"
)

func CreateProposeBFTMessage(block types.BlockInterface, peerID string) (*BFTPropose, error) {
	blockData, _ := json.Marshal(block)
	var bftPropose = new(BFTPropose)
	bftPropose.Block = blockData
	bftPropose.PeerID = peerID
	return bftPropose, nil
}

//check if node should propose in this timeslot
//if yes, then create and send propose block message
func (a *actorV3) maybeProposeBlock() error {
	time1 := time.Now()
	var err error
	bestView := a.chain.GetBestView()
	round := a.currentTimeSlot - common.CalculateTimeSlot(bestView.GetBlock().GetProposeTime())
	monitor.SetGlobalParam("RoundKey", fmt.Sprintf("%d_%d", bestView.GetHeight(), round))

	signingCommittees, committees, proposerPk, committeeViewHash, err := a.getCommitteesAndCommitteeViewHash()
	if err != nil {
		a.logger.Info(err)
		return err
	}
	b58Str, _ := proposerPk.ToBase58()
	userKeySet := a.getUserKeySetForSigning(signingCommittees, a.userKeySet)
	shouldListen, shouldPropose, userProposeKey := a.isUserKeyProposer(
		common.CalculateTimeSlot(bestView.GetBlock().GetProposeTime()),
		proposerPk,
		userKeySet,
	)

	if shouldListen {
		a.logger.Infof("%v TS: %v, LISTEN BLOCK %v, Round %v\n", a.chainKey, common.CalculateTimeSlot(a.currentTime), bestView.GetHeight()+1, round)
	}
	if !shouldPropose {
		return nil
	}

	a.logger.Infof("%v TS: %v, PROPOSE BLOCK %v, Round %v\n", a.chainKey, common.CalculateTimeSlot(a.currentTime), bestView.GetHeight()+1, round)
	if err := a.AddCurrentTimeSlotProposeHistory(); err != nil {
		a.logger.Errorf("add current time slot propose history")
	}

	// get block that we already send vote message (blockhash that is lock for this height)
	var block types.BlockInterface = nil
	lockBlockHash := a.getLockBlockHash(a.chain.GetBestViewHeight())
	if lockBlockHash != nil {
		block = lockBlockHash.block
	}

	if block == nil || block.GetVersion() < types.INSTANT_FINALITY_VERSION_V2 {
		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, (time.Duration(common.TIMESLOT)*time.Second)/2)
		defer cancel()
		a.logger.Info("CreateNewBlock version", a.blockVersion)
		block, err = a.chain.CreateNewBlock(a.blockVersion, b58Str, 1, a.currentTime, committees, committeeViewHash)
		if err != nil {
			return NewConsensusError(BlockCreationError, err)
		}
	} else {
		a.logger.Infof("CreateNewBlockFromOldBlock, Block Height %+v")
		block, err = a.chain.CreateNewBlockFromOldBlock(block, b58Str, a.currentTime, true)
		if err != nil {
			return NewConsensusError(BlockCreationError, err)
		}
	}

	//this actor v3, finality height is always current height, and committeefromblock is updated
	switch block.(type) {
	case *types.ShardBlock:
		block.(*types.ShardBlock).Header.FinalityHeight = block.GetHeight()
		block.(*types.ShardBlock).Header.CommitteeFromBlock = committeeViewHash
	case *types.BeaconBlock:
		block.(*types.BeaconBlock).Header.FinalityHeight = block.GetHeight()
	}

	if block != nil {
		a.logger.Infof("create block %v hash %v, propose time %v, produce time %v", block.GetHeight(), block.FullHashString(), block.(types.BlockInterface).GetProposeTime(), block.(types.BlockInterface).GetProduceTime())
	} else {
		a.logger.Infof("create block fail, time: %v", time.Since(time1).Seconds())
		return NewConsensusError(BlockCreationError, errors.New("block is nil"))
	}

	block, err = a.addValidationData(userProposeKey, block)
	if err != nil {
		a.logger.Errorf("Add validation data for new block failed", err)
	}
	bftProposeMessage, err := CreateProposeBFTMessage(block, a.node.GetSelfPeerID().String())
	err = a.sendBFTProposeMsg(bftProposeMessage)
	if err != nil {
		a.logger.Error("Send BFT Propose Message Failed", err)
	}
	return nil
}

//on receive propose message, store it into mem and db
func (a *actorV3) handleProposeMsg(proposeMsg BFTPropose) error {

	blockInfo, err := a.chain.UnmarshalBlock(proposeMsg.Block)
	if err != nil || blockInfo == nil {
		return err
	}

	block := blockInfo.(types.BlockInterface)
	blockHash := block.ProposeHash().String()

	_, ok := a.GetReceiveBlockByHash(blockHash)
	if ok {
		return errors.New("Already receive block")
	}

	if _, err := a.chain.GetBlockByHash(block.GetPrevHash()); err != nil {
		a.logger.Infof("Request sync block from node %s from %s to %s", proposeMsg.PeerID, block.GetPrevHash().String(), block.GetPrevHash().Bytes())
		a.node.RequestMissingViewViaStream(proposeMsg.PeerID, [][]byte{block.GetPrevHash().Bytes()}, a.chain.GetShardID(), a.chain.GetChainName())
		return err
	}

	if block.GetHeight() <= a.chain.GetBestViewHeight() {
		return errors.New("Receive block create from old view. Rejected!")
	}

	proposerCommitteePublicKey := incognitokey.CommitteePublicKey{}
	proposerCommitteePublicKey.FromBase58(block.GetProposer())
	proposerMiningKeyBase58 := proposerCommitteePublicKey.GetMiningKeyBase58(a.GetConsensusName())
	signingCommittees, committees, err := a.getCommitteeForNewBlock(block)
	if err != nil {
		return err
	}
	userKeySet := a.getUserKeySetForSigning(signingCommittees, a.userKeySet)

	proposeBlockInfo := &ProposeBlockInfo{
		block:                   block,
		ReceiveTime:             time.Now(),
		Votes:                   make(map[string]*BFTVote),
		PreVotes:                make(map[string]*BFTVote),
		Committees:              incognitokey.DeepCopy(committees),
		SigningCommittees:       incognitokey.DeepCopy(signingCommittees),
		UserKeySet:              signatureschemes2.DeepCopyMiningKeyArray(userKeySet),
		ProposerMiningKeyBase58: proposerMiningKeyBase58,
	}

	//get vote for this propose block (case receive vote faster)
	votes, prevotes, err := GetVotesByBlockHashFromDB(block.ProposeHash().String())
	if err != nil {
		a.logger.Error("Cannot get vote by block hash for rebuild", err)
		return err
	}
	proposeBlockInfo.Votes = votes
	proposeBlockInfo.PreVotes = prevotes

	if err := a.AddReceiveBlockByHash(blockHash, proposeBlockInfo); err != nil {
		a.logger.Errorf("add receive block by hash error %+v", err)
	}
	a.logger.Info("Receive block ", block.FullHashString(), "height", block.GetHeight(), ",block timeslot ", common.CalculateTimeSlot(block.GetProposeTime()))

	return nil
}
