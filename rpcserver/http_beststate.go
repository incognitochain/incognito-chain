package rpcserver

import (
	"errors"
	"github.com/incognitochain/incognito-chain/blockchain"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)

/*
handleGetBeaconBestState - RPC get beacon best state
*/
func (httpServer *HttpServer) handleGetBeaconBestState(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetBeaconBestState params: %+v", params)
	if httpServer.config.BlockChain.BestState.Beacon == nil {
		Logger.log.Infof("handleGetBeaconBestState result: %+v", nil)
		return nil, NewRPCError(ErrUnexpected, errors.New("Best State beacon not existed"))
	}

	result := *httpServer.config.BlockChain.BestState.Beacon
	result.BestBlock = blockchain.BeaconBlock{}
	Logger.log.Infof("handleGetBeaconBestState result: %+v", result)
	return result, nil
}

/*
handleGetShardBestState - RPC get shard best state
*/
func (httpServer *HttpServer) handleGetShardBestState(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetShardBestState params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("Shard ID empty"))
	}
	shardIdParam, ok := arrayParams[0].(float64)
	if !ok {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("Shard ID component invalid"))
	}
	shardID := byte(shardIdParam)
	if httpServer.config.BlockChain.BestState.Shard == nil || len(httpServer.config.BlockChain.BestState.Shard) <= 0 {
		return nil, NewRPCError(ErrUnexpected, errors.New("Best State shard not existed"))
	}
	result, ok := httpServer.config.BlockChain.BestState.Shard[shardID]
	if !ok || result == nil {
		return nil, NewRPCError(ErrUnexpected, errors.New("Best State shard given by ID not existed"))
	}
	valueResult := *result
	valueResult.BestBlock = nil
	Logger.log.Infof("handleGetShardBestState result: %+v", result)
	return valueResult, nil
}

// handleGetCandidateList - return list candidate of committee
func (httpServer *HttpServer) handleGetCandidateList(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetCandidateList params: %+v", params)
	CSWFCR := httpServer.config.BlockChain.BestState.Beacon.CandidateShardWaitingForCurrentRandom
	CSWFNR := httpServer.config.BlockChain.BestState.Beacon.CandidateShardWaitingForNextRandom
	CBWFCR := httpServer.config.BlockChain.BestState.Beacon.CandidateBeaconWaitingForCurrentRandom
	CBWFNR := httpServer.config.BlockChain.BestState.Beacon.CandidateBeaconWaitingForNextRandom
	epoch := httpServer.config.BlockChain.BestState.Beacon.Epoch
	result := jsonresult.CandidateListsResult{
		Epoch:                                  epoch,
		CandidateShardWaitingForCurrentRandom:  CSWFCR,
		CandidateBeaconWaitingForCurrentRandom: CBWFCR,
		CandidateShardWaitingForNextRandom:     CSWFNR,
		CandidateBeaconWaitingForNextRandom:    CBWFNR,
	}
	Logger.log.Infof("handleGetCandidateList result: %+v", result)
	return result, nil
}

// handleGetCommitteeList - return current committee in network
func (httpServer *HttpServer) handleGetCommitteeList(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetCommitteeList params: %+v", params)
	beaconCommittee := httpServer.config.BlockChain.BestState.Beacon.BeaconCommittee
	beaconPendingValidator := httpServer.config.BlockChain.BestState.Beacon.BeaconPendingValidator
	shardCommittee := httpServer.config.BlockChain.BestState.Beacon.GetShardCommittee()
	shardPendingValidator := httpServer.config.BlockChain.BestState.Beacon.GetShardPendingValidator()
	epoch := httpServer.config.BlockChain.BestState.Beacon.Epoch
	result := jsonresult.CommitteeListsResult{
		Epoch:                  epoch,
		BeaconCommittee:        beaconCommittee,
		BeaconPendingValidator: beaconPendingValidator,
		ShardCommittee:         shardCommittee,
		ShardPendingValidator:  shardPendingValidator,
	}
	Logger.log.Infof("handleGetCommitteeList result: %+v", result)
	return result, nil
}

/*
	Tell a public key can stake or not
	Compare this public key with database only
	param #1: public key
	return #1: true (can stake), false (can't stake)
	return #2: error
*/
func (httpServer *HttpServer) handleCanPubkeyStake(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleCanPubkeyStake params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	publicKey, ok := arrayParams[0].(string)
	if !ok {
		Logger.log.Infof("handleCanPubkeyStake result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("Pub key is invalid"))
	}
	temp := httpServer.config.BlockChain.BestState.Beacon.GetValidStakers([]string{publicKey})
	if len(temp) == 0 {
		result := jsonresult.StakeResult{PublicKey: publicKey, CanStake: false}
		Logger.log.Infof("handleCanPubkeyStake result: %+v", result)
		return result, nil
	}
	if common.IndexOfStrInHashMap(publicKey, httpServer.config.TxMemPool.CandidatePool) > 0 {
		result := jsonresult.StakeResult{PublicKey: publicKey, CanStake: false}
		Logger.log.Infof("handleCanPubkeyStake result: %+v", result)
		return result, nil
	}
	result := jsonresult.StakeResult{PublicKey: publicKey, CanStake: true}
	Logger.log.Infof("handleCanPubkeyStake result: %+v", result)
	return result, nil
}

func (httpServer *HttpServer) handleGetTotalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetTotalTransaction params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		Logger.log.Infof("handleGetTotalTransaction result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("Shard ID empty"))
	}
	shardIdParam, ok := arrayParams[0].(float64)
	if !ok {
		Logger.log.Infof("handleGetTotalTransaction result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("Shard ID invalid"))
	}
	shardID := byte(shardIdParam)
	if httpServer.config.BlockChain.BestState.Shard == nil || len(httpServer.config.BlockChain.BestState.Shard) <= 0 {
		Logger.log.Infof("handleGetTotalTransaction result: %+v", nil)
		return nil, NewRPCError(ErrUnexpected, errors.New("Best State shard not existed"))
	}
	shardBeststate, ok := httpServer.config.BlockChain.BestState.Shard[shardID]
	if !ok || shardBeststate == nil {
		Logger.log.Infof("handleGetTotalTransaction result: %+v", nil)
		return nil, NewRPCError(ErrUnexpected, errors.New("Best State shard given by ID not existed"))
	}
	result := jsonresult.TotalTransactionInShard{
		TotalTransactions:                 shardBeststate.TotalTxns,
		TotalTransactionsExcludeSystemTxs: shardBeststate.TotalTxnsExcludeSalary,
		SalaryTransaction:                 shardBeststate.TotalTxns - shardBeststate.TotalTxnsExcludeSalary,
	}
	Logger.log.Infof("handleGetTotalTransaction result: %+v", result)
	return result, nil
}
