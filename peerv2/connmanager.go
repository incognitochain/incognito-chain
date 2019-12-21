package peerv2

import (
	"context"
	"encoding/hex"
	"math/rand"
	"reflect"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/peerv2/proto"
	"github.com/incognitochain/incognito-chain/wire"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/stathat/consistent"
)

var HighwayBeaconID = byte(255)

func NewConnManager(
	host *Host,
	dpa string,
	ikey *incognitokey.CommitteePublicKey,
	cd ConsensusData,
	dispatcher *Dispatcher,
	nodeMode string,
	relayShard []byte,
) *ConnManager {
	pubkey, _ := ikey.ToBase58()
	return &ConnManager{
		info: info{
			consensusData: cd,
			pubkey:        pubkey,
			relayShard:    relayShard,
			nodeMode:      nodeMode,
			peerID:        host.Host.ID(),
		},
		LocalHost:            host,
		DiscoverPeersAddress: dpa,
		disp:                 dispatcher,
		IsMasterNode:         false,
		registerRequests:     make(chan int, 100),
		stop:                 make(chan int),
	}
}

func (cm *ConnManager) PublishMessage(msg wire.Message) error {
	var topic string
	publishable := []string{wire.CmdBlockShard, wire.CmdBFT, wire.CmdBlockBeacon, wire.CmdTx, wire.CmdCustomToken, wire.CmdPrivacyCustomToken, wire.CmdPeerState, wire.CmdBlkShardToBeacon, wire.CmdCrossShard}

	// msgCrossShard := msg.(wire.MessageCrossShard)
	msgType := msg.MessageType()
	subs := cm.subscriber.GetMsgToTopics()
	for _, p := range publishable {
		topic = ""
		if msgType == p {
			for _, availableTopic := range subs[msgType] {
				// Logger.Info("[hy]", availableTopic)
				if (availableTopic.Act == proto.MessageTopicPair_PUB) || (availableTopic.Act == proto.MessageTopicPair_PUBSUB) {
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
	subs := cm.subscriber.GetMsgToTopics()
	for _, p := range publishable {
		if msgType == p {
			// Get topic for mess
			for _, availableTopic := range subs[msgType] {
				Logger.Info(availableTopic)
				cID := GetCommitteeIDOfTopic(availableTopic.Name)
				if (byte(cID) == shardID) && ((availableTopic.Act == proto.MessageTopicPair_PUB) || (availableTopic.Act == proto.MessageTopicPair_PUBSUB)) {
					return broadcastMessage(msg, availableTopic.Name, cm.ps)
				}
			}
		}
	}

	Logger.Warn("Cannot publish message", msgType)
	return nil
}

func (cm *ConnManager) Start(ns NetSync) {
	// Pubsub
	var err error
	cm.ps, err = pubsub.NewFloodSub(context.Background(), cm.LocalHost.Host)
	if err != nil {
		panic(err)
	}
	cm.messages = make(chan *pubsub.Message, 1000)

	// Wait until connection to highway is established to make sure gRPC won't fail
	// NOTE: must Connect after creating FloodSub
	go cm.keepHighwayConnection()

	cm.Requester, err = NewRequester(cm.LocalHost.GRPC, peer.ID(""))
	if err != nil {
		panic(err)
	}

	cm.subscriber = NewSubManager(cm.info, cm.ps, cm.Requester, cm.messages)
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

type ForcedSubscriber interface {
	Subscribe(forced bool) error
	GetMsgToTopics() msgToTopics
}

type ConnManager struct {
	info         // info of running node
	LocalHost    *Host
	subscriber   ForcedSubscriber
	disconnected bool

	DiscoverPeersAddress string
	IsMasterNode         bool

	ps               *pubsub.PubSub
	messages         chan *pubsub.Message // queue messages from all topics
	registerRequests chan int

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
	// Init list of highways
	var currentHighway *peer.AddrInfo
	hwAddrs := []HighwayAddr{
		HighwayAddr{
			Libp2pAddr: "",
			RPCUrl:     cm.DiscoverPeersAddress,
		},
	}

	watchTimestep := time.Tick(10 * time.Second)   // Check connection every 10s
	refreshTimestep := time.Tick(30 * time.Minute) // RPC to update list of highways every 30 mins
	cm.disconnected = true                         // Init, to make first connection to highway
	pid := cm.LocalHost.Host.ID()
	for {
		select {
		case <-watchTimestep:
			if currentHighway == nil {
				var err error
				hwAddrs, err = updateHighwayAddrs(hwAddrs)
				if err != nil {
					Logger.Errorf("Fail updating highway addresses: %v", err)
					continue
				}

				currentHighway, err = chooseHighway(hwAddrs, pid)
				if err != nil {
					Logger.Errorf("Fail choosing highway: %v", err)
					continue
				}
			}

			cm.checkConnection(currentHighway)

		case <-refreshTimestep:
			var err error
			hwAddrs, err = updateHighwayAddrs(hwAddrs)
			if err != nil {
				Logger.Errorf("Fail updating highway addresses: %v", err)
				continue
			}

			newHighway, err := chooseHighway(hwAddrs, pid)
			if err != nil {
				Logger.Errorf("Fail choosing highway: %v", err)
				continue
			}

			if newHighway.ID != currentHighway.ID {
				err := cm.LocalHost.Host.Network().ClosePeer(currentHighway.ID) // Close connection to current highway
				if err != nil {
					Logger.Errorf("Failed closing connection to highway: %v %v %+v", currentHighway.ID, newHighway.ID, err)
				}
				currentHighway = newHighway
			}

		case <-cm.stop:
			Logger.Info("Stop keeping connection to highway")
			break
		}
	}
}

func (cm *ConnManager) checkConnection(addrInfo *peer.AddrInfo) {
	net := cm.LocalHost.Host.Network()
	// Reconnect if not connected
	if net.Connectedness(addrInfo.ID) != network.Connected {
		cm.disconnected = true
		Logger.Info("Not connected to highway, connecting")
		ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
		if err := cm.LocalHost.Host.Connect(ctx, *addrInfo); err != nil {
			Logger.Errorf("Could not connect to highway: %v %v", err, addrInfo)
		}
	}

	if cm.disconnected && net.Connectedness(addrInfo.ID) == network.Connected {
		// Register again since this might be a new highway
		Logger.Info("Connected to highway, sending register request")
		cm.registerRequests <- 1
		cm.disconnected = false
	}
}

func chooseHighway(hwAddrs []HighwayAddr, pid peer.ID) (*peer.AddrInfo, error) {
	if len(hwAddrs) == 0 {
		return nil, errors.New("cannot choose highway from empty list")
	}

	// Filter out bootnode addresss (only rpcUrl, no libp2p address)
	filterAddrs := []HighwayAddr{}
	for _, addr := range hwAddrs {
		if len(addr.Libp2pAddr) != 0 {
			filterAddrs = append(filterAddrs, addr)
		}
	}

	addr, err := choosePeer(filterAddrs, pid)
	if err != nil {
		return nil, err
	}
	return getAddressInfo(addr.Libp2pAddr)
}

// choosePeer picks a peer from a list using consistent hashing
func choosePeer(peers []HighwayAddr, id peer.ID) (HighwayAddr, error) {
	cst := consistent.New()
	for _, p := range peers {
		cst.Add(p.Libp2pAddr)
	}

	closest, err := cst.Get(string(id))
	if err != nil {
		return HighwayAddr{}, errors.Errorf("could not get consistent-hashing peer %v %v", peers, id)
	}

	for _, p := range peers {
		if p.Libp2pAddr == closest {
			return p, nil
		}
	}
	return HighwayAddr{}, errors.Errorf("could not find closest peer %v %v %v", peers, id, closest)
}

func updateHighwayAddrs(hwAddrs []HighwayAddr) ([]HighwayAddr, error) {
	// Pick random peer to get new list of highways
	if len(hwAddrs) == 0 {
		return nil, errors.New("No peer to get list of highways")
	}
	addr := hwAddrs[rand.Intn(len(hwAddrs))]
	return getHighwayAddrs(addr.RPCUrl)
}

func getHighwayAddrs(rpcUrl string) ([]HighwayAddr, error) {
	mapHWPerShard, err := DiscoverHighWay(rpcUrl, []string{"all"})
	if err != nil {
		return nil, err
	}
	Logger.Infof("Got %v from bootnode", mapHWPerShard)
	return mapHWPerShard["all"], nil
}

func getAddressInfo(libp2pAddr string) (*peer.AddrInfo, error) {
	addr, err := multiaddr.NewMultiaddr(libp2pAddr)
	if err != nil {
		return nil, errors.WithMessagef(err, "invalid libp2p address: %s", libp2pAddr)
	}
	hwPeerInfo, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		return nil, errors.WithMessagef(err, "invalid multi address: %s", addr)
	}
	return hwPeerInfo, nil
}

// manageRoleSubscription: polling current role periodically and subscribe to relevant topics
func (cm *ConnManager) manageRoleSubscription() {
	forced := false // only subscribe when role changed or last forced subscribe failed
	var err error
	for {
		select {
		case <-time.Tick(1 * time.Second):
			err = cm.subscriber.Subscribe(forced)
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
