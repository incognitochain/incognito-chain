package wire

import (
	"fmt"
	"reflect"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/transaction"
)

// list message type
const (
	MessageHeaderSize  = 24
	MessageCmdTypeSize = 12

	CmdBlock              = "block"
	CmdTx                 = "tx"
	CmdRegisteration      = "registeration"
	CmdCustomToken        = "txtoken"
	CmdCLoanRequestToken  = "txloanreq"
	CmdCLoanResponseToken = "txloanres"
	CmdCLoanWithdrawToken = "txloanwith"
	CmdCLoanPayToken      = "txloanpay"
	CmdGetBlocks          = "getblocks"
	CmdInv                = "inv"
	CmdGetData            = "getdata"
	CmdVersion            = "version"
	CmdVerack             = "verack"
	CmdGetAddr            = "getaddr"
	CmdAddr               = "addr"
	CmdPing               = "ping"

	// POS Cmd
	CmdBFTPropose    = "bftpropose"
	CmdBFTPrepare    = "bftprepare"
	CmdBFTCommit     = "bftcommit"
	CmdBFTReply      = "bftreply"
	CmdInvalidBlock  = "invalidblock"
	CmdGetChainState = "getchstate"
	CmdChainState    = "chainstate"

	// SWAP Cmd
	CmdSwapRequest = "swaprequest"
	CmdSwapSig     = "swapsig"
	CmdSwapUpdate  = "swapupdate"

	// heavy message check cmd
	CmdMsgCheck     = "msgcheck"
	CmdMsgCheckResp = "msgcheckresp"
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
	case CmdCustomToken:
		msg = &MessageTx{
			Transaction: &transaction.TxCustomToken{},
		}
		break
		/*case CmdCLoanRequestToken:
			  msg = &MessageTx{
				  Transaction: &transaction.TxLoanRequest{},
			  }
			  break
		  case CmdCLoanResponseToken:
			  msg = &MessageTx{
				  Transaction: &transaction.TxLoanResponse{},
			  }
			  break
		  case CmdCLoanWithdrawToken:
			  msg = &MessageTx{
				  Transaction: &transaction.TxLoanWithdraw{},
			  }
			  break
		  case CmdCLoanPayToken:
			  msg = &MessageTx{
				  Transaction: &transaction.TxLoanPayment{},
			  }
			  break*/
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
	case CmdBFTPropose:
		msg = &MessageBFTPropose{}
	case CmdBFTPrepare:
		msg = &MessageBFTPrepare{}
	case CmdBFTCommit:
		msg = &MessageBFTCommit{}
	case CmdBFTReply:
		msg = &MessageBFTReply{}
	case CmdInvalidBlock:
		msg = &MessageInvalidBlock{}
		break
	case CmdGetChainState:
		msg = &MessageGetChainState{}
	case CmdChainState:
		msg = &MessageChainState{}
	case CmdGetAddr:
		msg = &MessageGetAddr{}
		break
	case CmdAddr:
		msg = &MessageAddr{}
		break
	case CmdPing:
		msg = &MessagePing{}
		break
	case CmdSwapRequest:
		msg = &MessageSwapRequest{}
		break
	case CmdSwapSig:
		msg = &MessageSwapSig{}
		break
	case CmdSwapUpdate:
		msg = &MessageSwapUpdate{
			Signatures: make(map[string]string),
		}
		break
	case CmdMsgCheck:
		msg = &MessageMsgCheck{}
		break
	case CmdMsgCheckResp:
		msg = &MessageMsgCheckResp{}
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
		/*case reflect.TypeOf(&MessageRegistration{}):
		  return CmdRegisteration, nil*/
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
	case reflect.TypeOf(&MessageBFTPropose{}):
		return CmdBFTPropose, nil
	case reflect.TypeOf(&MessageBFTPrepare{}):
		return CmdBFTPrepare, nil
	case reflect.TypeOf(&MessageBFTCommit{}):
		return CmdBFTCommit, nil
	case reflect.TypeOf(&MessageBFTReply{}):
		return CmdBFTReply, nil
	case reflect.TypeOf(&MessageInvalidBlock{}):
		return CmdInvalidBlock, nil
	case reflect.TypeOf(&MessageGetChainState{}):
		return CmdGetChainState, nil
	case reflect.TypeOf(&MessageChainState{}):
		return CmdChainState, nil
	case reflect.TypeOf(&MessageSwapRequest{}):
		return CmdSwapRequest, nil
	case reflect.TypeOf(&MessageSwapSig{}):
		return CmdSwapSig, nil
	case reflect.TypeOf(&MessageSwapUpdate{}):
		return CmdSwapUpdate, nil
	case reflect.TypeOf(&MessageMsgCheck{}):
		return CmdMsgCheck, nil
	case reflect.TypeOf(&MessageMsgCheckResp{}):
		return CmdMsgCheckResp, nil
	default:
		return "", fmt.Errorf("unhandled this message type [%s]", msgType)
	}
}
