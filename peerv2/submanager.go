package peerv2

import (
	"context"
	"fmt"
	"sort"

	"github.com/incognitochain/incognito-chain/common/consensus"
	"github.com/incognitochain/incognito-chain/incognitokey"

	pubsub "github.com/incognitochain/go-libp2p-pubsub"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/peerv2/proto"
	"github.com/incognitochain/incognito-chain/wire"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
)

type ConsensusData interface {
	GetValidators() []*consensus.Validator
	GetOneValidator() *consensus.Validator
	GetOneValidatorForEachConsensusProcess() map[int]*consensus.Validator
}

// SubManager manages pubsub subscription of highway's topics
type SubManager struct {
	info
	messages chan *pubsub.Message // channel to put subscribed messages to

	registerer Registerer
	subscriber Subscriber

	topics msgToTopics
	subs   msgToTopics // mapping from message to topic's subscription

	rolehash string

	disp *Dispatcher
}

type info struct {
	consensusData ConsensusData
	pubkey        string
	syncMode      string
	relayShard    []byte
	peerID        peer.ID
}

type Subscriber interface {
	Subscribe(topic string, opts ...pubsub.SubOpt) (*pubsub.Subscription, error)
}

type Registerer interface {
	Register(context.Context, string, []string, []byte, peer.ID, string) ([]*proto.MessageTopicPair, *proto.UserRole, error)
	Target() string
	UpdateTarget(peer.ID)
}

func NewSubManager(
	info info,
	subscriber Subscriber,
	registerer Registerer,
	messages chan *pubsub.Message,
	dispatcher *Dispatcher,
) *SubManager {
	return &SubManager{
		info:       info,
		subscriber: subscriber,
		registerer: registerer,
		messages:   messages,
		//role:       newUserRole("dummyLayer", "dummyRole", -1000),
		topics: msgToTopics{},
		subs:   msgToTopics{},
		disp:   dispatcher,
	}
}

func (sub *SubManager) GetMsgToTopics() msgToTopics {
	return sub.subs // no need to make a copy since topics rarely changed (when role changed)
}

func (sub *SubManager) SetSyncMode(s string) {
	sub.syncMode = s
}

