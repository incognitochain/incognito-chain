package rpcserver

import (
	"github.com/ninjadotorg/constant/rpcserver/jsonresult"
	"strconv"
	"encoding/hex"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/transaction"
	"encoding/json"
	"errors"
	"fmt"
)

// handleGetBestBlock implements the getbestblock command.
func (rpcServer RpcServer) handleGetBestBlock(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	result := jsonresult.GetBestBlockResult{
		BestBlocks: make(map[string]jsonresult.GetBestBlockItem),
	}
	for chainID, best := range rpcServer.config.BlockChain.BestState {
		result.BestBlocks[strconv.Itoa(chainID)] = jsonresult.GetBestBlockItem{
			Height:   best.BestBlock.Header.Height,
			Hash:     best.BestBlockHash.String(),
			TotalTxs: best.TotalTxns,
		}
	}
	return result, nil
}

// handleGetBestBlock implements the getbestblock command.
func (rpcServer RpcServer) handleGetBestBlockHash(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	result := jsonresult.GetBestBlockHashResult{
		BestBlockHashes: make(map[string]string),
	}
	for chainID, best := range rpcServer.config.BlockChain.BestState {
		result.BestBlockHashes[strconv.Itoa(chainID)] = best.BestBlockHash.String()
	}
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
		block, errD := rpcServer.config.BlockChain.GetBlockByBlockHash(hash)
		if errD != nil {
			return nil, NewRPCError(ErrUnexpected, errD)
		}
		result := jsonresult.GetBlockResult{}

		verbosity := paramsT[1].(string)

		chainId := block.Header.ChainID

		if verbosity == "0" {
			data, err := json.Marshal(block)
			if err != nil {
				return nil, NewRPCError(ErrUnexpected, err)
			}
			result.Data = hex.EncodeToString(data)
		} else if verbosity == "1" {
			best := rpcServer.config.BlockChain.BestState[chainId]

			blockHeight := block.Header.Height
			// Get next block hash unless there are none.
			var nextHashString string
			if blockHeight < best.Height {
				nextHash, err := rpcServer.config.BlockChain.GetBlockByBlockHeight(blockHeight+1, chainId)
				if err != nil {
					return nil, NewRPCError(ErrUnexpected, err)
				}
				nextHashString = nextHash.Hash().String()
			}

			result.Hash = block.Hash().String()
			result.Confirmations = int64(1 + best.Height - blockHeight)
			result.Height = block.Header.Height
			result.Version = block.Header.Version
			result.MerkleRoot = block.Header.MerkleRoot.String()
			result.Time = block.Header.Timestamp
			result.ChainID = block.Header.ChainID
			result.PreviousBlockHash = block.Header.PrevBlockHash.String()
			result.NextBlockHash = nextHashString
			result.TxHashes = []string{}
			result.BlockProducerSign = block.BlockProducerSig
			for _, tx := range block.Transactions {
				result.TxHashes = append(result.TxHashes, tx.Hash().String())
			}
		} else if verbosity == "2" {
			best := rpcServer.config.BlockChain.BestState[chainId]

			blockHeight := block.Header.Height
			// Get next block hash unless there are none.
			var nextHashString string
			if blockHeight < best.Height {
				nextHash, err := rpcServer.config.BlockChain.GetBlockByBlockHeight(blockHeight+1, chainId)
				if err != nil {
					return nil, NewRPCError(ErrUnexpected, err)
				}
				nextHashString = nextHash.Hash().String()
			}

			result.Hash = block.Hash().String()
			result.Confirmations = int64(1 + best.Height - blockHeight)
			result.Height = block.Header.Height
			result.Version = block.Header.Version
			result.MerkleRoot = block.Header.MerkleRoot.String()
			result.Time = block.Header.Timestamp
			result.ChainID = block.Header.ChainID
			result.PreviousBlockHash = block.Header.PrevBlockHash.String()
			result.NextBlockHash = nextHashString
			result.Txs = make([]jsonresult.GetBlockTxResult, 0)
			for _, tx := range block.Transactions {
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
	numBlock := int(arrayParams[0].(float64))
	chainID := int(arrayParams[1].(float64))
	bestBlock := rpcServer.config.BlockChain.BestState[chainID].BestBlock
	previousHash := bestBlock.Hash()
	for numBlock > 0 {
		numBlock--
		block, errD := rpcServer.config.BlockChain.GetBlockByBlockHash(previousHash)
		if errD != nil {
			return nil, NewRPCError(ErrUnexpected, errD)
		}
		blockResult := jsonresult.GetBlockResult{}
		blockResult.Init(block)
		result = append(result, blockResult)
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
		ChainName:  rpcServer.config.ChainParams.Name,
		BestBlocks: make(map[string]jsonresult.GetBestBlockItem),
	}
	for chainID, bestState := range rpcServer.config.BlockChain.BestState {
		result.BestBlocks[strconv.Itoa(chainID)] = jsonresult.GetBestBlockItem{
			Height:           bestState.BestBlock.Header.Height,
			Hash:             bestState.BestBlockHash.String(),
			TotalTxs:         bestState.TotalTxns,
			SalaryFund:       bestState.BestBlock.Header.SalaryFund,
			BasicSalary:      bestState.BestBlock.Header.GOVConstitution.GOVParams.BasicSalary,
			SalaryPerTx:      bestState.BestBlock.Header.GOVConstitution.GOVParams.SalaryPerTx,
			BlockProducer:    bestState.BestBlock.BlockProducer,
			BlockProducerSig: bestState.BestBlock.BlockProducerSig,
		}
	}
	return result, nil
}

/*
getblockcount RPC return information fo blockchain node
*/
func (rpcServer RpcServer) handleGetBlockCount(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	chainId := byte(int(params.(float64)))
	if rpcServer.config.BlockChain.BestState != nil && rpcServer.config.BlockChain.BestState[chainId] != nil && rpcServer.config.BlockChain.BestState[chainId].BestBlock != nil {
		return rpcServer.config.BlockChain.BestState[chainId].BestBlock.Header.Height + 1, nil
	}
	return 0, nil
}

/*
getblockhash RPC return information fo blockchain node
*/
func (rpcServer RpcServer) handleGetBlockHash(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	chainId := byte(int(arrayParams[0].(float64)))
	height := int32(arrayParams[1].(float64))
	hash, err := rpcServer.config.BlockChain.GetBlockByBlockHeight(height, chainId)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	return hash.Hash().String(), nil
}

// handleGetBlockHeader - return block header data
func (rpcServer RpcServer) handleGetBlockHeader(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Info(params)
	result := jsonresult.GetHeaderResult{}

	arrayParams := common.InterfaceSlice(params)
	Logger.log.Info(arrayParams)
	getBy := arrayParams[0].(string)
	block := arrayParams[1].(string)
	chainID := arrayParams[2].(float64)
	switch getBy {
	case "blockhash":
		bhash := common.Hash{}
		err := bhash.Decode(&bhash, block)
		Logger.log.Info(bhash)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, errors.New("Invalid blockhash format"))
		}
		block, err := rpcServer.config.BlockChain.GetBlockByBlockHash(&bhash)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, errors.New("Block not exist"))
		}
		result.Header = block.Header
		result.BlockNum = int(block.Header.Height) + 1
		result.ChainID = uint8(chainID)
		result.BlockHash = bhash.String()
	case "blocknum":
		bnum, err := strconv.Atoi(block)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, errors.New("Invalid blocknum format"))
		}
		fmt.Println(chainID)
		if int32(bnum-1) > rpcServer.config.BlockChain.BestState[uint8(chainID)].Height || bnum <= 0 {
			return nil, NewRPCError(ErrUnexpected, errors.New("Block not exist"))
		}
		block, _ := rpcServer.config.BlockChain.GetBlockByBlockHeight(int32(bnum-1), uint8(chainID))
		result.Header = block.Header
		result.BlockNum = bnum
		result.ChainID = uint8(chainID)
		result.BlockHash = block.Hash().String()
	default:
		return nil, NewRPCError(ErrUnexpected, errors.New("Wrong request format"))
	}

	return result, nil
}
