package peerv2

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wire"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
)

func NewConnManager(
	host *Host,
	dpa string,
	ikey *incognitokey.CommitteePublicKey,
	cd ConsensusData,
	dispatcher *Dispatcher,
) *ConnManager {
	return &ConnManager{
		LocalHost:            host,
		DiscoverPeersAddress: dpa,
		IdentityKey:          ikey,
		cd:                   cd,
		disp:                 dispatcher,
	}
}

func (cm *ConnManager) PublishMessage(msg wire.Message) error {
	publishable := []string{wire.CmdBlockShard, wire.CmdBFT, wire.CmdBlockBeacon, wire.CmdPeerState}
	msgType := msg.MessageType()
	for _, p := range publishable {
		if msgType == p {
			fmt.Println("[db] Publishing message", msgType)
			return cm.encodeAndPublish(msg)
		}
	}

	log.Println("Cannot publish message", msgType)
	return nil
}

func (cm *ConnManager) Start() {
	// connect to proxy node
	proxyIP, proxyPort := ParseListenner(cm.DiscoverPeersAddress, "127.0.0.1", 9330)
	ipfsaddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", proxyIP, proxyPort))
	if err != nil {
		panic(err)
	}
	peerid, err := peer.IDB58Decode("QmbV4AAHWFFEtE67qqmNeEYXs5Yw5xNMS75oEKtdBvfoKN")

	// Pubsub
	// TODO(@0xbunyip): handle error
	cm.ps, _ = pubsub.NewFloodSub(context.Background(), cm.LocalHost.Host)
	cm.subs = map[string]Topic{}
	cm.messages = make(chan *pubsub.Message, 1000)

	// Must Connect after creating FloodSub
	must(cm.LocalHost.Host.Connect(context.Background(), peer.AddrInfo{peerid, append([]multiaddr.Multiaddr{}, ipfsaddr)}))

	go cm.manageRoleSubscription()

	cm.process()
}

type ConsensusData interface {
	GetUserRole() (string, string, int)
}

type Topic struct {
	Name string
	Sub  *pubsub.Subscription
}

type ConnManager struct {
	LocalHost            *Host
	DiscoverPeersAddress string
	IdentityKey          *incognitokey.CommitteePublicKey

	ps       *pubsub.PubSub
	subs     map[string]Topic     // mapping from message to topic's subscription
	messages chan *pubsub.Message // queue messages from all topics

	cd   ConsensusData
	disp *Dispatcher
}

func (cm *ConnManager) process() {
	for {
		select {
		case msg := <-cm.messages:
			fmt.Println("[db] go cm.disp.processInMessageString(string(msg.Data))")
			// go cm.disp.processInMessageString(string(msg.Data))
			err := cm.disp.processInMessageString(string(msg.Data))
			fmt.Printf("err: %+v\n", err)
		}
	}
}

func (cm *ConnManager) encodeAndPublish(msg wire.Message) error {
	// NOTE: copy from peerConn.outMessageHandler
	// Create and send messageHex
	messageBytes, err := msg.JsonSerialize()
	if err != nil {
		fmt.Println("Can not serialize json format for messageHex:" + msg.MessageType())
		fmt.Println(err)
		return err
	}

	// Add 24 bytes headerBytes into messageHex
	headerBytes := make([]byte, wire.MessageHeaderSize)
	// add command type of message
	cmdType, messageErr := wire.GetCmdType(reflect.TypeOf(msg))
	if messageErr != nil {
		fmt.Println("Can not get cmd type for " + msg.MessageType())
		fmt.Println(messageErr)
		return err
	}
	copy(headerBytes[:], []byte(cmdType))
	// add forward type of message at 13st byte
	forwardType := byte('s')
	forwardValue := byte(0)
	copy(headerBytes[wire.MessageCmdTypeSize:], []byte{forwardType})
	copy(headerBytes[wire.MessageCmdTypeSize+1:], []byte{forwardValue})
	messageBytes = append(messageBytes, headerBytes...)
	fmt.Printf("[db] OutMessageHandler TYPE %s CONTENT %s\n", cmdType, string(messageBytes))

	// zip data before send
	messageBytes, err = common.GZipFromBytes(messageBytes)
	if err != nil {
		fmt.Println("Can not gzip for messageHex:" + msg.MessageType())
		fmt.Println(err)
		return err
	}
	messageHex := hex.EncodeToString(messageBytes)
	//log.Debugf("Content in hex encode: %s", string(messageHex))
	// add end character to messageHex (delim '\n')
	// messageHex += "\n"

	// Publish
	topic := cm.subs[msg.MessageType()].Name
	if isJustPubOrSub(msg.MessageType()) {
		topic = topic + "-nodepub"
	}
	fmt.Printf("[db] Publishing to topic %s\n", topic)
	return cm.ps.Publish(topic, []byte(messageHex))
}

