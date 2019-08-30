package rpcservice

import (
	"encoding/hex"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/transaction"
)

type BlockService struct {
	BlockChain *blockchain.BlockChain
}

func (blockService BlockService) GetShardBestStates() map[byte]*blockchain.ShardBestState {
	shards := blockService.BlockChain.BestState.GetClonedAllShardBestState()
	return shards
}

func (blockService BlockService) GetShardBestStateByShardID(shardID byte) (*blockchain.ShardBestState, error) {
	shard, err := blockService.BlockChain.BestState.GetClonedAShardBestState(shardID)
	return shard, err
}

func (blockService BlockService) GetShardBestBlocks() map[byte]blockchain.ShardBlock {
	bestBlocks := make(map[byte]blockchain.ShardBlock)
	shards := blockService.BlockChain.BestState.GetClonedAllShardBestState()
	for shardID, best := range shards {
		bestBlocks[shardID] = *best.BestBlock
	}
	return bestBlocks
}

func (blockService BlockService) GetShardBestBlockByShardID(shardID byte) (blockchain.ShardBlock, common.Hash, error) {
	shard, err := blockService.BlockChain.BestState.GetClonedAShardBestState(shardID)
	return *shard.BestBlock, shard.BestBlockHash, err
}

func (blockService BlockService) GetShardBestBlockHashes() map[int]common.Hash {
	bestBlockHashes := make(map[int]common.Hash)
	shards := blockService.BlockChain.BestState.GetClonedAllShardBestState()
	for shardID, best := range shards {
		bestBlockHashes[int(shardID)] = best.BestBlockHash
	}
	return bestBlockHashes
}

func (blockService BlockService) GetShardBestBlockHashByShardID(shardID byte) common.Hash {
	shards := blockService.BlockChain.BestState.GetClonedAllShardBestState()
	return shards[shardID].BestBlockHash
}

func (blockService BlockService) GetBeaconBestStates() (*blockchain.BeaconBestState, error) {
	beacon, err := blockService.BlockChain.BestState.GetClonedBeaconBestState()
	return beacon, err
}

func (blockService BlockService) GetBeaconBestBlock() (*blockchain.BeaconBlock, error) {
	clonedBeaconBestState, err := blockService.BlockChain.BestState.GetClonedBeaconBestState()
	if err != nil {
		return nil, err
	}
	return &clonedBeaconBestState.BestBlock, nil
}

func (blockService BlockService) GetBeaconBestBlockHash() (*common.Hash, error) {
	clonedBeaconBestState, err := blockService.BlockChain.BestState.GetClonedBeaconBestState()
	if err != nil {
		return nil, err
	}
	return &clonedBeaconBestState.BestBlockHash, nil
}

