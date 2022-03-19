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

func GetEVMHeader(
	evmBlockHash rCommon.Hash,
	host string,
) (*types.Header, bool, error) {
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
		return nil, false, NewEVMCallerError(GetEVMHeaderByHashError, fmt.Errorf("An error occured during calling eth_getBlockByHash: %s", err))
	}
	if getEVMHeaderByHashRes.RPCError != nil {
		Logger.log.Warnf("WARNING: an error occured during calling eth_getBlockByHash: %s", getEVMHeaderByHashRes.RPCError.Message)
		return nil, false, NewEVMCallerError(GetEVMHeaderByHashError, fmt.Errorf("An error occured during calling eth_getBlockByHash: %s", getEVMHeaderByHashRes.RPCError.Message))
	}
	if getEVMHeaderByHashRes.Result == nil {
		Logger.log.Warnf("WARNING: an error occured during calling eth_getBlockByHash: result is nil")
		return nil, false, NewEVMCallerError(GetEVMHeaderByHashError, fmt.Errorf("An error occured during calling eth_getBlockByHash: result is nil"))
	}

	evmHeaderByHash := getEVMHeaderByHashRes.Result
	headerNum := evmHeaderByHash.Number

	getEVMHeaderByNumberParams := []interface{}{fmt.Sprintf("0x%x", headerNum), false}
	var getEVMHeaderByNumberRes GetEVMHeaderByNumberRes
	err = rpcClient.RPCCall(
		"",
		host,
		"",
		"eth_getBlockByNumber",
		getEVMHeaderByNumberParams,
		&getEVMHeaderByNumberRes,
	)
	if err != nil {
		Logger.log.Warnf("WARNING: an error occured during calling eth_getBlockByNumber: %v", err)
		return nil, false, NewEVMCallerError(GetEVMHeaderByHeightError, fmt.Errorf("An error occured during calling eth_getBlockByNumber: %v", err))
	}
	if getEVMHeaderByNumberRes.RPCError != nil {
		Logger.log.Warnf("WARNING: an error occured during calling eth_getBlockByNumber: %s", getEVMHeaderByNumberRes.RPCError.Message)
		return nil, false, NewEVMCallerError(GetEVMHeaderByHeightError, fmt.Errorf("An error occured during calling eth_getBlockByNumber: %s", getEVMHeaderByNumberRes.RPCError.Message))
	}

	if getEVMHeaderByNumberRes.Result == nil {
		Logger.log.Warnf("WARNING: an error occured during calling eth_getBlockByNumber: result is nil")
		return nil, false, NewEVMCallerError(GetEVMHeaderByHeightError, fmt.Errorf("An error occured during calling eth_getBlockByNumber: result is nil"))
	}

	evmHeaderByNum := getEVMHeaderByNumberRes.Result
	if evmHeaderByNum.Hash().String() != evmHeaderByHash.Hash().String() {
		Logger.log.Warnf("WARNING: The requested eth BlockHash is being on fork branch, rejected!")
		return nil, true, nil
	}
	return evmHeaderByHash, false, nil
}

// getMostRecentEVMBlockHeight get most recent block height on Ethereum/BSC/PLG
func getMostRecentEVMBlockHeight(host string) (*big.Int, error) {
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

func checkBlockFinality(
	evmBlockHeight *big.Int,
	hosts []string,
	minConfirmationBlocks int,
) (bool, error) {
	// try with multiple hosts
	for _, host := range hosts {
		currentBlockHeight, err := getMostRecentEVMBlockHeight(host)
		if err != nil {
			continue
		}
		if currentBlockHeight.Cmp(big.NewInt(0).Add(evmBlockHeight, big.NewInt(int64(minConfirmationBlocks)))) == -1 {
			Logger.log.Warnf("WARNING: It needs %v confirmation blocks for the process, "+
				"the requested block (%s) but the latest block (%s)", minConfirmationBlocks,
				evmBlockHeight.String(), currentBlockHeight.String())
			return false, nil
		}
		return true, nil
	}
	return false, NewEVMCallerError(GetEVMBlockHeightError, errors.New("Can not get latest EVM block height from multiple hosts"))
}

func getEVMHeaderResultMultipleHosts(
	evmBlockHash rCommon.Hash,
	hosts []string,
	minConfirmationBlocks int,
) (*EVMHeaderResult, error) {
	evmHeaderResult := NewEVMHeaderResult()
	// try with multiple hosts
	for _, host := range hosts {
		Logger.log.Infof("EVMHeader Call request with host: %v", host)
		// get evm header and check fork
		evmHeader, isForked, err := GetEVMHeader(evmBlockHash, host)
		if err != nil {
			continue
		}
		if isForked {
			evmHeaderResult.IsForked = true
			return evmHeaderResult, nil
		}
		evmHeaderResult.Header = *evmHeader

		// check finality
		currentBlockHeight, err := getMostRecentEVMBlockHeight(host)
		if err != nil {
			continue
		}
		if currentBlockHeight.Cmp(big.NewInt(0).Add(evmHeaderResult.Header.Number, big.NewInt(int64(minConfirmationBlocks)))) == -1 {
			Logger.log.Warnf("WARNING: It needs %v confirmation blocks for the process, "+
				"the requested block (%s) but the latest block (%s)", minConfirmationBlocks,
				evmHeaderResult.Header.Number.String(), currentBlockHeight.String())
			evmHeaderResult.IsFinalized = false
		} else {
			evmHeaderResult.IsFinalized = true
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
	// check EVM header is existed from DB or not
	evmHeaderRes, isExisted := handleGetEVMHeaderResult(networkPrefix, evmBlockHash.String())

	isReplaceCache := false
	if isExisted {
		// if existed, re-check block finality (if not finalized before)
		if evmHeaderRes.IsFinalized || evmHeaderRes.IsForked {
			return evmHeaderRes, nil
		}

		isFinalized, err := checkBlockFinality(evmHeaderRes.Header.Number, hosts, minConfirmationBlocks)
		if err != nil {
			Logger.log.Errorf("An error occured during re-checking block finality: %v", err)
			return evmHeaderRes, NewEVMCallerError(GetEVMHeaderResultFromDBError, fmt.Errorf("An error occured during re-checking block finality: %v", err))
		}
		if !isFinalized {
			return evmHeaderRes, nil
		}
		evmHeaderRes.IsFinalized = isFinalized
		isReplaceCache = true
	} else {
		// if not existed, call RPC to EVM's node to get EVM block
		evmHeaderRes, err = getEVMHeaderResultMultipleHosts(evmBlockHash, hosts, minConfirmationBlocks)
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
