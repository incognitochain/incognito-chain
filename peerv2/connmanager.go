package peerv2

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/rpc"
	"reflect"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wire"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
)

var HighwayBeaconID = byte(255)

func DiscoverHighWay(
	discoverPeerAddress string,
	shardsStr []string,
) (
	map[string][]string,
	error,
) {
	if discoverPeerAddress == common.EmptyString {
		return nil, errors.Errorf("Not config discovery peer")
	}
	client := new(rpc.Client)
	var err error
	for {
		client, err = rpc.Dial("tcp", discoverPeerAddress)
		Logger.Info("Dialing...")
		if err != nil {
			// can not create connection to rpc server with
			// provided "discover peer address" in config
			Logger.Errorf("Connect to discover peer %v return error %v:", discoverPeerAddress, err)
			time.Sleep(2 * time.Second)
		} else {
			Logger.Info("Connected to %v", discoverPeerAddress)
			break
		}
	}
	if client != nil {
		defer client.Close()

		req := Request{
			Shard: shardsStr,
		}
		var res Response
		Logger.Infof("Start dialing RPC server with param %v\n", req)

		err = client.Call("Handler.GetPeers", req, &res)

		if err != nil {
			Logger.Errorf("Call Handler.GetPeers return error %v", err)
			return nil, err
		} else {
			Logger.Infof("Bootnode return %v\n", res.PeerPerShard)
			return res.PeerPerShard, nil
		}
	}
	return nil, errors.Errorf("Can not get any from bootnode")
}

func NewConnManager(
	host *Host,
	dpa string,
	ikey *incognitokey.CommitteePublicKey,
	cd ConsensusData,
	dispatcher *Dispatcher,
	nodeMode string,
	relayShard []byte,
) *ConnManager {
	return &ConnManager{
		LocalHost:            host,
		DiscoverPeersAddress: dpa,
		IdentityKey:          ikey,
		cd:                   cd,
		disp:                 dispatcher,
		IsMasterNode:         false,
		registerRequests:     make(chan int, 100),
		relayShard:           relayShard,
		nodeMode:             nodeMode,
		stop:                 make(chan int),
	}
}

func (cm *ConnManager) PublishMessage(msg wire.Message) error {
	var topic string
	publishable := []string{wire.CmdBlockShard, wire.CmdBFT, wire.CmdBlockBeacon, wire.CmdTx, wire.CmdCustomToken, wire.CmdPrivacyCustomToken, wire.CmdPeerState, wire.CmdBlkShardToBeacon, wire.CmdCrossShard}

	// msgCrossShard := msg.(wire.MessageCrossShard)
	msgType := msg.MessageType()
	for _, p := range publishable {
		topic = ""
		if msgType == p {
			for _, availableTopic := range cm.subs[msgType] {
				// Logger.Info("[hy]", availableTopic)
				if (availableTopic.Act == MessageTopicPair_PUB) || (availableTopic.Act == MessageTopicPair_PUBSUB) {
					topic = availableTopic.Name
					err := broadcastMessage(msg, topic, cm.ps)
					if err != nil {
						Logger.Errorf("Broadcast to topic %v error %v", topic, err)
						return err
					}
				}

			}
			if topic == "" {
				return errors.New("Can not find topic of this message type " + msgType + "for publish")
			}

			// return broadcastMessage(msg, topic, cm.ps)
		}
	}

	return nil
}

func (cm *ConnManager) PublishMessageToShard(msg wire.Message, shardID byte) error {
	publishable := []string{wire.CmdBlockShard, wire.CmdCrossShard, wire.CmdBFT}
	msgType := msg.MessageType()
	for _, p := range publishable {
		if msgType == p {
			// Get topic for mess
			for _, availableTopic := range cm.subs[msgType] {
				Logger.Info(availableTopic)
				cID := GetCommitteeIDOfTopic(availableTopic.Name)
				if (byte(cID) == shardID) && ((availableTopic.Act == MessageTopicPair_PUB) || (availableTopic.Act == MessageTopicPair_PUBSUB)) {
					return broadcastMessage(msg, availableTopic.Name, cm.ps)
				}
			}
		}
	}

	Logger.Warn("Cannot publish message", msgType)
	return nil
}

