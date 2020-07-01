package rpcserver

import (
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
)

/*
handleGetBeaconBestState - RPC get beacon best state
*/
func (httpServer *HttpServer) handleGetBeaconBestState(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {

	beaconBestState, err := httpServer.blockService.GetBeaconBestState()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetClonedBeaconBestStateError, err)
	}

	err = beaconBestState.InitStateRootHash(httpServer.config.BlockChain)
	if err != nil {
		panic(err)
	}

	// TODO: re-produce field that not marshal
	//best block
	block, _, err := httpServer.config.BlockChain.GetBeaconBlockByHash(beaconBestState.BestBlockHash)
	if err != nil || block == nil {
		fmt.Println("block ", block)
		panic(err)
	}
	beaconBestState.BestBlock = *block
	//@tin beacon committee
	//@tin shard committee
	if beaconBestState.RewardReceiver == nil {
		beaconBestState.RewardReceiver = make(map[string]privacy.PaymentAddress)
	}
	err = beaconBestState.RestoreBeaconCommittee()
	if err != nil {
		panic(err)
	}

	err = beaconBestState.RestoreShardCommittee()
	if err != nil {
		panic(err)
	}

	err = beaconBestState.RestoreBeaconPendingValidator()
	if err != nil {
		panic(err)
	}

	err = beaconBestState.RestoreShardPendingValidator()
	if err != nil {
		panic(err)
	}

	err = beaconBestState.RestoreCandidateBeaconWaitingForCurrentRandom()
	if err != nil {
		panic(err)
	}

	err = beaconBestState.RestoreCandidateBeaconWaitingForNextRandom()
	if err != nil {
		panic(err)
	}

	err = beaconBestState.RestoreCandidateShardWaitingForCurrentRandom()
	if err != nil {
		panic(err)
	}

	err = beaconBestState.RestoreCandidateShardWaitingForNextRandom()
	if err != nil {
		panic(err)
	}

	result := jsonresult.NewGetBeaconBestState(beaconBestState)
	return result, nil
}

/*
handleGetShardBestState - RPC get shard best state
*/
func (httpServer *HttpServer) handleGetShardBestState(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
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

	block, _, err := httpServer.config.BlockChain.GetShardBlockByHash(shardBestState.BestBlockHash)
	if err != nil || block == nil {
		fmt.Println("block ", block)
		return nil, rpcservice.NewRPCError(rpcservice.RPCInternalError, err)
	}
	shardBestState.BestBlock = block

	err = shardBestState.InitStateRootHash(httpServer.config.BlockChain.GetShardChainDatabase(shardID), httpServer.config.BlockChain)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInternalError, err)
	}

	err = shardBestState.RestoreCommittee(shardID, httpServer.config.BlockChain)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInternalError, err)
	}

	beaconConsensusRootHash, err := httpServer.config.BlockChain.GetBeaconConsensusRootHash(httpServer.config.BlockChain.GetBeaconBestState(), shardBestState.BeaconHeight)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInternalError, err)
	}

	beaconConsensusStateDB, err := statedb.NewWithPrefixTrie(beaconConsensusRootHash, statedb.NewDatabaseAccessWarper(httpServer.config.BlockChain.GetBeaconChainDatabase()))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInternalError, err)
	}
	mapStakingTx := statedb.GetMapStakingTx(beaconConsensusStateDB, httpServer.config.BlockChain.GetShardChainDatabase(shardID), httpServer.config.BlockChain.GetShardIDs(), int(shardID))
	shardBestState.StakingTx = mapStakingTx

	err = shardBestState.RestorePendingValidators(shardID, httpServer.config.BlockChain)
	if err != nil {
		panic(err)
	}

	result := jsonresult.NewGetShardBestState(shardBestState)
	return result, nil
}

// handleGetCandidateList - return list candidate of committee
func (httpServer *HttpServer) handleGetCandidateList(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {

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
	return result, nil
}

// handleGetCommitteeList - return current committee in network
func (httpServer *HttpServer) handleGetCommitteeList(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
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
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Shard ID empty"))
	}

	publicKey, ok := arrayParams[0].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Pub key is invalid"))
	}

	canStake, err := httpServer.blockService.CanPubkeyStake(publicKey)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	result := jsonresult.NewStakeResult(publicKey, canStake)
	return result, nil
}

func (httpServer *HttpServer) handleGetTotalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Shard ID empty"))
	}

	shardIdParam, ok := arrayParams[0].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Shard ID invalid"))
	}
	shardID := byte(shardIdParam)

	clonedShardBestState, err := httpServer.blockService.GetShardBestStateByShardID(shardID)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetClonedShardBestStateError, err)
	}

	result := jsonresult.NewTotalTransactionInShard(clonedShardBestState)
	return result, nil
}

func (httpServer *HttpServer) handleGetBeaconBestStateDetail(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {

	clonedBeaconBestState, err := httpServer.blockService.GetBeaconBestState()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetClonedBeaconBestStateError, err)
	}

	result := jsonresult.NewGetBeaconBestStateDetail(clonedBeaconBestState)
	return result, nil
}

/*
handleGetShardBestState - RPC get shard best state
*/
func (httpServer *HttpServer) handleGetShardBestStateDetail(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
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
	return result, nil
}
