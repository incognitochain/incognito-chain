package peer

import (
	"sync"

	"github.com/incognitochain/incognito-chain/wire"
	peer "github.com/libp2p/go-libp2p-peer"
)

type PeerMessageInOut struct {
	PeerID  peer.ID      `json:"PeerID"`
	Message wire.Message `json:"Message"`
	Time    int64        `json:"Time"`
}

var inboundPeerMessage = map[string][]PeerMessageInOut{}
var outboundPeerMessage = map[string][]PeerMessageInOut{}
var inMutex = &sync.Mutex{}
var outMutex = &sync.Mutex{}

func storeInboundPeerMessage(msg wire.Message, time int64, peerID peer.ID) {
	messageType := msg.MessageType()
	inMutex.Lock()
	defer inMutex.Unlock()
	existingMessages := inboundPeerMessage[messageType]
	if len(existingMessages) == 0 {
		inboundPeerMessage[messageType] = []PeerMessageInOut{
			{Message: msg, Time: time, PeerID: peerID},
		}
		return
	}

	messages := []PeerMessageInOut{
		{peerID, msg, time},
	}
	for _, message := range existingMessages {
		if message.Time < time-10 {
			continue
		}
		messages = append(messages, message)
	}
	inboundPeerMessage[messageType] = messages
}

func GetInboundPeerMessages() map[string][]PeerMessageInOut {
	return inboundPeerMessage
}

func GetInboundMessagesByPeer() map[string]int {
	result := map[string]int{}
	for _, inboundMessages := range inboundPeerMessage {
		for _, message := range inboundMessages {
			result[message.PeerID.Pretty()]++
		}
	}
	return result
}

func storeOutboundPeerMessage(msg wire.Message, time int64, peerID peer.ID) {
	messageType := msg.MessageType()
	outMutex.Lock()
	defer outMutex.Unlock()
	existingMessages := outboundPeerMessage[messageType]
	if len(existingMessages) == 0 {
		outboundPeerMessage[messageType] = []PeerMessageInOut{
			{Message: msg, Time: time, PeerID: peerID},
		}
		return
	}
	messages := []PeerMessageInOut{
		{peerID, msg, time},
	}
	for _, message := range existingMessages {
		if message.Time < time-10 {
			continue
		}
		messages = append(messages, message)
	}
	outboundPeerMessage[messageType] = messages
}

func GetOutboundPeerMessages() map[string][]PeerMessageInOut {
	return outboundPeerMessage
}

func GetOutboundMessagesByPeer() map[string]int {
	result := map[string]int{}
	for _, outboundMessages := range outboundPeerMessage {
		for _, message := range outboundMessages {
			result[message.PeerID.Pretty()]++
		}
	}
	return result
}
