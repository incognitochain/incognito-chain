package pubsub

import (
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"sync"
	"time"
)

// This package provide an Event Channel for internal pub sub of this application
// It will issue a list of pre-defined topic
// Publisher will register to be publisher with a particular topic,
// then it will be able to publish message of this topic into Event Channel
// Subcriber will register to be Subcribe with a particular topic,
// when new message of this topic come to Event Channel,
// then Event Channel will fire this message to subcriber
type PubSubManager struct {
	TopicList      []string                         // only allow registered Topic
	SubscriberList map[string]map[uint]EventChannel // List of Subscriber
	MessageBroker  map[string][]*Message            // Message pool
	IdGenerator    uint                             // id generator for event
	cond           *sync.Cond
}

func NewPubSubManager() *PubSubManager {
	pubSubManager := &PubSubManager{
		TopicList:      Topics,
		SubscriberList: make(map[string]map[uint]EventChannel),
		MessageBroker:  make(map[string][]*Message),
		IdGenerator:    0,
		cond:           sync.NewCond(&sync.Mutex{}),
	}
	for _, topic := range pubSubManager.TopicList {
		pubSubManager.SubscriberList[topic] = make(map[uint]EventChannel)
	}
	return pubSubManager
}

// Forever Loop play as an Event Channel
func (pubSubManager *PubSubManager) Start() {
	for {
		pubSubManager.cond.L.Lock()
		for topic, messages := range pubSubManager.MessageBroker {
			for _, message := range messages {
				if subMap, ok := pubSubManager.SubscriberList[topic]; ok {
					for _, event := range subMap {
						go event.NotifyMessage(message)
					}
				}
			}
			// delete message (if no thing subscribe for it then delete msg too)
			pubSubManager.MessageBroker[topic] = []*Message{}
		}
		pubSubManager.cond.Wait()
		pubSubManager.cond.L.Unlock()
		time.Sleep(1 * time.Microsecond)
	}
}

// Subcriber register with wanted topic
// Return Event and Id of that Event
// Event Channel using event to signal subcriber new message
func (pubSubManager *PubSubManager) RegisterNewSubscriber(topic string) (uint, EventChannel, error) {
	pubSubManager.cond.L.Lock()
	defer pubSubManager.cond.L.Unlock()
	cSubscribe := make(chan *Message, ChanWorkLoad)
	if !pubSubManager.HasTopic(topic) {
		return 0, cSubscribe, NewPubSubError(UnregisteredTopicError, errors.New(topic))
	}
	if _, ok := pubSubManager.SubscriberList[topic]; !ok {
		pubSubManager.SubscriberList[topic] = make(map[uint]EventChannel)
	}
	id := pubSubManager.IdGenerator
	pubSubManager.SubscriberList[topic][id] = cSubscribe
	pubSubManager.IdGenerator = id + 1
	return id, cSubscribe, nil
}

// Publisher public message to EventChannel
func (pubSubManager *PubSubManager) PublishMessage(message *Message) {
	pubSubManager.cond.L.Lock()
	defer pubSubManager.cond.L.Unlock()
	pubSubManager.MessageBroker[message.Topic] = append(pubSubManager.MessageBroker[message.Topic], message)
	pubSubManager.cond.Signal()
}

func (pubSubManager *PubSubManager) Unsubscribe(topic string, subId uint) {
	pubSubManager.cond.L.Lock()
	defer pubSubManager.cond.L.Unlock()
	if subMap, ok := pubSubManager.SubscriberList[topic]; ok {
		if _, ok := subMap[subId]; ok {
			delete(subMap, subId)
		}
	}
}

func (pubSubManager *PubSubManager) HasTopic(topic string) bool {
	if common.IndexOfStr(topic, pubSubManager.TopicList) > -1 {
		return true
	}
	return false
}

func (pubSubManager *PubSubManager) AddTopic(topic string) {
	pubSubManager.cond.L.Lock()
	defer pubSubManager.cond.L.Unlock()
	if !pubSubManager.HasTopic(topic) {
		pubSubManager.TopicList = append(pubSubManager.TopicList, topic)
	}
}
