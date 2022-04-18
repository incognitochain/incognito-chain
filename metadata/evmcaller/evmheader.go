package evmcaller

import (
	"errors"
	"fmt"
	"math/big"

	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/incognitochain/incognito-chain/metadata/rpccaller"
)

type GetEVMHeaderByHashRes struct {
	rpccaller.RPCBaseRes
	Result *types.Header `json:"result"`
}

type GetEVMHeaderByNumberRes struct {
	rpccaller.RPCBaseRes
	Result *types.Header `json:"result"`
}

type GetEVMBlockNumRes struct {
	rpccaller.RPCBaseRes
	Result string `json:"result"`
}

type EVMHeaderResult struct {
	Header      types.Header
	IsForked    bool
	IsFinalized bool
}

func NewEVMHeaderResult() *EVMHeaderResult {
	return &EVMHeaderResult{
		Header:      types.Header{},
		IsForked:    false,
		IsFinalized: false,
	}
}

func GetEVMHeaderByHash(
	evmBlockHash rCommon.Hash,
	host string,
) (*types.Header, error) {
	rpcClient := rpccaller.NewRPCClient()
	getEVMHeaderByHashParams := []interface{}{evmBlockHash, false}
	var getEVMHeaderByHashRes GetEVMHeaderByHashRes
	err := rpcClient.RPCCall(
		"",
		host,
		"",
		"eth_getBlockByHash",
		getEVMHeaderByHashParams,
		&getEVMHeaderByHashRes,
	)
	if err != nil {
		Logger.log.Warnf("WARNING: an error occured during calling eth_getBlockByHash: %s", err)
		return nil, NewEVMCallerError(GetEVMHeaderByHashError, fmt.Errorf("An error occured during calling eth_getBlockByHash: %s", err))
	}
	if getEVMHeaderByHashRes.RPCError != nil {
		Logger.log.Warnf("WARNING: an error occured during calling eth_getBlockByHash: %s", getEVMHeaderByHashRes.RPCError.Message)
		return nil, NewEVMCallerError(GetEVMHeaderByHashError, fmt.Errorf("An error occured during calling eth_getBlockByHash: %s", getEVMHeaderByHashRes.RPCError.Message))
	}
	if getEVMHeaderByHashRes.Result == nil {
		Logger.log.Warnf("WARNING: an error occured during calling eth_getBlockByHash: result is nil")
		return nil, NewEVMCallerError(GetEVMHeaderByHashError, fmt.Errorf("An error occured during calling eth_getBlockByHash: result is nil"))
	}

	return getEVMHeaderByHashRes.Result, nil
}

func GetEVMHeaderByNumber(blockNumber *big.Int, host string) (*types.Header, error) {
	rpcClient := rpccaller.NewRPCClient()
	getEVMHeaderByNumberParams := []interface{}{fmt.Sprintf("0x%x", blockNumber), false}
	var getEVMHeaderByNumberRes GetEVMHeaderByNumberRes
	err := rpcClient.RPCCall(
		"",
		host,
		"",
		"eth_getBlockByNumber",
		getEVMHeaderByNumberParams,
		&getEVMHeaderByNumberRes,
	)
	if err != nil {
		return nil, err
	}
	if getEVMHeaderByNumberRes.RPCError != nil {
		Logger.log.Warnf("WARNING: an error occured during calling eth_getBlockByNumber: %s", getEVMHeaderByNumberRes.RPCError.Message)
		return nil, errors.New(fmt.Sprintf("An error occured during calling eth_getBlockByNumber: %s", getEVMHeaderByNumberRes.RPCError.Message))
	}

	if getEVMHeaderByNumberRes.Result == nil {
		Logger.log.Warnf("WARNING: an error occured during calling eth_getBlockByNumber: result is nil")
		return nil, errors.New(fmt.Sprintf("An error occured during calling eth_getBlockByNumber: result is nil"))
	}
	return getEVMHeaderByNumberRes.Result, nil
}

// GetMostRecentEVMBlockHeight get most recent block height on Ethereum/BSC/PLG
func GetMostRecentEVMBlockHeight(host string) (*big.Int, error) {
	rpcClient := rpccaller.NewRPCClient()
	params := []interface{}{}
	var getEVMBlockNumRes GetEVMBlockNumRes
	err := rpcClient.RPCCall(
		"",
		host,
		"",
		"eth_blockNumber",
		params,
		&getEVMBlockNumRes,
	)
	if err != nil {
		return nil, err
	}
	if getEVMBlockNumRes.RPCError != nil {
		Logger.log.Warnf("WARNING: an error occured during calling eth_blockNumber: %s", getEVMBlockNumRes.RPCError.Message)
		return nil, NewEVMCallerError(GetEVMBlockHeightError, fmt.Errorf("An error occured during calling eth_blockNumber: %s", getEVMBlockNumRes.RPCError.Message))
	}

	if len(getEVMBlockNumRes.Result) < 2 {
		Logger.log.Warnf("WARNING: invalid response calling eth_blockNumber: %s", getEVMBlockNumRes.Result)
		return nil, NewEVMCallerError(GetEVMBlockHeightError, fmt.Errorf("Invalid response calling eth_blockNumber: %s", getEVMBlockNumRes.Result))
	}

	blockNumber := new(big.Int)
	_, ok := blockNumber.SetString(getEVMBlockNumRes.Result[2:], 16)
	if !ok {
		Logger.log.Warnf("WARNING: Cannot convert blockNumber into integer: %s", getEVMBlockNumRes.Result)
		return nil, NewEVMCallerError(GetEVMBlockHeightError, fmt.Errorf("Cannot convert blockNumber into integer: %s", getEVMBlockNumRes.Result))
	}
	return blockNumber, nil
}

