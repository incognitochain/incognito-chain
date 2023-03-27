package rpcserver

import (
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/blockchain/committeestate"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
)

/*
handleGetBeaconBestState - RPC get beacon best state
*/
func (httpServer *HttpServer) handleGetBeaconBestState(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	result := jsonresult.NewGetBeaconBestState(httpServer.blockService.BlockChain.GetBeaconBestState())
	return result, nil
}

/*
handleGetBeaconViewByHash
*/
func (httpServer *HttpServer) handleGetBeaconViewByHash(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Not enough param"))
	}
	blockHashStr, ok := arrayParams[0].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Block Hash is invalid"))
	}
	blockHash, err := common.Hash{}.NewHashFromStr(blockHashStr)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Block Hash format is invalid"))
	}
	bView, err := httpServer.blockService.BlockChain.GetBeaconViewStateDataFromBlockHash(*blockHash, true, false, false)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetBeaconViewByBlockHashError, err)
	}
	return jsonresult.NewGetBeaconBestState(bView), nil

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
	result := jsonresult.NewGetShardBestState(httpServer.blockService.BlockChain.GetBestStateShard(shardID))
	return result, nil
}

// handleGetCandidateList - return list candidate of committee
func (httpServer *HttpServer) handleGetCandidateList(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {

	beacon, err := httpServer.blockService.GetBeaconBestState()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetClonedBeaconBestStateError, err)
	}

	CSWFCR := beacon.GetCandidateShardWaitingForCurrentRandom()
	CSWFNR := beacon.GetCandidateShardWaitingForNextRandom()
	CBWFCR := beacon.GetCandidateBeaconWaitingForCurrentRandom()
	CBWFNR := beacon.GetCandidateBeaconWaitingForNextRandom()
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

	beaconCommittee := clonedBeaconBestState.GetBeaconCommittee()
	beaconPendingValidator := clonedBeaconBestState.GetBeaconPendingValidator()
	shardCommittee := clonedBeaconBestState.GetShardCommittee()
	shardPendingValidator := clonedBeaconBestState.GetShardPendingValidator()
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

func (httpServer *HttpServer) handleGetTotalStaker(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	total, err := httpServer.config.BlockChain.GetTotalStaker()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetTotalStakerError, err)
	}
	result := jsonresult.NewGetTotalStaker(total)
	return result, nil
}

func (httpServer *HttpServer) handleGetConnectionStatus(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return httpServer.blockService.BlockChain.GetConfig().Highway.GetConnectionStatus(), nil
}

func (httpServer *HttpServer) handleGetBeaconStakerInfo(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	height := uint64(arrayParams[0].(float64))
	stakerPubkey := arrayParams[1].(string)

	beaconConsensusStateRootHash, err := httpServer.config.BlockChain.GetBeaconRootsHashFromBlockHeight(
		height,
	)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	stateDB, err := statedb.NewWithPrefixTrie(beaconConsensusStateRootHash.ConsensusStateDBRootHash,
		statedb.NewDatabaseAccessWarper(httpServer.config.BlockChain.GetBeaconChainDatabase()))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	res, found, err := statedb.GetBeaconStakerInfo(stateDB, stakerPubkey)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	if !found {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, fmt.Errorf("Not found"))
	}
	return res, nil
}

func (httpServer *HttpServer) handleGetBeaconCandidateUID(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	stakerPubkey := arrayParams[0].(string)
	bcBestState := httpServer.GetBlockchain().GetBeaconBestState()
	uid, err := bcBestState.GetBeaconCandidateUID(stakerPubkey)

	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	return uid, nil

}

func (httpServer *HttpServer) handleGetShardStakerInfo(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	height := uint64(arrayParams[0].(float64))
	stakerPubkey := arrayParams[1].(string)
	stateDB := httpServer.config.BlockChain.GetBeaconBestState().GetBeaconConsensusStateDB()
	if height != 0 {
		beaconConsensusStateRootHash, err := httpServer.config.BlockChain.GetBeaconRootsHashFromBlockHeight(
			height,
		)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
		}
		stateDB, err = statedb.NewWithPrefixTrie(beaconConsensusStateRootHash.ConsensusStateDBRootHash,
			statedb.NewDatabaseAccessWarper(httpServer.config.BlockChain.GetBeaconChainDatabase()))
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
		}
	}
	res, found, err := statedb.GetShardStakerInfo(stateDB, stakerPubkey)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	if !found {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, fmt.Errorf("Not found"))
	}
	return res, nil
}

func (httpServer *HttpServer) handleGetBeaconCommitteeState(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	height := uint64(arrayParams[0].(float64))
	beacon_stakingflowv4_enable_height := httpServer.config.BlockChain.GetBeaconBestState().TriggeredFeature[blockchain.BEACON_STAKING_FLOW_V4]
	if beacon_stakingflowv4_enable_height == 0 {
		return nil, nil
	} else {
		if height != 0 && height < beacon_stakingflowv4_enable_height {
			return nil, nil
		}
	}

	if height == 0 {
		return httpServer.config.BlockChain.GetBeaconBestState().GetBeaconCommitteeState().(*committeestate.BeaconCommitteeStateV4).DebugBeaconCommitteeState(), nil
	}
	beaconConsensusStateRootHash, err := httpServer.config.BlockChain.GetBeaconRootsHashFromBlockHeight(
		height,
	)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	stateDB, err := statedb.NewWithPrefixTrie(beaconConsensusStateRootHash.ConsensusStateDBRootHash,
		statedb.NewDatabaseAccessWarper(httpServer.config.BlockChain.GetBeaconChainDatabase()))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	allBeaconBlockInEpoch := []types.BeaconBlock{}
	currentBlock, _ := httpServer.config.BlockChain.GetBeaconBlockByHeight(height)
	tempBeaconBlock := *currentBlock[0]
	tempBeaconHeight := tempBeaconBlock.GetBeaconHeight()
	firstBeaconHeightOfEpoch := blockchain.GetFirstBeaconHeightInEpoch(tempBeaconBlock.GetCurrentEpoch())
	for tempBeaconHeight > firstBeaconHeightOfEpoch { //dont need to get the first block, we dont count prevValidation data for this block
		allBeaconBlockInEpoch = append([]types.BeaconBlock{tempBeaconBlock}, allBeaconBlockInEpoch...)
		previousBeaconBlock, _, err := httpServer.config.BlockChain.GetBeaconBlockByHash(tempBeaconBlock.Header.PreviousBlockHash)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
		}
		tempBeaconBlock = *previousBeaconBlock
		tempBeaconHeight = tempBeaconBlock.GetBeaconHeight()
	}

	stateV4 := committeestate.NewBeaconCommitteeStateV4()

	stateV4.RestoreBeaconCommitteeFromDB(stateDB, httpServer.config.BlockChain.GetBeaconBestState().MinBeaconCommitteeSize, allBeaconBlockInEpoch)

	return stateV4.DebugBeaconCommitteeState(), nil
}
