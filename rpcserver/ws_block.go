package rpcserver

import (
	"encoding/json"
	"errors"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)

func (wsServer *WsServer) handleSubcribeNewBlock(params interface{}, subcription string, cResult chan RpcSubResult, closeChan <-chan struct{}) {
	Logger.log.Info("Handle Subcribe New Block", params, subcription)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		err := NewRPCError(ErrRPCInvalidParams, errors.New("Methods should only contain 1 params"))
		cResult <- RpcSubResult{Error: err}
		return
	}
	shardID := byte(arrayParams[0].(float64))
	var cShardBlock = make(chan *blockchain.ShardBlock, 10)
	id := wsServer.config.BlockChain.SubcribeNewShardBlock(cShardBlock)
	defer func() {
		wsServer.config.BlockChain.UnsubcribeNewShardBlock(id)
		close(cResult)
	}()
	for {
		select {
		case shardBlock := <-cShardBlock:
			{
				if shardBlock.Header.ShardID != shardID {
					continue
				}
				blockResult := jsonresult.GetBlockResult{}
				blockBytes, err := json.Marshal(shardBlock)
				if err != nil {
					cResult <- RpcSubResult{Error: NewRPCError(ErrUnexpected, err)}
					return
				}
				blockResult.Init(shardBlock, uint64(len(blockBytes)))
				cResult <- RpcSubResult{Result: blockResult, Error: nil}
			}
		case <-closeChan:
			{
				return
			}
		}
	}
}
