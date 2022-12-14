package blsbft

import (
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

func (a *actorV3) maybePreVoteMsg() {
	for _, proposeBlockInfo := range a.receiveBlockByHash {
		if a.shouldPrevote(proposeBlockInfo) {
			a.sendVote(proposeBlockInfo, "prevote")
			proposeBlockInfo.IsPreVoted = true
		}
	}
}

/*
send prevote for propose block that
- link to best view (including next height)
- not yet prevote
- valid block
- in current timeslot
- not lock, or having same blockhash with lock blockhash or having POLC > lockTS (validPOLC)
*/
func (a *actorV3) shouldPrevote(proposeBlockInfo *ProposeBlockInfo) bool {
	bestView := a.chain.GetBestView()

	if proposeBlockInfo.block.GetPrevHash().String() != bestView.GetHash().String() {
		return false
	}

	if !proposeBlockInfo.IsValid {
		return false
	}

	if proposeBlockInfo.IsPreVoted {
		return false
	}

	if a.currentTimeSlot != bestView.CalculateTimeSlot(proposeBlockInfo.block.GetProposeTime()) {
		return false
	}

	lockBlockHash := a.getLockBlockHash(a.currentBestViewHeight + 1)

	if proposeBlockInfo.ValidPOLC || lockBlockHash == nil || lockBlockHash.block.Hash().String() == proposeBlockInfo.block.Hash().String() {
		return true
	}
	return false
}

// VoteValidBlock this function should be use to vote for valid block only
func (a *actorV3) sendVote(
	proposeBlockInfo *ProposeBlockInfo, phase string,
) error {
	//if valid then vote
	committeeBLSString, _ := incognitokey.ExtractPublickeysFromCommitteeKeyList(proposeBlockInfo.SigningCommittees, common.BlsConsensus)
	for _, userKey := range proposeBlockInfo.UserKeySet {
		pubKey := userKey.GetPublicKey()
		if common.IndexOfStr(pubKey.GetMiningKeyBase58(a.GetConsensusName()), committeeBLSString) != -1 {
			switch phase {
			case "prevote":
				err := a.createAndSendPreVote(&userKey, proposeBlockInfo.block, proposeBlockInfo.SigningCommittees)
				if err != nil {
					a.logger.Error(err)
					return NewConsensusError(UnExpectedError, err)
				}
			case "vote":
				//set isVote = true (lock), so that, if at same block height next time, we dont pre vote for different block hash
				proposeBlockInfo.IsVoted = true
				if err := a.AddReceiveBlockByHash(proposeBlockInfo.block.ProposeHash().String(), proposeBlockInfo); err != nil {
					return NewConsensusError(UnExpectedError, err)
				}
				//send
				err := a.createAndSendVote(&userKey, proposeBlockInfo.block, proposeBlockInfo.SigningCommittees, a.chain.GetPortalParamsV4(0))
				if err != nil {
					a.logger.Error(err)
					return NewConsensusError(UnExpectedError, err)
				}
			default:
				a.logger.Errorf("Send vote phase is not correct: %v", phase)
			}
		}
	}

	return nil
}

func (a *actorV3) createAndSendPreVote(
	userKey *signatureschemes2.MiningKey,
	block types.BlockInterface,
	signingCommittees []incognitokey.CommitteePublicKey,
) error {

	vote, err := a.CreatePreVote(userKey, block)
	if err != nil {
		return NewConsensusError(UnExpectedError, err)
	}

	msg, err := a.makeBFTVoteMsg(vote, a.chainKey, a.currentTimeSlot, block.GetHeight())
	if err != nil {
		a.logger.Error(err)
		return NewConsensusError(UnExpectedError, err)
	}

	a.logger.Info(a.chainKey, "sending pre vote...", block.FullHashString())

	go a.node.PushMessageToChain(msg, a.chain)

	return nil
}

func (a actorV3) CreatePreVote(
	userKey *signatureschemes2.MiningKey,
	block types.BlockInterface,
) (*BFTVote, error) {
	previousView := a.chain.GetViewByHash(block.GetPrevHash())
	var vote = new(BFTVote)
	userBLSPk := userKey.GetPublicKey().GetMiningKeyBase58(common.BlsConsensus)
	vote.Phase = "prevote"
	vote.BlockHash = block.ProposeHash().String()
	vote.Hash = block.Hash().String()
	vote.Validator = userBLSPk
	vote.ChainID = block.GetShardID()
	vote.ProposeTimeSlot = previousView.CalculateTimeSlot(block.GetProposeTime())
	err := vote.signVote(userKey)
	if err != nil {
		return nil, NewConsensusError(UnExpectedError, err)
	}
	return vote, nil
}

func (a *actorV3) handlePreVoteMsg(voteMsg BFTVote) error {
	voteMsg.IsValid = 0
	if proposeBlockInfo, ok := a.GetReceiveBlockByHash(voteMsg.BlockHash); ok { //if received block is already initiated
		if _, ok := proposeBlockInfo.PreVotes[voteMsg.Validator]; !ok { // and not receive validatorA vote
			proposeBlockInfo.PreVotes[voteMsg.Validator] = &voteMsg // store it
			vid, v := a.getValidatorIndex(proposeBlockInfo.SigningCommittees, voteMsg.Validator)
			if v != nil {
				vbase58, _ := v.ToBase58()
				a.logger.Infof("%v Receive prevote (%d) for block %s from validator %d %v", a.chainKey, len(a.receiveBlockByHash[voteMsg.BlockHash].PreVotes), voteMsg.BlockHash, vid, vbase58)
			} else {
				a.logger.Infof("%v Receive prevote (%d) for block %v from unknown validator %v", a.chainKey, len(a.receiveBlockByHash[voteMsg.BlockHash].PreVotes), voteMsg.BlockHash, voteMsg.Validator)
			}
			proposeBlockInfo.HasNewPreVote = true
		}
	}
	// record new votes for restore
	if err := AddVoteByBlockHashToDB(voteMsg.BlockHash, voteMsg); err != nil {
		a.logger.Errorf("add receive block by hash error %+v", err)
	}

	return nil
}

func (a *actorV3) validatePreVote(proposeBlockInfo *ProposeBlockInfo) {
	if !proposeBlockInfo.HasNewPreVote {
		return
	}

	validVote := 0
	errVote := 0

	committees := make(map[string]int)
	if len(proposeBlockInfo.PreVotes) != 0 {
		for i, v := range proposeBlockInfo.SigningCommittees {
			committees[v.GetMiningKeyBase58(common.BlsConsensus)] = i
		}
	}

	for id, vote := range proposeBlockInfo.PreVotes {
		dsaKey := []byte{}
		switch vote.IsValid {
		case 0:
			if value, ok := committees[vote.Validator]; ok {
				dsaKey = proposeBlockInfo.SigningCommittees[value].MiningPubKey[common.BridgeConsensus]
			} else {
				a.logger.Error("Receive prevote from nonCommittee member")
				continue
			}
			if len(dsaKey) == 0 {
				a.logger.Error("canot find dsa key")
				continue
			}

			err := vote.validateVoteOwner(dsaKey)
			if err != nil {
				a.logger.Error(dsaKey)
				a.logger.Error(err)
				proposeBlockInfo.PreVotes[id].IsValid = -1
				errVote++
			} else {
				proposeBlockInfo.PreVotes[id].IsValid = 1
				validVote++
			}
		case 1:
			validVote++
		case -1:
			errVote++
		}
	}

	a.logger.Info("Number of Valid Pre Vote", validVote, "| Number Of Error Pre Vote", errVote)
	proposeBlockInfo.HasNewPreVote = false
	proposeBlockInfo.ValidPreVotes = validVote
	proposeBlockInfo.ErrPreVotes = errVote

	for key, value := range proposeBlockInfo.Votes {
		if value.IsValid == -1 {
			delete(proposeBlockInfo.Votes, key)
		}
	}

	return
}
