package wire

import "github.com/internet-cash/prototype/blockchain"

const (
	MaxBlockPayload = 4000000
)

type MessageBlock struct {
	Block *blockchain.MsgBlock
}

func (msg *MessageBlock) MessageType() string {
	return CmdBlock
}

func (msg *MessageBlock) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}