// manageRoleSubscription: polling current role every minute and subscribe to relevant topics
func (cm *ConnManager) manageRoleSubscription() {
	peerid, _ := peer.IDB58Decode("QmbV4AAHWFFEtE67qqmNeEYXs5Yw5xNMS75oEKtdBvfoKN")
	pubkey, _ := cm.IdentityKey.ToBase58()

	lastRole := newUserRole("dummyLayer", "dummyRole", -1000)
	lastTopics := m2t{}
	for range time.Tick(5 * time.Second) {
		// Update when role changes
		newRole := newUserRole(cm.cd.GetUserRole())
		if newRole == lastRole {
			continue
		}
		log.Printf("Role changed: %v -> %v", lastRole, newRole)
		lastRole = newRole

		// TODO(@0xbunyip): Pending & Waiting roles?
		if newRole.role != common.CommitteeRole {
			continue
		}

		// Get new topics
		topics, err := cm.registerToProxy(peerid, pubkey, newRole.layer, newRole.shardID)
		if err != nil {
			log.Println(err)
			continue
		}
		if err := cm.subscribeNewTopics(topics, lastTopics); err != nil {
			log.Println(err)
			continue
		}
		lastTopics = topics
	}
}

type userRole struct {
	layer   string
	role    string
	shardID int
}

func newUserRole(layer, role string, shardID int) *userRole {
	return &userRole{
		layer:   layer,
		role:    role,
		shardID: shardID,
	}
}

// subscribeNewTopics subscribes to new topics and unsubcribes any topics that aren't needed anymore
func (cm *ConnManager) subscribeNewTopics(topics, subscribed m2t) error {
	found := func(s string, m m2t) bool {
		for _, v := range m {
			if s == v {
				return true
			}
		}
		return false
	}

	// Subscribe to new topics
	for m, t := range topics {
		topic4Subs := t
		if isJustPubOrSub(t) {
			topic4Subs = topic4Subs + "_nodesub"
		}
		if found(topic4Subs, subscribed) {
			continue
		}

		fmt.Println("[db] subscribing", m, topic4Subs)

		s, err := cm.ps.Subscribe(topic4Subs)
		if err != nil {
			return err
		}
		cm.subs[m] = Topic{Name: t, Sub: s}
		go processSubscriptionMessage(cm.messages, s)
	}

	// Unsubscribe to old ones
	for m, t := range subscribed {
		topic4Subs := t
		if isJustPubOrSub(t) {
			topic4Subs = topic4Subs + "_nodesub"
		}
		if found(topic4Subs, topics) {
			continue
		}

		fmt.Println("[db] unsubscribing", m, t)
		cm.subs[m].Sub.Cancel() // TODO(@0xbunyip): lock
		delete(cm.subs, m)
	}
	return nil
}

// processSubscriptionMessage listens to a topic and pushes all messages to a queue to be processed later
func processSubscriptionMessage(inbox chan *pubsub.Message, sub *pubsub.Subscription) {
	ctx := context.Background()
	for {
		msg, err := sub.Next(ctx)
		fmt.Println("[db] Found new msg")
		_ = err
		// if err != nil {
		// 	log.Println(err)
		// 	return
		// 	// TODO(@0xbunyip): check if topic is unsubbed then return, otherwise just continue
		// }

		inbox <- msg
	}
}

type m2t map[string]string // Message to topic name

func (cm *ConnManager) registerToProxy(
	peerID peer.ID,
	pubkey string,
	layer string,
	shardID int,
) (m2t, error) {
	// Client on this node
	client := GRPCService_Client{cm.LocalHost.GRPC}
	messagesWanted := getMessagesForLayer(layer, shardID)
	pairs, err := client.ProxyRegister(
		context.Background(),
		peerID,
		pubkey,
		messagesWanted,
	)
	if err != nil {
		return nil, err
	}

	// Mapping from message to topic name
	topics := m2t{}
	for _, p := range pairs {
		topics[p.Message] = p.Topic
	}
	return topics, nil
}

func getMessagesForLayer(layer string, shardID int) []string {
	if layer == common.ShardRole {
		return []string{
			wire.CmdBlockShard,
			wire.CmdBlockBeacon,
			wire.CmdBFT,
			wire.CmdPeerState,
			wire.CmdCrossShard,
			wire.CmdBlkShardToBeacon,
		}
	} else if layer == common.BeaconRole {
		return []string{
			wire.CmdBlockBeacon,
			wire.CmdBFT,
			wire.CmdPeerState,
			wire.CmdBlkShardToBeacon,
		}
	}
	return []string{}
}

func isJustPubOrSub(message string) bool {
	if message == wire.CmdPeerState {
		return true
	}
	return false
}

//go run *.go --listen "127.0.0.1:9433" --externaladdress "127.0.0.1:9433" --datadir "/data/fullnode" --discoverpeersaddress "127.0.0.1:9330" --loglevel debug
