package peerv2

import (
	"context"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/wire"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"
)

type ConsensusData interface {
	GetUserRole() (string, string, int)
}

// SubManager manages pubsub subscription of highway's topics
type SubManager struct {
	consensusData ConsensusData

	pubkey     string
	nodeMode   string
	relayShard []byte

	role   userRole
	topics msgToTopics
	subs   msgToTopics // mapping from message to topic's subscription
}

func NewSubManager(
	consensusData ConsensusData,
	nodeMode string,
	pubkey string,
	relayShard []byte,
) *SubManager {
	return &SubManager{
		consensusData: consensusData,
		nodeMode:      nodeMode,
		pubkey:        pubkey,
		relayShard:    relayShard,
		role:          newUserRole("dummyLayer", "dummyRole", -1000),
		topics:        msgToTopics{},
	}
}

// Subscribe registers to proxy and save the list of new topics if needed
func (sub *SubManager) Subscribe(forced bool) error {
	newRole := newUserRole(sub.consensusData.GetUserRole())
	if newRole == sub.role && !forced { // Not forced => no need to subscribe when role stays the same
		sub.role = newRole // Save new role and return
		return nil
	}
	Logger.Infof("Role changed: %v -> %v", sub.role, newRole)

	// Registering
	roleSID := newRole.shardID
	if roleSID == -2 { // normal node
		roleSID = -1
	}
	shardIDs := []byte{}
	if sub.nodeMode == common.NodeModeRelay {
		shardIDs = sub.relayShard
		shardIDs = append(shardIDs, HighwayBeaconID)
	} else {
		shardIDs = append(shardIDs, byte(roleSID))
	}
	newTopics, roleOfTopics, err := sub.registerToProxy(sub.pubkey, newRole.layer, newRole.role, shardIDs)
	if err != nil {
		return err // Don't save new role and topics since we need to retry later
	}

	// NOTE: disabled, highway always return the same role
	_ = roleOfTopics
	// if newRole != roleOfTopics {
	// 	return role, topics, errors.Errorf("lole not matching with highway, local = %+v, highway = %+v", newRole, roleOfTopics)
	// }

	Logger.Infof("Received topics = %+v, oldTopics = %+v", newTopics, sub.topics)

	// Subscribing
	if err := sub.subscribeNewTopics(newTopics, sub.topics); err != nil {
		return err
	}

	sub.role = newRole
	sub.topics = newTopics
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
func (sub *SubManager) subscribeNewTopics(newTopics, subscribed msgToTopics) error {
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

	// Subscribe to new topics
	for m, topicList := range newTopics {
		Logger.Infof("Process message %v and topic %v", m, topicList)
		for _, t := range topicList {

			if found(t.Name, subscribed) {
				Logger.Infof("Countinue 1 %v %v", t.Name, subscribed)
				continue
			}

			if t.Act == MessageTopicPair_PUB {
				sub.subs[m] = append(sub.subs[m], Topic{Name: t.Name, Sub: nil, Act: t.Act})
				Logger.Infof("Countinue 2 %v %v", t.Name, subscribed)
				continue
			}

			Logger.Info("subscribing", m, t.Name)

			s, err := sub.ps.Subscribe(t.Name)
			if err != nil {
				return errors.WithStack(err)
			}
			sub.subs[m] = append(sub.subs[m], Topic{Name: t.Name, Sub: s, Act: t.Act})
			go processSubscriptionMessage(sub.messages, s)
		}
	}

	// Unsubscribe to old ones
	for m, topicList := range subscribed {
		for _, t := range topicList {
			if found(t.Name, newTopics) {
				continue
			}

			if t.Act == MessageTopicPair_PUB {
				continue
			}

			Logger.Info("unsubscribing", m, t.Name)
			for _, s := range sub.subs[m] {
				if s.Name == t.Name {
					s.Sub.Cancel() // TODO(@0xbunyip): lock
				}
			}
			delete(sub.subs, m)
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

type msgToTopics map[string][]Topic // Message to topics

func (sub *SubManager) registerToProxy(
	pubkey string,
	layer string,
	role string,
	shardID []byte,
) (msgToTopics, userRole, error) {
	messagesWanted := getMessagesForLayer(sub.nodeMode, layer, shardID)
	pid := sub.LocalHost.Host.ID()
	Logger.Infof("Registering: message: %v", messagesWanted)
	Logger.Infof("Registering: nodeMode: %v", sub.nodeMode)
	Logger.Infof("Registering: layer: %v", layer)
	Logger.Infof("Registering: role: %v", role)
	Logger.Infof("Registering: shardID: %v", shardID)
	Logger.Infof("Registering: peerID: %v", pid.String())
	Logger.Infof("Registering: pubkey: %v", pubkey)
	pairs, topicRole, err := sub.Requester.Register(
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
