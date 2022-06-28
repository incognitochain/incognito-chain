package blsbft

import (
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	"github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes/blsmultisig"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
)

/*
send prevote for propose block that
- not yet prevote
- valid block
- in current timeslot
- not lock, or having same blockhash with lock blockhash
*/
func (a *actorV3) maybePreVoteMsg() {
	for _, proposeBlockInfo := range a.receiveBlockByHash {
		lockBlockHash := a.lockBlockHashByHeight[a.currentTimeSlot]
		if proposeBlockInfo.IsValid &&
			!proposeBlockInfo.IsPreVoted &&
			a.currentTimeSlot == common.CalculateTimeSlot(proposeBlockInfo.block.GetProposeTime()) &&
			(lockBlockHash == "" || lockBlockHash == proposeBlockInfo.block.Hash().String()) {
			a.sendVote(proposeBlockInfo, "prevote")
			proposeBlockInfo.IsPreVoted = true
		}
	}
}

//VoteValidBlock this function should be use to vote for valid block only
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
				} else {
					if !proposeBlockInfo.IsPreVoted { //not update database if field is already set
						proposeBlockInfo.IsPreVoted = true
						if err := a.AddReceiveBlockByHash(proposeBlockInfo.block.ProposeHash().String(), proposeBlockInfo); err != nil {
							return err
						}
					}
				}
			case "vote":
				err := a.createAndSendVote(&userKey, proposeBlockInfo.block, proposeBlockInfo.SigningCommittees, a.chain.GetPortalParamsV4(0))
				if err != nil {
					a.logger.Error(err)
					return NewConsensusError(UnExpectedError, err)
				} else {
					if !proposeBlockInfo.IsVoted { //not update database if field is already set
						proposeBlockInfo.IsVoted = true
						if err := a.AddReceiveBlockByHash(proposeBlockInfo.block.ProposeHash().String(), proposeBlockInfo); err != nil {
							return err
						}
					}
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

	vote, err := a.CreatePreVote(userKey, block, signingCommittees)
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
	committees []incognitokey.CommitteePublicKey,
) (*BFTVote, error) {
	var vote = new(BFTVote)
	bytelist := []blsmultisig.PublicKey{}
	selfIdx := 0
	userBLSPk := userKey.GetPublicKey().GetMiningKeyBase58(common.BlsConsensus)
	for i, v := range committees {
		if v.GetMiningKeyBase58(common.BlsConsensus) == userBLSPk {
			selfIdx = i
		}
		bytelist = append(bytelist, v.MiningPubKey[common.BlsConsensus])
	}

	blsSig, err := userKey.BLSSignData(block.ProposeHash().GetBytes(), selfIdx, bytelist)
	if err != nil {
		return nil, NewConsensusError(UnExpectedError, err)
	}

	bridgeSig := []byte{}
	if metadata.HasBridgeInstructions(block.GetInstructions()) {
		bridgeSig, err = userKey.BriSignData(block.Hash().GetBytes()) //proof is agg sig on block hash (not propose hash)
		if err != nil {
			return nil, NewConsensusError(UnExpectedError, err)
		}
	}
	vote.Phase = "prevote"
	vote.BLS = blsSig
	vote.BRI = bridgeSig
	vote.BlockHash = block.Hash().String()
	vote.Validator = userBLSPk
	vote.ProduceTimeSlot = common.CalculateTimeSlot(block.GetProduceTime())
	vote.ProposeTimeSlot = common.CalculateTimeSlot(block.GetProposeTime())
	vote.PrevBlockHash = block.GetPrevHash().String()
	vote.BlockHeight = block.GetHeight()
	vote.CommitteeFromBlock = block.CommitteeFromBlock()
	vote.ChainID = block.GetShardID()
	err = vote.signVote(userKey)
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

		if !proposeBlockInfo.ProposerSendPreVote {
			for _, userKey := range a.userKeySet {
				pubKey := userKey.GetPublicKey()
				if proposeBlockInfo.block != nil && pubKey.GetMiningKeyBase58(a.GetConsensusName()) == proposeBlockInfo.ProposerMiningKeyBase58 { // if this node is proposer and not sending vote
					var err error
					if err = a.validateBlock(a.chain.GetBestView().GetHeight(), proposeBlockInfo); err == nil {
						bestViewHeight := a.chain.GetBestView().GetHeight()
						if proposeBlockInfo.block.GetHeight() == bestViewHeight+1 { // and if the propose block is still connected to bestview
							err := a.createAndSendPreVote(&userKey, proposeBlockInfo.block, proposeBlockInfo.SigningCommittees) // => send vote
							if err != nil {
								a.logger.Error(err)
							} else {
								proposeBlockInfo.ProposerSendPreVote = true
							}
						}
					} else {
						a.logger.Error(err)
					}
				}
			}
		}
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
