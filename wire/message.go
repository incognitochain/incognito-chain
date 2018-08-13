package wire

import (
	"fmt"
)

// list message type
const (
	CmdBlock = "block"
	CmdTx    = "tx"
)

// Interface for message wire on P2P network
type Message interface {
	MessageType() string
	MaxPayloadLength(int) int
}

func MakeEmptyMessage(messageType string) (Message, error) {
	var msg Message
	switch messageType {
	case CmdBlock:
		msg = &MessageBlock{}
	case CmdTx:
		msg = &MessageTransaction{}
	default:
		return nil, fmt.Errorf("unhandled this message type [%s]", messageType)
	}
	return msg, nil
}
