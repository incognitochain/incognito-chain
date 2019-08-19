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
	topicList      []string                         // only allow registered Topic
	subscriberList map[string]map[uint]EventChannel // List of Subscriber
	messageBroker  map[string][]*Message            // Message pool
	idGenerator    uint                             // id generator for event
	cond           *sync.Cond
}

func NewPubSubManager() *PubSubManager {
	pubSubManager := &PubSubManager{
		topicList:      Topics,
		subscriberList: make(map[string]map[uint]EventChannel),
		messageBroker:  make(map[string][]*Message),
		idGenerator:    0,
		cond:           sync.NewCond(&sync.Mutex{}),
	}
	for _, topic := range pubSubManager.topicList {
		pubSubManager.subscriberList[topic] = make(map[uint]EventChannel)
	}
	return pubSubManager
}

// Forever Loop play as an Event Channel
func (pubSubManager *PubSubManager) Start() {
	for {
		pubSubManager.cond.L.Lock()
		for topic, messages := range pubSubManager.messageBroker {
			for _, message := range messages {
				if subMap, ok := pubSubManager.subscriberList[topic]; ok {
					for _, event := range subMap {
						go event.NotifyMessage(message)
					}
				}
			}
			// delete message (if no thing subscribe for it then delete msg too)
			pubSubManager.messageBroker[topic] = []*Message{}
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
	if _, ok := pubSubManager.subscriberList[topic]; !ok {
		pubSubManager.subscriberList[topic] = make(map[uint]EventChannel)
	}
	id := pubSubManager.idGenerator
	pubSubManager.subscriberList[topic][id] = cSubscribe
	pubSubManager.idGenerator = id + 1
	return id, cSubscribe, nil
}

// Publisher public message to EventChannel
func (pubSubManager *PubSubManager) PublishMessage(message *Message) {
	pubSubManager.cond.L.Lock()
	defer pubSubManager.cond.L.Unlock()
	pubSubManager.messageBroker[message.topic] = append(pubSubManager.messageBroker[message.topic], message)
	pubSubManager.cond.Signal()
}

func (pubSubManager *PubSubManager) Unsubscribe(topic string, subId uint) {
	pubSubManager.cond.L.Lock()
	defer pubSubManager.cond.L.Unlock()
	if subMap, ok := pubSubManager.subscriberList[topic]; ok {
		if _, ok := subMap[subId]; ok {
			delete(subMap, subId)
		}
	}
}

func (pubSubManager *PubSubManager) HasTopic(topic string) bool {
	if common.IndexOfStr(topic, pubSubManager.topicList) > -1 {
		return true
	}
	return false
}

func (pubSubManager *PubSubManager) AddTopic(topic string) {
	pubSubManager.cond.L.Lock()
	defer pubSubManager.cond.L.Unlock()
	if !pubSubManager.HasTopic(topic) {
		pubSubManager.topicList = append(pubSubManager.topicList, topic)
	}
}
