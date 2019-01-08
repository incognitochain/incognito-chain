package rpcserver

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/rpcserver/jsonresult"
	"github.com/ninjadotorg/constant/transaction"
)

// handleGetBestBlock implements the getbestblock command.
func (self RpcServer) handleGetBestBlock(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	result := jsonresult.GetBestBlockResult{
		BestBlocks: make(map[string]jsonresult.GetBestBlockItem),
	}
	for shardID, best := range self.config.BlockChain.BestState.Shard {
		result.BestBlocks[strconv.Itoa(int(shardID))] = jsonresult.GetBestBlockItem{
			// Height:   best.BestBlock.Header.GetHeight(),
			Height:   best.BestBlock.Header.Height,
			Hash:     best.BestBlockHash.String(),
			TotalTxs: best.TotalTxns,
		}
	}
	return result, nil
}

// handleGetBestBlock implements the getbestblock command.
func (self RpcServer) handleGetBestBlockHash(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	result := jsonresult.GetBestBlockHashResult{
		BestBlockHashes: make(map[byte]string),
	}
	for shardID, best := range self.config.BlockChain.BestState.Shard {
		result.BestBlockHashes[shardID] = best.BestBlockHash.String()
	}
	return result, nil
}

/*
getblockcount RPC return information fo blockchain node
*/
func (self RpcServer) handleRetrieveBlock(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	paramsT, ok := params.([]interface{})
	if ok && len(paramsT) >= 2 {
		hashString := paramsT[0].(string)
		hash, errH := common.Hash{}.NewHashFromStr(hashString)
		if errH != nil {
			return nil, NewRPCError(ErrUnexpected, errH)
		}
		// block, errD := self.config.BlockChain.GetBlockByHash(hash)
		block, errD := self.config.BlockChain.GetShardBlockByHash(hash)
		if errD != nil {
			return nil, NewRPCError(ErrUnexpected, errD)
		}
		result := jsonresult.GetBlockResult{}

		verbosity := paramsT[1].(string)

		// shardID := block.Header.(*blockchain.BlockHeaderShard).ShardID
		shardID := block.Header.ShardID

		if verbosity == "0" {
			data, err := json.Marshal(block)
			if err != nil {
				return nil, NewRPCError(ErrUnexpected, err)
			}
			result.Data = hex.EncodeToString(data)
		} else if verbosity == "1" {
			best := self.config.BlockChain.BestState.Shard[shardID].BestBlock

			// blockHeight := block.Header.GetHeight()
			blockHeight := block.Header.Height
			// Get next block hash unless there are none.
			var nextHashString string
			// if blockHeight < best.Header.GetHeight() {
			if blockHeight < best.Header.Height {
				nextHash, err := self.config.BlockChain.GetShardBlockByHeight(blockHeight+1, shardID)
				if err != nil {
					return nil, NewRPCError(ErrUnexpected, err)
				}
				nextHashString = nextHash.Hash().String()
			}

			result.Hash = block.Hash().String()
			// result.Confirmations = int64(1 + best.Header.GetHeight() - blockHeight)
			result.Confirmations = int64(1 + best.Header.Height - blockHeight)
			// result.Height = block.Header.GetHeight()
			result.Height = block.Header.Height
			// result.Version = block.Header.(*blockchain.BlockHeaderShard).Version
			result.Version = block.Header.Version
			// result.MerkleRoot = block.Header.(*blockchain.BlockHeaderShard).MerkleRoot.String()
			result.MerkleRoot = block.Header.MerkleRoot.String()
			// result.Time = block.Header.(*blockchain.BlockHeaderShard).Timestamp
			result.Time = block.Header.Timestamp
			// result.ShardID = block.Header.(*blockchain.BlockHeaderShard).ShardID
			result.ShardID = block.Header.ShardID
			// result.PreviousBlockHash = block.Header.(*blockchain.BlockHeaderShard).PrevBlockHash.String()
			result.PreviousBlockHash = block.Header.PrevBlockHash.String()
			result.NextBlockHash = nextHashString
			result.TxHashes = []string{}
			result.BlockProducerSign = block.ProducerSig
			// for _, tx := range block.Body.(*blockchain.BlockBodyShard).Transactions {
			for _, tx := range block.Body.Transactions {
				result.TxHashes = append(result.TxHashes, tx.Hash().String())
			}
		} else if verbosity == "2" {
			best := self.config.BlockChain.BestState.Shard[shardID].BestBlock

			// blockHeight := block.Header.GetHeight()
			blockHeight := block.Header.Height
			// Get next block hash unless there are none.
			var nextHashString string
			// if blockHeight < best.Header.GetHeight() {
			if blockHeight < best.Header.Height {
				nextHash, err := self.config.BlockChain.GetShardBlockByHeight(blockHeight+1, shardID)
				if err != nil {
					return nil, NewRPCError(ErrUnexpected, err)
				}
				nextHashString = nextHash.Hash().String()
			}

			result.Hash = block.Hash().String()
			// result.Confirmations = int64(1 + best.Header.GetHeight() - blockHeight)
			result.Confirmations = int64(1 + best.Header.Height - blockHeight)
			// result.Height = block.Header.GetHeight()
			result.Height = block.Header.Height
			// result.Version = block.Header.(*blockchain.BlockHeaderShard).Version
			// result.MerkleRoot = block.Header.(*blockchain.BlockHeaderShard).MerkleRoot.String()
			// result.Time = block.Header.(*blockchain.BlockHeaderShard).Timestamp
			// result.ShardID = block.Header.(*blockchain.BlockHeaderShard).ShardID
			// result.PreviousBlockHash = block.Header.(*blockchain.BlockHeaderShard).PrevBlockHash.String()
			result.Version = block.Header.Version
			result.MerkleRoot = block.Header.MerkleRoot.String()
			result.Time = block.Header.Timestamp
			result.ShardID = block.Header.ShardID
			result.PreviousBlockHash = block.Header.PrevBlockHash.String()
			result.NextBlockHash = nextHashString
			result.Txs = make([]jsonresult.GetBlockTxResult, 0)
			// for _, tx := range block.Body.(*blockchain.BlockBodyShard).Transactions {
			for _, tx := range block.Body.Transactions {
				transactionT := jsonresult.GetBlockTxResult{}

				transactionT.Hash = tx.Hash().String()
				if tx.GetType() == common.TxNormalType || tx.GetType() == common.TxSalaryType {
					txN := tx.(*transaction.Tx)
					data, err := json.Marshal(txN)
					if err != nil {
						return nil, NewRPCError(ErrUnexpected, err)
					}
					transactionT.HexData = hex.EncodeToString(data)
					transactionT.Locktime = txN.LockTime
				}
				result.Txs = append(result.Txs, transactionT)
			}
		}

		return result, nil
	}
	return nil, nil
}

