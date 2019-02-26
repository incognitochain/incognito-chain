package rpcserver

import (
	"errors"
	"fmt"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/rpcserver/jsonresult"
)

/*
handleGetBeaconBestState - RPC get beacon best state
*/
func (rpcServer RpcServer) handleGetBeaconBestState(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	if rpcServer.config.BlockChain.BestState.Beacon == nil {
		return nil, NewRPCError(ErrUnexpected, errors.New("Best State beacon not existed"))
	}

	result := *rpcServer.config.BlockChain.BestState.Beacon
	result.BestBlock = nil
	return result, nil
}

/*
handleGetShardBestState - RPC get shard best state
*/
func (rpcServer RpcServer) handleGetShardBestState(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("Shard ID params empty"))
	}
	shardIdParam, ok := arrayParams[0].(float64)
	if !ok {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("Shard ID params invalid"))
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
	return valueResult, nil
}

func (rpcServer RpcServer) handleGetCandidateList(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	CSWFCR := rpcServer.config.BlockChain.BestState.Beacon.CandidateShardWaitingForCurrentRandom
	CSWFNR := rpcServer.config.BlockChain.BestState.Beacon.CandidateShardWaitingForNextRandom
	CBWFCR := rpcServer.config.BlockChain.BestState.Beacon.CandidateBeaconWaitingForCurrentRandom
	CBWFNR := rpcServer.config.BlockChain.BestState.Beacon.CandidateBeaconWaitingForNextRandom
	epoch := rpcServer.config.BlockChain.BestState.Beacon.Epoch
	result := jsonresult.CandidateListsResult{
		Epoch: epoch,
		CandidateShardWaitingForCurrentRandom:  CSWFCR,
		CandidateBeaconWaitingForCurrentRandom: CBWFCR,
		CandidateShardWaitingForNextRandom:     CSWFNR,
		CandidateBeaconWaitingForNextRandom:    CBWFNR,
	}
	return result, nil
}
func (rpcServer RpcServer) handleGetCommitteeList(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	beaconCommittee := rpcServer.config.BlockChain.BestState.Beacon.BeaconCommittee
	beaconPendingValidator := rpcServer.config.BlockChain.BestState.Beacon.BeaconPendingValidator
	shardCommittee := rpcServer.config.BlockChain.BestState.Beacon.ShardCommittee
	shardPendingValidator := rpcServer.config.BlockChain.BestState.Beacon.ShardPendingValidator
	epoch := rpcServer.config.BlockChain.BestState.Beacon.Epoch
	result := jsonresult.CommitteeListsResult{
		Epoch:                  epoch,
		BeaconCommittee:        beaconCommittee,
		BeaconPendingValidator: beaconPendingValidator,
		ShardCommittee:         shardCommittee,
		ShardPendingValidator:  shardPendingValidator,
	}
	return result, nil
}

/*
	Tell a public key can stake or not
	Compare this public key with database only (TODO: compare with mempool also)
	param #1: public key
	return #1: true (can stake), false (can't stake)
	return #2: error
*/
func (rpcServer RpcServer) handleCanPubkeyStake(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	pubkey := arrayParams[0].(string)
	temp := rpcServer.config.BlockChain.BestState.Beacon.GetValidStakers([]string{pubkey})
	fmt.Println("alksjdklajsdkljaskldjkasjdlkasjdkl ", temp)
	if len(temp) == 0 {

		return jsonresult.StakeResult{PublicKey: pubkey, CanStake: false}, nil
	}
	return jsonresult.StakeResult{PublicKey: pubkey, CanStake: true}, nil
}
func (self RpcServer) handleRetrieveCommiteeCandidate(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	// candidateInfo := self.config.BlockChain.GetCommitteCandidate(params.(string))
	// if candidateInfo == nil {
	// 	return nil, nil
	// }
	// result := jsonresult.RetrieveCommitteecCandidateResult{}
	// result.Init(candidateInfo)
	// return result, nil
	return nil, nil
}
