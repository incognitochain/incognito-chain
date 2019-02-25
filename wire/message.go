package wire

import (
	"fmt"
	"reflect"

	peer "github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/common"

	"time"

	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/transaction"
)

// list message type
const (
	TypicalMessageSize = 1000
	MessageHeaderSize  = 24
	MessageCmdTypeSize = 12

	CmdBlockShard         = "blockshard"
	CmdBlockBeacon        = "blockbeacon"
	CmdGetCrossShard      = "getcrossshd"
	CmdCrossShard         = "crossshard"
	CmdGetShardToBeacon   = "getshdtobcn"
	CmdBlkShardToBeacon   = "blkshdtobcn"
	CmdTx                 = "tx"
	CmdCustomToken        = "txtoken"
	CmdPrivacyCustomToken = "txprivacytok"
	CmdCLoanRequestToken  = "txloanreq"
	CmdCLoanResponseToken = "txloanres"
	CmdCLoanWithdrawToken = "txloanwith"
	CmdCLoanPayToken      = "txloanpay"
	CmdGetBlockBeacon     = "getblkbeacon"
	CmdGetBlockShard      = "getblkshard"
	CmdInv                = "inv"
	CmdGetData            = "getdata"
	CmdVersion            = "version"
	CmdVerack             = "verack"
	CmdGetAddr            = "getaddr"
	CmdAddr               = "addr"
	CmdPing               = "ping"

	// POS Cmd
	CmdBFTPropose   = "bftpropose"
	CmdBFTPrepare   = "bftprepare"
	CmdBFTCommit    = "bftcommit"
	CmdBFTReady     = "bftready"
	CmdInvalidBlock = "invalidblock"
	CmdPeerState    = "peerstate"

	// SWAP Cmd
	//CmdSwapRequest = "swaprequest"
	//CmdSwapSig     = "swapsig"
	//CmdSwapUpdate  = "swapupdate"

	// heavy message check cmd
	CmdMsgCheck     = "msgcheck"
	CmdMsgCheckResp = "msgcheckresp"
)

// Interface for message wire on P2P network
type Message interface {
	Hash() string
	MessageType() string
	MaxPayloadLength(int) int
	JsonSerialize() ([]byte, error)
	JsonDeserialize(string) error
	SetSenderID(peer.ID) error

	//SignMsg sig this msg with a keyset
	SignMsg(*cashec.KeySet) error

	//VerifyMsgSanity verify msg before push it to final handler
	VerifyMsgSanity() error
}