func CheckBlockFinality(
	evmHeader types.Header,
	evmBlockHash rCommon.Hash,
	hosts []string,
	minConfirmationBlocks int,
) (bool, bool, error) {
	isFinalized := false
	isForked := false
	// try with multiple hosts
	for _, host := range hosts {
		// re-check finality
		currentBlockHeight, err := GetMostRecentEVMBlockHeight(host)
		if err != nil {
			continue
		}
		if currentBlockHeight.Cmp(big.NewInt(0).Add(evmHeader.Number, big.NewInt(int64(minConfirmationBlocks)))) == -1 {
			Logger.log.Warnf("WARNING: It needs %v confirmation blocks for the process, "+
				"the requested block (%s) but the latest block (%s)", minConfirmationBlocks,
				evmHeader.Number.String(), currentBlockHeight.String())
			return isFinalized, isForked, nil
		}
		isFinalized = true

		// re-check fork, because the unfinalized blocks still have possibility of fork
		evmHeaderByNumber, err := GetEVMHeaderByNumber(evmHeader.Number, host)
		if err != nil {
			continue
		}
		if evmHeaderByNumber.Hash().String() != evmHeader.Hash().String() {
			Logger.log.Errorf("The requested evm BlockHash %v is being on fork branch!", evmBlockHash.String())
			isForked = true
		}
		return isFinalized, isForked, nil
	}
	return isFinalized, isForked, NewEVMCallerError(GetEVMBlockHeightError, errors.New("Can not get latest EVM block height from multiple hosts"))
}

func GetEVMHeaderResultMultipleHosts(
	evmBlockHash rCommon.Hash,
	hosts []string,
	minConfirmationBlocks int,
) (*EVMHeaderResult, error) {
	evmHeaderResult := NewEVMHeaderResult()
	// try with multiple hosts
	for _, host := range hosts {
		Logger.log.Infof("EVMHeader Call request with host: %v for block hash %v", host, evmBlockHash)
		// get evm header
		evmHeader, err := GetEVMHeaderByHash(evmBlockHash, host)
		if err != nil {
			continue
		}
		evmHeaderResult.Header = *evmHeader

		// check finality
		currentBlockHeight, err := GetMostRecentEVMBlockHeight(host)
		if err != nil {
			continue
		}
		if currentBlockHeight.Cmp(big.NewInt(0).Add(evmHeaderResult.Header.Number, big.NewInt(int64(minConfirmationBlocks)))) == -1 {
			Logger.log.Warnf("WARNING: It needs %v confirmation blocks for the process, "+
				"the requested block (%s) but the latest block (%s)", minConfirmationBlocks,
				evmHeaderResult.Header.Number.String(), currentBlockHeight.String())
			evmHeaderResult.IsFinalized = false
			return evmHeaderResult, nil
		}
		evmHeaderResult.IsFinalized = true

		// check fork
		evmHeaderByNumber, err := GetEVMHeaderByNumber(evmHeader.Number, host)
		if err != nil {
			continue
		}
		if evmHeaderByNumber.Hash().String() != evmHeader.Hash().String() {
			Logger.log.Errorf("The requested evm BlockHash %v is being on fork branch!", evmBlockHash.String())
			evmHeaderResult.IsForked = true
		} else {
			evmHeaderResult.IsForked = false
		}
		return evmHeaderResult, nil
	}
	return nil, NewEVMCallerError(GetEVMBlockHeightError, errors.New("Can not get EVM header from multiple hosts"))
}

func GetEVMHeaderResult(
	evmBlockHash rCommon.Hash,
	hosts []string,
	minConfirmationBlocks int,
	networkPrefix string,
) (*EVMHeaderResult, error) {
	var err error
	// check EVM header is existed from cache
	evmHeaderRes, isExisted := handleGetEVMHeaderResult(networkPrefix, evmBlockHash.String())

	isReplaceCache := false
	if isExisted {
		// if existed, re-check block finality (if not finalized before)
		if evmHeaderRes.IsFinalized || evmHeaderRes.IsForked {
			return evmHeaderRes, nil
		}

		isFinalized, isForked, err := CheckBlockFinality(evmHeaderRes.Header, evmBlockHash, hosts, minConfirmationBlocks)
		if err != nil {
			Logger.log.Errorf("An error occured during re-checking block finality: %v", err)
			return evmHeaderRes, NewEVMCallerError(GetEVMHeaderResultFromDBError, fmt.Errorf("An error occured during re-checking block finality: %v", err))
		}
		if isFinalized || isForked {
			evmHeaderRes.IsFinalized = isFinalized
			evmHeaderRes.IsForked = isForked
			isReplaceCache = true
		} else {
			return evmHeaderRes, nil
		}
	} else {
		// if not existed, call RPC to EVM's node to get EVM block
		evmHeaderRes, err = GetEVMHeaderResultMultipleHosts(evmBlockHash, hosts, minConfirmationBlocks)
		if err != nil {
			Logger.log.Errorf("An error occured during getting evm header result from APIs : %v", err)
			return nil, NewEVMCallerError(GetEVMHeaderResultFromDBError, fmt.Errorf("An error occured during getting evm header result from APIs : %v", err))
		}
	}

	// store EVM header result to DB
	err = handleCacheEVMHeaderResult(networkPrefix, evmBlockHash.String(), *evmHeaderRes, isReplaceCache)
	if err != nil {
		Logger.log.Errorf("An error occured during caching evm header result: %v", err)
	}

	return evmHeaderRes, nil
}
