package random

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/incognitochain/incognito-chain/metadata/rpccaller"
)

type ETHClient struct {
	Protocol string

	IP   string
	Port string
}
type GetBlockNumberResult struct {
	rpccaller.RPCBaseRes
	Result string `json:"result"`
}

type GetBlockHeaderResult struct {
	rpccaller.RPCBaseRes
	Result *types.Header `json:"result"`
}

func NewETHClient(protocol string, ip string, port string) *ETHClient {
	return &ETHClient{
		Protocol: protocol,
		IP:       ip,
		Port:     port,
	}
}
func (ethClient *ETHClient) GetCurrentChainTimeStamp() (int64, error) {
	_, timestamp, _, err := ethClient.GetChainTimeStampAndNonce()
	if err != nil {
		return -1, NewRandomClientError(GetCurrentChainTimestampError, err)
	}
	return timestamp, nil
}
func (ethClient *ETHClient) GetChainTimeStampAndNonce() (int, int64, int64, error) {
	rpcClient := rpccaller.NewRPCClient()
	getBlockNumberResult := &GetBlockNumberResult{}
	err := rpcClient.RPCCall(ethClient.Protocol, ethClient.IP, ethClient.Port, "eth_blockNumber", []interface{}{}, getBlockNumberResult)
	if err != nil {
		return -1, -1, -1, NewRandomClientError(GetBlockNumberResultError, err)
	}
	if getBlockNumberResult.RPCError != nil {
		return -1, -1, -1, NewRandomClientError(GetBlockNumberResultError, err)
	}
	getBlockHeaderResult := &GetBlockHeaderResult{}
	err = rpcClient.RPCCall(ethClient.Protocol, ethClient.IP, ethClient.Port, "eth_getBlockByNumber", []interface{}{getBlockNumberResult.Result, false}, getBlockHeaderResult)
	if err != nil {
		return -1, -1, -1, NewRandomClientError(GetBlockHeaderResultError, err)
	}
	if getBlockHeaderResult.RPCError != nil {
		return -1, -1, -1, NewRandomClientError(GetBlockHeaderResultError, err)
	}
	chainHeight, err := hexutil.DecodeUint64(getBlockNumberResult.Result)
	if err != nil {
		return -1, -1, -1, NewRandomClientError(DecodeHexStringError, fmt.Errorf("Failed to parse chain height with error %+v", err))
	}
	nonce := getBlockHeaderResult.Result.Nonce.Uint64()
	return int(chainHeight), int64(getBlockHeaderResult.Result.Time), int64(nonce), nil
}
func (ethClient *ETHClient) GetTimeStampAndNonceByBlockHeight(blockHeight int) (int64, int64, error) {
	rpcClient := rpccaller.NewRPCClient()
	getBlockHeaderResult := &GetBlockHeaderResult{}
	err := rpcClient.RPCCall(ethClient.Protocol, ethClient.IP, ethClient.Port, "eth_getBlockByNumber", []interface{}{hexutil.EncodeUint64(uint64(blockHeight)), false}, getBlockHeaderResult)
	if err != nil {
		return -1, -1, NewRandomClientError(GetBlockHeaderResultError, err)
	}
	if getBlockHeaderResult.RPCError != nil {
		return -1, -1, NewRandomClientError(GetBlockHeaderResultError, err)
	}
	nonce := getBlockHeaderResult.Result.Nonce.Uint64()
	return int64(getBlockHeaderResult.Result.Time), int64(nonce), nil
}
func (ethClient *ETHClient) VerifyNonceWithTimestamp(timestamp int64, nonce int64) (bool, error) {
	_, _, tempNonce, err := ethClient.GetNonceByTimestamp(timestamp)
	if err != nil {
		return false, err
	}
	return tempNonce == nonce, nil
}
func (ethClient *ETHClient) GetNonceByTimestamp(timestamp int64) (int, int64, int64, error) {
	var (
		chainHeight    int
		chainTimestamp int64
		nonce          int64
		err            error
	)
	chainHeight, chainTimestamp, nonce, err = ethClient.GetChainTimeStampAndNonce()
	if err != nil {
		return 0, 0, -1, err
	}
	blockHeight, err := estimateBlockHeight(ethClient, timestamp, chainHeight, chainTimestamp, EthereumKovanEsitmateTime)
	if err != nil {
		return 0, 0, -1, err
	}
	blockTimestamp, _, err = ethClient.GetTimeStampAndNonceByBlockHeight(blockHeight)
	if err != nil {
		return 0, 0, -1, err
	}
	if blockTimestamp == MaxTimeStamp {
		return 0, 0, -1, NewRandomClientError(APIError, errors.New("Can't get result from API"))
	}
	if blockTimestamp > timestamp {
		for blockTimestamp > timestamp {
			blockHeight--
			blockTimestamp, _, err = ethClient.GetTimeStampAndNonceByBlockHeight(blockHeight)
			if err != nil {
				return 0, 0, -1, err
			}
			if blockTimestamp == MaxTimeStamp {
				return 0, 0, -1, NewRandomClientError(APIError, errors.New("Can't get result from API"))
			}
			if blockTimestamp <= timestamp {
				blockHeight++
				break
			}
		}
	} else {
		for blockTimestamp <= timestamp {
			blockHeight++
			if blockHeight > chainHeight {
				return 0, 0, -1, NewRandomClientError(APIError, errors.New("Timestamp is greater than timestamp of highest block"))
			}
			blockTimestamp, _, err = ethClient.GetTimeStampAndNonceByBlockHeight(blockHeight)
			if err != nil {
				return 0, 0, -1, err
			}
			if blockTimestamp == MaxTimeStamp {
				return 0, 0, -1, NewRandomClientError(APIError, errors.New("Can't get result from API"))
			}
			if blockTimestamp > timestamp {
				break
			}
		}
	}
	timestamp, nonce, err = ethClient.GetTimeStampAndNonceByBlockHeight(blockHeight)
	if err != nil {
		return 0, 0, -1, err
	}
	return blockHeight, timestamp, nonce, nil
}
