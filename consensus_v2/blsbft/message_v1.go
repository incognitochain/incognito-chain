package blsbft

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/wire"
)

func (actorV1 *actorV1) makeBFTProposeMsg(block []byte, chainKey string, userKeySet *signatureschemes.MiningKey) (wire.Message, error) {
	var proposeCtn BFTPropose
	proposeCtn.Block = block
	proposeCtnBytes, err := json.Marshal(proposeCtn)
	if err != nil {
		return nil, NewConsensusError(UnExpectedError, err)
	}
	msg, _ := wire.MakeEmptyMessage(wire.CmdBFT)
	msg.(*wire.MessageBFT).ChainKey = chainKey
	msg.(*wire.MessageBFT).Content = proposeCtnBytes
	msg.(*wire.MessageBFT).Type = MsgPropose
	return msg, nil
}

func (actorV1 *actorV1) makeBFTVoteMsg(userPublicKey string, chainKey, roundKey string, vote vote) (wire.Message, error) {
	var voteCtn BFTVote
	voteCtn.RoundKey = roundKey
	voteCtn.Validator = userPublicKey
	voteCtn.Bls = vote.BLS
	voteCtn.Bri = vote.BRI
	voteCtn.Confirmation = vote.Confirmation
	voteCtnBytes, err := json.Marshal(voteCtn)
	if err != nil {
		return nil, NewConsensusError(UnExpectedError, err)
	}
	msg, _ := wire.MakeEmptyMessage(wire.CmdBFT)
	msg.(*wire.MessageBFT).ChainKey = chainKey
	msg.(*wire.MessageBFT).Content = voteCtnBytes
	msg.(*wire.MessageBFT).Type = MsgVote
	return msg, nil
}

func (actorV1 *actorV1) confirmVote(Vote *vote, userKey signatureschemes.MiningKey) error {
	data := actorV1.roundData.block.Hash().GetBytes()
	data = append(data, Vote.BLS...)
	data = append(data, Vote.BRI...)
	data = common.HashB(data)
	var err error
	Vote.Confirmation, err = userKey.BriSignData(data)
	return err
}

func (actorV1 *actorV1) sendVote() error {
	var Vote vote
	for _, userKey := range actorV1.userKeySet {
		pubKey := userKey.GetPublicKey()
		if common.IndexOfStr(pubKey.GetMiningKeyBase58(consensusName), actorV1.roundData.committeeBLS.stringList) != -1 {
			selfIdx := common.IndexOfStr(pubKey.GetMiningKeyBase58(consensusName), actorV1.roundData.committeeBLS.stringList)
			blsSig, err := userKey.BLSSignData(actorV1.roundData.block.Hash().GetBytes(), selfIdx, actorV1.roundData.committeeBLS.byteList)
			if err != nil {
				return NewConsensusError(UnExpectedError, err)
			}
			bridgeSig := []byte{}
			if metadata.HasBridgeInstructions(actorV1.roundData.block.GetInstructions()) || metadata.HasPortalInstructions(actorV1.roundData.block.GetInstructions()) {
				bridgeSig, err = userKey.BriSignData(actorV1.roundData.block.Hash().GetBytes())
				if err != nil {
					return NewConsensusError(UnExpectedError, err)
				}
			}

			Vote.BLS = blsSig
			Vote.BRI = bridgeSig

			err = actorV1.confirmVote(&Vote, userKey)
			if err != nil {
				return NewConsensusError(UnExpectedError, err)
			}
			key := userKey.GetPublicKey()

			msg, err := actorV1.makeBFTVoteMsg(key.GetMiningKeyBase58(consensusName), actorV1.chainKey, getRoundKey(actorV1.roundData.nextHeight, actorV1.roundData.round), Vote)
			if err != nil {
				return NewConsensusError(UnExpectedError, err)
			}
			actorV1.roundData.votes[pubKey.GetMiningKeyBase58(consensusName)] = Vote
			actorV1.logger.Info("sending vote...", getRoundKey(actorV1.roundData.nextHeight, actorV1.roundData.round))
			go actorV1.node.PushMessageToChain(msg, actorV1.chain)
		}
	}

	actorV1.roundData.notYetSendVote = false
	return nil
}
