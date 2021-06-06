package jsonresult

import (
	"github.com/incognitochain/incognito-chain/peer"
	"github.com/incognitochain/incognito-chain/utils"
	"github.com/incognitochain/incognito-chain/wire"
)

type GetInOutMessageResult struct {
	InboundMessages  map[string]interface{} `json:"Inbounds"`
	OutboundMessages map[string]interface{} `json:"Outbounds"`
}

func NewGetInOutMessageResult(paramsArray []interface{}) (*GetInOutMessageResult, error) {
	inboundMessages := peer.GetInboundPeerMessages()
	outboundMessages := peer.GetOutboundPeerMessages()
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
		peerID = utils.EmptyString
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

type GetInOutMessageCountResult struct {
	InboundMessages  interface{} `json:"Inbounds"`
	OutboundMessages interface{} `json:"Outbounds"`
}

func NewGetInOutMessageCountResult(paramsArray []interface{}) (*GetInOutMessageCountResult, error) {
	result := &GetInOutMessageCountResult{}
	inboundMessageByPeers := peer.GetInboundMessagesByPeer()
	outboundMessageByPeers := peer.GetOutboundMessagesByPeer()

	if len(paramsArray) == 0 {
		result.InboundMessages = inboundMessageByPeers
		result.OutboundMessages = outboundMessageByPeers
		return result, nil
	}

	peerID, ok := paramsArray[0].(string)
	if !ok {
		peerID = ""
	}
	result.InboundMessages = inboundMessageByPeers[peerID]
	result.OutboundMessages = outboundMessageByPeers[peerID]
	return result, nil
}