func (cm *ConnManager) Start(ns NetSync) {
	mapHWPerShard := map[string][]string{}
	var err error
	for {
		mapHWPerShard, err = DiscoverHighWay(cm.DiscoverPeersAddress, []string{"all"})
		if err != nil {
			Logger.Errorf("DiscoverHighWay return erro: %v", err)
			time.Sleep(5 * time.Second)
			Logger.Infof("Re connect to bootnode!")
		} else {
			Logger.Infof("Got %v from bootnode", mapHWPerShard)
			break
		}
	}

	// TODO remove hardcode here
	hwPeerIDForAllShard := mapHWPerShard["all"][0]
	cm.HighwayAddress = hwPeerIDForAllShard

	// connect to highway
	addr, err := multiaddr.NewMultiaddr(cm.HighwayAddress)
	if err != nil {
		panic(err)
	}

	addrInfo, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		panic(err)
	}

	// Pubsub
	// TODO(@0xbunyip): handle error
	cm.ps, _ = pubsub.NewFloodSub(context.Background(), cm.LocalHost.Host)
	cm.messages = make(chan *pubsub.Message, 1000)

	// Wait until connection to highway is established to make sure gRPC won't fail
	// NOTE: must Connect after creating FloodSub
	go cm.keepHighwayConnection()

	req, err := NewRequester(cm.LocalHost.GRPC, addrInfo.ID)
	if err != nil {
		panic(err)
	}
	cm.Requester = req

	cm.Provider = NewBlockProvider(cm.LocalHost.GRPC, ns)

	go cm.manageRoleSubscription()

	cm.process()
}

// BroadcastCommittee floods message to topic `chain_committee` for highways
// Only masternode actually does the broadcast, other's messages will be ignored by highway
func (cm *ConnManager) BroadcastCommittee(
	epoch uint64,
	newBeaconCommittee []incognitokey.CommitteePublicKey,
	newAllShardCommittee map[byte][]incognitokey.CommitteePublicKey,
	newAllShardPending map[byte][]incognitokey.CommitteePublicKey,
) {
	// NOTE: disabled feature, always return for now
	if !cm.IsMasterNode {
		return
	}

	Logger.Info("Broadcasting committee to highways!!!")
	cc := &incognitokey.ChainCommittee{
		Epoch:             epoch,
		BeaconCommittee:   newBeaconCommittee,
		AllShardCommittee: newAllShardCommittee,
		AllShardPending:   newAllShardPending,
	}
	data, err := cc.ToByte()
	if err != nil {
		Logger.Error(err)
		return
	}

	topic := "chain_committee"
	err = cm.ps.Publish(topic, data)
	if err != nil {
		Logger.Error(err)
	}
}

type Topic struct {
	Name string
	Sub  *pubsub.Subscription
	Act  MessageTopicPair_Action
}

type Subscriber interface {
	Subscribe(forced bool) error
}

type ConnManager struct {
	LocalHost  *Host
	Subscriber Subscriber

	DiscoverPeersAddress string
	HighwayAddress       string
	IdentityKey          *incognitokey.CommitteePublicKey
	IsMasterNode         bool

	ps               *pubsub.PubSub
	messages         chan *pubsub.Message // queue messages from all topics
	registerRequests chan int

	nodeMode   string
	relayShard []byte

	cd        ConsensusData
	disp      *Dispatcher
	Requester *BlockRequester
	Provider  *BlockProvider

	stop chan int
}

func (cm *ConnManager) PutMessage(msg *pubsub.Message) {
	cm.messages <- msg
}

func (cm *ConnManager) process() {
	for {
		select {
		case msg := <-cm.messages:
			err := cm.disp.processInMessageString(string(msg.Data))
			if err != nil {
				Logger.Warn(err)
			}
		}
	}
}

