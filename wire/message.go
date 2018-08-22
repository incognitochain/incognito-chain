package wire

import (
	"fmt"
	"github.com/ninjadotorg/cash-prototype/transaction"
)

// list message type
const (
	MessageHeaderSize = 24

	CmdBlock     = "block"
	CmdTx        = "tx"
	CmdGetBlocks = "getblocks"
	CmdInv       = "inv"
	CmdGetData   = "getdata"
)

// Interface for message wire on P2P network
type Message interface {
	MessageType() string
	MaxPayloadLength(int) int
	JsonSerialize() (string, error)
	JsonDeserialize(string) error
}

func MakeEmptyMessage(messageType string) (Message, error) {
	var msg Message
	switch messageType {
	case CmdBlock:
		msg = &MessageBlock{}
	case CmdTx:
		msg = &MessageTx{
			Transaction: &transaction.Tx{},
		}
	default:
		return nil, fmt.Errorf("unhandled this message type [%s]", messageType)
	}
	return msg, nil
}
