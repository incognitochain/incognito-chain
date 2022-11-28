package blsbft

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metrics/monitor"
	"github.com/incognitochain/incognito-chain/multiview"
)

func CreateProposeBFTMessage(block types.BlockInterface, peerID string) (*BFTPropose, error) {
	blockData, _ := json.Marshal(block)
	var bftPropose = new(BFTPropose)
	bftPropose.Block = blockData
	bftPropose.PeerID = peerID
	return bftPropose, nil
}

func (a *actorV3) getBlockForPropose(proposeBlockHeight uint64) types.BlockInterface {
	// get block that we already send vote message (blockhash that is lock for this height)
	var block types.BlockInterface = nil
	lockBlockHash := a.getLockBlockHash(proposeBlockHeight)
	if lockBlockHash != nil {
		block = lockBlockHash.block
		a.validatePreVote(lockBlockHash)
	} else { //or previous valid block
		for _, v := range a.GetSortedReceiveBlockByHeight(proposeBlockHeight) {
			if v.IsValid {
				block = v.block
				break
			}
		}
	}
	return block
}

// check if node should propose in this timeslot
// if yes, then create and send propose block message
func (a *actorV3) maybeProposeBlock() error {
	time1 := time.Now()
	var err error
	bestView := a.chain.GetBestView()
	round := a.currentTimeSlot - bestView.CalculateTimeSlot(bestView.GetBlock().GetProposeTime())
	monitor.SetGlobalParam("RoundKey", fmt.Sprintf("%d_%d", bestView.GetHeight(), round))

	signingCommittees, committees, proposerPk, committeeViewHash, err := a.getCommitteesAndCommitteeViewHash()
	if err != nil {
		a.logger.Info(err)
		return err
	}
	b58Str, _ := proposerPk.ToBase58()
	userKeySet := a.getUserKeySetForSigning(signingCommittees, a.userKeySet)
	shouldListen, shouldPropose, userProposeKey := a.isUserKeyProposer(
		bestView.CalculateTimeSlot(bestView.GetBlock().GetProposeTime()),
		proposerPk,
		userKeySet,
	)

	if shouldListen {
		a.logger.Infof("%v TS: %v, LISTEN BLOCK %v, Round %v\n", a.chainKey, bestView.CalculateTimeSlot(a.currentTime), bestView.GetHeight()+1, round)
		a.logger.Info("")
	}
	if !shouldPropose {
		return nil
	}

	a.logger.Infof("%v TS: %v, PROPOSE BLOCK %v, Round %v\n", a.chainKey, bestView.CalculateTimeSlot(a.currentTime), bestView.GetHeight()+1, round)
	a.logger.Info("")
	if err := a.AddCurrentTimeSlotProposeHistory(); err != nil {
		a.logger.Errorf("add current time slot propose history")
	}

	block := a.getBlockForPropose(bestView.GetHeight() + 1)
	rawPreviousValidationData := ""
	if block == nil || block.GetVersion() < types.INSTANT_FINALITY_VERSION_V2 {
		if block.Type() == common.BeaconChainKey {
			previousBlock, _ := a.chain.GetBlockByHash(*bestView.GetHash())
			if previousBlock != nil {
				if previousProposeBlockInfo, ok := a.GetReceiveBlockByHash(previousBlock.ProposeHash().String()); ok &&
					previousProposeBlockInfo != nil && previousProposeBlockInfo.block != nil {
					a.validateVote(previousProposeBlockInfo)
					rawPreviousValidationData, err = a.createBLSAggregatedSignatures(
						previousProposeBlockInfo.SigningCommittees,
						previousProposeBlockInfo.block.ProposeHash(),
						previousProposeBlockInfo.block.GetValidationField(),
						previousProposeBlockInfo.Votes)
					if err != nil {
						a.logger.Error("Create BLS Aggregated Signature for previous block propose info, height ", previousProposeBlockInfo.block.GetHeight(), " error", err)
					}
				}
			}
		}
		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, (time.Duration(common.TIMESLOT)*time.Second)/2)
		defer cancel()
		a.logger.Info("CreateNewBlock version", a.blockVersion)
		block, err = a.chain.CreateNewBlock(a.blockVersion, b58Str, 1, a.currentTime, committees, committeeViewHash, rawPreviousValidationData)
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
	// bftProposeMessage.PrevValidationData = rawPreviousValidationData
	if lockBlockHash := a.getLockBlockHash(block.GetHeight()); lockBlockHash != nil {
		bftProposeMessage.POLC, err = buildPOLCFromPreVote(bestView, lockBlockHash)
		if err != nil {
			a.logger.Error("buildPOLCFromPreVote Failed", err)
			return NewConsensusError(BlockCreationError, errors.New("buildPOLCFromPreVote fail"))
		}
		a.logger.Info("Build propose block message with POLC")
	}

	err = a.sendBFTProposeMsg(bftProposeMessage)
	if err != nil {
		a.logger.Error("Send BFT Propose Message Failed", err)
	}
	return nil
}

