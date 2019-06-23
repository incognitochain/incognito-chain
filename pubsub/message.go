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

type Event chan *Message

func (event Event) NotifyMessage(message *Message) {
	event <- message
}
