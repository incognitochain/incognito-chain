package rpcserver

import (
	"errors"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
)

func (wsServer *WsServer) handleSubcribePendingTransaction(params interface{}, subcription string, cResult chan RpcSubResult, closeChan <-chan struct{}) {
	Logger.log.Info("Handle Subcribe Pending Transaction", params, subcription)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		err := NewRPCError(ErrRPCInvalidParams, errors.New("Methods should only contain 1 params"))
		cResult <- RpcSubResult{Error: err}
		return
	}
	txHashTemp,ok := arrayParams[0].(string)
	if !ok {
		err := NewRPCError(ErrRPCInvalidParams, errors.New("Invalid Tx Hash"))
		cResult <- RpcSubResult{Error: err}
	}
	txHash, _ := common.Hash{}.NewHashFromStr(txHashTemp)
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
				for index, tx := range shardBlock.Body.Transactions {
					if tx.Hash().IsEqual(txHash) {
						res, err := (&HttpServer{}).revertTxToResponseObject(tx, shardBlock.Hash(), shardBlock.Header.Height, index, shardBlock.Header.ShardID)
						cResult <- RpcSubResult{Result:res, Error:err}
						return
					}
				}
			}
		case <-closeChan:
			{
				return
			}
		}
	}
}
