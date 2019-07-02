package rpcserver

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wire"
	"github.com/pkg/errors"
	"io/ioutil"
	"time"
)

type txs struct {
	Txs []string `json:"Txs"`
}

/*
For testing and benchmark only
*/
type CountResult struct {
	Success int
	Fail    int
}

func (httpServer *HttpServer) handleUnlockMempool(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	httpServer.config.TxMemPool.SendTransactionToBlockGen()
	return nil, nil
}
func (httpServer *HttpServer) handleGetAndSendTxsFromFile(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	Logger.log.Critical(arrayParams)
	shardIDParam := int(arrayParams[0].(float64))
	txType := arrayParams[1].(string)
	isSent := arrayParams[2].(bool)
	interval := int64(arrayParams[3].(float64))
	Logger.log.Criticalf("Interval between transactions %+v \n", interval)
	datadir := "./bin/"
	filename := ""
	success := 0
	fail := 0
	switch txType {
	case "privacy3000_1":
		filename = "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-privacy-3000.1.json"
	case "privacy3000_2":
		filename = "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-privacy-3000.2.json"
	case "privacy3000_3":
		filename = "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-privacy-3000.3.json"
	case "noprivacy9000":
		filename = "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-noprivacy-9000.json"
	case "noprivacy10000_2":
		filename = "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-noprivacy-10000.2.json"
	case "noprivacy10000_3":
		filename = "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-noprivacy-10000.3.json"
	case "noprivacy10000_4":
		filename = "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-noprivacy-10000.4.json"
	case "noprivacy":
		filename = "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-noprivacy-5000.json"
	case "privacy":
		filename = "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-privacy-5000.json"
	case "cstoken":
		filename = "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-cstoken-5000.json"
	case "cstokenprivacy":
		filename = "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-cstokenprivacy-5000.json"
	default:
		return CountResult{}, NewRPCError(ErrUnexpected, errors.New("Can't find file"))
	}

	Logger.log.Critical("Getting Transactions from file: ", datadir+filename)
	file, err := ioutil.ReadFile(datadir + filename)
	if err != nil {
		Logger.log.Error("Fail to get Transactions from file: ", err)
	}
	data := txs{}
	count := 0
	_ = json.Unmarshal([]byte(file), &data)
	Logger.log.Criticalf("Get %+v Transactions from file \n", len(data.Txs))
	intervalDuration := time.Duration(interval) * time.Millisecond
	for index, txBase58Data := range data.Txs {
		<-time.Tick(intervalDuration)
		Logger.log.Critical("Number of Transaction: ", index)
		//<-time.Tick(50*time.Millisecond)
		rawTxBytes, _, err := base58.Base58Check{}.Decode(txBase58Data)
		if err != nil {
			fail++
			continue
		}
		switch txType {
		case "cstoken":
			{
				var tx transaction.TxCustomToken
				err = json.Unmarshal(rawTxBytes, &tx)
				if err != nil {
					fail++
					continue
				}
				if !isSent {
					_, _, err = httpServer.config.TxMemPool.MaybeAcceptTransaction(&tx)
					if err != nil {
						fail++
						continue
					} else {
						success++
						continue
					}
				} else {
					_, _, err = httpServer.config.TxMemPool.MaybeAcceptTransaction(&tx)
					//httpServer.config.NetSync.HandleCacheTxHash(*tx.Hash())
					if err != nil {
						fail++
						continue
					}
					txMsg, err := wire.MakeEmptyMessage(wire.CmdCustomToken)
					if err != nil {
						fail++
						continue
					}
					txMsg.(*wire.MessageTxToken).Transaction = &tx
					err = httpServer.config.Server.PushMessageToAll(txMsg)
					if err != nil {
						fail++
						continue
					}
				}
				if err == nil {
					count++
					success++
				}
			}
		case "cstokenprivacy":
			{
				var tx transaction.TxCustomTokenPrivacy
				err = json.Unmarshal(rawTxBytes, &tx)
				if err != nil {
					fail++
					continue
				}
				if !isSent {
					_, _, err = httpServer.config.TxMemPool.MaybeAcceptTransaction(&tx)
					if err != nil {
						fail++
						continue
					} else {
						success++
						continue
					}
				} else {
					_, _, err = httpServer.config.TxMemPool.MaybeAcceptTransaction(&tx)
					//httpServer.config.NetSync.HandleCacheTxHash(*tx.Hash())
					if err != nil {
						fail++
						continue
					}
					txMsg, err := wire.MakeEmptyMessage(wire.CmdPrivacyCustomToken)
					if err != nil {
						fail++
						continue
					}
					txMsg.(*wire.MessageTxPrivacyToken).Transaction = &tx
					err = httpServer.config.Server.PushMessageToAll(txMsg)
					if err != nil {
						fail++
						continue
					}
				}
				if err == nil {
					count++
					success++
				}
			}
		default:
			var tx transaction.Tx
			err = json.Unmarshal(rawTxBytes, &tx)
			if err != nil {
				fail++
				continue
			}
			if !isSent {
				_, _, err = httpServer.config.TxMemPool.MaybeAcceptTransaction(&tx)
				if err != nil {
					fail++
					continue
				} else {
					success++
					continue
				}
			} else {
				_, _, err = httpServer.config.TxMemPool.MaybeAcceptTransaction(&tx)
				//httpServer.config.NetSync.HandleCacheTxHash(*tx.Hash())
				if err != nil {
					fail++
					continue
				}
				txMsg, err := wire.MakeEmptyMessage(wire.CmdTx)
				if err != nil {
					fail++
					continue
				}
				txMsg.(*wire.MessageTx).Transaction = &tx
				err = httpServer.config.Server.PushMessageToAll(txMsg)
				if err != nil {
					fail++
					continue
				}
			}
		}
		if err == nil {
			count++
			success++
		}
	}
	return CountResult{Success: success, Fail: fail}, nil
}

