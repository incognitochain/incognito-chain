package rpcserver

import (
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
)

/*
handleGetBeaconBestState - RPC get beacon best state
*/
func (httpServer *HttpServer) handleGetBeaconBestState(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("handleGetBeaconBestState params: %+v", params)

	beaconBestState, err := httpServer.blockService.GetBeaconBestState()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetClonedBeaconBestStateError, err)
	}

	result := jsonresult.NewGetBeaconBestState(beaconBestState)
	Logger.log.Debugf("Get Beacon BestState: %+v", beaconBestState)
	return result, nil
}

/*
handleGetShardBestState - RPC get shard best state
*/
func (httpServer *HttpServer) handleGetShardBestState(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("handleGetShardBestState params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Shard ID empty"))
	}
	shardIdParam, ok := arrayParams[0].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Shard ID component invalid"))
	}
	shardID := byte(shardIdParam)

	shardBestState, err := httpServer.blockService.GetShardBestStateByShardID(shardID)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetClonedShardBestStateError, err)
	}

	result := jsonresult.NewGetShardBestState(shardBestState)
	Logger.log.Debugf("Get Shard BestState result: %+v", result)
	return result, nil
}

// handleGetCandidateList - return list candidate of committee
func (httpServer *HttpServer) handleGetCandidateList(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("handleGetCandidateList params: %+v", params)

	beacon, err := httpServer.blockService.GetBeaconBestState()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetClonedBeaconBestStateError, err)
	}

	CSWFCR := beacon.CandidateShardWaitingForCurrentRandom
	CSWFNR := beacon.CandidateShardWaitingForNextRandom
	CBWFCR := beacon.CandidateBeaconWaitingForCurrentRandom
	CBWFNR := beacon.CandidateBeaconWaitingForNextRandom
	epoch := beacon.Epoch
	result := jsonresult.CandidateListsResult{
		Epoch:                                  epoch,
		CandidateShardWaitingForCurrentRandom:  CSWFCR,
		CandidateBeaconWaitingForCurrentRandom: CBWFCR,
		CandidateShardWaitingForNextRandom:     CSWFNR,
		CandidateBeaconWaitingForNextRandom:    CBWFNR,
	}
	Logger.log.Debugf("handleGetCandidateList result: %+v", result)
	return result, nil
}

// handleGetCommitteeList - return current committee in network
func (httpServer *HttpServer) handleGetCommitteeList(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("handleGetCommitteeList params: %+v", params)
	clonedBeaconBestState, err := httpServer.blockService.GetBeaconBestState()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetClonedBeaconBestStateError, err)
	}

	beaconCommittee := clonedBeaconBestState.BeaconCommittee
	beaconPendingValidator := clonedBeaconBestState.BeaconPendingValidator
	shardCommittee := clonedBeaconBestState.ShardCommittee
	shardPendingValidator := clonedBeaconBestState.ShardPendingValidator
	epoch := clonedBeaconBestState.Epoch
	result := jsonresult.NewCommitteeListsResult(epoch, shardCommittee, shardPendingValidator, beaconCommittee, beaconPendingValidator)
	Logger.log.Debugf("handleGetCommitteeList result: %+v", result)
	return result, nil
}

/*
	Tell a public key can stake or not
	Compare this public key with database only
	param #1: public key
	return #1: true (can stake), false (can't stake)
	return #2: error
*/
func (httpServer *HttpServer) handleCanPubkeyStake(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("handleCanPubkeyStake params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Shard ID empty"))
	}

	publicKey, ok := arrayParams[0].(string)
	if !ok {
		Logger.log.Debugf("handleCanPubkeyStake result: %+v", nil)
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Pub key is invalid"))
	}

	canStake, err := httpServer.blockService.CanPubkeyStake(publicKey)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	result := jsonresult.NewStakeResult(publicKey, canStake)
	Logger.log.Debugf("handleCanPubkeyStake result: %+v", result)
	return result, nil
}

func (httpServer *HttpServer) handleGetTotalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("handleGetTotalTransaction params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 1 {
		Logger.log.Debugf("handleGetTotalTransaction result: %+v", nil)
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Shard ID empty"))
	}

	shardIdParam, ok := arrayParams[0].(float64)
	if !ok {
		Logger.log.Debugf("handleGetTotalTransaction result: %+v", nil)
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Shard ID invalid"))
	}
	shardID := byte(shardIdParam)

	clonedShardBestState, err := httpServer.blockService.GetShardBestStateByShardID(shardID)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetClonedShardBestStateError, err)
	}

	result := jsonresult.NewTotalTransactionInShard(clonedShardBestState)
	Logger.log.Debugf("handleGetTotalTransaction result: %+v", result)
	return result, nil
}

func (httpServer *HttpServer) handleGetBeaconBestStateDetail(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("handleGetBeaconBestState params: %+v", params)

	clonedBeaconBestState, err := httpServer.blockService.GetBeaconBestState()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetClonedBeaconBestStateError, err)
	}

	result := jsonresult.NewGetBeaconBestStateDetail(clonedBeaconBestState)
	Logger.log.Debugf("Get Beacon BestState: %+v", clonedBeaconBestState)
	return result, nil
}

/*
handleGetShardBestState - RPC get shard best state
*/
func (httpServer *HttpServer) handleGetShardBestStateDetail(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("handleGetShardBestStateDetail params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Shard ID empty"))
	}
	shardIdParam, ok := arrayParams[0].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Shard ID component invalid"))
	}
	shardID := byte(shardIdParam)

	shardBestState, err := httpServer.blockService.GetShardBestStateByShardID(shardID)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetClonedShardBestStateError, err)
	}

	result := jsonresult.NewGetShardBestStateDetail(shardBestState)
	Logger.log.Debugf("Get Shard BestState result: %+v", result)
	return result, nil
}
