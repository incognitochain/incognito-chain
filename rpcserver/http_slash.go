package rpcserver

import (
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
)

func (httpServer *HttpServer) handleGetProducersBlackList(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	producersBlackList, err := httpServer.databaseService.GetProducersBlackList()
	if err != nil {
		return false, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	return producersBlackList, nil
}

func (httpServer *HttpServer) handleGetProducersBlackListDetail(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	producersBlackList, err := httpServer.databaseService.GetProducersBlackList()
	if err != nil {
		return false, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	type producerBlacklistString struct {
		IncPubKey    string
		MiningPubKey map[string]string
		Epochs       uint8
	}
	var result []producerBlacklistString
	for k, v := range producersBlackList {
		var keySet incognitokey.CommitteePublicKey
		keySet.FromString(k)

		var keyMap producerBlacklistString
		keyMap.IncPubKey = keySet.GetIncKeyBase58()
		keyMap.MiningPubKey = make(map[string]string)
		for keyType := range keySet.MiningPubKey {
			keyMap.MiningPubKey[keyType] = keySet.GetMiningKeyBase58(keyType)
		}
		keyMap.Epochs = v
		result = append(result, keyMap)
	}

	return result, nil
}