// handleGetBlocks - get n top blocks from chain ID
func (self RpcServer) handleGetBlocks(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	result := make([]jsonresult.GetBlockResult, 0)
	arrayParams := common.InterfaceSlice(params)
	numBlock := int(arrayParams[0].(float64))
	shardID := byte(arrayParams[1].(float64))
	bestBlock := self.config.BlockChain.BestState.Shard[shardID].BestBlock
	previousHash := bestBlock.Hash()
	for numBlock > 0 {
		numBlock--
		// block, errD := self.config.BlockChain.GetBlockByHash(previousHash)
		block, errD := self.config.BlockChain.GetShardBlockByHash(previousHash)
		if errD != nil {
			return nil, NewRPCError(ErrUnexpected, errD)
		}
		blockResult := jsonresult.GetBlockResult{}
		blockResult.Init(block)
		result = append(result, blockResult)
		// previousHash = &block.Header.(*blockchain.BlockHeaderShard).PrevBlockHash
		previousHash = &block.Header.PrevBlockHash
		if previousHash.String() == (common.Hash{}).String() {
			break
		}
	}
	return result, nil
}

/*
getblockchaininfo RPC return information fo blockchain node
*/
func (self RpcServer) handleGetBlockChainInfo(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	result := jsonresult.GetBlockChainInfoResult{
		ChainName:  self.config.ChainParams.Name,
		BestBlocks: make(map[byte]jsonresult.GetBestBlockItem),
	}
	for shardID, bestState := range self.config.BlockChain.BestState.Shard {
		result.BestBlocks[shardID] = jsonresult.GetBestBlockItem{
			// Height:   bestState.BestBlock.Header.GetHeight(),
			Height:   bestState.BestBlock.Header.Height,
			Hash:     bestState.BestBlockHash.String(),
			TotalTxs: bestState.TotalTxns,
			// SalaryFund:       bestState.BestBlock.Header.SalaryFund,
			// BasicSalary:      bestState.BestBlock.Header.GOVConstitution.GOVParams.BasicSalary,
			// SalaryPerTx:      bestState.BestBlock.Header.GOVConstitution.GOVParams.SalaryPerTx,
			// BlockProducer:    bestState.BestBlock.Producer,
			BlockProducerSig: bestState.BestBlock.ProducerSig,
		}
	}
	return result, nil
}

