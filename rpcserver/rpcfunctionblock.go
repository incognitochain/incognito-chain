package rpcserver

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/rpcserver/jsonresult"
	"github.com/ninjadotorg/constant/transaction"
)

// handleGetBestBlock implements the getbestblock command.
func (rpcServer RpcServer) handleGetBestBlock(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	result := jsonresult.GetBestBlockResult{
		// BestBlocks: make(map[string]jsonresult.GetBestBlockItem),
		BestBlocks: make(map[int]jsonresult.GetBestBlockItem),
	}
	for shardID, best := range rpcServer.config.BlockChain.BestState.Shard {
		// result.BestBlocks[strconv.Itoa(int(shardID))] = jsonresult.GetBestBlockItem{
		result.BestBlocks[int(shardID)] = jsonresult.GetBestBlockItem{
			Height:   best.BestBlock.Header.Height,
			Hash:     best.BestBlockHash.String(),
			TotalTxs: best.TotalTxns,
		}
	}
	beaconBestState := rpcServer.config.BlockChain.BestState.Beacon
	if beaconBestState == nil {
		return result, nil
	}
	result.BestBlocks[-1] = jsonresult.GetBestBlockItem{
		Height: beaconBestState.BestBlock.Header.Height,
		Hash:   beaconBestState.BestBlockHash.String(),
	}

	return result, nil
}

// handleGetBestBlock implements the getbestblock command.
func (rpcServer RpcServer) handleGetBestBlockHash(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	result := jsonresult.GetBestBlockHashResult{
		// BestBlockHashes: make(map[byte]string),
		BestBlockHashes: make(map[int]string),
	}
	for shardID, best := range rpcServer.config.BlockChain.BestState.Shard {
		result.BestBlockHashes[int(shardID)] = best.BestBlockHash.String()
	}
	beaconBestState := rpcServer.config.BlockChain.BestState.Beacon
	if beaconBestState == nil {
		return result, nil
	}
	result.BestBlockHashes[-1] = beaconBestState.BestBlockHash.String()
	return result, nil
}

