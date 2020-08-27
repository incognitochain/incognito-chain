package peerv2

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"reflect"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/peer"
	"github.com/incognitochain/incognito-chain/wire"
	libp2p "github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
)

type Dispatcher struct {
	MessageListeners   *MessageListeners
	PublishableMessage []string
	BC                 *blockchain.BlockChain
	CurrentHWPeerID    libp2p.ID
}

// Just for consensus v1
func (d *Dispatcher) processStreamBlk(blktype byte, data []byte) error {
	//switch blktype {
	//case byte(proto.BlkType_BlkBc):
	//	newBlk := new(blockchain.BeaconBlock)
	//	err := wrapper.DeCom(data, newBlk)
	//	if err != nil {
	//		Logger.Errorf("[stream] process stream beacon block return error %v", err)
	//		return err
	//	}
	//	Logger.Infof("[stream] Got beacon block %v", newBlk.GetHeight())
	//	d.BC.OnBlockBeaconReceived(newBlk)
	//case byte(proto.BlkType_BlkShard):
	//	newBlk := new(blockchain.ShardBlock)
	//	err := wrapper.DeCom(data, newBlk)
	//	if err != nil {
	//		Logger.Errorf("[stream] process stream block return error %v", err)
	//		return err
	//	}
	//	Logger.Infof("[stream] Got Shard Block height %v, shard %v", newBlk.GetHeight(), newBlk.Header.ShardID)
	//	d.BC.OnBlockShardReceived(newBlk)
	//case byte(proto.BlkType_BlkS2B):
	//	newBlk := new(blockchain.ShardToBeaconBlock)
	//	err := wrapper.DeCom(data, newBlk)
	//	if err != nil {
	//		Logger.Errorf("[stream] process stream S2B block return error %v", err)
	//		return err
	//	}
	//	Logger.Infof("[stream] Got S2B block height %v shard %v", newBlk.GetHeight(), newBlk.Header.ShardID)
	//	d.BC.OnShardToBeaconBlockReceived(newBlk)
	//case byte(proto.BlkType_BlkXShard):
	//	newBlk := new(blockchain.CrossShardBlock)
	//	err := wrapper.DeCom(data, newBlk)
	//	Logger.Infof("[stream] Got block %v", newBlk.GetHeight())
	//	if err != nil {
	//		Logger.Errorf("[stream] process stream Cross shard block return error %v", err)
	//		return err
	//	}
	//	Logger.Infof("[stream] Got Cross block height %v shard %v to shard %v", newBlk.GetHeight(), newBlk.Header.ShardID, newBlk.ToShardID)
	//	d.BC.OnCrossShardBlockReceived(newBlk)
	//default:
	//	return errors.Errorf("[stream] Not implement for this block type %v", blktype)
	//}
	return nil
}

