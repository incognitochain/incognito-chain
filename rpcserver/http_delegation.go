package rpcserver

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
)

func (httpServer *HttpServer) handleGetDelegationDetail(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	shardIDs := []int{}
	for i := 0; i < config.Param().ActiveShards; i++ {
		shardIDs = append(shardIDs, i)
	}
	arrayParams := common.InterfaceSlice(params)
	height := uint64(arrayParams[0].(float64))
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
	_, _, _,
		_, _, _, _,
		_, _, _, beaconDelegate := statedb.GetAllCandidateSubstituteCommittee(stateDB, shardIDs)

	return beaconDelegate, nil
}
