package rpcserver

import (
	"errors"
	"fmt"
	"github.com/ninjadotorg/constant/database"

	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/rpcserver/jsonresult"
)

func iPlusPlus(x *int) int {
	*x += 1
	return *x - 1
}

func (rpcServer RpcServer) handleGetAmountVoteToken(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	paymentAddressData := arrayParams[0].(string)
	paymentAddress, err := metadata.GetPaymentAddressFromSenderKeyParams(paymentAddressData)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	db := *rpcServer.config.Database
	result := jsonresult.ListCustomTokenBalance{ListCustomTokenBalance: []jsonresult.CustomTokenBalance{}}

	// For DCB voting token
	result.PaymentAddress = string(paymentAddressData)
	item := jsonresult.CustomTokenBalance{}
	item.Name = "DCB voting token"
	item.Symbol = "DCB Voting Token"
	TokenID := &common.Hash{}
	TokenID.SetBytes(common.DCBVotingTokenID[:])
	item.TokenID = TokenID.String()
	item.TokenImage = common.Render([]byte(item.TokenID))
	amount, err := db.GetVoteTokenAmount(metadata.DCBBoard.BoardTypeDB(), rpcServer.config.BlockChain.GetCurrentBoardIndex(blockchain.DCBConstitutionHelper{}), *paymentAddress)
	if err != nil {
		Logger.log.Error(err)
	}
	item.Amount = uint64(amount)
	result.ListCustomTokenBalance = append(result.ListCustomTokenBalance, item)

	// For GOV voting token
	item = jsonresult.CustomTokenBalance{}
	item.Name = "GOV voting token"
	item.Symbol = "GOV Voting Token"
	TokenID = &common.Hash{}
	TokenID.SetBytes(common.GOVVotingTokenID[:])
	item.TokenID = TokenID.String()
	item.TokenImage = common.Render([]byte(item.TokenID))
	amount, err = db.GetVoteTokenAmount(metadata.GOVBoard.BoardTypeDB(), rpcServer.config.BlockChain.GetCurrentBoardIndex(blockchain.GOVConstitutionHelper{}), *paymentAddress)
	if err != nil {
		Logger.log.Error(err)
	}
	item.Amount = uint64(amount)
	result.ListCustomTokenBalance = append(result.ListCustomTokenBalance, item)

	return result, nil
}

func (rpcServer RpcServer) handleSetAmountVoteToken(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	paymentAddressSenderKey := arrayParams[0].(string)
	paymentAddress, err1 := metadata.GetPaymentAddressFromSenderKeyParams(paymentAddressSenderKey)
	if err1 != nil {
		return nil, NewRPCError(ErrUnexpected, err1)
	}
	p2, _ := metadata.GetPaymentAddressFromSenderKeyParams(string(paymentAddressSenderKey))
	_ = p2
	db := *rpcServer.config.Database

	amountDCBVote := uint32(arrayParams[1].(float64))
	amountGOVVote := uint32(arrayParams[2].(float64))

	err := db.SetVoteTokenAmount(metadata.DCBBoard.BoardTypeDB(), rpcServer.config.BlockChain.GetCurrentBoardIndex(blockchain.DCBConstitutionHelper{}), *paymentAddress, amountDCBVote)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	err = db.SetVoteTokenAmount(metadata.GOVBoard.BoardTypeDB(), rpcServer.config.BlockChain.GetCurrentBoardIndex(blockchain.DCBConstitutionHelper{}), *paymentAddress, amountGOVVote)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	return nil, nil
}

// ============================== VOTE PROPOSAL

func (rpcServer RpcServer) handleGetEncryptionFlag(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	db := *rpcServer.config.Database
	dcbEncryptionFlag, _ := db.GetEncryptFlag(metadata.DCBBoard.BoardTypeDB())
	govEncryptionFlag, _ := db.GetEncryptFlag(metadata.GOVBoard.BoardTypeDB())
	return jsonresult.GetEncryptionFlagResult{
		DCBFlag: dcbEncryptionFlag,
		GOVFlag: govEncryptionFlag,
	}, nil
}

func (rpcServer RpcServer) handleSetEncryptionFlag(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	fmt.Print("delete me, use only for test purpose!!!")
	db := *rpcServer.config.Database
	dcbEncryptionFlag, _ := db.GetEncryptFlag(metadata.DCBBoard.BoardTypeDB())
	govEncryptionFlag, _ := db.GetEncryptFlag(metadata.GOVBoard.BoardTypeDB())
	db.SetEncryptFlag(metadata.DCBBoard.BoardTypeDB(), (dcbEncryptionFlag+1)%4)
	db.SetEncryptFlag(metadata.GOVBoard.BoardTypeDB(), (govEncryptionFlag+1)%4)
	return dcbEncryptionFlag, nil
}

func (rpcServer RpcServer) handleGetEncryptionLastBlockHeightFlag(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	boardType := database.BoardTypeDB(arrayParams[0].([]byte)[0])
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

func setBuildRawBurnTransactionParams(params interface{}, fee float64) interface{} {
	arrayParams := common.InterfaceSlice(params)
	x := make(map[string]interface{})
	x[common.BurningAddress] = fee
	arrayParams[1] = x
	return arrayParams
}