//TODO hy parse msg here
// processInMessageString - this is sub-function of InMessageHandler
// after receiving a good message from stream,
// we need analyze it and process with corresponding message type
func (d *Dispatcher) processInMessageString(msgStr string) error {
	ctx := Logger.NewContext(common.GenUUID())
	// NOTE: copy from peerConn.processInMessageString
	// Parse Message header from last 24 bytes header message
	jsonDecodeBytesRaw, err := hex.DecodeString(msgStr)
	if err != nil {
		Logger.Errorc(ctx, err)
		return errors.Wrapf(err, "msgStr: %v", msgStr)
	}

	// TODO(0xbunyip): separate caching from peerConn
	// // cache message hash
	// hashMsgRaw := common.HashH(jsonDecodeBytesRaw).String()
	// if peerConn.listenerPeer != nil {
	// 	if err := peerConn.listenerPeer.HashToPool(hashMsgRaw); err != nil {
	// 		Logger.Error(err)
	// 		return NewPeerError(HashToPoolError, err, nil)
	// 	}
	// }
	// unzip data before process
	jsonDecodeBytes, err := common.GZipToBytes(jsonDecodeBytesRaw)
	if err != nil {
		Logger.Errorc(ctx, err)
		return errors.WithStack(err)
	}

	// fmt.Printf("In message content : %s", string(jsonDecodeBytes))

	// Parse Message body
	messageBody := jsonDecodeBytes[:len(jsonDecodeBytes)-wire.MessageHeaderSize]

	messageHeader := jsonDecodeBytes[len(jsonDecodeBytes)-wire.MessageHeaderSize:]

	// get cmd type in header message
	commandInHeader := bytes.Trim(messageHeader[:wire.MessageCmdTypeSize], "\x00")
	commandType := string(messageHeader[:len(commandInHeader)])
	// convert to particular message from message cmd type
	message, err := wire.MakeEmptyMessage(string(commandType))
	if err != nil {
		Logger.Errorc(ctx, err)
		return errors.WithStack(err)
	}

	if len(jsonDecodeBytes) > message.MaxPayloadLength(wire.Version) {
		err := errors.Errorf("Message size too lagre %v, it must be less than %v", len(jsonDecodeBytes), message.MaxPayloadLength(wire.Version))
		Logger.Errorc(ctx, err)
		return err
	}

	err = json.Unmarshal(messageBody, &message)
	if err != nil {
		Logger.Errorc(ctx, err)
		return errors.WithStack(err)
	}
	// check forward TODO
	/*if peerConn.config.MessageListeners.GetCurrentRoleShard != nil {
		cRole, cShard := peerConn.config.MessageListeners.GetCurrentRoleShard()
		if cShard != nil {
			fT := messageHeader[wire.MessageCmdTypeSize]
			if fT == MessageToShard {
				fS := messageHeader[wire.MessageCmdTypeSize+1]
				if *cShard != fS {
					if peerConn.config.MessageListeners.PushRawBytesToShard != nil {
						err1 := peerConn.config.MessageListeners.PushRawBytesToShard(ctx, peerConn, &jsonDecodeBytesRaw, *cShard)
						if err1 != nil {
							Logger.Error(err1)
						}
					}
					return NewPeerError(CheckForwardError, err, nil)
				}
			}
		}
		if cRole != "" {
			fT := messageHeader[wire.MessageCmdTypeSize]
			if fT == MessageToBeacon && cRole != "beacon" {
				if peerConn.config.MessageListeners.PushRawBytesToBeacon != nil {
					err1 := peerConn.config.MessageListeners.PushRawBytesToBeacon(ctx, peerConn, &jsonDecodeBytesRaw)
					if err1 != nil {
						Logger.Error(err1)
					}
				}
				return NewPeerError(CheckForwardError, err, nil)
			}
		}
	}*/

	realType := reflect.TypeOf(message)
	// fmt.Printf("Cmd message type of struct %s", realType.String())

	// // cache message hash
	// if peerConn.listenerPeer != nil {
	// 	hashMsg := message.Hash()
	// 	if err := peerConn.listenerPeer.HashToPool(hashMsg); err != nil {
	// 		Logger.Error(err)
	// 		return NewPeerError(CacheMessageHashError, err, nil)
	// 	}
	// }

	// process message for each of message type
	errProcessMessage := d.processMessageForEachType(ctx, realType, message)
	if errProcessMessage != nil {
		return errors.WithStack(errProcessMessage)
	}

	// MONITOR INBOUND MESSAGE
	//storeInboundPeerMessage(message, time.Now().Unix(), peerConn.remotePeer.GetPeerID())
	return nil
}

