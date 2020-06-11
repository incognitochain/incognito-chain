package rpcserver

import (
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
)

/*
handleGetInOutPeerMessageCount - return all inbound/outbound message count by peer which this node connected
*/
func (httpServer *HttpServer) handleGetInOutMessageCount(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	paramsArray := common.InterfaceSlice(params)
	if paramsArray == nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, nil)
	}

	result, err := jsonresult.NewGetInOutMessageCountResult(paramsArray)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	return result, nil
}

/*
handleGetInOutPeerMessages - return all inbound/outbound messages peer which this node connected
*/
func (httpServer *HttpServer) handleGetInOutMessages(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	paramsArray := common.InterfaceSlice(params)
	if paramsArray == nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, nil)
	}

	result, err := jsonresult.NewGetInOutMessageResult(paramsArray)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	return result, nil
}

/*
handleGetAllConnectedPeers - return all connnected peers which this node connected
*/
func (httpServer *HttpServer) handleGetAllConnectedPeers(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	result := jsonresult.NewGetAllConnectedPeersResult(*httpServer.config.ConnMgr, httpServer.config.BlockChain)
	return result, nil
}

/*
handleGetAllPeers - return all peers which this node connected
*/
func (httpServer *HttpServer) handleGetAllPeers(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	result := jsonresult.NewGetAllPeersResult(*httpServer.config.AddrMgr)
	return result, nil
}

func (httpServer *HttpServer) handleGetNodeRole(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return httpServer.config.Server.GetNodeRole(), nil
}

func (httpServer *HttpServer) handleGetNetWorkInfo(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	result, err := jsonresult.NewGetNetworkInfoResult(httpServer.config.ProtocolVersion, *httpServer.config.ConnMgr, httpServer.config.Wallet)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	return result, nil
}

func (httpServer *HttpServer) handleCheckHashValue(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) == 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Expected array component"))
	}
	hashParams, ok := arrayParams[0].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Expected hash string value"))
	}
	// param #1: transaction Hash

	isTransaction, isShardBlock, isBeaconBlock, err := httpServer.blockService.CheckHashValue(hashParams)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}

	result := jsonresult.HashValueDetail{
		IsBlock:       isShardBlock,
		IsTransaction: isTransaction,
		IsBeaconBlock: isBeaconBlock,
	}

	return result, nil
}

/*
handleGetConnectionCount - RPC returns the number of connections to other nodes.
*/
func (httpServer *HttpServer) handleGetConnectionCount(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	result := httpServer.networkService.GetConnectionCount()
	return result, nil
}

// handleGetActiveShards - return active shard num
func (httpServer *HttpServer) handleGetActiveShards(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	activeShards := httpServer.blockService.GetActiveShards()
	return activeShards, nil
}

func (httpServer *HttpServer) handleGetMaxShardsNumber(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	result := common.MaxShardNumber
	return result, nil
}

func (httpServer *HttpServer) handleGetStakingAmount(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) <= 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("ErrRPCInvalidParams"))
	}

	stakingTypeParam, ok := arrayParams[0].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("staking type is invalid"))
	}
	stakingType := int(stakingTypeParam)

	amount := rpcservice.GetStakingAmount(stakingType, httpServer.config.ChainParams.StakingAmountShard)
	return amount, nil
}

func (httpServer *HttpServer) handleHashToIdenticon(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("ErrRPCInvalidParams"))
	}

	result, err := rpcservice.HashToIdenticon(arrayParams)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	return result, nil
}

// handleGetPublicKeyMining - return publickey mining which be used to verify block
func (httpServer *HttpServer) handleGetPublicKeyMining(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	keys := httpServer.config.ConsensusEngine.GetAllMiningPublicKeys()
	return keys, nil
}

func (httpServer *HttpServer) handleGenerateTokenID(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 2 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 2 elements"))
	}

	network, ok := arrayParams[0].(string)
	if !ok {
		rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("network invalid"))
	}

	tokenName, ok := arrayParams[1].(string)
	if !ok {
		rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("token name invalid"))
	}

	tokenID, err := rpcservice.GenerateTokenID(network, tokenName)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	} else {
		return tokenID.String(), nil
	}
}
