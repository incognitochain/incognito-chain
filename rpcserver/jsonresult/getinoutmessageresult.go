package jsonresult

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/peer"
	"github.com/incognitochain/incognito-chain/wire"
)

type GetInOutMessageResult struct {
	InboundMessages  map[string]interface{} `json:"Inbounds"`
	OutboundMessages map[string]interface{} `json:"Outbounds"`
}

func NewGetInOutMessageResult(paramsArray []interface{},
	inboundMessages map[string][]peer.PeerMessageInOut,
	outboundMessages map[string][]peer.PeerMessageInOut) (*GetInOutMessageResult, error) {
	result := &GetInOutMessageResult{
		InboundMessages:  map[string]interface{}{},
		OutboundMessages: map[string]interface{}{},
	}
	if len(paramsArray) == 0 {
		for messageType, messagePeers := range inboundMessages {
			result.InboundMessages[messageType] = len(messagePeers)
		}
		for messageType, messagePeers := range outboundMessages {
			result.OutboundMessages[messageType] = len(messagePeers)
		}
		return result, nil
	}
	peerID, ok := paramsArray[0].(string)
	if !ok {
		peerID = common.EmptyString
	}

	for messageType, messagePeers := range inboundMessages {
		messages := []wire.Message{}
		for _, m := range messagePeers {
			if m.PeerID.Pretty() != peerID {
				continue
			}
			messages = append(messages, m.Message)
		}
		result.InboundMessages[messageType] = messages
	}
	for messageType, messagePeers := range outboundMessages {
		messages := []wire.Message{}
		for _, m := range messagePeers {
			if m.PeerID.Pretty() != peerID {
				continue
			}
			messages = append(messages, m.Message)
		}
		result.OutboundMessages[messageType] = messages
	}
	return result, nil
}
