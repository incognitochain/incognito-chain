package wire

import (
	"fmt"
	"reflect"

	"github.com/ninjadotorg/cash-prototype/blockchain"
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
	CmdVersion   = "version"
	CmdVerack    = "verack"

	// POS Cmd
	CmdGetBlockHeader = "getheader"
	CmdBlockHeader    = "header"
	CmdSignedBlock    = "signedblock"
	CmdVoteCandidate  = "votecandidate"
	CmdRequestSign    = "requestsign"
	CmdInvalidBlock   = "invalidblock"
)

// Interface for message wire on P2P network
type Message interface {
	MessageType() string
	MaxPayloadLength(int) int
	JsonSerialize() ([]byte, error)
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
	case CmdGetBlocks:
		msg = &MessageGetBlocks{}
	case CmdTx:
		msg = &MessageTx{
			Transaction: &transaction.Tx{},
		}
	case CmdVersion:
		msg = &MessageVersion{}
	case CmdVerack:
		msg = &MessageVerAck{}
	// POS
	case CmdGetBlockHeader:
		msg = &MessageGetBlockHeader{}
	case CmdBlockHeader:
		msg = &MessageBlockHeader{}
	case CmdSignedBlock:
		msg = &MessageSignedBlock{}
	case CmdRequestSign:
		msg = &MessageRequestSign{}
	case CmdVoteCandidate:
		msg = &MessageVoteCandidate{}
	case CmdInvalidBlock:
		msg = &MessageInvalidBlock{}
	default:
		return nil, fmt.Errorf("unhandled this message type [%s]", messageType)
	}
	return msg, nil
}

func GetCmdType(msgType reflect.Type) (string, error) {
	switch msgType {
	case reflect.TypeOf(&MessageBlock{}):
		return CmdBlock, nil
	case reflect.TypeOf(&MessageGetBlocks{}):
		return CmdGetBlocks, nil
	case reflect.TypeOf(&MessageTx{}):
		return CmdTx, nil
	case reflect.TypeOf(&MessageVersion{}):
		return CmdVersion, nil
	case reflect.TypeOf(&MessageVerAck{}):
		return CmdVerack, nil
	// POS
	case reflect.TypeOf(&MessageGetBlockHeader{}):
		return CmdGetBlockHeader, nil
	case reflect.TypeOf(&MessageBlockHeader{}):
		return CmdBlockHeader, nil
	case reflect.TypeOf(&MessageSignedBlock{}):
		return CmdSignedBlock, nil
	case reflect.TypeOf(&MessageRequestSign{}):
		return CmdRequestSign, nil
	case reflect.TypeOf(&MessageVoteCandidate{}):
		return CmdVoteCandidate, nil
	case reflect.TypeOf(&MessageInvalidBlock{}):
		return CmdInvalidBlock, nil
	default:
		return "", fmt.Errorf("unhandled this message type [%s]", msgType)
	}
}
