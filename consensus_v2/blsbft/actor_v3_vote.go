package blsbft

import (
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	"github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes/blsmultisig"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/portal/portalv4"
	portalprocessv4 "github.com/incognitochain/incognito-chain/portal/portalv4/portalprocess"
)

func (a *actorV3) maybeVoteMsg() {
	for _, proposeBlockInfo := range a.receiveBlockByHash {
		if a.shouldVote(proposeBlockInfo) {
			a.sendVote(proposeBlockInfo, "vote")
		}
	}
}

/*
send vote for propose block that
- not yet vote
- block valid
- receive 2/3 prevote
*/
func (a *actorV3) shouldVote(proposeBlockInfo *ProposeBlockInfo) bool {
	if a.currentTimeSlot == common.CalculateTimeSlot(proposeBlockInfo.block.GetProposeTime()) &&
		!proposeBlockInfo.IsVoted &&
		proposeBlockInfo.IsValid {
		if proposeBlockInfo.ValidPreVotes > 2*len(proposeBlockInfo.SigningCommittees)/3 {
			return true
		}
	}
	return false
}

func (a *actorV3) handleVoteMsg(voteMsg BFTVote) error {
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

	vote, err := a.CreateVote(userKey, block, signingCommittees, portalParamV4)
	if err != nil {
		return NewConsensusError(UnExpectedError, err)
	}

	msg, err := a.makeBFTVoteMsg(vote, a.chainKey, a.currentTimeSlot, block.GetHeight())
	if err != nil {
		a.logger.Error(err)
		return NewConsensusError(UnExpectedError, err)
	}

	a.logger.Info(a.chainKey, "sending vote...", block.FullHashString())

	go a.node.PushMessageToChain(msg, a.chain)

	return nil
}

func (a *actorV3) validateVote(proposeBlockInfo *ProposeBlockInfo) {
	if !proposeBlockInfo.HasNewVote {
		return
	}

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

	return
}

func (a actorV3) CreateVote(
	userKey *signatureschemes2.MiningKey,
	block types.BlockInterface,
	committees []incognitokey.CommitteePublicKey,
	portalParamsV4 portalv4.PortalParams,
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

	// check and sign on unshielding external tx for Portal v4
	portalSigs, err := portalprocessv4.CheckAndSignPortalUnshieldExternalTx(userKey.PriKey[common.BridgeConsensus], block.GetInstructions(), portalParamsV4)
	if err != nil {
		return nil, NewConsensusError(UnExpectedError, err)
	}
	vote.Phase = "vote"
	vote.BLS = blsSig
	vote.BRI = bridgeSig
	vote.PortalSigs = portalSigs
	vote.BlockHash = block.ProposeHash().String()
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
