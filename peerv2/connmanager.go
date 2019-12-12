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
	cm.subs = m2t{}
	cm.messages = make(chan *pubsub.Message, 1000)

	// Wait until connection to highway is established to make sure gRPC won't fail
	// NOTE: must Connect after creating FloodSub
	connected := make(chan error)
	go cm.keepHighwayConnection(connected)
	<-connected

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

type ConsensusData interface {
	GetUserRole() (string, string, int)
}

type Topic struct {
	Name string
	Sub  *pubsub.Subscription
	Act  MessageTopicPair_Action
}

type ConnManager struct {
	LocalHost            *Host
	DiscoverPeersAddress string
	HighwayAddress       string
	IdentityKey          *incognitokey.CommitteePublicKey
	IsMasterNode         bool

	ps               *pubsub.PubSub
	subs             m2t                  // mapping from message to topic's subscription
	messages         chan *pubsub.Message // queue messages from all topics
	registerRequests chan int

	nodeMode   string
	relayShard []byte

	cd        ConsensusData
	disp      *Dispatcher
	Requester *BlockRequester
	Provider  *BlockProvider
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
// The method push data to the given channel to signal that the first attempt had finished.
// Constructor can use this info to initialize other objects.
func (cm *ConnManager) keepHighwayConnection(connectedOnce chan error) {
	addr, err := multiaddr.NewMultiaddr(cm.HighwayAddress)
	if err != nil {
		panic(fmt.Sprintf("invalid discover peers address: %v", cm.HighwayAddress))
	}

	hwPeerInfo, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		panic(err)
	}
	hwPID := hwPeerInfo.ID

	first := true
	net := cm.LocalHost.Host.Network()
	disconnected := true
	for ; true; <-time.Tick(10 * time.Second) {
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

		// Notify that first attempt had finished
		if first {
			connectedOnce <- err
			first = false
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
	role := newUserRole("dummyLayer", "dummyRole", -1000)
	topics := m2t{}
	forced := false // only subscribe when role changed or last forced subscribe failed
	var err error
	for {
		select {
		case <-time.Tick(10 * time.Second):
			role, topics, err = cm.subscribe(role, topics, forced)
			if err != nil {
				Logger.Errorf("subscribe failed: %v %+v", forced, err)
			} else {
				forced = false
			}

		case <-cm.registerRequests:
			Logger.Info("Received request to register")
			forced = true // register no matter if role changed or not
		}
	}
}

func (cm *ConnManager) subscribe(role userRole, topics m2t, forced bool) (userRole, m2t, error) {
	newRole := newUserRole(cm.cd.GetUserRole())
	if newRole == role && !forced { // Not forced => no need to subscribe when role stays the same
		return newRole, topics, nil
	}
	Logger.Infof("Role changed: %v -> %v", role, newRole)

	// Registering
	pubkey, _ := cm.IdentityKey.ToBase58()
	roleSID := newRole.shardID
	if roleSID == -2 { // normal node
		roleSID = -1
	}
	shardIDs := []byte{}
	if cm.nodeMode == common.NodeModeRelay {
		shardIDs = cm.relayShard
		shardIDs = append(shardIDs, HighwayBeaconID)
	} else {
		shardIDs = append(shardIDs, byte(roleSID))
	}
	newTopics, roleOfTopics, err := cm.registerToProxy(pubkey, newRole.layer, newRole.role, shardIDs)
	if err != nil {
		return role, topics, err
	}

	// NOTE: disabled, highway always return the same role
	_ = roleOfTopics
	// if newRole != roleOfTopics {
	// 	return role, topics, errors.Errorf("lole not matching with highway, local = %+v, highway = %+v", newRole, roleOfTopics)
	// }

	Logger.Infof("Received topics = %+v, oldTopics = %+v", newTopics, topics)

	// Subscribing
	if err := cm.subscribeNewTopics(newTopics, topics); err != nil {
		return role, topics, err
	}

	return newRole, newTopics, nil
}

type userRole struct {
	layer   string
	role    string
	shardID int
}

func newUserRole(layer, role string, shardID int) userRole {
	return userRole{
		layer:   layer,
		role:    role,
		shardID: shardID,
	}
}

// subscribeNewTopics subscribes to new topics and unsubcribes any topics that aren't needed anymore
func (cm *ConnManager) subscribeNewTopics(newTopics, subscribed m2t) error {
	found := func(tName string, tmap m2t) bool {
		for _, topicList := range tmap {
			for _, t := range topicList {
				if tName == t.Name {
					return true
				}
			}
		}
		return false
	}

	// Subscribe to new topics
	for m, topicList := range newTopics {
		Logger.Infof("Process message %v and topic %v", m, topicList)
		for _, t := range topicList {

			if found(t.Name, subscribed) {
				Logger.Infof("Countinue 1 %v %v", t.Name, subscribed)
				continue
			}

			// TODO(@0xakk0r0kamui): check here
			if t.Act == MessageTopicPair_PUB {
				cm.subs[m] = append(cm.subs[m], Topic{Name: t.Name, Sub: nil, Act: t.Act})
				Logger.Infof("Countinue 2 %v %v", t.Name, subscribed)
				continue
			}

			Logger.Info("subscribing", m, t.Name)

			s, err := cm.ps.Subscribe(t.Name)
			if err != nil {
				return errors.WithStack(err)
			}
			cm.subs[m] = append(cm.subs[m], Topic{Name: t.Name, Sub: s, Act: t.Act})
			go processSubscriptionMessage(cm.messages, s)
		}
	}

	// Unsubscribe to old ones
	for m, topicList := range subscribed {
		for _, t := range topicList {
			if found(t.Name, newTopics) {
				continue
			}

			// TODO(@0xakk0r0kamui): check here
			if t.Act == MessageTopicPair_PUB {
				continue
			}

			Logger.Info("unsubscribing", m, t.Name)
			for _, s := range cm.subs[m] {
				if s.Name == t.Name {
					s.Sub.Cancel() // TODO(@0xbunyip): lock
				}
			}
			delete(cm.subs, m)
		}
	}
	return nil
}

// processSubscriptionMessage listens to a topic and pushes all messages to a queue to be processed later
func processSubscriptionMessage(inbox chan *pubsub.Message, sub *pubsub.Subscription) {
	ctx := context.Background()
	for {
		// TODO(@0xbunyip): check if topic is unsubbed then return, otherwise just continue
		msg, err := sub.Next(ctx)
		if err != nil { // Subscription might have been cancelled
			Logger.Warn(err)
			return
		}

		inbox <- msg
	}
}

type m2t map[string][]Topic // Message to topics

func (cm *ConnManager) registerToProxy(
	pubkey string,
	layer string,
	role string,
	shardID []byte,
) (m2t, userRole, error) {
	messagesWanted := getMessagesForLayer(cm.nodeMode, layer, shardID)
	pid := cm.LocalHost.Host.ID()
	Logger.Infof("Registering: message: %v", messagesWanted)
	Logger.Infof("Registering: nodeMode: %v", cm.nodeMode)
	Logger.Infof("Registering: layer: %v", layer)
	Logger.Infof("Registering: role: %v", role)
	Logger.Infof("Registering: shardID: %v", shardID)
	Logger.Infof("Registering: peerID: %v", pid.String())
	Logger.Infof("Registering: pubkey: %v", pubkey)
	pairs, topicRole, err := cm.Requester.Register(
		context.Background(),
		pubkey,
		messagesWanted,
		shardID,
		pid,
		role,
	)
	if err != nil {
		return nil, userRole{}, err
	}

	// Mapping from message to list of topics
	topics := m2t{}
	for _, p := range pairs {
		for i, t := range p.Topic {
			topics[p.Message] = append(topics[p.Message], Topic{
				Name: t,
				Act:  p.Act[i],
			})
		}
	}
	r := userRole{
		layer:   topicRole.Layer,
		role:    topicRole.Role,
		shardID: int(topicRole.Shard),
	}
	return topics, r, nil
}

func getMessagesForLayer(mode, layer string, shardID []byte) []string {
	switch mode {
	case common.NodeModeAuto:
		if layer == common.ShardRole {
			return []string{
				wire.CmdBlockShard,
				wire.CmdBlockBeacon,
				wire.CmdBFT,
				wire.CmdPeerState,
				wire.CmdCrossShard,
				wire.CmdBlkShardToBeacon,
				wire.CmdTx,
				wire.CmdPrivacyCustomToken,
				wire.CmdCustomToken,
			}
		} else if layer == common.BeaconRole {
			return []string{
				wire.CmdBlockBeacon,
				wire.CmdBFT,
				wire.CmdPeerState,
				wire.CmdBlkShardToBeacon,
			}
		} else {
			return []string{
				wire.CmdBlockBeacon,
				wire.CmdPeerState,
				wire.CmdTx,
				wire.CmdPrivacyCustomToken,
				wire.CmdCustomToken,
			}
		}
	case common.NodeModeRelay:
		return []string{
			wire.CmdTx,
			wire.CmdBlockShard,
			wire.CmdBlockBeacon,
			wire.CmdPeerState,
			wire.CmdPrivacyCustomToken,
			wire.CmdCustomToken,
		}
	}
	return []string{}
}

//go run *.go --listen "127.0.0.1:9433" --externaladdress "127.0.0.1:9433" --datadir "/data/fullnode" --discoverpeersaddress "127.0.0.1:9330" --loglevel debug
