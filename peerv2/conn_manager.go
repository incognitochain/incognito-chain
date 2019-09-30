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
) *ConnManager {
	return &ConnManager{
		LocalHost:            host,
		DiscoverPeersAddress: dpa,
		IdentityKey:          ikey,
		cd:                   cd,
	}
}

func (cm *ConnManager) PublishMessage(msg wire.Message) error {
	publishable := []string{wire.CmdBlockShard, wire.CmdBFT}
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
	messageHex += "\n"

	// Publish
	topic := cm.subs[msg.MessageType()].Name
	fmt.Printf("[db] Publishing to topic %s\n", topic)
	return cm.ps.Publish(topic, []byte(messageHex))
}

func (cm *ConnManager) Start() {
	////connect to proxy node
	//proxyIP, proxyPort := ParseListenner(cm.DiscoverPeersAddress, "127.0.0.1", 9300)
	//ipfsaddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", proxyIP, proxyPort))
	//if err != nil {
	//	panic(err)
	//}
	//peerid, err := peer.IDB58Decode("QmbV4AAHWFFEtE67qqmNeEYXs5Yw5xNMS75oEKtdBvfoKN")
	//must(cm.LocalHost.Host.Connect(context.Background(), peer.AddrInfo{peerid, append([]multiaddr.Multiaddr{}, ipfsaddr)}))

	// Pubsub
	// TODO(@0xbunyip): handle error
	cm.ps, _ = pubsub.NewFloodSub(context.Background(), cm.LocalHost.Host)
	cm.subs = map[string]Topic{}
	cm.messages = make(chan *pubsub.Message, 1000)

	go cm.manageRoleSubscription()
}

type ConsensusData interface {
	GetUserRole() (string, int)
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

	cd ConsensusData
}

// manageRoleSubscription: polling current role every minute and subscribe to relevant topics
func (cm *ConnManager) manageRoleSubscription() {
	cd := cm.cd
	peerid, _ := peer.IDB58Decode("QmbV4AAHWFFEtE67qqmNeEYXs5Yw5xNMS75oEKtdBvfoKN")
	pubkey, _ := cm.IdentityKey.ToBase58()

	lastRole := "dummy"
	lastShardID := -1000 // dummy value
	lastTopics := m2t{}
	for range time.Tick(5 * time.Second) {
		// Update when role changes
		role, shardID := cd.GetUserRole()
		if role == lastRole && shardID == lastShardID {
			continue
		}

		// Get new topics
		topics := lastTopics
		if role == common.ShardRole || role == common.BeaconRole {
			var err error
			topics, err = cm.registerToProxy(peerid, pubkey, role, shardID)
			if err != nil {
				log.Println(err)
				continue
			}
		}

		err := cm.subscribeNewTopics(topics, lastTopics)
		if err != nil {
			log.Println(err)
			continue
		}

		//connect to proxy node
		proxyIP, proxyPort := ParseListenner(cm.DiscoverPeersAddress, "127.0.0.1", 9300)
		ipfsaddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", proxyIP, proxyPort))
		if err != nil {
			panic(err)
		}
		peerid, err := peer.IDB58Decode("QmbV4AAHWFFEtE67qqmNeEYXs5Yw5xNMS75oEKtdBvfoKN")
		must(cm.LocalHost.Host.Connect(context.Background(), peer.AddrInfo{peerid, append([]multiaddr.Multiaddr{}, ipfsaddr)}))
		fmt.Println("[db] Connected\n\n\n")

		lastRole = role
		lastShardID = shardID
		lastTopics = topics
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
		if found(t, subscribed) {
			continue
		}

		fmt.Println("[db] subscribing", m, t)
		s, err := cm.ps.Subscribe(t)
		if err != nil {
			return err
		}
		cm.subs[m] = Topic{Name: t, Sub: s}
		go processSubscriptionMessage(cm.messages, s)
	}

	// Unsubscribe to old ones
	for m, t := range subscribed {
		if found(t, topics) {
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
		if err != nil {
			log.Println(err)
			return
			// TODO(@0xbunyip): check if topic is unsubbed then return, otherwise just continue
		}

		inbox <- msg
	}
}

type m2t map[string]string // Message to topic name

func (cm *ConnManager) registerToProxy(
	peerID peer.ID,
	pubkey string,
	role string,
	shardID int,
) (m2t, error) {
	fmt.Println("[db] Registering to Proxy\n\n\n")
	// Client on this node
	// client := GRPCService_Client{cm.LocalHost.GRPC}
	// messagesWanted := getMessagesForRole(role, shardID)
	// pairs, err := client.ProxyRegister(
	// 	context.Background(),
	// 	peerID,
	// 	pubkey,
	// 	messagesWanted,
	// )
	// if err != nil {
	// 	fmt.Println("[db] err:", err, pairs)
	// 	// return nil, err
	// }

	// // Mapping from message to topic name
	// topics := m2t{}
	// for _, p := range pairs {
	// 	topics[p.Message] = p.Topic
	// }
	// fmt.Println("[db]", topics)
	tmp := m2t{
		"blockshard":  "PROXYblockshard",
		"blockbeacon": "PROXYblockbeacon",
		"bft":         "PROXYbft",
		"peerstate":   "PROXYpeerstate",
		"crossshard":  "PROXYcrossshard",
		"blkshdtobcn": "PROXYblkshdtobcn",
	}
	return tmp, nil
}

func getMessagesForRole(role string, shardID int) []string {
	if role == common.ShardRole {
		return []string{
			wire.CmdBlockShard,
			wire.CmdBlockBeacon,
			wire.CmdBFT,
			wire.CmdPeerState,
			wire.CmdCrossShard,
			wire.CmdBlkShardToBeacon,
		}
	} else if role == common.BeaconRole {
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
