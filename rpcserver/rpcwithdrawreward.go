package rpcserver

import (
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/metadata"
)

func (rpcServer RpcServer) handleCreateRawWithDrawTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	//VoteProposal - Step 2: Create Raw vote proposal transaction
	// params = setBuildRawBurnTransactionParams(params, FeeVote)
	arrayParams := common.InterfaceSlice(params)
	arrayParams[1] = nil
	return rpcServer.createRawTxWithMetadata(
		arrayParams,
		closeChan,
		metadata.NewWithDrawRewardRequestFromRPC,
	)
}

func (rpcServer RpcServer) handleCreateAndSendWithDrawTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	//VoteProposal - Step 1: Client call rpc function to create vote proposal transaction
	return rpcServer.createAndSendTxWithMetadata(
		params,
		closeChan,
		RpcServer.handleCreateRawWithDrawTransaction,
		RpcServer.handleSendRawTransaction,
	)
}
