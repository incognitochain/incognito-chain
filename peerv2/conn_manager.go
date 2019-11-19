package peerv2

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wire"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
)

// TODO REMOVE HARDCODE
var HighwayPeerID = "QmSPa4gxx6PRmoNRu6P2iFwEwmayaoLdR5By3i3MgM9gMv"
var MasterNodeID = "QmVsCnV9kRZ182MX11CpcHMyFAReyXV49a599AbqmwtNrV"

func NewConnManager(
	host *Host,
	dpa string,
	ikey *incognitokey.CommitteePublicKey,
	cd ConsensusData,
	dispatcher *Dispatcher,
	nodeMode *string,
	relayShard *[]byte,
) *ConnManager {
	master := peer.IDB58Encode(host.Host.ID()) == MasterNodeID
	log.Println("IsMasterNode:", master)
	return &ConnManager{
		LocalHost:            host,
		DiscoverPeersAddress: dpa,
		IdentityKey:          ikey,
		cd:                   cd,
		disp:                 dispatcher,
		IsMasterNode:         master,
		registerRequests:     make(chan int, 100),
		relayShard:           relayShard,
		nodeMode:             nodeMode,
	}
}

func (cm *ConnManager) PublishMessage(msg wire.Message) error {
	var topic string
	publishable := []string{wire.CmdBlockShard, wire.CmdBFT, wire.CmdBlockBeacon, wire.CmdTx, wire.CmdCustomToken, wire.CmdPeerState, wire.CmdBlkShardToBeacon}
	// msgCrossShard := msg.(wire.MessageCrossShard)
	msgType := msg.MessageType()
	for _, p := range publishable {
		topic = ""
		if msgType == p {
			for _, availableTopic := range cm.subs[msgType] {
				fmt.Println(availableTopic)
				if (availableTopic.Act == MessageTopicPair_PUB) || (availableTopic.Act == MessageTopicPair_PUBSUB) {
					topic = availableTopic.Name
				}

			}
			if topic == "" {
				return errors.New("Can not find topic of this message type " + msgType + "for publish")
			}
			return broadcastMessage(msg, topic, cm.ps)
		}
	}

	log.Println("Cannot publish message", msgType)
	return nil
}

func (cm *ConnManager) PublishMessageToShard(msg wire.Message, shardID byte) error {
	publishable := []string{wire.CmdCrossShard, wire.CmdBFT}
	msgType := msg.MessageType()
	for _, p := range publishable {
		if msgType == p {
			// Get topic for mess
			//TODO hy add more logic
			if msgType == wire.CmdCrossShard {
				// TODO(@0xakk0r0kamui): implicit order of subscriptions?
				return broadcastMessage(msg, cm.subs[msgType][shardID].Name, cm.ps)
			} else {
				for _, availableTopic := range cm.subs[msgType] {
					fmt.Println(availableTopic)
					if (availableTopic.Act == MessageTopicPair_PUB) || (availableTopic.Act == MessageTopicPair_PUBSUB) {
						return broadcastMessage(msg, availableTopic.Name, cm.ps)
					}
				}
			}
		}
	}

	log.Println("Cannot publish message", msgType)
	return nil
}

func (cm *ConnManager) Start(ns NetSync) {
	// connect to proxy node
	peerid, err := peer.IDB58Decode(HighwayPeerID)

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

	req, err := NewRequester(cm.LocalHost.GRPC, peerid)
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
	if !cm.IsMasterNode {
		return
	}

	log.Println("Broadcasting committee to highways!!!")
	cc := &incognitokey.ChainCommittee{
		Epoch:             epoch,
		BeaconCommittee:   newBeaconCommittee,
		AllShardCommittee: newAllShardCommittee,
		AllShardPending:   newAllShardPending,
	}
	data, err := cc.ToByte()
	if err != nil {
		log.Println(err)
		return
	}

	topic := "chain_committee"
	err = cm.ps.Publish(topic, data)
	if err != nil {
		log.Println(err)
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
	IdentityKey          *incognitokey.CommitteePublicKey
	IsMasterNode         bool

	ps               *pubsub.PubSub
	subs             m2t                  // mapping from message to topic's subscription
	messages         chan *pubsub.Message // queue messages from all topics
	registerRequests chan int

	nodeMode   *string
	relayShard *[]byte

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
			// fmt.Println("[db] go cm.disp.processInMessageString(string(msg.Data))")
			// go cm.disp.processInMessageString(string(msg.Data))
			err := cm.disp.processInMessageString(string(msg.Data))
			if err != nil {
				log.Println(err)
			}
		}
	}
}

