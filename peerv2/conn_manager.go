package peerv2

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wire"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
)

type ConsensusData interface {
	GetUserRole() (string, int)
}

type ConnManager struct {
	LocalHost            *Host
	DiscoverPeersAddress string
	IdentityKey          *incognitokey.CommitteePublicKey

	ps       *pubsub.PubSub
	subs     map[string]*pubsub.Subscription
	messages chan *pubsub.Message

	cd ConsensusData
}

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

func (cm *ConnManager) Start() {
	//connect to proxy node
	proxyIP, proxyPort := ParseListenner(cm.DiscoverPeersAddress, "127.0.0.1", 9300)
	ipfsaddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", proxyIP, proxyPort))
	if err != nil {
		panic(err)
	}
	peerid, err := peer.IDB58Decode("QmbV4AAHWFFEtE67qqmNeEYXs5Yw5xNMS75oEKtdBvfoKN")
	must(cm.LocalHost.Host.Connect(context.Background(), peer.AddrInfo{peerid, append([]multiaddr.Multiaddr{}, ipfsaddr)}))

	//server service on this node
	gRPCService := GRPCService_Server{}
	gRPCService.registerServices(cm.LocalHost.GRPC.GetGRPCServer())

	// Pubsub
	// TODO(@0xbunyip): handle error
	cm.ps, _ = pubsub.NewFloodSub(context.Background(), cm.LocalHost.Host)
	cm.subs = map[string]*pubsub.Subscription{}
	cm.messages = make(chan *pubsub.Message, 1000)

	go cm.manageRoleSubscription()
}

// manageRoleSubscription: polling current role every minute and subscribe to relevant topics
func (cm *ConnManager) manageRoleSubscription() {
	cd := cm.cd
	peerid, _ := peer.IDB58Decode("QmbV4AAHWFFEtE67qqmNeEYXs5Yw5xNMS75oEKtdBvfoKN")
	pubkey, _ := cm.IdentityKey.ToBase58()

	lastRole := "dummy"
	lastShardID := -1000 // dummy value
	lastTopics := []string{}
	for range time.Tick(10 * time.Second) {
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

		lastRole = role
		lastShardID = shardID
		lastTopics = topics
	}
}

// subscribeNewTopics subscribes to new topics and unsubcribes any topics that aren't needed anymore
func (cm *ConnManager) subscribeNewTopics(topics []string, subscribed []string) error {
	found := func(s string, l []string) bool {
		for _, m := range l {
			if s == m {
				return true
			}
		}
		return false
	}

	// Subscribe to new topics
	for _, t := range topics {
		if found(t, subscribed) {
			continue
		}

		fmt.Println("subscribing", t)
		s, err := cm.ps.Subscribe(t)
		if err != nil {
			return err
		}
		cm.subs[t] = s
		go processSubscriptionMessage(cm.messages, s)
	}

	// Unsubscribe to old ones
	for _, t := range subscribed {
		if found(t, topics) {
			continue
		}

		fmt.Println("unsubscribing", t)
		s := cm.subs[t]
		s.Cancel()
		delete(cm.subs, t)
	}
	return nil
}

// processSubscriptionMessage listens to a topic and pushes all messages to a queue to be processed later
func processSubscriptionMessage(inbox chan *pubsub.Message, sub *pubsub.Subscription) {
	ctx := context.Background()
	for {
		msg, err := sub.Next(ctx)
		if err != nil {
			log.Println(err)
			return
			// TODO(@0xbunyip): check if topic is unsubbed then return, otherwise just continue
		}

		inbox <- msg
	}
}

func (cm *ConnManager) registerToProxy(
	peerID peer.ID,
	pubkey string,
	role string,
	shardID int,
) ([]string, error) {
	// Client on this node
	client := GRPCService_Client{cm.LocalHost.GRPC}
	return client.ProxyRegister(
		context.Background(),
		peerID,
		pubkey,
		getMessagesForRole(role, shardID),
	)
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
