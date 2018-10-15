package wire

import (
	"encoding/json"

	peer "github.com/libp2p/go-libp2p-peer"
)

type MessageCandidateProposal struct {
}

func (self MessageCandidateProposal) MessageType() string {
	return CmdBlock
}

func (self MessageCandidateProposal) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (self MessageCandidateProposal) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self MessageCandidateProposal) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}
func (self MessageCandidateProposal) SetSenderID(senderID peer.ID) error {
	return nil
}
