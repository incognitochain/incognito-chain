package wire

import (
	"fmt"
	"github.com/ninjadotorg/cash-prototype/transaction"
	"github.com/ninjadotorg/cash-prototype/blockchain"
)

// list message type
const (
	MessageHeaderSize = 24

	CmdBlock     = "block"
	CmdTx        = "tx"
	CmdGetBlocks = "getblocks"
	CmdInv       = "inv"
	CmdGetData   = "getdata"
	CmdVersion   = "version"
	Cmdveack     = "verack"
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
		msg = &MessageBlock{
			Block: blockchain.Block{
				Transactions: make([]transaction.Transaction, 0),
			},
		}
	case CmdTx:
		msg = &MessageTx{
			Transaction: &transaction.Tx{},
		}
	case CmdVersion:
		msg = &MessageVersion{}
	case Cmdveack:
		msg = &MessageVerAck{}
	default:
		return nil, fmt.Errorf("unhandled this message type [%s]", messageType)
	}
	return msg, nil
}
