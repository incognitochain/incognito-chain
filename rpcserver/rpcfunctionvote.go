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

	meta := metadata.Metadata()
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

func (self RpcServer) handleCreateRawSealLv2VoteDCBProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	tx, err := self.buildRawSealLv2VoteDCBProposalTransaction(params)
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
func (self RpcServer) handleCreateAndSendSealLv2VoteDCBProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawSealLv2VoteDCBProposalTransaction(params, closeChan)
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

func (self RpcServer) buildRawSealLv1VoteDCBProposalTransaction(
	params interface{},
) (*transaction.Tx, *RPCError) {
	arrayParams := common.InterfaceSlice(params)

	PointerToLv3VoteProposal := arrayParams[len(arrayParams)-7]
	Pointer3 := common.Hash{}
	copy(Pointer3[:], []byte(PointerToLv3VoteProposal.(string)))

	PointerToLv2VoteProposal := arrayParams[len(arrayParams)-6]
	Pointer2 := common.Hash{}
	copy(Pointer2[:], []byte(PointerToLv2VoteProposal.(string)))

	Seal2Data := arrayParams[len(arrayParams)-5]
	firstPubKey := arrayParams[len(arrayParams)-4]
	secondPubKey := arrayParams[len(arrayParams)-3]
	thirdPubKey := arrayParams[len(arrayParams)-2]
	secondPrivateKey := arrayParams[len(arrayParams)-1]
	Seal1Data := common.Decrypt(Seal2Data, secondPrivateKey)
	meta := metadata.NewSealedLv1DCBVoteProposalMetadata([]byte(Seal1Data.(string)), [][]byte{[]byte(firstPubKey.(string)), []byte(secondPubKey.(string)), []byte(thirdPubKey.(string))}, Pointer2, Pointer3)
	tx, err := self.buildRawTransaction(params, meta)
	return tx, err
}

func (self RpcServer) handleCreateRawSealLv1VoteDCBProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	tx, err := self.buildRawSealLv1VoteDCBProposalTransaction(params)
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

func (self RpcServer) handleCreateAndSendSealLv1VoteDCBProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawSealLv1VoteDCBProposalTransaction(params, closeChan)
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

func (self RpcServer) buildRawNormalVoteDCBProposalTransactionFromSealer(
	params interface{},
) (*transaction.Tx, *RPCError) {
	arrayParams := common.InterfaceSlice(params)

	PointerToLv3VoteProposal := arrayParams[len(arrayParams)-7]
	Pointer3 := common.Hash{}
	copy(Pointer3[:], []byte(PointerToLv3VoteProposal.(string)))

	PointerToLv1VoteProposal := arrayParams[len(arrayParams)-6]
	Pointer1 := common.Hash{}
	copy(Pointer1[:], []byte(PointerToLv1VoteProposal.(string)))

	Seal1Data := arrayParams[len(arrayParams)-5]
	firstPubKey := arrayParams[len(arrayParams)-4]
	secondPubKey := arrayParams[len(arrayParams)-3]
	thirdPubKey := arrayParams[len(arrayParams)-2]
	thirdPrivateKey := arrayParams[len(arrayParams)-1]
	normalVoteProposal := common.Decrypt(Seal1Data, thirdPrivateKey)
	meta := metadata.NewNormalDCBVoteProposalFromSealerMetadata(normalVoteProposal.([]byte), [][]byte{firstPubKey.([]byte), secondPubKey.([]byte), thirdPubKey.([]byte)}, Pointer1, Pointer3)
	tx, err := self.buildRawTransaction(params, meta)
	return tx, err
}

func (self RpcServer) handleCreateRawNormalVoteDCBProposalTransactionFromSealer(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	tx, err := self.buildRawNormalVoteDCBProposalTransactionFromSealer(params)
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

func (self RpcServer) handleCreateAndSendNormalVoteDCBProposalTransactionFromSealer(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawNormalVoteDCBProposalTransactionFromSealer(params, closeChan)
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

func (self RpcServer) buildRawNormalVoteDCBProposalTransactionFromOwner(
	params interface{},
) (*transaction.Tx, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	PointerToLv3VoteProposal := arrayParams[len(arrayParams)-5]
	Pointer := common.Hash{}
	copy(Pointer[:], []byte(PointerToLv3VoteProposal.(string)))
	firstPubKey := arrayParams[len(arrayParams)-4]
	secondPubKey := arrayParams[len(arrayParams)-3]
	thirdPubKey := arrayParams[len(arrayParams)-2]
	normalVoteProposal := arrayParams[len(arrayParams)-1]
	meta := metadata.NewNormalDCBVoteProposalFromOwnerMetadata(normalVoteProposal.([]byte), [][]byte{[]byte(firstPubKey.(string)), []byte(secondPubKey.(string)), []byte(thirdPubKey.(string))}, Pointer)
	tx, err := self.buildRawTransaction(params, meta)
	return tx, err
}

func (self RpcServer) handleCreateRawNormalVoteDCBProposalTransactionFromOwner(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	tx, err1 := self.buildRawNormalVoteDCBProposalTransactionFromOwner(params)
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

func (self RpcServer) handleCreateAndSendNormalVoteDCBProposalTransactionFromOwner(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawNormalVoteDCBProposalTransactionFromOwner(params, closeChan)
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
