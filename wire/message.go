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

func makeEmptyMessage(command string) (Message, error) {
	var msg Message
	switch command {
	case CmdBlock:
		msg = &MessageBlock{}
	case CmdTx:
		msg = &MessageTransaction{}
	default:
		return nil, fmt.Errorf("unhandled command [%s]", command)
	}
	return msg, nil
}
