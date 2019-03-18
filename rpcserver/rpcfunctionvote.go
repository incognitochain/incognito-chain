package rpcserver

import (
	"fmt"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/rpcserver/jsonresult"
)

func iPlusPlus(x *int) int {
	*x += 1
	return *x - 1
}

// ============================== VOTE PROPOSAL

func (rpcServer RpcServer) handleGetEncryptionFlag(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	db := *rpcServer.config.Database
	dcbEncryptionFlag, _ := db.GetEncryptFlag(common.DCBBoard)
	govEncryptionFlag, _ := db.GetEncryptFlag(common.GOVBoard)
	return jsonresult.GetEncryptionFlagResult{
		DCBFlag: dcbEncryptionFlag,
		GOVFlag: govEncryptionFlag,
	}, nil
}

func (rpcServer RpcServer) handleSetEncryptionFlag(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	fmt.Print("delete me, use only for test purpose!!!")
	db := *rpcServer.config.Database
	dcbEncryptionFlag, _ := db.GetEncryptFlag(common.DCBBoard)
	govEncryptionFlag, _ := db.GetEncryptFlag(common.GOVBoard)
	db.SetEncryptFlag(common.DCBBoard, (dcbEncryptionFlag+1)%4)
	db.SetEncryptFlag(common.GOVBoard, (govEncryptionFlag+1)%4)
	return dcbEncryptionFlag, nil
}

func (rpcServer RpcServer) handleGetEncryptionLastBlockHeightFlag(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	boardType := common.NewBoardTypeFromString(arrayParams[0].(string))
	db := *rpcServer.config.Database
	blockHeight, _ := db.GetEncryptionLastBlockHeight(boardType)
	return jsonresult.GetEncryptionLastBlockHeightResult{blockHeight}, nil
}

func (rpcServer RpcServer) handleCreateRawVoteProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	//VoteProposal - Step 2: Create Raw vote proposal transaction
	params, err := rpcServer.buildParamsVoteProposal(params)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	return rpcServer.createRawTxWithMetadata(
		params,
		closeChan,
		metadata.NewNormalVoteProposalMetadataFromRPC,
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