func (httpServer *HttpServer) handleGetAndSendTxsFromFileV2(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	Logger.log.Critical(arrayParams)
	shardIDParam := int(arrayParams[0].(float64))
	txType := arrayParams[1].(string)
	isSent := arrayParams[2].(bool)
	interval := int64(arrayParams[3].(float64))
	Logger.log.Criticalf("Interval between transactions %+v \n", interval)
	datadir := "./utility/"
	Txs := []string{}
	filename := ""
	filenames := []string{}
	success := 0
	fail := 0
	count := 0
	switch txType {
	case "noprivacy":
		filename = "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-noprivacy-9000.json"
		filenames = append(filenames, filename)
		for i := 2; i <= 3; i++ {
			filename := "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-noprivacy-10000." + fmt.Sprintf("%d", i) + ".json"
			filenames = append(filenames, filename)
		}
	default:
		return CountResult{}, NewRPCError(ErrUnexpected, errors.New("Can't find file"))
	}
	for _, filename := range filenames {
		Logger.log.Critical("Getting Transactions from file: ", datadir+filename)
		file, err := ioutil.ReadFile(datadir + filename)
		if err != nil {
			Logger.log.Error("Fail to get Transactions from file: ", err)
			continue
		}
		data := txs{}
		_ = json.Unmarshal([]byte(file), &data)
		Logger.log.Criticalf("Get %+v Transactions from file %+v \n", len(data.Txs), filename)
		Txs = append(Txs, data.Txs...)
	}
	intervalDuration := time.Duration(interval) * time.Millisecond
	for index, txBase58Data := range Txs {
		<-time.Tick(intervalDuration)
		Logger.log.Critical("Number of Transaction: ", index)
		//<-time.Tick(50*time.Millisecond)
		rawTxBytes, _, err := base58.Base58Check{}.Decode(txBase58Data)
		if err != nil {
			fail++
			continue
		}
		switch txType {
		case "cstoken":
			{
				var tx transaction.TxCustomToken
				err = json.Unmarshal(rawTxBytes, &tx)
				if err != nil {
					fail++
					continue
				}
				if !isSent {
					_, _, err = httpServer.config.TxMemPool.MaybeAcceptTransaction(&tx)
					if err != nil {
						fail++
						continue
					} else {
						success++
						continue
					}
				} else {
					_, _, err = httpServer.config.TxMemPool.MaybeAcceptTransaction(&tx)
					//httpServer.config.NetSync.HandleCacheTxHash(*tx.Hash())
					if err != nil {
						fail++
						continue
					}
					txMsg, err := wire.MakeEmptyMessage(wire.CmdCustomToken)
					if err != nil {
						fail++
						continue
					}
					txMsg.(*wire.MessageTxToken).Transaction = &tx
					err = httpServer.config.Server.PushMessageToAll(txMsg)
					if err != nil {
						fail++
						continue
					}
				}
				if err == nil {
					count++
					success++
				}
			}
		case "cstokenprivacy":
			{
				var tx transaction.TxCustomTokenPrivacy
				err = json.Unmarshal(rawTxBytes, &tx)
				if err != nil {
					fail++
					continue
				}
				if !isSent {
					_, _, err = httpServer.config.TxMemPool.MaybeAcceptTransaction(&tx)
					if err != nil {
						fail++
						continue
					} else {
						success++
						continue
					}
				} else {
					_, _, err = httpServer.config.TxMemPool.MaybeAcceptTransaction(&tx)
					//httpServer.config.NetSync.HandleCacheTxHash(*tx.Hash())
					if err != nil {
						fail++
						continue
					}
					txMsg, err := wire.MakeEmptyMessage(wire.CmdPrivacyCustomToken)
					if err != nil {
						fail++
						continue
					}
					txMsg.(*wire.MessageTxPrivacyToken).Transaction = &tx
					err = httpServer.config.Server.PushMessageToAll(txMsg)
					if err != nil {
						fail++
						continue
					}
				}
				if err == nil {
					count++
					success++
				}
			}
		default:
			var tx transaction.Tx
			err = json.Unmarshal(rawTxBytes, &tx)
			if err != nil {
				fail++
				continue
			}
			if !isSent {
				_, _, err = httpServer.config.TxMemPool.MaybeAcceptTransaction(&tx)
				if err != nil {
					fail++
					continue
				} else {
					success++
					continue
				}
			} else {
				_, _, err = httpServer.config.TxMemPool.MaybeAcceptTransaction(&tx)
				//httpServer.config.NetSync.HandleCacheTxHash(*tx.Hash())
				if err != nil {
					fail++
					continue
				}
				txMsg, err := wire.MakeEmptyMessage(wire.CmdTx)
				if err != nil {
					fail++
					continue
				}
				txMsg.(*wire.MessageTx).Transaction = &tx
				err = httpServer.config.Server.PushMessageToAll(txMsg)
				if err != nil {
					fail++
					continue
				}
			}
		}
		if err == nil {
			count++
			success++
		}
	}
	return CountResult{Success: success, Fail: fail}, nil
}