func (blockService BlockService) RetrieveBlock(hashString string, verbosity string) (*jsonresult.GetBlockResult, *RPCError) {
	hash, errH := common.Hash{}.NewHashFromStr(hashString)
	if errH != nil {
		Logger.log.Debugf("handleRetrieveBlock result: %+v, err: %+v", nil, errH)
		return nil, NewRPCError(RPCInvalidParamsError, errH)
	}
	block, _, errD := blockService.BlockChain.GetShardBlockByHash(*hash)
	if errD != nil {
		Logger.log.Debugf("handleRetrieveBlock result: %+v, err: %+v", nil, errD)
		return nil, NewRPCError(GetShardBlockByHashError, errD)
	}
	result := jsonresult.GetBlockResult{}

	shardID := block.Header.ShardID

	if verbosity == "0" {
		data, err := json.Marshal(block)
		if err != nil {
			Logger.log.Debugf("handleRetrieveBlock result: %+v, err: %+v", nil, err)
			return nil, NewRPCError(JsonError, err)
		}
		result.Data = hex.EncodeToString(data)
	} else if verbosity == "1" {
		best := blockService.BlockChain.BestState.Shard[shardID].BestBlock

		blockHeight := block.Header.Height
		// Get next block hash unless there are none.
		var nextHashString string
		// if blockHeight < best.Header.GetHeight() {
		if blockHeight < best.Header.Height {
			nextHash, err := blockService.BlockChain.GetShardBlockByHeight(blockHeight+1, shardID)
			if err != nil {
				return nil, NewRPCError(GetShardBlockByHeightError, err)
			}
			nextHashString = nextHash.Hash().String()
		}

		result.Hash = block.Hash().String()
		result.Confirmations = int64(1 + best.Header.Height - blockHeight)
		result.Height = block.Header.Height
		result.Version = block.Header.Version
		result.TxRoot = block.Header.TxRoot.String()
		result.Time = block.Header.Timestamp
		result.ShardID = block.Header.ShardID
		result.PreviousBlockHash = block.Header.PreviousBlockHash.String()
		result.NextBlockHash = nextHashString
		result.TxHashes = []string{}
		result.BlockProducerSign = block.ProducerSig
		result.BlockProducer = block.Header.ProducerAddress.String()
		result.AggregatedSig = block.AggregatedSig
		result.BeaconHeight = block.Header.BeaconHeight
		result.BeaconBlockHash = block.Header.BeaconHash.String()
		result.R = block.R
		result.Round = block.Header.Round
		result.CrossShardBitMap = []int{}
		result.Instruction = block.Body.Instructions
		if len(block.Header.CrossShardBitMap) > 0 {
			for _, shardID := range block.Header.CrossShardBitMap {
				result.CrossShardBitMap = append(result.CrossShardBitMap, int(shardID))
			}
		}
		result.Epoch = block.Header.Epoch

		for _, tx := range block.Body.Transactions {
			result.TxHashes = append(result.TxHashes, tx.Hash().String())
		}
	} else if verbosity == "2" {
		best := blockService.BlockChain.BestState.Shard[shardID].BestBlock

		blockHeight := block.Header.Height
		// Get next block hash unless there are none.
		var nextHashString string
		if blockHeight < best.Header.Height {
			nextHash, err := blockService.BlockChain.GetShardBlockByHeight(blockHeight+1, shardID)
			if err != nil {
				Logger.log.Debugf("handleRetrieveBlock result: %+v, err: %+v", nil, err)
				return nil, NewRPCError(GetShardBlockByHeightError, err)
			}
			nextHashString = nextHash.Hash().String()
		}

		result.Hash = block.Hash().String()
		result.Confirmations = int64(1 + best.Header.Height - blockHeight)
		result.Height = block.Header.Height
		result.Version = block.Header.Version
		result.TxRoot = block.Header.TxRoot.String()
		result.Time = block.Header.Timestamp
		result.ShardID = block.Header.ShardID
		result.PreviousBlockHash = block.Header.PreviousBlockHash.String()
		result.NextBlockHash = nextHashString
		result.BlockProducerSign = block.ProducerSig
		result.BlockProducer = block.Header.ProducerAddress.String()
		result.AggregatedSig = block.AggregatedSig
		result.BeaconHeight = block.Header.BeaconHeight
		result.BeaconBlockHash = block.Header.BeaconHash.String()
		result.R = block.R
		result.Round = block.Header.Round
		result.CrossShardBitMap = []int{}
		result.Instruction = block.Body.Instructions
		if len(block.Header.CrossShardBitMap) > 0 {
			for _, shardID := range block.Header.CrossShardBitMap {
				result.CrossShardBitMap = append(result.CrossShardBitMap, int(shardID))
			}
		}
		result.Epoch = block.Header.Epoch

		result.Txs = make([]jsonresult.GetBlockTxResult, 0)
		for _, tx := range block.Body.Transactions {
			transactionT := jsonresult.GetBlockTxResult{}

			transactionT.Hash = tx.Hash().String()

			switch tx.GetType() {
			case common.TxNormalType, common.TxRewardType, common.TxReturnStakingType:
				txN := tx.(*transaction.Tx)
				data, err := json.Marshal(txN)
				if err != nil {
					return nil, NewRPCError(JsonError, err)
				}
				transactionT.HexData = hex.EncodeToString(data)
				transactionT.Locktime = txN.LockTime
			}

			result.Txs = append(result.Txs, transactionT)
		}
	}
	return &result, nil
}