// Subscribe registers to proxy and save the list of new topics if needed
func (sub *SubManager) Subscribe(forced bool) error {
	rolehash := ""
	relayShardIDs := sub.relayShard
	var newTopics = make(msgToTopics)
	var err error
	shardIDs := []int{}
	nodePK, _ := new(incognitokey.CommitteePublicKey).ToBase58()

	if sub.syncMode == "" {
		newRole := sub.consensusData.GetOneValidatorForEachConsensusProcess()

		Logger.Infof("[debugSubs] relayShardIDs %+v", sub.relayShard)
		for _, sid := range relayShardIDs {
			if newRole[int(sid)] == nil {
				newRole[int(sid)] = &consensus.Validator{}
			} else {
				Logger.Infof("[debugSubs] newRole sID %v role %v", sid, newRole[int(sid)].State.Role)
			}
		}

		//recalculate rolehash
		for k := range newRole {
			shardIDs = append(shardIDs, int(k))
		}
		sort.Ints(shardIDs)
		str := ""
		for _, chainID := range shardIDs {
			Logger.Infof("[debugSubs] newRole sID %v role %v", chainID, newRole[int(chainID)].State.Role)
			if newRole[chainID] != nil {
				str += fmt.Sprintf("%v-%v", chainID, newRole[chainID].State.Role)
			} else {
				str += fmt.Sprintf("%v-", chainID)
			}
		}
		rolehash = common.HashH([]byte(str)).String()

		//check if role hash is changed
		if rolehash == sub.rolehash && !forced { // Not forced => no need to subscribe when role stays the same
			return nil
		}

		Logger.Infof("Role changed %+v", rolehash)

		// Registering relay
		if len(relayShardIDs) == 0 {
			relayShardIDs = []byte{255}
		}

		validator := sub.consensusData.GetOneValidator()
		if validator != nil {
			nodePK = validator.MiningKey.GetPublicKeyBase58()
		}
		newTopics, _, err = sub.registerToProxy(
			nodePK,
			"",
			"",
			relayShardIDs,
		)
		if err != nil {
			return err
		}

		// Registering mining
		for chainID, validator := range newRole {
			if validator.PrivateSeed != "" {
				topics, _, err := sub.registerToProxy(
					validator.MiningKey.GetPublicKeyBase58(),
					validator.State.Layer,
					validator.State.Role,
					[]byte{byte(chainID)},
				)
				if err != nil {
					return err // Don't save new role and topics since we need to retry later
				}
				for msg, topic := range topics {
					newTopics[msg] = append(newTopics[msg], topic...)
				}
			}
		}
	} else if sub.syncMode == "netmonitor" {
		rolehash = common.HashH([]byte("netmonitor")).String()

		//check if role hash is changed
		if rolehash == sub.rolehash && !forced { // Not forced => no need to subscribe when role stays the same
			return nil
		}

		topics, _, err := sub.registerToProxy(
			nodePK,
			"",
			"netmonitor",
			[]byte{255},
		)
		if err != nil {
			return err // Don't save new role and topics since we need to retry later
		}
		for msg, topic := range topics {
			newTopics[msg] = append(newTopics[msg], topic...)
		}

	}

	Logger.Infof("newTopics %v", newTopics)
	Logger.Infof("sub topics %v", sub.topics)

	// Subscribing
	if err := sub.subscribeNewTopics(newTopics, sub.topics, forced); err != nil {
		Logger.Error(err)
		return err
	}

	sub.topics = newTopics
	sub.rolehash = rolehash
	return nil
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
func (sub *SubManager) subscribeNewTopics(newTopics, subscribed msgToTopics, forced bool) error {
	found := func(tName string, tmap msgToTopics) bool {
		for _, topicList := range tmap {
			for _, t := range topicList {
				if tName == t.Name {
					return true
				}
			}
		}
		return false
	}

	// Unsubscribe to old ones
	for m, topicList := range subscribed {
		for _, t := range topicList {
			if !forced && found(t.Name, newTopics) { // Unsub and sub again if forced subscribe
				continue
			}

			idx := -1
			for i, s := range sub.subs[m] {
				if s.Name == t.Name {
					if t.Act != proto.MessageTopicPair_PUB {
						Logger.Info("unsubscribing", m, t.Name)
						if s.Sub != nil {
							go s.Sub.Cancel()
						}
					}
					idx = i
					break
				}
			}

			if idx < 0 {
				continue
			}

			if len(sub.subs[m]) == 1 {
				delete(sub.subs, m)
			} else {
				sub.subs[m] = append(sub.subs[m][:idx], sub.subs[m][idx+1:]...)
			}
		}
	}

	// Subscribe to new topics
	for m, topicList := range newTopics {
		Logger.Infof("Process message %v and topic %v", m, topicList)
		for _, t := range topicList {
			if !forced && found(t.Name, subscribed) {
				Logger.Infof("Skip subscribed topic: %v %v", t.Name, subscribed)
				continue
			}

			if t.Act == proto.MessageTopicPair_PUB {
				sub.subs[m] = append(sub.subs[m], Topic{Name: t.Name, Sub: nil, Act: t.Act})
				Logger.Infof("Skip subscribing to PUB topic: %v", t.Name)
				continue
			}

			Logger.Infof("subscribing message: %v, topic: %v", m, t.Name)

			s, err := sub.subscriber.Subscribe(t.Name)
			if err != nil {
				return errors.WithStack(err)
			}
			sub.subs[m] = append(sub.subs[m], Topic{Name: t.Name, Sub: s, Act: t.Act})
			go sub.processSubscriptionMessage(m, sub.messages, s)
		}
	}
	return nil
}

func getMsgTxsFromSub(txSubs *pubsub.Subscription, msgCh chan *pubsub.Message) {
	defer close(msgCh)
	ctx := context.Background()
	for {
		msg, err := txSubs.Next(ctx)
		if err != nil { // Subscription might have been cancelled
			Logger.Warn(err)
			return
		}
		msgCh <- msg
	}
}

func (sub *SubManager) processIncomingTxs(txSubs *pubsub.Subscription) {
	msgCh := make(chan *pubsub.Message, 1000)
	go getMsgTxsFromSub(txSubs, msgCh)
	for msg := range msgCh {
		go func(msg *pubsub.Message) {
			Logger.Infof("Received msg tx, data hash %v", common.HashH(msg.Data).String())
			err := sub.disp.processInMessageString(string(msg.Data))
			if err != nil {
				Logger.Errorf("Process msg %v return error %v", common.HashH(msg.Data).String(), err)
			}
		}(msg)
	}
}

// processSubscriptionMessage listens to a topic and pushes all messages to a queue to be processed later
func (sub *SubManager) processSubscriptionMessage(msgName string, inbox chan *pubsub.Message, subs *pubsub.Subscription) {
	if (msgName == wire.CmdTx) || (msgName == wire.CmdPrivacyCustomToken) {
		sub.processIncomingTxs(subs)
		return
	}
	ctx := context.Background()
	for {
		// TODO(@0xbunyip): check if topic is unsubbed then return, otherwise just continue
		msg, err := subs.Next(ctx)
		if err != nil { // Subscription might have been cancelled
			Logger.Warn(err)
			return
		}
		//Logger.Info("[dcs] sub.Topic():", sub.Topic())
		inbox <- msg
	}
}

type Topic struct {
	Name string
	Sub  *pubsub.Subscription
	Act  proto.MessageTopicPair_Action
}

type msgToTopics map[string][]Topic // Message to topics

func (sub *SubManager) registerToProxy(
	pubkey string,
	layer string,
	role string,
	shardID []byte,
) (msgToTopics, userRole, error) {
	messagesWanted := getMessagesForLayer(layer, role, shardID)
	// Logger.Infof("Registering: nodeMode: %v", sub.nodeMode)
	Logger.Infof("Registering: layer: %v", layer)
	Logger.Infof("Registering: role: %v", role)
	Logger.Infof("Registering: shardID: %v", shardID)
	Logger.Infof("Registering: peerID: %v", sub.peerID.String())
	Logger.Infof("Registering: pubkey: %v", pubkey)
	Logger.Infof("Registering: wantedMessages: %v", messagesWanted)

	pairs, topicRole, err := sub.registerer.Register(
		context.Background(),
		pubkey,
		messagesWanted,
		shardID,
		sub.peerID,
		role,
	)
	if err != nil {
		return nil, userRole{}, err
	}

	// Mapping from message to list of topics
	topics := msgToTopics{}
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

func getMessagesForLayer(layer, role string, shardID []byte) []string {
	msgs := []string{}
	switch layer {
	case common.ShardRole:
		switch role {
		case common.CommitteeRole:
			msgs = []string{
				wire.CmdBlockShard,
				wire.CmdBlockBeacon,
				wire.CmdMsgFinishSync,
				wire.CmdBFT,
				wire.CmdPeerState,
				wire.CmdCrossShard,
				wire.CmdTx,
				wire.CmdPrivacyCustomToken,
				wire.CmdMsgFeatureStat,
			}

		case common.SyncingRole:
			msgs = []string{
				wire.CmdBlockBeacon,
				wire.CmdMsgFinishSync,
				wire.CmdMsgFeatureStat,
				wire.CmdPeerState,
				wire.CmdTx,
				wire.CmdPrivacyCustomToken,
			}

		case common.PendingRole:
			msgs = []string{
				wire.CmdBlockBeacon,
				wire.CmdMsgFinishSync,
				wire.CmdPeerState,
				wire.CmdTx,
				wire.CmdPrivacyCustomToken,
				wire.CmdMsgFeatureStat,
			}
		}

	case common.BeaconRole:
		msgs = []string{
			wire.CmdBlockBeacon,
			wire.CmdBFT,
			wire.CmdPeerState,
			wire.CmdBlockShard,
			wire.CmdMsgFinishSync,
			wire.CmdMsgFeatureStat,
		}
	default:
		containShard := false
		for _, s := range shardID {
			if s != HighwayBeaconID {
				containShard = true
			}
		}
		msgs = []string{
			wire.CmdBlockBeacon,
			wire.CmdPeerState,
			wire.CmdTx,
			wire.CmdPrivacyCustomToken,
		}
		if containShard {
			msgs = append(msgs, wire.CmdBlockShard)
		}
	}
	return msgs
}

func getMessagesForLayer2(layer string, pubKey string, shardID []byte) []string {
	if pubKey == "1Eh5UZjDSUsdofnToKSELNz24rUfKgdsf3JubsetDQYscHHWERnQcVUxee9C68UFK1C1iTaHE2HxtyrkEvzsjN68LfEnqtnV1TEy3MQpkF2hmGFiExwGw4MbgApx7Vg6aYehqys3Xsgt2nxD4sG9iiNAQgLPkiDwGTj39yDciHqsETHrqUGAU4bFAm4mSnwEeTXo2FvQq7WrKsfAcg8nBWrNMG7UhZbsZ9hYrnc1WGxXy3y3DQ43B2RXat3BHvsz8fU6U1aNsPF5NH9AMDwU3ZDQrwGA5tsrANLfs33Y72rX5z1LRtApczEdVa8Xxvsxt3HE8rMtL7gcRzmRwHCXtbi9dEKfoDdMEXpjubbjFevNRR9LVH1jg6" {
		return []string{
			wire.CmdBlockBeacon,
			wire.CmdBFT,
			wire.CmdPeerState,
			// wire.CmdBlockShard,
		}
	}
	if layer == common.ShardRole {
		return []string{
			wire.CmdBlockShard,
			wire.CmdBlockBeacon,
			wire.CmdBFT,
			wire.CmdPeerState,
			wire.CmdCrossShard,
			wire.CmdTx,
			wire.CmdPrivacyCustomToken,
		}
	} else if layer == common.BeaconRole {
		return []string{
			wire.CmdBlockBeacon,
			wire.CmdBFT,
			wire.CmdPeerState,
			wire.CmdBlockShard,
		}
	} else {
		containShard := false
		for _, s := range shardID {
			if s != HighwayBeaconID {
				containShard = true
			}
		}
		msgs := []string{
			wire.CmdBlockBeacon,
			wire.CmdPeerState,
			wire.CmdTx,
			wire.CmdPrivacyCustomToken,
		}
		if containShard {
			msgs = append(msgs, wire.CmdBlockShard)
		}
		return msgs
	}

	return []string{}
}
