package rpcserver

import (
	"encoding/json"
	"errors"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
)

/*
handleGetBeaconBestState - RPC get beacon best state
*/
func (httpServer *HttpServer) handleGetBeaconBestState(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	allViews := []*blockchain.BeaconBestState{}
	beaconBestState := &blockchain.BeaconBestState{}
	beaconDB := httpServer.blockService.BlockChain.GetBeaconChainDatabase()
	beaconViews, err := rawdbv2.GetBeaconViews(beaconDB)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetAllBeaconViews, err)
	}

	err = json.Unmarshal(beaconViews, &allViews)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetAllBeaconViews, err)
	}

	sID := []int{}
	for i := 0; i < httpServer.config.ChainParams.ActiveShards; i++ {
		sID = append(sID, i)
	}

	for _, v := range allViews {
		err := v.RestoreBeaconViewStateFromHash(httpServer.GetBlockchain())
		if err != nil {
			continue
		}
		beaconConsensusStateDB, err := statedb.NewWithPrefixTrie(v.ConsensusStateDBRootHash, statedb.NewDatabaseAccessWarper(beaconDB))
		if err != nil {
			continue
		}
		v.AutoStaking = blockchain.NewMapStringBool()

		mapAutoStaking := statedb.GetMapAutoStaking(beaconConsensusStateDB, sID)

		for hash, value := range mapAutoStaking {
			v.AutoStaking.Set(hash, value)
		}
		// finish reproduce
		// if !blockchain.BeaconChain.multiView.AddView(v) {
		// 	continue
		// }
		beaconBestState = v
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