/*
getblockcount RPC return information fo blockchain node
*/
func (self RpcServer) handleGetBlockCount(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	shardID := byte(int(params.(float64)))
	// if self.config.BlockChain.BestState != nil && self.config.BlockChain.BestState[shardID] != nil && self.config.BlockChain.BestState[shardID].BestBlock != nil {
	// 	return self.config.BlockChain.BestState[shardID].BestBlock.Header.Height + 1, nil
	// }
	if self.config.BlockChain.BestState != nil && self.config.BlockChain.BestState.Shard[shardID] != nil && self.config.BlockChain.BestState.Shard[shardID].BestBlock != nil {
		return self.config.BlockChain.BestState.Shard[shardID].BestBlock.Header.Height + 1, nil
	}
	return 0, nil
}

/*
getblockhash RPC return information fo blockchain node
*/
func (self RpcServer) handleGetBlockHash(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	shardID := byte(int(arrayParams[0].(float64)))
	height := uint64(arrayParams[1].(float64))
	hash, err := self.config.BlockChain.GetShardBlockByHeight(height, shardID)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	return hash.Hash().String(), nil
}

// handleGetBlockHeader - return block header data
func (self RpcServer) handleGetBlockHeader(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Info(params)
	result := jsonresult.GetHeaderResult{}

	arrayParams := common.InterfaceSlice(params)
	Logger.log.Info(arrayParams)
	getBy := arrayParams[0].(string)
	block := arrayParams[1].(string)
	shardID := arrayParams[2].(float64)
	switch getBy {
	case "blockhash":
		bhash := common.Hash{}
		err := bhash.Decode(&bhash, block)
		Logger.log.Info(bhash)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, errors.New("Invalid blockhash format"))
		}
		// block, err := self.config.BlockChain.GetBlockByHash(&bhash)
		block, err := self.config.BlockChain.GetShardBlockByHash(&bhash)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, errors.New("Block not exist"))
		}
		result.Header = block.Header
		// result.BlockNum = int(block.Header.GetHeight()) + 1
		result.BlockNum = int(block.Header.Height) + 1
		result.ShardID = uint8(shardID)
		result.BlockHash = bhash.String()
	case "blocknum":
		bnum, err := strconv.Atoi(block)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, errors.New("Invalid blocknum format"))
		}
		fmt.Println(shardID)
		// if uint64(bnum-1) > self.config.BlockChain.BestState.Shard[uint8(shardID)].BestBlock.Header.GetHeight() || bnum <= 0 {
		if uint64(bnum-1) > self.config.BlockChain.BestState.Shard[uint8(shardID)].BestBlock.Header.Height || bnum <= 0 {
			return nil, NewRPCError(ErrUnexpected, errors.New("Block not exist"))
		}
		block, _ := self.config.BlockChain.GetShardBlockByHeight(uint64(bnum-1), uint8(shardID))
		result.Header = block.Header
		result.BlockNum = bnum
		result.ShardID = uint8(shardID)
		result.BlockHash = block.Hash().String()
	default:
		return nil, NewRPCError(ErrUnexpected, errors.New("Wrong request format"))
	}

	return result, nil
}
