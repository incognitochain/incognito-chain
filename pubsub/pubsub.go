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
	SubcriberList map[string][]Event
	// Message pool
	MessagePool map[string][]*Message

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
		SubcriberList: make(map[string][]Event),
		MessagePool:   make(map[string][]*Message),
		cond:          sync.NewCond(&sync.Mutex{}),
	}
}
func (pubsubManager *PubsubManager) Start() {
	for {
		pubsubManager.cond.L.Lock()
		for topic, messages := range pubsubManager.MessagePool {
			for _, message := range messages {
				if subList, ok := pubsubManager.SubcriberList[topic]; ok {
					i := 0
					for i < len(subList) {
						event := subList[i]
						// detect closed event
						if !event.IsClosed() {
							// if not closed then notify message
							go event.NotifyMessage(message)
							i++
						} else {
							// else delete event out of subcribe list
							subList = append(subList[:i], subList[i+1:]...)
						}
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
func (pubsubManager *PubsubManager) RegisterNewSubcriber(topic string) Event {
	pubsubManager.cond.L.Lock()
	defer pubsubManager.cond.L.Unlock()
	cSubcribe := make(chan *Message, ChanWorkLoad)
	if _, ok := pubsubManager.SubcriberList[topic]; !ok {
		pubsubManager.SubcriberList[topic] = []Event{}
	}
	pubsubManager.SubcriberList[topic] = append(pubsubManager.SubcriberList[topic], cSubcribe)
	return cSubcribe
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
func (event Event) Unsubcribe() {
	close(event)
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
