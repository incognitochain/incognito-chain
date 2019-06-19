package pubusb

import (
	"sync"
)

type PubsubManager struct {
	// only allow registered topic
	TopicList []string
	//List of sender
	SenderList map[string]chan interface{}
	//List of Subcriber
	SubcriberList map[string][]chan interface{}
	// Message pool
	MessagePool map[string][]Message
	
	cond *sync.Cond
}
type Message struct {
	topic   string
	value   interface{}
	unSendSubcribe []chan interface{}
}

func NewPubsubManager() *PubsubManager{
	return &PubsubManager{
		TopicList: Topics,
		SenderList: make(map[string]chan interface{}),
		SubcriberList: make(map[string][]chan interface{}),
		MessagePool: make(map[string][]Message),
		cond: sync.NewCond(&sync.Mutex{}),
	}
}
func (pubsubManager *PubsubManager) Start() {

}
func (pubsubManager *PubsubManager) RegisterNewPublisher(topic string, pub chan interface{}) {
	pubsubManager.cond.L.Lock()
	defer pubsubManager.cond.L.Unlock()
}
func (pubsubManager *PubsubManager) RegisterNewSubcriber(topic string, sub chan interface{}) {

}
func PublishMessage(message *Message, pub chan interface{}) {
	pub <- message
}
func NotifyMessage(message *Message, sub chan interface{}) {
	sub <- message
}