/*
getblockcount RPC return information fo blockchain node
*/
func (rpcServer RpcServer) handleRetrieveBlock(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	paramsT, ok := params.([]interface{})
	if ok && len(paramsT) >= 2 {
		hashString := paramsT[0].(string)
		hash, errH := common.Hash{}.NewHashFromStr(hashString)
		if errH != nil {
			return nil, NewRPCError(ErrUnexpected, errH)
		}
		// block, errD := rpcServer.config.BlockChain.GetBlockByHash(hash)
		block, errD := rpcServer.config.BlockChain.GetShardBlockByHash(hash)
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
			best := rpcServer.config.BlockChain.BestState.Shard[shardID].BestBlock

			// blockHeight := block.Header.GetHeight()
			blockHeight := block.Header.Height
			// Get next block hash unless there are none.
			var nextHashString string
			// if blockHeight < best.Header.GetHeight() {
			if blockHeight < best.Header.Height {
				nextHash, err := rpcServer.config.BlockChain.GetShardBlockByHeight(blockHeight+1, shardID)
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
			result.MerkleRoot = block.Header.TxRoot.String()
			// result.Time = block.Header.(*blockchain.BlockHeaderShard).Timestamp
			result.Time = block.Header.Timestamp
			// result.ShardID = block.Header.(*blockchain.BlockHeaderShard).ShardID
			result.ShardID = block.Header.ShardID
			// result.PreviousBlockHash = block.Header.(*blockchain.BlockHeaderShard).PrevBlockHash.String()
			result.PreviousBlockHash = block.Header.PrevBlockHash.String()
			result.NextBlockHash = nextHashString
			result.TxHashes = []string{}
			result.BlockProducerSign = block.ProducerSig
			result.BlockProducer = block.Header.Producer
			// for _, tx := range block.Body.(*blockchain.BlockBodyShard).Transactions {
			for _, tx := range block.Body.Transactions {
				result.TxHashes = append(result.TxHashes, tx.Hash().String())
			}
		} else if verbosity == "2" {
			best := rpcServer.config.BlockChain.BestState.Shard[shardID].BestBlock

			// blockHeight := block.Header.GetHeight()
			blockHeight := block.Header.Height
			// Get next block hash unless there are none.
			var nextHashString string
			// if blockHeight < best.Header.GetHeight() {
			if blockHeight < best.Header.Height {
				nextHash, err := rpcServer.config.BlockChain.GetShardBlockByHeight(blockHeight+1, shardID)
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
			result.MerkleRoot = block.Header.TxRoot.String()
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
func (rpcServer RpcServer) handleGetBlocks(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	result := make([]jsonresult.GetBlockResult, 0)
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) != 2 {
		arrayParams = append(arrayParams, 0.0, 0.0)
	}
	numBlock := int(arrayParams[0].(float64))
	shardID := byte(arrayParams[1].(float64))
	bestBlock := rpcServer.config.BlockChain.BestState.Shard[shardID].BestBlock
	previousHash := bestBlock.Hash()
	for numBlock > 0 {
		numBlock--
		// block, errD := rpcServer.config.BlockChain.GetBlockByHash(previousHash)
		block, errD := rpcServer.config.BlockChain.GetShardBlockByHash(previousHash)
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
func (rpcServer RpcServer) handleGetBlockChainInfo(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	result := jsonresult.GetBlockChainInfoResult{
		ChainName:    rpcServer.config.ChainParams.Name,
		BestBlocks:   make(map[int]jsonresult.GetBestBlockItem),
		ActiveShards: rpcServer.config.ChainParams.ActiveShards,
	}
	for shardID, bestState := range rpcServer.config.BlockChain.BestState.Shard {
		result.BestBlocks[int(shardID)] = jsonresult.GetBestBlockItem{
			Height:           bestState.BestBlock.Header.Height,
			Hash:             bestState.BestBlockHash.String(),
			TotalTxs:         bestState.TotalTxns,
			SalaryFund:       bestState.BestBlock.Header.SalaryFund,
			BlockProducer:    bestState.BestBlock.Header.Producer,
			BlockProducerSig: bestState.BestBlock.ProducerSig,
		}
	}
	beaconBestState := rpcServer.config.BlockChain.BestState.Beacon
	result.BestBlocks[-1] = jsonresult.GetBestBlockItem{
		Height:           beaconBestState.BestBlock.Header.Height,
		Hash:             beaconBestState.BestBlockHash.String(),
		BlockProducer:    beaconBestState.BestBlock.Header.Producer,
		BlockProducerSig: beaconBestState.BestBlock.ProducerSig,
	}
	return result, nil
}

/*
getblockcount RPC return information fo blockchain node
*/
func (rpcServer RpcServer) handleGetBlockCount(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("params empty"))
	}
	params, ok := arrayParams[0].(float64)
	// params, ok := params.(float64)
	if !ok {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("Expected get float number params"))
	}
	paramNumber := int(params.(float64))
	shardID := byte(paramNumber)
	isGetBeacon := paramNumber == -1
	if isGetBeacon {
		if rpcServer.config.BlockChain.BestState != nil && rpcServer.config.BlockChain.BestState.Beacon != nil && rpcServer.config.BlockChain.BestState.Beacon.BestBlock != nil {
			return rpcServer.config.BlockChain.BestState.Beacon.BestBlock.Header.Height, nil
		}
	}

	if rpcServer.config.BlockChain.BestState != nil && rpcServer.config.BlockChain.BestState.Shard[shardID] != nil && rpcServer.config.BlockChain.BestState.Shard[shardID].BestBlock != nil {
		return rpcServer.config.BlockChain.BestState.Shard[shardID].BestBlock.Header.Height + 1, nil
	}
	return 0, nil
}

/*
getblockhash RPC return information fo blockchain node
*/
func (rpcServer RpcServer) handleGetBlockHash(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) != 2 {
		arrayParams = []interface{}{
			0.0,
			1.0,
		}
	}

	shardID := int(arrayParams[0].(float64))
	height := uint64(arrayParams[1].(float64))

	var hash *common.Hash
	var err error
	var beaconBlock *blockchain.BeaconBlock
	var shardBlock *blockchain.ShardBlock

	isGetBeacon := shardID == -1

	if isGetBeacon {
		beaconBlock, err = rpcServer.config.BlockChain.GetBeaconBlockByHeight(height)
	} else {
		shardBlock, err = rpcServer.config.BlockChain.GetShardBlockByHeight(height, byte(shardID))
	}

	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	if isGetBeacon {
		hash = beaconBlock.Hash()
	} else {
		hash = shardBlock.Hash()
	}
	// return hash.Hash().String(), nil
	return hash.String(), nil
}

// handleGetBlockHeader - return block header data
func (rpcServer RpcServer) handleGetBlockHeader(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	// Logger.log.Info(params)
	log.Printf("%+v", params)
	result := jsonresult.GetHeaderResult{}

	arrayParams := common.InterfaceSlice(params)
	// Logger.log.Info(arrayParams)
	log.Printf("arrayParams: %+v", arrayParams)
	if arrayParams == nil || len(arrayParams) == 0 || len(arrayParams) <= 3 {
		arrayParams = append(arrayParams, "", "", 0.0)
	}
	getBy := arrayParams[0].(string)
	block := arrayParams[1].(string)
	shardID := arrayParams[2].(float64)
	switch getBy {
	case "blockhash":
		bhash := common.Hash{}
		err := bhash.Decode(&bhash, block)
		// Logger.log.Info(bhash)
		log.Printf("%+v", bhash)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, errors.New("invalid blockhash format"))
		}
		// block, err := rpcServer.config.BlockChain.GetBlockByHash(&bhash)
		block, err := rpcServer.config.BlockChain.GetShardBlockByHash(&bhash)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, errors.New("block not exist"))
		}
		result.Header = block.Header
		// result.BlockNum = int(block.Header.GetHeight()) + 1
		result.BlockNum = int(block.Header.Height) + 1
		result.ShardID = uint8(shardID)
		result.BlockHash = bhash.String()
	case "blocknum":
		bnum, err := strconv.Atoi(block)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, errors.New("invalid blocknum format"))
		}
		fmt.Println(shardID)
		// if uint64(bnum-1) > rpcServer.config.BlockChain.BestState.Shard[uint8(shardID)].BestBlock.Header.GetHeight() || bnum <= 0 {
		if uint64(bnum-1) > rpcServer.config.BlockChain.BestState.Shard[uint8(shardID)].BestBlock.Header.Height || bnum <= 0 {
			return nil, NewRPCError(ErrUnexpected, errors.New("Block not exist"))
		}
		block, _ := rpcServer.config.BlockChain.GetShardBlockByHeight(uint64(bnum-1), uint8(shardID))

		if block != nil {
			result.Header = block.Header
			result.BlockHash = block.Hash().String()
		}
		result.BlockNum = bnum
		result.ShardID = uint8(shardID)
	default:
		return nil, NewRPCError(ErrUnexpected, errors.New("wrong request format"))
	}

	return result, nil
}
