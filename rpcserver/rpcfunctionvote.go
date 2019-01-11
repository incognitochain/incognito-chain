package rpcserver

import (
	"encoding/json"
	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/rpcserver/jsonresult"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/wallet"
)

func (self RpcServer) handleGetAmountVoteToken(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	paymentAddress := arrayParams[0].(string)
	pubKey := wallet.GetPubKeyFromPaymentAddress(paymentAddress)
	db := *self.config.Database
	result := jsonresult.ListCustomTokenBalance{ListCustomTokenBalance: []jsonresult.CustomTokenBalance{}}

	// For DCB voting token
	result.PaymentAddress = paymentAddress
	item := jsonresult.CustomTokenBalance{}
	item.Name = "DCB voting token"
	item.Symbol = "DCB Voting Token"
	TokenID := &common.Hash{}
	TokenID.SetBytes(common.DCBVotingTokenID[:])
	item.TokenID = TokenID.String()
	item.TokenImage = common.Render([]byte(item.TokenID))
	amount, err := db.GetVoteTokenAmount("dcb", self.config.BlockChain.GetCurrentBoardIndex(blockchain.DCBConstitutionHelper{}), pubKey)
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
	amount, err = db.GetVoteTokenAmount("gov", self.config.BlockChain.GetCurrentBoardIndex(blockchain.GOVConstitutionHelper{}), pubKey)
	if err != nil {
		Logger.log.Error(err)
	}
	item.Amount = uint64(amount)
	result.ListCustomTokenBalance = append(result.ListCustomTokenBalance, item)

	return result, nil
}

// ============================== VOTE PROPOSAL

func (self RpcServer) handleGetEncryptionFlag(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	boardType := arrayParams[0].(string)
	db := *self.config.Database
	encryptionFlag, _ := db.GetEncryptFlag(boardType)
	return jsonresult.GetEncryptionFlagResult{encryptionFlag}, nil
}

func (self RpcServer) handleGetEncryptionLastBlockHeightFlag(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	boardType := arrayParams[0].(string)
	db := *self.config.Database
	blockHeight, _ := db.GetEncryptionLastBlockHeight(boardType)
	return jsonresult.GetEncryptionLastBlockHeightResult{blockHeight}, nil
}

func (self RpcServer) buildRawSealLv3VoteProposalTransaction(
	params interface{},
) (*transaction.Tx, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	nParams := len(arrayParams)

	boardType := arrayParams[nParams-3].(string)
	voteInfo := arrayParams[len(arrayParams)-2]
	pubKey := arrayParams[len(arrayParams)-1].([]interface{}) // firstPubKey is pubkey of itself
	Seal3Data := common.Encrypt(common.Encrypt(common.Encrypt(voteInfo, pubKey[0]), pubKey[1]), pubKey[2])

	var meta metadata.Metadata
	if boardType == "dcb" {
		meta = metadata.NewSealedLv3DCBVoteProposalMetadata(Seal3Data, common.SliceInterfaceToSliceSliceByte(pubKey))
	} else {
		meta = metadata.NewSealedLv3GOVVoteProposalMetadata(Seal3Data, common.SliceInterfaceToSliceSliceByte(pubKey))
	}
	tx, err := self.buildRawTransaction(params, meta)
	return tx, err
}

