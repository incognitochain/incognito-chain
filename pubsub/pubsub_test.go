package pubsub

import (
	"reflect"
	"sync"
	"testing"
)
func TestNewMessage(t *testing.T) {
	msg := NewMessage(TestTopic, 1)
	if msg.Topic != TestTopic {
		t.Error("Wrong Topic")
	}
	value, ok := msg.Value.(int)
	if !ok {
		t.Error("Wrong value type")
	}
	if value != 1 {
		t.Error("Wrong value")
	}
}

func TestRegisterNewSubcriber(t *testing.T) {
	var pubsubManager = NewPubsubManager()
	id, event, err := pubsubManager.RegisterNewSubcriber(TestTopic)
	if err != nil {
		t.Errorf("Counter error %+v \n", err)
	}
	subMap, ok := pubsubManager.SubcriberList[TestTopic]
	if !ok {
		t.Error("Can not get subcribe map by topic")
	}
	if subChan, ok := subMap[id]; !ok {
		t.Error("Can not get sub chan")
	} else {
		if !reflect.DeepEqual(event, subChan) {
			t.Error("Wrong Subchan")
		}
	}
}
func TestRegisterNewSubcribeWithUnregisteredTopic(t *testing.T) {
	var pubsubManager = NewPubsubManager()
	id, _, err := pubsubManager.RegisterNewSubcriber("ajsdkl;awjdkl")
	if id != 0 {
		t.Error("Wrong Event ID")
	}
	if pubsubErr, ok := err.(*PubsubError); !ok {
		t.Error("Wrong error type")
	} else {
		if pubsubErr.Code != -1002 {
			t.Error("Wrong Error code")
		}
	}
}
func TestUnsubcribe(t *testing.T) {
	var pubsubManager = NewPubsubManager()
	id, _, _ := pubsubManager.RegisterNewSubcriber(TestTopic)
	pubsubManager.Unsubcribe(TestTopic, id)
	subMap, ok := pubsubManager.SubcriberList[TestTopic]
	if !ok {
		t.Error("Can not get subcribe map by topic")
	}
	if _, ok := subMap[id]; ok {
		t.Error("Should have no sub chan")
	}
}
func TestPublishMessage(t *testing.T) {
	var pubsubManager = NewPubsubManager()
	pubsubManager.PublishMessage(NewMessage(TestTopic, "abc"))
	msgs, ok := pubsubManager.MessageBroker[TestTopic]
	if !ok {
		t.Error("No Message found with this topic")
	}
	if len(msgs) != 1 {
		t.Errorf("Should have only 1 message %+v \n", len(pubsubManager.MessageBroker[TestTopic]))
	}
	if msgs[0].Topic != TestTopic {
		t.Error("Wrong Topic")
	}
	valueInterface := msgs[0].Value
	if value, ok := valueInterface.(string); !ok {
		t.Error("Wrong msg type")
	} else {
		if value != "abc" {
			t.Error("Wrong msg value")
		}
	}
}
func TestMessageBroken(t *testing.T) {
	var pubsubManager = NewPubsubManager()
	var wg sync.WaitGroup
	go pubsubManager.Start()
	id, event, err := pubsubManager.RegisterNewSubcriber(TestTopic)
	if err != nil {
		t.Error("Error when subcription")
	}
	wg.Add(1)
	go func(event chan *Message) {
		defer wg.Done()
		for msg := range event {
			topic := msg.Topic
			if topic != TestTopic {
				t.Error("Wrong subcription topic")
			}
			if value, ok := msg.Value.(string); !ok {
				t.Error("Wrong value type")
			} else {
				if value != "abc" {
					t.Error("Unexpected value")
				}
			}
			close(event)
		}
	}(event)
	pubsubManager.PublishMessage(NewMessage(TestTopic, "abc"))
	wg.Wait()
	pubsubManager.Unsubcribe(TestTopic, id)
	return
}
func TestHasTopic(t *testing.T) {
	var pubsubManager = NewPubsubManager()
	if !pubsubManager.HasTopic(NewBeaconBlockTopic) {
		t.Error("Pubsub manager should have this topic")
	}
	if pubsubManager.HasTopic("lajsdlkjaskldj") {
		t.Error("Pubsub manager should not have this topic")
	}
}

func TestAddTopic(t *testing.T) {
	var pubsubManager = NewPubsubManager()
	if pubsubManager.HasTopic("lajsdlkjaskldj") {
		t.Error("Pubsub manager should not have this topic")
	}
	pubsubManager.AddTopic("lajsdlkjaskldj")
	if !pubsubManager.HasTopic("lajsdlkjaskldj") {
		t.Error("Pubsub manager should have this topic")
	}
}