// keepHighwayConnection periodically checks liveliness of connection to highway
// and try to connect if it's not available.
func (cm *ConnManager) keepHighwayConnection() {
	addr, err := multiaddr.NewMultiaddr(cm.HighwayAddress)
	if err != nil {
		panic(fmt.Sprintf("invalid discover peers address: %v", cm.HighwayAddress))
	}

	hwPeerInfo, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		panic(err)
	}
	hwPID := hwPeerInfo.ID

	net := cm.LocalHost.Host.Network()
	disconnected := true
	for ; true; <-time.Tick(1 * time.Second) {
		// Reconnect if not connected
		var err error
		if net.Connectedness(hwPID) != network.Connected {
			disconnected = true
			Logger.Info("Not connected to highway, connecting")
			ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
			if err = cm.LocalHost.Host.Connect(ctx, *hwPeerInfo); err != nil {
				Logger.Errorf("Could not connect to highway: %v %v", err, hwPeerInfo)
			}
		}

		if disconnected && net.Connectedness(hwPID) == network.Connected {
			// Register again since this might be a new highway
			Logger.Info("Connected to highway, sending register request")
			cm.registerRequests <- 1
			disconnected = false
		}

		select {
		case <-cm.stop:
			Logger.Info("Stop keeping connection to highway")
			break

		default:
		}
	}
}

func encodeMessage(msg wire.Message) (string, error) {
	// NOTE: copy from peerConn.outMessageHandler
	// Create messageHex
	messageBytes, err := msg.JsonSerialize()
	if err != nil {
		Logger.Error("Can not serialize json format for messageHex:"+msg.MessageType(), err)
		return "", err
	}

	// Add 24 bytes headerBytes into messageHex
	headerBytes := make([]byte, wire.MessageHeaderSize)
	// add command type of message
	cmdType, messageErr := wire.GetCmdType(reflect.TypeOf(msg))
	if messageErr != nil {
		Logger.Error("Can not get cmd type for "+msg.MessageType(), messageErr)
		return "", err
	}
	copy(headerBytes[:], []byte(cmdType))
	// add forward type of message at 13st byte
	forwardType := byte('s')
	forwardValue := byte(0)
	copy(headerBytes[wire.MessageCmdTypeSize:], []byte{forwardType})
	copy(headerBytes[wire.MessageCmdTypeSize+1:], []byte{forwardValue})
	messageBytes = append(messageBytes, headerBytes...)
	// Logger.Infof("Encoded message TYPE %s CONTENT %s", cmdType, string(messageBytes))

	// zip data before send
	messageBytes, err = common.GZipFromBytes(messageBytes)
	if err != nil {
		Logger.Error("Can not gzip for messageHex:"+msg.MessageType(), err)
		return "", err
	}
	messageHex := hex.EncodeToString(messageBytes)
	//log.Debugf("Content in hex encode: %s", string(messageHex))
	// add end character to messageHex (delim '\n')
	// messageHex += "\n"
	return messageHex, nil
}

func broadcastMessage(msg wire.Message, topic string, ps *pubsub.PubSub) error {
	// Encode message to string first
	messageHex, err := encodeMessage(msg)
	if err != nil {
		return err
	}

	// Broadcast
	Logger.Infof("Publishing to topic %s", topic)
	return ps.Publish(topic, []byte(messageHex))
}

// manageRoleSubscription: polling current role every minute and subscribe to relevant topics
func (cm *ConnManager) manageRoleSubscription() {
	forced := false // only subscribe when role changed or last forced subscribe failed
	var err error
	for {
		select {
		case <-time.Tick(10 * time.Second):
			err = cm.Subscriber.Subscribe(forced)
			if err != nil {
				Logger.Errorf("subscribe failed: %v %+v", forced, err)
			} else {
				forced = false
			}

		case <-cm.registerRequests:
			Logger.Info("Received request to register")
			forced = true // register no matter if role changed or not

		case <-cm.stop:
			Logger.Info("Stop managing role subscription")
			break
		}
	}
}