func buildPOLCFromPreVote(bestView multiview.View, info *ProposeBlockInfo) (POLC, error) {
	committeeBLSString, err := incognitokey.ExtractPublickeysFromCommitteeKeyList(info.SigningCommittees, common.BlsConsensus)
	if err != nil {
		return POLC{}, err
	}

	idx := []int{}
	sigs := [][]byte{}
	for pk, vote := range info.PreVotes {
		if vote.IsValid != 1 {
			continue
		}
		index := common.IndexOfStr(pk, committeeBLSString)
		idx = append(idx, index)
		sigs = append(sigs, vote.Confirmation)
	}
	res := POLC{
		idx, sigs, info.block.ProposeHash().String(), info.block.Hash().String(), bestView.CalculateTimeSlot(info.block.GetProposeTime()),
	}

	return res, nil
}

func (a *actorV3) verifyPOLCFromPreVote(info *ProposeBlockInfo, polc POLC, lock *ProposeBlockInfo) bool {
	if len(polc.Idx) == 0 {
		a.logger.Info("Empty polc")
		return false
	}

	if info.block.Hash().String() != polc.BlockHash {
		a.logger.Info("Should repropose polc blockhash")
		return false
	}

	if lock != nil {
		previousView := a.chain.GetViewByHash(lock.block.GetPrevHash())
		if polc.Timeslot < previousView.CalculateTimeSlot(lock.block.GetProposeTime()) {
			a.logger.Info("Not a new POLC")
			return false
		}
	}

	committeeBLSString, err := incognitokey.ExtractPublickeysFromCommitteeKeyList(info.SigningCommittees, common.BlsConsensus)
	if err != nil {
		return false
	}
	for i, index := range polc.Idx {
		vote := new(BFTVote)
		vote.BlockHash = polc.BlockProposeHash
		vote.Hash = info.block.Hash().String()
		vote.Validator = committeeBLSString[index]
		vote.ChainID = info.block.GetShardID()
		vote.ProposeTimeSlot = polc.Timeslot
		vote.Confirmation = polc.Sig[i]

		dsaKey := info.SigningCommittees[index].MiningPubKey[common.BridgeConsensus]
		if len(dsaKey) == 0 {
			a.logger.Info("Verify dsa error 1")
			return false
		}

		err := vote.validateVoteOwner(dsaKey)
		if err != nil {
			a.logger.Info("Verify dsa error 2")
			return false
		}
	}
	if len(polc.Idx) <= 2*len(info.SigningCommittees)/3 {
		a.logger.Info("Not enough signing signature")
		return false
	}

	return true
}

// on receive propose message, store it into mem and db
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
	previousView := a.chain.GetViewByHash(block.GetPrevHash())

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
		NumberOfFixNode:         previousView.GetProposerLength(),
	}

	//get vote for this propose block (case receive vote faster)
	votes, prevotes, err := GetVotesByBlockHashFromDB(block.ProposeHash().String())
	if err != nil {
		a.logger.Error("Cannot get vote by block hash for rebuild", err)
		return err
	}
	proposeBlockInfo.Votes = votes
	proposeBlockInfo.PreVotes = prevotes

	// handle Proof-of-lock-change (POLC: 2/3+ prevote signature) -> help node to unlock outdated locking blockhash (locking TS < POLC TS)
	if !a.verifyPOLCFromPreVote(proposeBlockInfo, proposeMsg.POLC, a.getLockBlockHash(block.GetHeight())) {
		a.logger.Info("Current propose block message dont have valid POLC")
	} else {
		proposeBlockInfo.ValidPOLC = true
	}

	if err := a.AddReceiveBlockByHash(blockHash, proposeBlockInfo); err != nil {
		a.logger.Errorf("add receive block by hash error %+v", err)
	}

	a.logger.Info("Receive block ", block.FullHashString(), "height", block.GetHeight(), ",block timeslot ", previousView.CalculateTimeSlot(block.GetProposeTime()))

	return nil
}
