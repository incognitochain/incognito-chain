package pubsub

type Message struct {
	Topic           string
	Value           interface{}
	unSendSubscribe []chan interface{}
}

func NewMessage(topic string, value interface{}) *Message {
	return &Message{
		Topic: topic,
		Value: value,
	}
}

type EventChannel chan *Message

func (event EventChannel) NotifyMessage(message *Message) {
	event <- message
}
