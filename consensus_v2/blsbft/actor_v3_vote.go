package blsbft

import (
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/portal/portalv4"
)

/*
send vote for propose block that
- not yet vote
- receive 2/3 prevote
*/
func (a *actorV3) maybeVoteMsg() {
	for _, proposeBlockInfo := range a.receiveBlockByHash {
		if a.currentTimeSlot == common.CalculateTimeSlot(proposeBlockInfo.block.GetProposeTime()) &&
			!proposeBlockInfo.IsVoted {
			//no new pre vote
			if proposeBlockInfo.HasNewPreVote == false {
				return
			}

			if proposeBlockInfo.ValidPreVotes > 2*len(proposeBlockInfo.SigningCommittees)/3 {
				a.sendVote(proposeBlockInfo, "vote")
				proposeBlockInfo.IsVoted = true
				//TODO: lock this propose block info, so that, if same block height next time, we dont vote for different block hash
			}

		}
	}
}

func (a *actorV3) handleVoteMsg(voteMsg BFTVote) error {

	if a.chainID != common.BeaconChainID {
		if err := ByzantineDetectorObject.Validate(
			a.chain.GetBestViewHeight(),
			&voteMsg,
		); err != nil {
			a.logger.Errorf("Found byzantine validator %+v, err %+v", voteMsg.Validator, err)
			return err
		}
	}

	return a.processVoteMessage(voteMsg)
}

func (a *actorV3) processVoteMessage(voteMsg BFTVote) error {
	voteMsg.IsValid = 0
	if proposeBlockInfo, ok := a.GetReceiveBlockByHash(voteMsg.BlockHash); ok { //if received block is already initiated
		if _, ok := proposeBlockInfo.Votes[voteMsg.Validator]; !ok { // and not receive validatorA vote
			proposeBlockInfo.Votes[voteMsg.Validator] = &voteMsg // store it
			vid, v := a.getValidatorIndex(proposeBlockInfo.SigningCommittees, voteMsg.Validator)
			if v != nil {
				vbase58, _ := v.ToBase58()
				a.logger.Infof("%v Receive vote (%d) for block %s from validator %d %v", a.chainKey, len(a.receiveBlockByHash[voteMsg.BlockHash].Votes), voteMsg.BlockHash, vid, vbase58)
			} else {
				a.logger.Infof("%v Receive vote (%d) for block %v from unknown validator %v", a.chainKey, len(a.receiveBlockByHash[voteMsg.BlockHash].Votes), voteMsg.BlockHash, voteMsg.Validator)
			}
			proposeBlockInfo.HasNewVote = true
		}

		if !proposeBlockInfo.ProposerSendVote {
			for _, userKey := range a.userKeySet {
				pubKey := userKey.GetPublicKey()
				if proposeBlockInfo.block != nil && pubKey.GetMiningKeyBase58(a.GetConsensusName()) == proposeBlockInfo.ProposerMiningKeyBase58 { // if this node is proposer and not sending vote
					var err error
					if err = a.validateBlock(a.chain.GetBestView().GetHeight(), proposeBlockInfo); err == nil {
						bestViewHeight := a.chain.GetBestView().GetHeight()
						if proposeBlockInfo.block.GetHeight() == bestViewHeight+1 { // and if the propose block is still connected to bestview
							err := a.createAndSendVote(&userKey, proposeBlockInfo.block, proposeBlockInfo.SigningCommittees, a.chain.GetPortalParamsV4(0)) // => send vote
							if err != nil {
								a.logger.Error(err)
							} else {
								proposeBlockInfo.ProposerSendVote = true
								if err := a.AddReceiveBlockByHash(proposeBlockInfo.block.ProposeHash().String(), proposeBlockInfo); err != nil {
									return err
								}
							}
						}
					} else {
						a.logger.Debug(err)
					}
				}
			}
		}

	}

	// record new votes for restore
	if err := AddVoteByBlockHashToDB(voteMsg.BlockHash, voteMsg); err != nil {
		a.logger.Errorf("add receive block by hash error %+v", err)
	}
	return nil
}

func (a *actorV3) createAndSendVote(
	userKey *signatureschemes2.MiningKey,
	block types.BlockInterface,
	signingCommittees []incognitokey.CommitteePublicKey,
	portalParamV4 portalv4.PortalParams,
) error {

	vote, err := CreateVote(userKey, block, signingCommittees, portalParamV4)
	if err != nil {
		return NewConsensusError(UnExpectedError, err)
	}

	msg, err := a.makeBFTVoteMsg(vote, a.chainKey, a.currentTimeSlot, block.GetHeight())
	if err != nil {
		a.logger.Error(err)
		return NewConsensusError(UnExpectedError, err)
	}

	if err := a.AddVoteHistory(block.GetHeight(), block); err != nil {
		a.logger.Errorf("add vote history error %+v", err)
	}

	a.logger.Info(a.chainKey, "sending vote...", block.FullHashString())

	go a.node.PushMessageToChain(msg, a.chain)

	return nil
}

func (a *actorV3) validateVote(proposeBlockInfo *ProposeBlockInfo) *ProposeBlockInfo {
	validVote := 0
	errVote := 0

	committees := make(map[string]int)
	if len(proposeBlockInfo.Votes) != 0 {
		for i, v := range proposeBlockInfo.SigningCommittees {
			committees[v.GetMiningKeyBase58(common.BlsConsensus)] = i
		}
	}

	for id, vote := range proposeBlockInfo.Votes {
		dsaKey := []byte{}
		switch vote.IsValid {
		case 0:
			if value, ok := committees[vote.Validator]; ok {
				dsaKey = proposeBlockInfo.SigningCommittees[value].MiningPubKey[common.BridgeConsensus]
			} else {
				a.logger.Error("Receive vote from nonCommittee member")
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
				proposeBlockInfo.Votes[id].IsValid = -1
				errVote++
			} else {
				proposeBlockInfo.Votes[id].IsValid = 1
				validVote++
			}
		case 1:
			validVote++
		case -1:
			errVote++
		}
	}

	a.logger.Info("Number of Valid Vote", validVote, "| Number Of Error Vote", errVote)
	proposeBlockInfo.HasNewVote = false
	proposeBlockInfo.ValidVotes = validVote
	proposeBlockInfo.ErrVotes = errVote

	for key, value := range proposeBlockInfo.Votes {
		if value.IsValid == -1 {
			delete(proposeBlockInfo.Votes, key)
		}
	}

	return proposeBlockInfo
}
