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
	CmdGetAddr   = "getaddr"
	CmdAddr      = "addr"

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
		break
	case CmdGetBlocks:
		msg = &MessageGetBlocks{}
		break
	case CmdTx:
		msg = &MessageTx{
			Transaction: &transaction.Tx{},
		}
		break
	case CmdVersion:
		msg = &MessageVersion{}
		break
	case CmdVerack:
		msg = &MessageVerAck{}
		break
		// POS
	case CmdGetBlockHeader:
		msg = &MessageGetBlockHeader{}
		break
	case CmdBlockHeader:
		msg = &MessageBlockHeader{}
		break
	case CmdSignedBlock:
		msg = &MessageSignedBlock{}
		break
	case CmdRequestSign:
		msg = &MessageRequestSign{}
		break
	case CmdVoteCandidate:
		msg = &MessageVoteCandidate{}
		break
	case CmdInvalidBlock:
		msg = &MessageInvalidBlock{}
		break
	case CmdGetAddr:
		msg = &MessageGetAddr{}
		break
	case CmdAddr:
		msg = &MessageAddr{}
		break
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
	case reflect.TypeOf(&MessageGetAddr{}):
		return CmdGetAddr, nil
	case reflect.TypeOf(&MessageAddr{}):
		return CmdAddr, nil
	default:
		return "", fmt.Errorf("unhandled this message type [%s]", msgType)
	}
}
