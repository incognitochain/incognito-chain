package rpcserver

import (
	"errors"
	"fmt"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
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

func (rpcServer RpcServer) handleCreateRawSealLv3VoteProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	params, err := rpcServer.buildParamsSealLv3VoteProposal(params)
	if err != nil {
		return nil, err
	}
	return rpcServer.createRawTxWithMetadata(
		params,
		closeChan,
		metadata.NewSealedLv3VoteProposalMetadataFromRPC,
	)
}

//create lv3 vote by 3 layer encrypt
func (rpcServer RpcServer) handleCreateAndSendSealLv3VoteProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.createAndSendTxWithMetadata(
		params,
		closeChan,
		RpcServer.handleCreateRawSealLv3VoteProposalTransaction,
		RpcServer.handleSendRawTransaction,
	)
}

func GetLockerPaymentAddresses(tx metadata.Transaction) []privacy.PaymentAddress {
	meta := tx.GetMetadata()
	if meta.GetType() == metadata.SealedLv3DCBVoteProposalMeta {
		newMeta := meta.(*metadata.SealedLv3DCBVoteProposalMetadata)
		return newMeta.SealedLv3VoteProposalMetadata.SealedVoteProposal.LockerPaymentAddresses
	} else {
		newMeta := meta.(*metadata.SealedLv3GOVVoteProposalMetadata)
		return newMeta.SealedLv3VoteProposalMetadata.SealedVoteProposal.LockerPaymentAddresses
	}
}

func GetSealLv3Data(tx metadata.Transaction) ([]byte, error) {
	meta := tx.GetMetadata()
	if meta.GetType() == metadata.SealedLv3DCBVoteProposalMeta {
		newMeta := meta.(*metadata.SealedLv3DCBVoteProposalMetadata)
		return newMeta.SealedLv3VoteProposalMetadata.SealedVoteProposal.SealVoteProposalData, nil
	} else if meta.GetType() == metadata.SealedLv3GOVVoteProposalMeta {
		newMeta := meta.(*metadata.SealedLv3GOVVoteProposalMetadata)
		return newMeta.SealedLv3VoteProposalMetadata.SealedVoteProposal.SealVoteProposalData, nil
	}
	return nil, errors.New("wrong type")
}

//Input metadataParam: {
//	Lv3TxID: string,
//	FirstPrivateKey: string,
//}
func (rpcServer RpcServer) handleCreateRawSealLv2VoteProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	params, err := rpcServer.buildParamsSealLv2VoteProposal(params)
	if err != nil {
		return nil, err
	}

	return rpcServer.createRawTxWithMetadata(
		params,
		closeChan,
		metadata.NewSealedLv2VoteProposalMetadataFromRPC,
	)
}

//create lv2 vote by decrypt A layer
func (rpcServer RpcServer) handleCreateAndSendSealLv2VoteProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.createAndSendTxWithMetadata(
		params,
		closeChan,
		RpcServer.handleCreateRawSealLv2VoteProposalTransaction,
		RpcServer.handleSendRawTransaction,
	)
}

func GetSealLv2Data(lv2tx metadata.Transaction) ([]byte, error) {
	meta := lv2tx.GetMetadata()
	if meta.GetType() == metadata.SealedLv2DCBVoteProposalMeta {
		newMeta := meta.(*metadata.SealedLv2DCBVoteProposalMetadata)
		return newMeta.SealedLv2VoteProposalMetadata.SealedVoteProposal.SealVoteProposalData, nil
	} else if meta.GetType() == metadata.SealedLv2GOVVoteProposalMeta {
		newMeta := meta.(*metadata.SealedLv2GOVVoteProposalMetadata)
		return newMeta.SealedLv2VoteProposalMetadata.SealedVoteProposal.SealVoteProposalData, nil
	}
	return nil, errors.New("wrong type")
}

func GetSealLv1Data(lv1tx metadata.Transaction) ([]byte, error) {
	meta := lv1tx.GetMetadata()
	if meta.GetType() == metadata.SealedLv1DCBVoteProposalMeta {
		newMeta := meta.(*metadata.SealedLv1DCBVoteProposalMetadata)
		return newMeta.SealedLv1VoteProposalMetadata.SealedVoteProposal.SealVoteProposalData, nil
	} else if meta.GetType() == metadata.SealedLv1GOVVoteProposalMeta {
		newMeta := meta.(*metadata.SealedLv1GOVVoteProposalMetadata)
		return newMeta.SealedLv1VoteProposalMetadata.SealedVoteProposal.SealVoteProposalData, nil
	}
	return nil, errors.New("wrong type")
}

func (rpcServer RpcServer) handleCreateRawSealLv1VoteProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	params, err := rpcServer.buildParamsSealLv1VoteProposal(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	return rpcServer.createRawTxWithMetadata(
		params,
		closeChan,
		metadata.NewSealedLv1VoteProposalMetadataFromRPC,
	)
}

func (rpcServer RpcServer) handleCreateAndSendSealLv1VoteProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.createAndSendTxWithMetadata(
		params,
		closeChan,
		RpcServer.handleCreateRawSealLv1VoteProposalTransaction,
		RpcServer.handleSendRawTransaction,
	)
}

func (rpcServer RpcServer) handleCreateRawNormalVoteProposalTransactionFromOwner(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	params, err := rpcServer.buildParamsNormalVoteProposalFromOwner(params)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	return rpcServer.createRawTxWithMetadata(
		params,
		closeChan,
		metadata.NewNormalVoteProposalFromOwnerMetadataFromRPC,
	)
}

func (rpcServer RpcServer) handleCreateAndSendNormalVoteProposalFromOwnerTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.createAndSendTxWithMetadata(
		params,
		closeChan,
		RpcServer.handleCreateRawNormalVoteProposalTransactionFromOwner,
		RpcServer.handleSendRawTransaction,
	)
}

func (rpcServer RpcServer) handleCreateRawNormalVoteProposalTransactionFromSealer(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	params, err := rpcServer.buildParamsNormalVoteProposalFromSealer(params)
	if err != nil {
		return nil, err
	}
	return rpcServer.createRawTxWithMetadata(
		params,
		closeChan,
		metadata.NewNormalVoteProposalFromSealerMetadataFromRPC,
	)
}

func (rpcServer RpcServer) handleCreateAndSendNormalVoteProposalFromSealerTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.createAndSendTxWithMetadata(
		params,
		closeChan,
		RpcServer.handleCreateRawNormalVoteProposalTransactionFromSealer,
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