func MakeEmptyMessage(messageType string) (Message, error) {
	var msg Message
	switch messageType {
	case CmdBlockBeacon:
		msg = &MessageBlockBeacon{}
		break
	case CmdBlockShard:
		msg = &MessageBlockShard{}
		break
	case CmdGetCrossShard:
		msg = &MessageGetCrossShard{}
		break
	case CmdCrossShard:
		msg = &MessageCrossShard{}
		break
	case CmdGetShardToBeacon:
		msg = &MessageGetShardToBeacon{}
		break
	case CmdBlkShardToBeacon:
		msg = &MessageShardToBeacon{}
		break
	case CmdCustomToken:
		msg = &MessageTx{
			Transaction: &transaction.TxCustomToken{},
		}
		break
	case CmdPrivacyCustomToken:
		msg = &MessageTx{
			Transaction: &transaction.TxCustomTokenPrivacy{},
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
			Transaction: &transaction.Tx{
				Metadata: &metadata.LoanWithdraw{},
			},
		}
		break
	case CmdCLoanPayToken:
		msg = &MessageTx{
			Transaction: &transaction.Tx{
				Metadata: &metadata.LoanPayment{},
			},
		}
		break
	case CmdGetBlockBeacon:
		msg = &MessageGetBlockBeacon{
			Timestamp: time.Now().Unix(),
		}
		break
	case CmdGetBlockShard:
		msg = &MessageGetBlockShard{}
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
	case CmdBFTReady:
		msg = &MessageBFTReady{
			Timestamp: time.Now().Unix(),
		}
		break
	case CmdBFTPropose:
		msg = &MessageBFTPropose{
			Timestamp: time.Now().Unix(),
		}
		break
	case CmdBFTPrepare:
		msg = &MessageBFTPrepare{
			Timestamp: time.Now().Unix(),
		}
		break
	case CmdBFTCommit:
		msg = &MessageBFTCommit{
			Timestamp: time.Now().Unix(),
		}
		break
	case CmdPeerState:
		msg = &MessagePeerState{
			Timestamp:         time.Now().Unix(),
			Shards:            make(map[byte]blockchain.ChainState),
			ShardToBeaconPool: make(map[byte][]common.Hash),
			CrossShardPool:    make(map[byte]map[byte][]common.Hash),
		}
		break
	case CmdGetAddr:
		msg = &MessageGetAddr{
			Timestamp: time.Now(),
		}
		break
	case CmdAddr:
		msg = &MessageAddr{
			Timestamp: time.Now(),
		}
		break
	case CmdPing:
		msg = &MessagePing{}
		break
		// case CmdSwapRequest:
		// 	msg = &MessageSwapRequest{}
		// 	break
		// case CmdSwapSig:
		// 	msg = &MessageSwapSig{}
		// 	break
		// case CmdSwapUpdate:
		// 	msg = &MessageSwapUpdate{
		// 		Signatures: make(map[string]string),
		// 	}
		// 	break
	case CmdMsgCheck:
		msg = &MessageMsgCheck{
			Timestamp: time.Now().UnixNano(),
		}
		break
	case CmdMsgCheckResp:
		msg = &MessageMsgCheckResp{
			Timestamp: time.Now().UnixNano(),
		}
		break
	default:
		return nil, fmt.Errorf("unhandled this message type [%s]", messageType)
	}
	return msg, nil
}

func GetCmdType(msgType reflect.Type) (string, error) {
	switch msgType {
	case reflect.TypeOf(&MessageBlockBeacon{}):
		return CmdBlockBeacon, nil
	case reflect.TypeOf(&MessageBlockShard{}):
		return CmdBlockShard, nil
	case reflect.TypeOf(&MessageGetCrossShard{}):
		return CmdGetCrossShard, nil
	case reflect.TypeOf(&MessageCrossShard{}):
		return CmdCrossShard, nil
	case reflect.TypeOf(&MessageGetShardToBeacon{}):
		return CmdGetShardToBeacon, nil
	case reflect.TypeOf(&MessageShardToBeacon{}):
		return CmdBlkShardToBeacon, nil
	case reflect.TypeOf(&MessageGetBlockBeacon{}):
		return CmdGetBlockBeacon, nil
	case reflect.TypeOf(&MessageGetBlockShard{}):
		return CmdGetBlockShard, nil
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
	case reflect.TypeOf(&MessageBFTReady{}):
		return CmdBFTReady, nil
	case reflect.TypeOf(&MessageInvalidBlock{}):
		return CmdInvalidBlock, nil
	case reflect.TypeOf(&MessagePeerState{}):
		return CmdPeerState, nil
		//case reflect.TypeOf(&MessageSwapRequest{}):
		//	return CmdSwapRequest, nil
		//case reflect.TypeOf(&MessageSwapSig{}):
		//	return CmdSwapSig, nil
		//case reflect.TypeOf(&MessageSwapUpdate{}):
		//	return CmdSwapUpdate, nil
	case reflect.TypeOf(&MessageMsgCheck{}):
		return CmdMsgCheck, nil
	case reflect.TypeOf(&MessageMsgCheckResp{}):
		return CmdMsgCheckResp, nil
	default:
		return "", fmt.Errorf("unhandled this message type [%s]", msgType)
	}
}
