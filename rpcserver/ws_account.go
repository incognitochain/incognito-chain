package rpcserver

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/pubsub"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wallet"
	"reflect"
)

var (
	ErrParseTransaction = errors.New("Parse transaction failed")
)

func (wsServer *WsServer) handleSubcribeCrossOutputCoinByPrivateKey(params interface{}, subcription string, cResult chan RpcSubResult, closeChan <-chan struct{}) {
	Logger.log.Info("Handle Subscribe New Block", params, subcription)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		err := rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Methods should only contain ONE params"))
		cResult <- RpcSubResult{Error: err}
		return
	}
	privateKey := arrayParams[0].(string)
	keyWallet, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		err := rpcservice.NewRPCError(rpcservice.SubcribeError, err)
		cResult <- RpcSubResult{Error: err}
		return
	}
	err = keyWallet.KeySet.InitFromPrivateKey(&keyWallet.KeySet.PrivateKey)
	if err != nil {
		err := rpcservice.NewRPCError(rpcservice.SubcribeError, err)
		cResult <- RpcSubResult{Error: err}
		return
	}
	subId, subChan, err := wsServer.config.PubSubManager.RegisterNewSubscriber(pubsub.NewShardblockTopic)
	if err != nil {
		err := rpcservice.NewRPCError(rpcservice.SubcribeError, err)
		cResult <- RpcSubResult{Error: err}
		return
	}
	defer func() {
		Logger.log.Info("Finish Subscribe New Shard Block")
		wsServer.config.PubSubManager.Unsubscribe(pubsub.NewShardblockTopic, subId)
		close(cResult)
	}()
	for {
		select {
		case msg := <-subChan:
			{
				shardBlock, ok := msg.Value.(*blockchain.ShardBlock)
				if !ok {
					Logger.log.Errorf("Wrong Message Type from Pubsub Manager, wanted *blockchain.BeaconBlock, have %+v", reflect.TypeOf(msg.Value))
					continue
				}
				m := make(map[byte]uint64)
				for senderShardID, crossTransactions := range shardBlock.Body.CrossTransactions {
					for _, crossTransaction := range crossTransactions {
						for _, crossOutputCoin := range crossTransaction.OutputCoin {
							processedOutputCoin := wsServer.config.BlockChain.DecryptOutputCoinByKey(&crossOutputCoin, &keyWallet.KeySet, senderShardID, &common.PRVCoinID)
							if value, ok := m[senderShardID]; ok {
								value += processedOutputCoin.CoinDetails.GetValue()
								m[senderShardID] = value
							} else {
								if processedOutputCoin.CoinDetails != nil {
									m[senderShardID] = processedOutputCoin.CoinDetails.GetValue()
								}
							}
						}
					}
				}
				if len(m) != 0 {
					for senderShardID, value := range m {
						cResult <- RpcSubResult{Result: jsonresult.CrossOutputCoinResult{
							SenderShardID:   senderShardID,
							ReceiverShardID: shardBlock.Header.ShardID,
							BlockHeight:     shardBlock.Header.Height,
							BlockHash:       shardBlock.Header.Hash().String(),
							PaymentAddress:  keyWallet.Base58CheckSerialize(wallet.PaymentAddressType),
							Value:           value,
						}, Error: nil}
					}
				}
			}
		case <-closeChan:
			{
				cResult <- RpcSubResult{Result: jsonresult.UnsubcribeResult{Message: "Unsubscribe New Beacon Block"}}
				return
			}
		}
	}
}
func (wsServer *WsServer) handleSubcribeCrossCustomTokenByPrivateKey(params interface{}, subcription string, cResult chan RpcSubResult, closeChan <-chan struct{}) {
	Logger.log.Info("Handle Subscribe New Block", params, subcription)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		err := rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Methods should only contain ONE params"))
		cResult <- RpcSubResult{Error: err}
		return
	}
	privateKey := arrayParams[0].(string)
	keyWallet, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		err := rpcservice.NewRPCError(rpcservice.SubcribeError, err)
		cResult <- RpcSubResult{Error: err}
		return
	}
	err = keyWallet.KeySet.InitFromPrivateKey(&keyWallet.KeySet.PrivateKey)
	if err != nil {
		err := rpcservice.NewRPCError(rpcservice.SubcribeError, err)
		cResult <- RpcSubResult{Error: err}
		return
	}
	subId, subChan, err := wsServer.config.PubSubManager.RegisterNewSubscriber(pubsub.NewShardblockTopic)
	if err != nil {
		err := rpcservice.NewRPCError(rpcservice.SubcribeError, err)
		cResult <- RpcSubResult{Error: err}
		return
	}
	defer func() {
		Logger.log.Info("Finish Subscribe New Shard Block")
		wsServer.config.PubSubManager.Unsubscribe(pubsub.NewShardblockTopic, subId)
		close(cResult)
	}()
	for {
		select {
		case msg := <-subChan:
			{
				shardBlock, ok := msg.Value.(*blockchain.ShardBlock)
				if !ok {
					Logger.log.Errorf("Wrong Message Type from Pubsub Manager, wanted *blockchain.BeaconBlock, have %+v", reflect.TypeOf(msg.Value))
					continue
				}
				m := make(map[common.Hash]uint64)
				for _, tx := range shardBlock.Body.Transactions {
					if tx.GetType() == common.TxCustomTokenType {
						txCustomToken, ok := tx.(*transaction.TxNormalToken)
						if !ok {
							err := rpcservice.NewRPCError(rpcservice.SubcribeError, fmt.Errorf("%+v, expect type %+v", ErrParseTransaction, common.TxCustomTokenType))
							cResult <- RpcSubResult{Error: err}
							return
						}
						if txCustomToken.TxTokenData.Type == transaction.TokenCrossShard {
							for _, vout := range txCustomToken.TxTokenData.Vouts {
								if bytes.Compare(keyWallet.KeySet.PaymentAddress.Bytes(), vout.PaymentAddress.Bytes()) == 0 {
									if value, ok := m[txCustomToken.TxTokenData.PropertyID]; ok {
										value += vout.Value
										m[txCustomToken.TxTokenData.PropertyID] = value
									} else {
										m[txCustomToken.TxTokenData.PropertyID] = vout.Value
									}
								}
							}
						}
					}
				}
				if len(m) != 0 {
					for tokenID, value := range m {
						cResult <- RpcSubResult{Result: jsonresult.CrossCustomTokenResult{
							ReceiverShardID: shardBlock.Header.ShardID,
							BlockHeight:     shardBlock.Header.Height,
							BlockHash:       shardBlock.Header.Hash().String(),
							PaymentAddress:  keyWallet.Base58CheckSerialize(wallet.PaymentAddressType),
							TokenID:         tokenID.String(),
							Value:           value,
						}, Error: nil}
					}
				}
			}
		case <-closeChan:
			{
				cResult <- RpcSubResult{Result: jsonresult.UnsubcribeResult{Message: "Unsubscribe New Beacon Block"}}
				return
			}
		}
	}
}

