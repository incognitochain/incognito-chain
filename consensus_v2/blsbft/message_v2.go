package blsbft

import (
	"encoding/json"
	"time"

	"github.com/incognitochain/incognito-chain/wire"
)

func (actorV2 *actorV2) makeBFTProposeMsg(proposeCtn *BFTPropose, chainKey string, ts int64, height uint64) (wire.Message, error) {
	proposeCtnBytes, err := json.Marshal(proposeCtn)
	if err != nil {
		return nil, NewConsensusError(UnExpectedError, err)
	}
	msg, _ := wire.MakeEmptyMessage(wire.CmdBFT)
	msg.(*wire.MessageBFT).ChainKey = chainKey
	msg.(*wire.MessageBFT).Content = proposeCtnBytes
	msg.(*wire.MessageBFT).Type = MsgPropose
	msg.(*wire.MessageBFT).TimeSlot = ts
	msg.(*wire.MessageBFT).Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
	msg.(*wire.MessageBFT).PeerID = proposeCtn.PeerID
	return msg, nil
}

func (actorV2 *actorV2) makeBFTVoteMsg(vote *BFTVote, chainKey string, ts int64, height uint64) (wire.Message, error) {
	voteCtnBytes, err := json.Marshal(vote)
	if err != nil {
		return nil, NewConsensusError(UnExpectedError, err)
	}
	msg, _ := wire.MakeEmptyMessage(wire.CmdBFT)
	msg.(*wire.MessageBFT).ChainKey = chainKey
	msg.(*wire.MessageBFT).Content = voteCtnBytes
	msg.(*wire.MessageBFT).Type = MsgVote
	msg.(*wire.MessageBFT).TimeSlot = ts
	msg.(*wire.MessageBFT).Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
	return msg, nil
}

func (actorV2 *actorV2) makeBFTRequestBlk(request BFTRequestBlock, peerID string, chainKey string) (wire.Message, error) {
	requestCtnBytes, err := json.Marshal(request)
	if err != nil {
		return nil, NewConsensusError(UnExpectedError, err)
	}
	msg, _ := wire.MakeEmptyMessage(wire.CmdBFT)
	msg.(*wire.MessageBFT).ChainKey = chainKey
	msg.(*wire.MessageBFT).Content = requestCtnBytes
	msg.(*wire.MessageBFT).Type = MsgRequestBlk
	return msg, nil
}
