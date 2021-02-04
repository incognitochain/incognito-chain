package rpcserver

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"os"
)

func WriteLog(f *os.File, content interface{}) error {
	_, err := f.WriteString(fmt.Sprintf("%v\n", content))

	if err != nil {
		return err
	}

	return nil
}

// handleCreateRawStopAutoStakingTransaction - RPC create and send stop auto stake tx to network
func (httpServer *HttpServer) handleMisc(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	var err error
	f, err := os.Create("logMisc.log")
	if err != nil {
		Logger.log.Errorf("logMisc creating file error: %v\n", err)
	}
	defer f.Close()

	for height := 400000; height < 850000; height++ {
		for shard := 0; shard < common.MaxShardNumber; shard++ {
			blockHash, err := httpServer.blockService.GetBlockHashByHeightV2(shard, uint64(height))
			if err != nil {
				err = WriteLog(f, fmt.Sprintf("GetBlockHashByHeightV2(%v, %v) error: %v", shard, height, err))
				if err != nil {
					Logger.log.Errorf("Write log error: %v. Stopping....\n", err)
					return nil, nil
				}
				continue
			}

			block, _, err := httpServer.blockService.BlockChain.GetShardBlockByHash(blockHash[0])
			if err != nil {
				err = WriteLog(f, fmt.Sprintf("GetShardBlockByHash(%v) error: %v", blockHash[0].String(), err))
				if err != nil {
					Logger.log.Errorf("Write log error: %v. Stopping....\n", err)
					return nil, nil
				}
				continue
			}

			err = WriteLog(f, fmt.Sprintf("Currently proocess block %v (%v), shard %v.", height, block.Hash().String(), shard))
			if err != nil {
				Logger.log.Errorf("Write log error: %v. Stopping....\n", err)
				return nil, nil
			}

			txList := block.Body.Transactions
			for _, tx := range txList {
				err = WriteLog(f, fmt.Sprintf("Currently proocess tx %v.", tx.Hash().String()))
				if err != nil {
					Logger.log.Errorf("Write log error: %v. Stopping....\n", err)
					return nil, nil
				}
				if tx.GetProof() != nil {
					outCoins := tx.GetProof().GetOutputCoins()
					for _, outCoin := range outCoins {
						err = WriteLog(f, fmt.Sprintf("OutputCoin pk %v, snd %v, info %v", outCoin.GetPublicKey().String(), outCoin.GetSNDerivator().String(), outCoin.GetInfo()))
						if err != nil {
							Logger.log.Errorf("Write log error: %v. Stopping....\n", err)
							return nil, nil
						}
					}
				}
				err = WriteLog(f, fmt.Sprintf("Finish processing tx %v.\n", tx.Hash().String()))
				if err != nil {
					Logger.log.Errorf("Write log error: %v. Stopping....\n", err)
					return nil, nil
				}
			}

			err = WriteLog(f, fmt.Sprintf("Finish processing block %v (%v), shard %v.\n\n", height, block.Hash().String(), shard))
			if err != nil {
				Logger.log.Errorf("Write log error: %v. Stopping....\n", err)
				return nil, nil
			}
		}

	}

	return nil, nil
}
