package pubsub

import (
	"github.com/incognitochain/incognito-chain/common"
	"sync"
)

type Event chan *Message
type PubsubManager struct {
	// only allow registered Topic
	TopicList []string
	//List of sender
	//PublisherList map[string]Publisher
	//List of Subcriber
	SubcriberList map[string]map[uint]Event
	// Message pool
	MessagePool map[string][]*Message
	IdGenerator uint
	cond *sync.Cond
}
type Message struct {
	Topic          string
	Value          interface{}
	unSendSubcribe []chan interface{}
}

func NewMessage(topic string, value interface{}) *Message {
	return &Message{
		Topic: topic,
		Value: value,
	}
}
func NewPubsubManager() *PubsubManager {
	return &PubsubManager{
		TopicList: Topics,
		//PublisherList: make(map[string]Publisher),
		SubcriberList: make(map[string]map[uint]Event),
		MessagePool:   make(map[string][]*Message),
		IdGenerator:   0,
		cond:          sync.NewCond(&sync.Mutex{}),
	}
}
func (pubsubManager *PubsubManager) Start() {
	for {
		pubsubManager.cond.L.Lock()
		for topic, messages := range pubsubManager.MessagePool {
			for _, message := range messages {
				if subMap, ok := pubsubManager.SubcriberList[topic]; ok {
					for _, event := range subMap{
						go event.NotifyMessage(message)
					}
				}
			}
			// delete message (if no thing subcribe for it then delete msg too)
			pubsubManager.MessagePool[topic] = []*Message{}
		}
		pubsubManager.cond.Wait()
		pubsubManager.cond.L.Unlock()
	}
}

//func (pubsubManager *PubsubManager) RegisterNewPublisher(Topic string) chan interface{} {
//	pubsubManager.cond.L.Lock()
//	defer pubsubManager.cond.L.Unlock()
//	cPublish := make(chan interface{}, ChanWorkLoad)
//	pubsubManager.PublisherList[Topic] = cPublish
//	return cPublish
//}
func (pubsubManager *PubsubManager) RegisterNewSubcriber(topic string) (uint, Event) {
	pubsubManager.cond.L.Lock()
	defer pubsubManager.cond.L.Unlock()
	cSubcribe := make(chan *Message, ChanWorkLoad)
	if _, ok := pubsubManager.SubcriberList[topic]; !ok {
		pubsubManager.SubcriberList[topic] = make(map[uint]Event)
	}
	id := pubsubManager.IdGenerator
	pubsubManager.SubcriberList[topic][id] = cSubcribe
	pubsubManager.IdGenerator = id + 1
	return id, cSubcribe
}
func (pubsubManager *PubsubManager) PublishMessage(message *Message) {
	pubsubManager.cond.L.Lock()
	defer pubsubManager.cond.L.Unlock()
	pubsubManager.MessagePool[message.Topic] = append(pubsubManager.MessagePool[message.Topic], message)
	pubsubManager.cond.Signal()
}
func (event Event) IsClosed() bool {
	msg, ok := <-event
	if ok {
		go event.NotifyMessage(msg)
		return false
	} else {
		return true
	}
}
func (event Event) NotifyMessage(message *Message) {
	event <- message
}
func (pubsubManager *PubsubManager) Unsubcribe(topic string, subId uint) {
	pubsubManager.cond.L.Lock()
	defer pubsubManager.cond.L.Unlock()
	if subMap, ok := pubsubManager.SubcriberList[topic]; ok {
		if _, ok := subMap[subId]; ok {
			delete(subMap, subId)
		}
	}
}
func (pubsubManager *PubsubManager) HasTopic(topic string) bool {
	if common.IndexOfStr(topic, pubsubManager.TopicList) > -1 {
		return true
	}
	return false
}
func (pubsubManager *PubsubManager) AddTopic(topic string) {
	pubsubManager.cond.L.Lock()
	defer pubsubManager.cond.L.Unlock()
	if !pubsubManager.HasTopic(topic) {
		pubsubManager.TopicList = append(pubsubManager.TopicList, topic)
	}
}

//TODO: remove message, delete closed chan
//TODO: unscibe, delete subcribeList if no one list