// keepHighwayConnection periodically checks liveliness of connection to highway
// and try to connect if it's not available.
// The method push data to the given channel to signal that the first attempt had finished.
// Constructor can use this info to initialize other objects.
func (cm *ConnManager) keepHighwayConnection(connectedOnce chan error) {
	pid, _ := peer.IDB58Decode(HighwayPeerID)
	ip, port := ParseListenner(cm.DiscoverPeersAddress, "127.0.0.1", 9330)
	ipfsaddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", ip, port))
	if err != nil {
		panic(fmt.Sprintf("invalid highway config:", err, pid, ip, port))
	}
	peerInfo := peer.AddrInfo{
		ID:    pid,
		Addrs: append([]multiaddr.Multiaddr{}, ipfsaddr),
	}

	first := true
	net := cm.LocalHost.Host.Network()
	disconnected := true
	for ; true; <-time.Tick(10 * time.Second) {
		// Reconnect if not connected
		var err error
		if net.Connectedness(pid) != network.Connected {
			disconnected = true
			log.Println("Not connected to highway, connecting")
			ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
			if err = cm.LocalHost.Host.Connect(ctx, peerInfo); err != nil {
				log.Println("Could not connect to highway:", err, peerInfo)
			}
		}

		if disconnected && net.Connectedness(pid) == network.Connected {
			// Register again since this might be a new highway
			log.Println("Connected to highway, sending register request")
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
		fmt.Println("Can not serialize json format for messageHex:" + msg.MessageType())
		fmt.Println(err)
		return "", err
	}

	// Add 24 bytes headerBytes into messageHex
	headerBytes := make([]byte, wire.MessageHeaderSize)
	// add command type of message
	cmdType, messageErr := wire.GetCmdType(reflect.TypeOf(msg))
	if messageErr != nil {
		fmt.Println("Can not get cmd type for " + msg.MessageType())
		fmt.Println(messageErr)
		return "", err
	}
	copy(headerBytes[:], []byte(cmdType))
	// add forward type of message at 13st byte
	forwardType := byte('s')
	forwardValue := byte(0)
	copy(headerBytes[wire.MessageCmdTypeSize:], []byte{forwardType})
	copy(headerBytes[wire.MessageCmdTypeSize+1:], []byte{forwardValue})
	messageBytes = append(messageBytes, headerBytes...)
	log.Printf("Encoded message TYPE %s CONTENT %s", cmdType, string(messageBytes))

	// zip data before send
	messageBytes, err = common.GZipFromBytes(messageBytes)
	if err != nil {
		fmt.Println("Can not gzip for messageHex:" + msg.MessageType())
		fmt.Println(err)
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
	fmt.Printf("[db] Publishing to topic %s\n", topic)
	return ps.Publish(topic, []byte(messageHex))
}

// manageRoleSubscription: polling current role every minute and subscribe to relevant topics
func (cm *ConnManager) manageRoleSubscription() {
	role := newUserRole("dummyLayer", "dummyRole", -1000)
	topics := m2t{}
	for {
		select {
		case <-time.Tick(10 * time.Second):
			forced := false // only subscribe when role changed
			role, topics = cm.subscribe(role, topics, forced)

		case <-cm.registerRequests:
			log.Println("Received request to register")
			forced := true // register no matter if role changed or not
			role, topics = cm.subscribe(role, topics, forced)
		}
	}
}

func (cm *ConnManager) subscribe(role userRole, topics m2t, forced bool) (userRole, m2t) {
	newRole := newUserRole(cm.cd.GetUserRole())
	if newRole == role && !forced { // Not forced => no need to subscribe when role stays the same
		return newRole, topics
	}
	log.Printf("Role changed: %v -> %v", role, newRole)

	if newRole.role == common.WaitingRole && !forced { // Not forced => no need to subscribe when role is Waiting
		return newRole, topics
	}

	// Registering
	pubkey, _ := cm.IdentityKey.ToBase58()
	shardIDs := []byte{byte(newRole.shardID)}
	if *cm.nodeMode == common.NodeModeRelay {
		shardIDs = *cm.relayShard
	}
	newTopics, roleOfTopics, err := cm.registerToProxy(pubkey, newRole.layer, shardIDs)
	if err != nil {
		return role, topics
	}

	if newRole != roleOfTopics {
		log.Printf("Role not matching with highway, local = %+v, highway = %+v", newRole, roleOfTopics)
		return role, topics
	}

	// Subscribing
	if err := cm.subscribeNewTopics(newTopics, topics); err != nil {
		return role, topics
	}

	return newRole, newTopics
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
		fmt.Printf("Process message %v and topic %v\n", m, topicList)
		for _, t := range topicList {

			if found(t.Name, subscribed) {
				fmt.Printf("Countinue 1 %v %v\n", t.Name, subscribed)
				continue
			}

			// TODO(@0xakk0r0kamui): check here
			if t.Act == MessageTopicPair_PUB {
				cm.subs[m] = append(cm.subs[m], Topic{Name: t.Name, Sub: nil, Act: t.Act})
				fmt.Printf("Countinue 2 %v %v\n", t.Name, subscribed)
				continue
			}

			fmt.Println("[db] subscribing", m, t.Name)

			s, err := cm.ps.Subscribe(t.Name)
			if err != nil {
				return err
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

			fmt.Println("[db] unsubscribing", m, t.Name)
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
		if err != nil {
			log.Println(err)
			continue
		}

		inbox <- msg
	}
}

type m2t map[string][]Topic // Message to topics

func (cm *ConnManager) registerToProxy(
	pubkey string,
	layer string,
	shardID []byte,
) (m2t, userRole, error) {
	messagesWanted := getMessagesForLayer(*cm.nodeMode, layer, shardID)
	fmt.Printf("-%v-;;;-%v-;;;-%v-;;;\n", messagesWanted, *cm.nodeMode, shardID)
	// os.Exit(9)
	pairs, role, err := cm.Requester.Register(
		context.Background(),
		pubkey,
		messagesWanted,
		shardID,
		cm.LocalHost.Host.ID(),
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
		layer:   role.Layer,
		role:    role.Role,
		shardID: int(role.Shard),
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
