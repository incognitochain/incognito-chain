package pubusb

import (
	"github.com/incognitochain/incognito-chain/common"
	"sync"
)
type Event chan interface{}
type PubsubManager struct {
	// only allow registered topic
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
	topic   string
	value   interface{}
	unSendSubcribe []chan interface{}
}
func NewMessage(topic string, value interface{}) *Message {
	return &Message{
		topic: topic,
		value: value,
	}
}
func NewPubsubManager() *PubsubManager{
	return &PubsubManager{
		TopicList: Topics,
		//PublisherList: make(map[string]Publisher),
		SubcriberList: make(map[string][]Event),
		MessagePool: make(map[string][]*Message),
		cond: sync.NewCond(&sync.Mutex{}),
	}
}
func (pubsubManager *PubsubManager) Start() {
	for {
		pubsubManager.cond.L.Lock()
		for topic, messages := range pubsubManager.MessagePool {
			for _, message := range messages {
				if subList, ok := pubsubManager.SubcriberList[topic]; ok {
					for _, event := range subList {
						event.NotifyMessage(message)
					}
				}
			}
		}
		pubsubManager.cond.L.Unlock()
		pubsubManager.cond.Wait()
	}
}
//func (pubsubManager *PubsubManager) RegisterNewPublisher(topic string) chan interface{} {
//	pubsubManager.cond.L.Lock()
//	defer pubsubManager.cond.L.Unlock()
//	cPublish := make(chan interface{}, ChanWorkLoad)
//	pubsubManager.PublisherList[topic] = cPublish
//	return cPublish
//}
func (pubsubManager *PubsubManager) RegisterNewSubcriber(topic string) chan interface{} {
	pubsubManager.cond.L.Lock()
	defer pubsubManager.cond.L.Unlock()
	cSubcribe := make(chan interface{}, ChanWorkLoad)
	pubsubManager.SubcriberList[topic] = append(pubsubManager.SubcriberList[topic],cSubcribe)
	return cSubcribe
}
func (pubsubManager *PubsubManager) PublishMessage(message *Message) {
	pubsubManager.cond.L.Lock()
	defer pubsubManager.cond.L.Unlock()
	pubsubManager.MessagePool[message.topic] = append(pubsubManager.MessagePool[message.topic], message)
	pubsubManager.cond.Signal()
}
func (event Event) NotifyMessage(message *Message) {
	event <- message
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
