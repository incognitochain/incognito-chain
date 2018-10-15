package wire

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
)

const (
	MaxCandidateVotePayload = 1000 // 1 Kb
)

type MessageCandidateVote struct {
}

func (self MessageCandidateVote) MessageType() string {
	return CmdRequestBlockSign
}

func (self MessageCandidateVote) MaxPayloadLength(pver int) int {
	return MaxCandidateVotePayload
}

func (self MessageCandidateVote) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self MessageCandidateVote) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self MessageCandidateVote) SetSenderID(senderID peer.ID) error {
	return nil
}