func (self RpcServer) handleCreateRawSealLv3VoteProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	tx, err := self.buildRawSealLv3VoteProposalTransaction(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}

	byteArrays, err1 := json.Marshal(tx)
	if err1 != nil {
		Logger.log.Error(err1)
		return nil, NewRPCError(ErrUnexpected, err1)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

//create lv3 vote by 3 layer encrypt
func (self RpcServer) handleCreateAndSendSealLv3VoteProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawSealLv3VoteProposalTransaction(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		return nil, err
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	txId, err := self.handleSendRawTransaction(newParam, closeChan)
	return txId, err
}

func (self RpcServer) buildRawSealLv2VoteProposalTransaction(
	params interface{},
) (*transaction.Tx, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	nParams := len(arrayParams)

	boardType := arrayParams[nParams-5]

	firstPrivateKey := arrayParams[nParams-4]
	Seal3Data := arrayParams[nParams-3]
	Seal2Data := common.Decrypt(Seal3Data, firstPrivateKey)

	pubKeys := arrayParams[nParams-2].([]interface{})

	Pointer := common.NewHash([]byte(arrayParams[nParams-1].(string)))

	var meta metadata.Metadata
	if boardType == "dcb" {
		meta = metadata.NewSealedLv2DCBVoteProposalMetadata(
			[]byte(Seal2Data.(string)),
			common.SliceInterfaceToSliceSliceByte(pubKeys),
			Pointer,
		)
	} else {
		meta = metadata.NewSealedLv2GOVVoteProposalMetadata(
			[]byte(Seal2Data.(string)),
			common.SliceInterfaceToSliceSliceByte(pubKeys),
			Pointer,
		)
	}
	tx, err := self.buildRawTransaction(params, meta)
	return tx, err
}

func (self RpcServer) handleCreateRawSealLv2VoteProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	tx, err := self.buildRawSealLv2VoteProposalTransaction(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}

	byteArrays, err1 := json.Marshal(tx)
	if err1 != nil {
		Logger.log.Error(err1)
		return nil, NewRPCError(ErrUnexpected, err1)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

//create lv2 vote by decrypt A layer
func (self RpcServer) handleCreateAndSendSealLv2VoteProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawSealLv2VoteProposalTransaction(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		return nil, err
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	txId, err := self.handleSendRawTransaction(newParam, closeChan)
	return txId, err
}

func (self RpcServer) buildRawSealLv1VoteProposalTransaction(
	params interface{},
) (*transaction.Tx, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	nParams := len(arrayParams)

	boardType := arrayParams[nParams-5].(string)

	pointer3 := common.NewHash([]byte(arrayParams[nParams-4].(string)))

	pointer2 := common.NewHash([]byte(arrayParams[nParams-3].(string)))

	Seal2Data := arrayParams[nParams-2]

	pubKeys := arrayParams[nParams-1].([]interface{})
	Seal1Data := common.Decrypt(Seal2Data, pubKeys[1])

	var meta metadata.Metadata
	if boardType == "dcb" {
		meta = metadata.NewSealedLv1DCBVoteProposalMetadata(
			[]byte(Seal1Data.(string)),
			common.SliceInterfaceToSliceSliceByte(pubKeys),
			pointer2,
			pointer3,
		)
	} else {
		meta = metadata.NewSealedLv1GOVVoteProposalMetadata(
			[]byte(Seal1Data.(string)),
			common.SliceInterfaceToSliceSliceByte(pubKeys),
			pointer2,
			pointer3,
		)
	}

	tx, err := self.buildRawTransaction(params, meta)
	return tx, err
}

func (self RpcServer) handleCreateRawSealLv1VoteProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	tx, err := self.buildRawSealLv1VoteProposalTransaction(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}

	byteArrays, err1 := json.Marshal(tx)
	if err1 != nil {
		Logger.log.Error(err1)
		return nil, NewRPCError(ErrUnexpected, err1)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (self RpcServer) handleCreateAndSendSealLv1VoteProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawSealLv1VoteProposalTransaction(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		return nil, err
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	txId, err := self.handleSendRawTransaction(newParam, closeChan)
	return txId, err
}

func (self RpcServer) buildRawNormalVoteProposalTransactionFromSealer(
	params interface{},
) (*transaction.Tx, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	nParams := len(arrayParams)

	boardType := arrayParams[nParams-6].(string)

	pointer3 := common.NewHash([]byte(arrayParams[nParams-5].(string)))

	pointer1 := common.NewHash([]byte(arrayParams[nParams-4].(string)))

	Seal1Data := arrayParams[nParams-3]

	pubKeys := arrayParams[nParams-2].([]interface{})

	thirdPrivateKey := arrayParams[nParams-1]

	normalVoteProposal := common.Decrypt(Seal1Data, thirdPrivateKey)
	var meta metadata.Metadata
	if boardType == "dcb" {
		meta = metadata.NewNormalDCBVoteProposalFromSealerMetadata(
			normalVoteProposal.([]byte),
			common.SliceInterfaceToSliceSliceByte(pubKeys),
			pointer1,
			pointer3,
		)
	} else {
		meta = metadata.NewNormalGOVVoteProposalFromSealerMetadata(
			normalVoteProposal.([]byte),
			common.SliceInterfaceToSliceSliceByte(pubKeys),
			pointer1,
			pointer3,
		)
	}
	tx, err := self.buildRawTransaction(params, meta)
	return tx, err
}

func (self RpcServer) handleCreateRawNormalVoteProposalTransactionFromSealer(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	tx, err := self.buildRawNormalVoteProposalTransactionFromSealer(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}

	byteArrays, err1 := json.Marshal(tx)
	if err1 != nil {
		Logger.log.Error(err1)
		return nil, NewRPCError(ErrUnexpected, err1)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (self RpcServer) handleCreateAndSendNormalVoteProposalTransactionFromSealer(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawNormalVoteProposalTransactionFromSealer(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		return nil, err
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	txId, err := self.handleSendRawTransaction(newParam, closeChan)
	return txId, err
}

func (self RpcServer) buildRawNormalVoteProposalTransactionFromOwner(
	params interface{},
) (*transaction.Tx, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	nParams := len(arrayParams)

	boardType := arrayParams[nParams-4].(string)

	pointer := common.NewHash([]byte(arrayParams[nParams-3].(string)))

	pubKeys := arrayParams[nParams-2].([]interface{})

	normalVoteProposal := arrayParams[nParams-1]

	var meta metadata.Metadata
	if boardType == "dcb" {
		meta = metadata.NewNormalDCBVoteProposalFromOwnerMetadata(
			normalVoteProposal.([]byte),
			common.SliceInterfaceToSliceSliceByte(pubKeys),
			pointer,
		)
	} else {
		meta = metadata.NewNormalGOVVoteProposalFromOwnerMetadata(
			normalVoteProposal.([]byte),
			common.SliceInterfaceToSliceSliceByte(pubKeys),
			pointer,
		)
	}
	tx, err := self.buildRawTransaction(params, meta)
	return tx, err
}

func (self RpcServer) handleCreateRawNormalVoteProposalTransactionFromOwner(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	tx, err1 := self.buildRawNormalVoteProposalTransactionFromOwner(params)
	if err1 != nil {
		Logger.log.Error(err1)
		return nil, NewRPCError(ErrUnexpected, err1)
	}

	byteArrays, err := json.Marshal(tx)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (self RpcServer) handleCreateAndSendNormalVoteProposalTransactionFromOwner(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawNormalVoteProposalTransactionFromOwner(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		return nil, err
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	txId, err := self.handleSendRawTransaction(newParam, closeChan)
	return txId, err
}
