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
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
)

var HighwayPeerID = "12D3KooW9sbQK4J64Qat5D9vQEhBCnDYD3WPqWmgUZD4M7CJ2rXS"
var MasterNodeID = "QmYMpCasu9oJSTmca9fotDs21fTC5jHEWQ2oUaytNXiRnT"

func NewConnManager(
	host *Host,
	dpa string,
	ikey *incognitokey.CommitteePublicKey,
	cd ConsensusData,
	dispatcher *Dispatcher,
) *ConnManager {
	master := peer.IDB58Encode(host.Host.ID()) == MasterNodeID
	return &ConnManager{
		LocalHost:            host,
		DiscoverPeersAddress: dpa,
		IdentityKey:          ikey,
		cd:                   cd,
		disp:                 dispatcher,
		IsMasterNode:         master,
	}
}

func (cm *ConnManager) PublishMessage(msg wire.Message) error {
	var topic string
	publishable := []string{wire.CmdBlockShard, wire.CmdBFT, wire.CmdBlockBeacon, wire.CmdPeerState, wire.CmdBlkShardToBeacon}
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
			fmt.Println("[db] Publishing message", msgType)
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
			fmt.Println("[db] Publishing message", msgType)
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
	proxyIP, proxyPort := ParseListenner(cm.DiscoverPeersAddress, "127.0.0.1", 9330)
	ipfsaddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", proxyIP, proxyPort))
	if err != nil {
		panic(err)
	}
	peerid, err := peer.IDB58Decode(HighwayPeerID)

	// Pubsub
	// TODO(@0xbunyip): handle error
	cm.ps, _ = pubsub.NewFloodSub(context.Background(), cm.LocalHost.Host)
	cm.subs = m2t{}
	cm.messages = make(chan *pubsub.Message, 1000)

	// Must Connect after creating FloodSub
	must(cm.LocalHost.Host.Connect(context.Background(), peer.AddrInfo{peerid, append([]multiaddr.Multiaddr{}, ipfsaddr)}))
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

	ps       *pubsub.PubSub
	subs     m2t                  // mapping from message to topic's subscription
	messages chan *pubsub.Message // queue messages from all topics

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
			fmt.Println("[db] go cm.disp.processInMessageString(string(msg.Data))")
			// go cm.disp.processInMessageString(string(msg.Data))
			err := cm.disp.processInMessageString(string(msg.Data))
			fmt.Printf("err: %+v\n", err)
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
	fmt.Printf("[db] OutMessageHandler TYPE %s CONTENT %s\n", cmdType, string(messageBytes))

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
	peerid, _ := peer.IDB58Decode(HighwayPeerID)
	pubkey, _ := cm.IdentityKey.ToBase58()

	lastRole := newUserRole("dummyLayer", "dummyRole", -1000)
	lastTopics := m2t{}
	for range time.Tick(5 * time.Second) {
		// Update when role changes
		newRole := newUserRole(cm.cd.GetUserRole())
		if *newRole == *lastRole {
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

type m2t map[string][]Topic // Message to topics

func (cm *ConnManager) registerToProxy(
	peerID peer.ID,
	pubkey string,
	layer string,
	shardID int,
) (m2t, error) {
	messagesWanted := getMessagesForLayer(layer, shardID)
	pairs, err := cm.Requester.Register(
		context.Background(),
		pubkey,
		messagesWanted,
		cm.LocalHost.Host.ID(),
	)
	if err != nil {
		return nil, err
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

//go run *.go --listen "127.0.0.1:9433" --externaladdress "127.0.0.1:9433" --datadir "/data/fullnode" --discoverpeersaddress "127.0.0.1:9330" --loglevel debug
