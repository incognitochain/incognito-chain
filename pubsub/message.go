package pubsub

type Message struct {
	Value interface{}

	topic           string
	unSendSubscribe []chan interface{}
}

func NewMessage(topic string, value interface{}) *Message {
	return &Message{
		topic: topic,
		Value: value,
	}
}

type EventChannel chan *Message

func (event EventChannel) NotifyMessage(message *Message) {
	event <- message
}
