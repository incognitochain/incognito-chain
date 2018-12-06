package wire

import (
	"fmt"
	"reflect"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/metadata"
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
	CmdBlockSigReq   = "blocksigreq"
	CmdBlockSig      = "blocksig"
	CmdInvalidBlock  = "invalidblock"
	CmdGetChainState = "getchstate"
	CmdChainState    = "chainstate"

	// SWAP Cmd
	CmdSwapRequest = "swaprequest"
	CmdSwapSig     = "swapsig"
	CmdSwapUpdate  = "swapupdate"
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
				Transactions: make([]metadata.Transaction, 0),
			},
		}
		break
	case CmdCustomToken:
		msg = &MessageTx{
			Transaction: &transaction.TxCustomToken{},
		}
		break
	case CmdCLoanRequestToken:
		msg = &MessageTx{
			Transaction: &transaction.Tx{
				Metadata: &metadata.LoanRequest{},
			},
		}
		break
	case CmdCLoanResponseToken:
		msg = &MessageTx{
			Transaction: &transaction.Tx{
				Metadata: &metadata.LoanResponse{},
			},
		}
		break
	case CmdCLoanWithdrawToken:
		msg = &MessageTx{
			Transaction: &transaction.TxLoanPayment{},
		}
		break
	case CmdCLoanPayToken:
		msg = &MessageTx{
			Transaction: &transaction.TxLoanPayment{},
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
	case CmdBlockSig:
		msg = &MessageBlockSig{}
		break
	case CmdBlockSigReq:
		msg = &MessageBlockSigReq{}
		break
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
	case reflect.TypeOf(&MessageBlockSig{}):
		return CmdBlockSig, nil
	case reflect.TypeOf(&MessageBlockSigReq{}):
		return CmdBlockSigReq, nil
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
	default:
		return "", fmt.Errorf("unhandled this message type [%s]", msgType)
	}
}
