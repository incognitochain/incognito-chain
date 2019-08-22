package blsbft

import (
	"encoding/json"
	"fmt"

	"github.com/incognitochain/incognito-chain/consensus/blsmultisig"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/wire"
)

const (
	MSG_PROPOSE = "propose"
	MSG_VOTE    = "vote"
)

type BFTPropose struct {
	Block json.RawMessage
}

type BFTVote struct {
	RoundKey  string
	Validator string
	Vote      vote
}

func MakeBFTProposeMsg(block []byte, chainKey string, userKeySet *MiningKey) (wire.Message, error) {
	var proposeCtn BFTPropose
	proposeCtn.Block = block
	proposeCtnBytes, err := json.Marshal(proposeCtn)
	if err != nil {
		return nil, err
	}
	msg, _ := wire.MakeEmptyMessage(wire.CmdBFT)
	msg.(*wire.MessageBFT).ChainKey = chainKey
	msg.(*wire.MessageBFT).Content = proposeCtnBytes
	msg.(*wire.MessageBFT).Type = MSG_PROPOSE
	return msg, nil
}

func MakeBFTVoteMsg(userKey *MiningKey, chainKey, roundKey string, vote vote) (wire.Message, error) {
	var voteCtn BFTVote
	voteCtn.RoundKey = roundKey
	voteCtn.Validator = userKey.GetPublicKeyBase58()
	voteCtn.Vote = vote
	voteCtnBytes, err := json.Marshal(voteCtn)
	if err != nil {
		return nil, err
	}
	msg, _ := wire.MakeEmptyMessage(wire.CmdBFT)
	msg.(*wire.MessageBFT).ChainKey = chainKey
	msg.(*wire.MessageBFT).Content = voteCtnBytes
	msg.(*wire.MessageBFT).Type = MSG_VOTE
	return msg, nil
}

func (e *BLSBFT) ProcessBFTMsg(msg *wire.MessageBFT) {
	switch msg.Type {
	case MSG_PROPOSE:
		var msgPropose BFTPropose
		err := json.Unmarshal(msg.Content, &msgPropose)
		if err != nil {
			fmt.Println(err)
			return
		}
		e.ProposeMessageCh <- msgPropose
	case MSG_VOTE:
		var msgVote BFTVote
		err := json.Unmarshal(msg.Content, &msgVote)
		if err != nil {
			fmt.Println(err)
			return
		}
		e.VoteMessageCh <- msgVote
	default:
		fmt.Println("???")
		return
	}
}

func (e *BLSBFT) sendVote() error {
	var Vote vote
	selfIdx := e.Chain.GetPubKeyCommitteeIndex(e.UserKeySet.GetPublicKeyBase58())
	listCommittee := []blsmultisig.PublicKey{}
	blsSig, err := e.UserKeySet.BLSSignData(e.RoundData.Block.Hash().GetBytes(), selfIdx, listCommittee)
	if err != nil {
		return err
	}
	bridgeSig := []byte{}
	if metadata.HasBridgeInstructions(e.RoundData.Block.GetInstructions()) {
		bridgeSig, err = e.UserKeySet.BriSignData(e.RoundData.Block.Hash().GetBytes())
		if err != nil {
			return err
		}
	}

	Vote.BLS = blsSig
	Vote.BRI = bridgeSig

	msg, err := MakeBFTVoteMsg(e.UserKeySet, e.ChainKey, getRoundKey(e.RoundData.NextHeight, e.RoundData.Round), Vote)
	if err != nil {
		return err
	}
	go e.Node.PushMessageToChain(msg, e.Chain)
	e.RoundData.NotYetSendVote = false
	return nil
}
