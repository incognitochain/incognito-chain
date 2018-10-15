package wire

import (
	"encoding/json"

	peer "github.com/libp2p/go-libp2p-peer"
)

type MessageCandidateVote struct {
}

func (self MessageCandidateVote) MessageType() string {
	return CmdRequestSign
}

func (self MessageCandidateVote) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
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