// process message for each of message type
func (d *Dispatcher) processMessageForEachType(ctx context.Context, messageType reflect.Type, message wire.Message) error {
	// NOTE: copy from peerConn.processInMessageString
	Logger.Infofc(ctx, "Processing msgType %s", message.MessageType())
	peerConn := &peer.PeerConn{}
	peerConn.SetRemotePeerID(d.CurrentHWPeerID)
	//fmt.Printf("[stream2] %v\n", peerConn.GetRemotePeerID())
	switch messageType {
	case reflect.TypeOf(&wire.MessageTx{}):
		if d.MessageListeners.OnTx != nil {
			d.MessageListeners.OnTx(ctx, peerConn, message.(*wire.MessageTx))
		}
	case reflect.TypeOf(&wire.MessageTxPrivacyToken{}):
		if d.MessageListeners.OnTxPrivacyToken != nil {
			d.MessageListeners.OnTxPrivacyToken(ctx, peerConn, message.(*wire.MessageTxPrivacyToken))
		}
	case reflect.TypeOf(&wire.MessageBlockShard{}):
		// Logger.Infof("Processing msgContent %+v", message.(*wire.MessageBlockShard).Block)
		if d.MessageListeners.OnBlockShard != nil {
			d.MessageListeners.OnBlockShard(ctx, peerConn, message.(*wire.MessageBlockShard))
		}
	case reflect.TypeOf(&wire.MessageBlockBeacon{}):
		// Logger.Infof("Processing msgContent %+v", message.(*wire.MessageBlockBeacon).Block)
		if d.MessageListeners.OnBlockBeacon != nil {
			d.MessageListeners.OnBlockBeacon(ctx, peerConn, message.(*wire.MessageBlockBeacon))
		}
	case reflect.TypeOf(&wire.MessageCrossShard{}):
		// Logger.Infof("Processing msgContent %+v", message.(*wire.MessageCrossShard).Block)
		if d.MessageListeners.OnCrossShard != nil {
			d.MessageListeners.OnCrossShard(ctx, peerConn, message.(*wire.MessageCrossShard))
		}
	case reflect.TypeOf(&wire.MessageShardToBeacon{}):
		// Logger.Infof("Processing msgContent %+v", message.(*wire.MessageShardToBeacon).Block)
		if d.MessageListeners.OnShardToBeacon != nil {
			d.MessageListeners.OnShardToBeacon(ctx, peerConn, message.(*wire.MessageShardToBeacon))
		}
	case reflect.TypeOf(&wire.MessageGetBlockBeacon{}):
		if d.MessageListeners.OnGetBlockBeacon != nil {
			d.MessageListeners.OnGetBlockBeacon(ctx, peerConn, message.(*wire.MessageGetBlockBeacon))
		}
	case reflect.TypeOf(&wire.MessageGetBlockShard{}):
		if d.MessageListeners.OnGetBlockShard != nil {
			d.MessageListeners.OnGetBlockShard(ctx, peerConn, message.(*wire.MessageGetBlockShard))
		}
	case reflect.TypeOf(&wire.MessageGetCrossShard{}):
		if d.MessageListeners.OnGetCrossShard != nil {
			d.MessageListeners.OnGetCrossShard(ctx, peerConn, message.(*wire.MessageGetCrossShard))
		}
	case reflect.TypeOf(&wire.MessageGetShardToBeacon{}):
		if d.MessageListeners.OnGetShardToBeacon != nil {
			d.MessageListeners.OnGetShardToBeacon(ctx, peerConn, message.(*wire.MessageGetShardToBeacon))
		}
	case reflect.TypeOf(&wire.MessageVersion{}):
		if d.MessageListeners.OnVersion != nil {
			d.MessageListeners.OnVersion(ctx, peerConn, message.(*wire.MessageVersion))
		}
	case reflect.TypeOf(&wire.MessageVerAck{}):
		// d.verAckReceived = true
		if d.MessageListeners.OnVerAck != nil {
			d.MessageListeners.OnVerAck(ctx, peerConn, message.(*wire.MessageVerAck))
		}
	case reflect.TypeOf(&wire.MessageGetAddr{}):
		if d.MessageListeners.OnGetAddr != nil {
			d.MessageListeners.OnGetAddr(ctx, peerConn, message.(*wire.MessageGetAddr))
		}
	case reflect.TypeOf(&wire.MessageAddr{}):
		if d.MessageListeners.OnGetAddr != nil {
			d.MessageListeners.OnAddr(ctx, peerConn, message.(*wire.MessageAddr))
		}
	case reflect.TypeOf(&wire.MessageBFT{}):
		if d.MessageListeners.OnBFTMsg != nil {
			d.MessageListeners.OnBFTMsg(ctx, peerConn, message.(*wire.MessageBFT))
		}
	case reflect.TypeOf(&wire.MessagePeerState{}):
		if d.MessageListeners.OnPeerState != nil {
			d.MessageListeners.OnPeerState(ctx, peerConn, message.(*wire.MessagePeerState))
		}

	// case reflect.TypeOf(&wire.MessageMsgCheck{}):
	// 	err1 := peerConn.handleMsgCheck(message.(*wire.MessageMsgCheck))
	// 	if err1 != nil {
	// 		Logger.Error(err1)
	// 	}
	// case reflect.TypeOf(&wire.MessageMsgCheckResp{}):
	// 	err1 := peerConn.handleMsgCheckResp(message.(*wire.MessageMsgCheckResp))
	// 	if err1 != nil {
	// 		Logger.Error(err1)
	// 	}
	default:
		return errors.Errorf("InMessageHandler Received unhandled message of type % from %v", messageType, peerConn)
	}
	return nil
}

type MessageListeners struct {
	OnTx               func(ctx context.Context, p *peer.PeerConn, msg *wire.MessageTx)
	OnTxPrivacyToken   func(ctx context.Context, p *peer.PeerConn, msg *wire.MessageTxPrivacyToken)
	OnBlockShard       func(ctx context.Context, p *peer.PeerConn, msg *wire.MessageBlockShard)
	OnBlockBeacon      func(ctx context.Context, p *peer.PeerConn, msg *wire.MessageBlockBeacon)
	OnCrossShard       func(ctx context.Context, p *peer.PeerConn, msg *wire.MessageCrossShard)
	OnShardToBeacon    func(ctx context.Context, p *peer.PeerConn, msg *wire.MessageShardToBeacon)
	OnGetBlockBeacon   func(ctx context.Context, p *peer.PeerConn, msg *wire.MessageGetBlockBeacon)
	OnGetBlockShard    func(ctx context.Context, p *peer.PeerConn, msg *wire.MessageGetBlockShard)
	OnGetCrossShard    func(ctx context.Context, p *peer.PeerConn, msg *wire.MessageGetCrossShard)
	OnGetShardToBeacon func(ctx context.Context, p *peer.PeerConn, msg *wire.MessageGetShardToBeacon)
	OnVersion          func(ctx context.Context, p *peer.PeerConn, msg *wire.MessageVersion)
	OnVerAck           func(ctx context.Context, p *peer.PeerConn, msg *wire.MessageVerAck)
	OnGetAddr          func(ctx context.Context, p *peer.PeerConn, msg *wire.MessageGetAddr)
	OnAddr             func(ctx context.Context, p *peer.PeerConn, msg *wire.MessageAddr)

	//PBFT
	OnBFTMsg    func(ctx context.Context, p *peer.PeerConn, msg wire.Message)
	OnPeerState func(ctx context.Context, p *peer.PeerConn, msg *wire.MessagePeerState)
}
