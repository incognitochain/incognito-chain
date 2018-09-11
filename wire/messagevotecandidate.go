package wire

import (
	"encoding/json"

	peer "github.com/libp2p/go-libp2p-peer"
)

type MessageVoteCandidate struct {
}

func (self MessageVoteCandidate) MessageType() string {
	return CmdRequestSign
}

func (self MessageVoteCandidate) MaxPayloadLength(pver int) int {
	return MaxHeaderPayload
}

func (self MessageVoteCandidate) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self MessageVoteCandidate) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self MessageVoteCandidate) SetSenderID(senderID peer.ID) error {
	return nil
}
