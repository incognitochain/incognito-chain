package rpcserver

import (
	"errors"
	"github.com/constant-money/constant-chain/blockchain"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/rpcserver/jsonresult"
)

/*
handleGetBeaconBestState - RPC get beacon best state
*/
func (rpcServer RpcServer) handleGetBeaconBestState(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetBeaconBestState params: %+v", params)
	if rpcServer.config.BlockChain.BestState.Beacon == nil {
		Logger.log.Infof("handleGetBeaconBestState result: %+v", nil)
		return nil, NewRPCError(ErrUnexpected, errors.New("Best State beacon not existed"))
	}

	result := *rpcServer.config.BlockChain.BestState.Beacon
	result.BestBlock = blockchain.BeaconBlock{}
	Logger.log.Infof("handleGetBeaconBestState result: %+v", result)
	return result, nil
}

/*
handleGetShardBestState - RPC get shard best state
*/
func (rpcServer RpcServer) handleGetShardBestState(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
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
	if rpcServer.config.BlockChain.BestState.Shard == nil || len(rpcServer.config.BlockChain.BestState.Shard) <= 0 {
		return nil, NewRPCError(ErrUnexpected, errors.New("Best State shard not existed"))
	}
	result, ok := rpcServer.config.BlockChain.BestState.Shard[shardID]
	if !ok || result == nil {
		return nil, NewRPCError(ErrUnexpected, errors.New("Best State shard given by ID not existed"))
	}
	valueResult := *result
	valueResult.BestBlock = nil
	Logger.log.Infof("handleGetShardBestState result: %+v", result)
	return valueResult, nil
}

func (rpcServer RpcServer) handleGetCandidateList(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetCandidateList params: %+v", params)
	CSWFCR := rpcServer.config.BlockChain.BestState.Beacon.CandidateShardWaitingForCurrentRandom
	CSWFNR := rpcServer.config.BlockChain.BestState.Beacon.CandidateShardWaitingForNextRandom
	CBWFCR := rpcServer.config.BlockChain.BestState.Beacon.CandidateBeaconWaitingForCurrentRandom
	CBWFNR := rpcServer.config.BlockChain.BestState.Beacon.CandidateBeaconWaitingForNextRandom
	epoch := rpcServer.config.BlockChain.BestState.Beacon.Epoch
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
func (rpcServer RpcServer) handleGetCommitteeList(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetCommitteeList params: %+v", params)
	beaconCommittee := rpcServer.config.BlockChain.BestState.Beacon.BeaconCommittee
	beaconPendingValidator := rpcServer.config.BlockChain.BestState.Beacon.BeaconPendingValidator
	shardCommittee := rpcServer.config.BlockChain.BestState.Beacon.GetShardCommittee()
	shardPendingValidator := rpcServer.config.BlockChain.BestState.Beacon.GetShardPendingValidator()
	epoch := rpcServer.config.BlockChain.BestState.Beacon.Epoch
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
func (rpcServer RpcServer) handleCanPubkeyStake(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleCanPubkeyStake params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	publicKey, ok := arrayParams[0].(string)
	if !ok {
		Logger.log.Infof("handleCanPubkeyStake result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("Pub key is invalid"))
	}
	temp := rpcServer.config.BlockChain.BestState.Beacon.GetValidStakers([]string{publicKey})
	if len(temp) == 0 {
		result := jsonresult.StakeResult{PublicKey: publicKey, CanStake: false}
		Logger.log.Infof("handleCanPubkeyStake result: %+v", result)
		return result, nil
	}
	if common.IndexOfStrInHashMap(publicKey, rpcServer.config.TxMemPool.CandidatePool) > 0 {
		result := jsonresult.StakeResult{PublicKey: publicKey, CanStake: false}
		Logger.log.Infof("handleCanPubkeyStake result: %+v", result)
		return result, nil
	}
	result := jsonresult.StakeResult{PublicKey: publicKey, CanStake: true}
	Logger.log.Infof("handleCanPubkeyStake result: %+v", result)
	return result, nil
}

func (rpcServer RpcServer) handleGetTotalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
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
	if rpcServer.config.BlockChain.BestState.Shard == nil || len(rpcServer.config.BlockChain.BestState.Shard) <= 0 {
		Logger.log.Infof("handleGetTotalTransaction result: %+v", nil)
		return nil, NewRPCError(ErrUnexpected, errors.New("Best State shard not existed"))
	}
	shardBeststate, ok := rpcServer.config.BlockChain.BestState.Shard[shardID]
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