func (wsServer *WsServer) handleSubcribeCrossCustomTokenPrivacyByPrivateKey(params interface{}, subcription string, cResult chan RpcSubResult, closeChan <-chan struct{}) {
	Logger.log.Info("Handle Subscribe New Block", params, subcription)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		err := rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Methods should only contain ONE params"))
		cResult <- RpcSubResult{Error: err}
		return
	}
	privateKey := arrayParams[0].(string)
	keyWallet, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		err := rpcservice.NewRPCError(rpcservice.SubcribeError, err)
		cResult <- RpcSubResult{Error: err}
		return
	}
	err = keyWallet.KeySet.InitFromPrivateKey(&keyWallet.KeySet.PrivateKey)
	if err != nil {
		err := rpcservice.NewRPCError(rpcservice.SubcribeError, err)
		cResult <- RpcSubResult{Error: err}
		return
	}
	subId, subChan, err := wsServer.config.PubSubManager.RegisterNewSubscriber(pubsub.NewShardblockTopic)
	if err != nil {
		err := rpcservice.NewRPCError(rpcservice.SubcribeError, err)
		cResult <- RpcSubResult{Error: err}
		return
	}
	defer func() {
		Logger.log.Info("Finish Subscribe New Shard Block")
		wsServer.config.PubSubManager.Unsubscribe(pubsub.NewShardblockTopic, subId)
		close(cResult)
	}()
	for {
		select {
		case msg := <-subChan:
			{
				shardBlock, ok := msg.Value.(*blockchain.ShardBlock)
				if !ok {
					Logger.log.Errorf("Wrong Message Type from Pubsub Manager, wanted *blockchain.BeaconBlock, have %+v", reflect.TypeOf(msg.Value))
					continue
				}
				m := make(map[byte]map[common.Hash]uint64)
				for senderShardID, crossTransactions := range shardBlock.Body.CrossTransactions {
					for _, crossTransaction := range crossTransactions {
						for _, crossTokenPrivacyData := range crossTransaction.TokenPrivacyData {
							for _, crossOutputCoin := range crossTokenPrivacyData.OutputCoin {
								proccessedOutputCoin := wsServer.config.BlockChain.DecryptOutputCoinByKey(&crossOutputCoin, &keyWallet.KeySet, senderShardID, &crossTokenPrivacyData.PropertyID)
								if proccessedOutputCoin != nil {
									if m[senderShardID] == nil {
										m[senderShardID] = make(map[common.Hash]uint64)
									}
									if value, ok := m[senderShardID][crossTokenPrivacyData.PropertyID]; ok {
										value += proccessedOutputCoin.CoinDetails.GetValue()
										m[senderShardID][crossTokenPrivacyData.PropertyID] = value
									} else {
										m[senderShardID][crossTokenPrivacyData.PropertyID] = proccessedOutputCoin.CoinDetails.GetValue()
									}
								}
							}
						}
					}
				}
				if len(m) != 0 {
					for senderShardID, tokenIDValue := range m {
						for tokenID, value := range tokenIDValue {
							cResult <- RpcSubResult{Result: jsonresult.CrossCustomTokenPrivacyResult{
								SenderShardID:   senderShardID,
								ReceiverShardID: shardBlock.Header.ShardID,
								BlockHeight:     shardBlock.Header.Height,
								BlockHash:       shardBlock.Header.Hash().String(),
								PaymentAddress:  keyWallet.Base58CheckSerialize(wallet.PaymentAddressType),
								TokenID:         tokenID.String(),
								Value:           value,
							}, Error: nil}
						}
					}
				}
			}
		case <-closeChan:
			{
				cResult <- RpcSubResult{Result: jsonresult.UnsubcribeResult{Message: "Unsubscribe New Beacon Block"}}
				return
			}
		}
	}
}
