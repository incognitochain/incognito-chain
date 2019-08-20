package blsbft

import (
	"encoding/json"
	"fmt"

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
	Sig       string
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

func MakeBFTVoteMsg(userPubKey *blsKeySet, chainKey, sig, roundKey string) (wire.Message, error) {
	var voteCtn BFTVote
	voteCtn.RoundKey = roundKey
	voteCtn.Validator = userPubKey.GetPublicKeyBase58()
	voteCtn.Sig = sig
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

func (e *BLSBFT) sendVote() {
	sig, _ := e.UserKeySet.SignData(e.RoundData.Block.Hash())
	MakeBFTVoteMsg(e.UserKeySet, e.ChainKey, sig, getRoundKey(e.RoundData.NextHeight, e.RoundData.Round))
	// go e.Node.PushMessageToChain(msg)
	e.RoundData.NotYetSendVote = false
}
