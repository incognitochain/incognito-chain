package wire

import (
	"fmt"
	"reflect"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/cash/blockchain"
	"github.com/ninjadotorg/cash/transaction"
)

// list message type
const (
	MessageHeaderSize  = 24
	MessageCmdTypeSize = 12

	CmdBlock         = "block"
	CmdTx            = "tx"
	CmdRegisteration = "registeration"
	CmdGetBlocks     = "getblocks"
	CmdInv           = "inv"
	CmdGetData       = "getdata"
	CmdVersion       = "version"
	CmdVerack        = "verack"
	CmdGetAddr       = "getaddr"
	CmdAddr          = "addr"
	CmdPing          = "ping"

	// POS Cmd
	CmdRequestBlockSign  = "rqblocksign"
	CmdInvalidBlock      = "invalidblock"
	CmdBlockSig          = "blocksig"
	CmdGetChainState     = "getchstate"
	CmdChainState        = "chainstate"
	CmdCandidateProposal = "cndproposal"
	CmdCandidateVote     = "cndvote"

	// SWAP Cmd
	CmdRequestSwap = "requestswap"
	CmdSignSwap    = "signswap"
)

// Interface for message wire on P2P network
type Message interface {
	MessageType() string
	MaxPayloadLength(int) int
	JsonSerialize() ([]byte, error)
	JsonDeserialize(string) error
	SetSenderID(peer.ID) error
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
	case CmdRegisteration:
		msg = &MessageRegisteration{
			Transaction: &transaction.TxVoting{},
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
		// POS start
	case CmdBlockSig:
		msg = &MessageBlockSig{}
		break
	case CmdRequestBlockSign:
		msg = &MessageRequestBlockSign{}
		break
	case CmdInvalidBlock:
		msg = &MessageInvalidBlock{}
		break
	case CmdGetChainState:
		msg = &MessageGetChainState{}
	case CmdChainState:
		msg = &MessageChainState{}
	case CmdCandidateVote:
		msg = &MessageCandidateVote{}
		break
	case CmdCandidateProposal:
		msg = &MessageCandidateProposal{}
		// POS end
	case CmdGetAddr:
		msg = &MessageGetAddr{}
		break
	case CmdAddr:
		msg = &MessageAddr{}
		break
	case CmdPing:
		msg = &MessagePing{}
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
	case reflect.TypeOf(&MessageRegisteration{}):
		return CmdRegisteration, nil
	case reflect.TypeOf(&MessageVersion{}):
		return CmdVersion, nil
	case reflect.TypeOf(&MessageVerAck{}):
		return CmdVerack, nil
	case reflect.TypeOf(&MessageGetAddr{}):
		return CmdGetAddr, nil
	case reflect.TypeOf(&MessageAddr{}):
		return CmdAddr, nil
	case reflect.TypeOf(&MessagePing{}):
		return CmdPing, nil

		// POS start
	case reflect.TypeOf(&MessageBlockSig{}):
		return CmdBlockSig, nil
	case reflect.TypeOf(&MessageRequestBlockSign{}):
		return CmdRequestBlockSign, nil
	case reflect.TypeOf(&MessageCandidateVote{}):
		return CmdCandidateVote, nil
	case reflect.TypeOf(&MessageCandidateProposal{}):
		return CmdCandidateProposal, nil
	case reflect.TypeOf(&MessageInvalidBlock{}):
		return CmdInvalidBlock, nil
	case reflect.TypeOf(&MessageGetChainState{}):
		return CmdGetChainState, nil
	case reflect.TypeOf(&MessageChainState{}):
		return CmdChainState, nil
		// POS end

	default:
		return "", fmt.Errorf("unhandled this message type [%s]", msgType)
	}
}
