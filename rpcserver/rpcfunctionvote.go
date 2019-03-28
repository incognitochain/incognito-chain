package rpcserver

import (
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/metadata"
)

func iPlusPlus(x *int) int {
	*x += 1
	return *x - 1
}

// ============================== VOTE PROPOSAL

func (rpcServer RpcServer) handleCreateRawVoteProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	//VoteProposal - Step 2: Create Raw vote proposal transaction
	params = setBuildRawBurnTransactionParams(params, FeeVote)
	return rpcServer.createRawTxWithMetadata(
		params,
		closeChan,
		metadata.NewVoteProposalMetadataFromRPC,
	)
}

func (rpcServer RpcServer) handleCreateAndSendVoteProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	//VoteProposal - Step 1: Client call rpc function to create vote proposal transaction
	return rpcServer.createAndSendTxWithMetadata(
		params,
		closeChan,
		RpcServer.handleCreateRawVoteProposalTransaction,
		RpcServer.handleSendRawTransaction,
	)
}

func (rpcServer RpcServer) handleGetDCBBoardIndex(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.config.BlockChain.BestState.Beacon.StabilityInfo.DCBGovernor.BoardIndex, nil
}
func (rpcServer RpcServer) handleGetGOVBoardIndex(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.config.BlockChain.BestState.Beacon.StabilityInfo.GOVGovernor.BoardIndex, nil
}

func setBuildRawBurnTransactionParams(params interface{}, fee float64) interface{} {
	arrayParams := common.InterfaceSlice(params)
	x := make(map[string]interface{})
	x[common.BurningAddress] = fee
	arrayParams[1] = x
	return arrayParams
}
