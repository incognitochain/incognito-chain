package wire

import (
	"fmt"
)

// list message type
const (
	CmdBlock = "block"
)

type Message interface {
	MessageType() string
	MaxPayloadLength(uint32) uint32
}

func makeEmptyMessage(command string) (Message, error) {
	var msg Message
	switch command {
	case CmdBlock:
		msg = &MsgBlock{}

	default:
		return nil, fmt.Errorf("unhandled command [%s]", command)
	}
	return msg, nil
}
