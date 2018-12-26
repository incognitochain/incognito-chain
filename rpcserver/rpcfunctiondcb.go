package rpcserver

import (
	"encoding/json"

	params2 "github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/rpcserver/jsonresult"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/wallet"
	"github.com/ninjadotorg/constant/wire"
)

// handleGetDCBParams - get dcb params
func (self RpcServer) handleGetDCBParams(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	dcbParam := self.config.BlockChain.BestState[0].BestBlock.Header.DCBConstitution.DCBParams
	return dcbParam, nil
}

// handleGetDCBConstitution - get dcb constitution
func (self RpcServer) handleGetDCBConstitution(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	constitution := self.config.BlockChain.BestState[0].BestBlock.Header.DCBConstitution
	return constitution, nil
}

// handleGetListDCBBoard - return list payment address of DCB board
func (self RpcServer) handleGetListDCBBoard(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return self.config.BlockChain.BestState[0].BestBlock.Header.DCBGovernor.DCBBoardPubKeys, nil
}

func (self RpcServer) handleCreateRawTxWithIssuingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	var err error
	arrayParams := common.InterfaceSlice(params)
	normalTx, err := self.buildRawTransaction(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	// Req param #4: issuing request info
	issuingReq := arrayParams[4].(map[string]interface{})
	depositedAmount := uint64(issuingReq["depositedAmount"].(float64))
	assetTypeBytes := []byte(issuingReq["assetType"].(string))
	assetType := common.Hash{}
	copy(assetType[:], assetTypeBytes)
	metaType := metadata.IssuingRequestMeta
	receiverAddressMap := issuingReq["receiverAddress"].(map[string]interface{})
	receiverAddress := privacy.PaymentAddress{
		Pk: []byte(receiverAddressMap["pk"].(string)),
		Tk: []byte(receiverAddressMap["tk"].(string)),
	}

	normalTx.Metadata = metadata.NewIssuingRequest(
		receiverAddress,
		depositedAmount,
		assetType,
		metaType,
	)

	byteArrays, err := json.Marshal(normalTx)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            normalTx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (self RpcServer) handleCreateAndSendTxWithIssuingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawTxWithIssuingRequest(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err := self.handleSendRawTransaction(newParam, closeChan)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID: sendResult.(jsonresult.CreateTransactionResult).TxID,
	}
	return result, nil
}

func (self RpcServer) handleCreateRawTxWithContractingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	var err error
	normalTx, err := self.buildRawTransaction(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	metaType := metadata.ContractingRequestMeta
	normalTx.Metadata = metadata.NewContractingRequest(metaType)
	byteArrays, err := json.Marshal(normalTx)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            normalTx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (self RpcServer) handleCreateAndSendTxWithContractingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawTxWithContractingRequest(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err := self.handleSendRawTransaction(newParam, closeChan)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID: sendResult.(jsonresult.CreateTransactionResult).TxID,
	}
	return result, nil
}

func (self RpcServer) buildRawSealLv3VoteDCBProposalTransaction(
	params interface{},
) (*transaction.Tx, *RPCError) {
	tx, err := self.buildRawTransaction(params)
	arrayParams := common.InterfaceSlice(params)
	voteInfo := arrayParams[len(arrayParams)-4]
	firstPubKey := arrayParams[len(arrayParams)-3] // firstPubKey is pubkey of itself
	secondPubKey := arrayParams[len(arrayParams)-2]
	thirdPubKey := arrayParams[len(arrayParams)-1]
	Seal3Data := common.Encrypt(common.Encrypt(common.Encrypt(voteInfo, thirdPubKey), secondPubKey), firstPubKey)
	tx.Metadata = metadata.NewSealedLv3DCBBallotMetadata([]byte(Seal3Data.(string)), [][]byte{[]byte(firstPubKey.(string)), []byte(secondPubKey.(string)), []byte(thirdPubKey.(string))})
	return tx, err
}

func (self RpcServer) handleCreateRawSealLv3VoteDCBProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	var err error
	tx, err := self.buildRawSealLv3VoteDCBProposalTransaction(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
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

//create lv3 vote by 3 layer encrypt
func (self RpcServer) handleCreateAndSendSealLv3VoteDCBProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawSealLv3VoteDCBProposalTransaction(params, closeChan)
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

func (self RpcServer) buildRawSealLv2VoteDCBProposalTransaction(
	params interface{},
) (*transaction.Tx, *RPCError) {
	tx, err := self.buildRawTransaction(params)
	arrayParams := common.InterfaceSlice(params)
	PointerToLv3Ballot := arrayParams[len(arrayParams)-6]
	Seal3Data := arrayParams[len(arrayParams)-5]
	firstPubKey := arrayParams[len(arrayParams)-4]
	secondPubKey := arrayParams[len(arrayParams)-3]
	thirdPubKey := arrayParams[len(arrayParams)-2]
	firstPrivateKey := arrayParams[len(arrayParams)-1]
	Seal2Data := common.Decrypt(Seal3Data, firstPrivateKey)
	Pointer := common.Hash{}
	copy(Pointer[:], []byte(PointerToLv3Ballot.(string)))
	tx.Metadata = metadata.NewSealedLv2DCBBallotMetadata([]byte(Seal2Data.(string)), [][]byte{[]byte(firstPubKey.(string)), []byte(secondPubKey.(string)), []byte(thirdPubKey.(string))}, Pointer)
	return tx, err
}

func (self RpcServer) handleCreateRawSealLv2VoteDCBProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	var err error
	tx, err := self.buildRawSealLv2VoteDCBProposalTransaction(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
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
	tx, err := self.buildRawTransaction(params)
	arrayParams := common.InterfaceSlice(params)

	PointerToLv3Ballot := arrayParams[len(arrayParams)-7]
	Pointer3 := common.Hash{}
	copy(Pointer3[:], []byte(PointerToLv3Ballot.(string)))

	PointerToLv2Ballot := arrayParams[len(arrayParams)-6]
	Pointer2 := common.Hash{}
	copy(Pointer2[:], []byte(PointerToLv2Ballot.(string)))

	Seal2Data := arrayParams[len(arrayParams)-5]
	firstPubKey := arrayParams[len(arrayParams)-4]
	secondPubKey := arrayParams[len(arrayParams)-3]
	thirdPubKey := arrayParams[len(arrayParams)-2]
	secondPrivateKey := arrayParams[len(arrayParams)-1]
	Seal1Data := common.Decrypt(Seal2Data, secondPrivateKey)
	tx.Metadata = metadata.NewSealedLv1DCBBallotMetadata([]byte(Seal1Data.(string)), [][]byte{[]byte(firstPubKey.(string)), []byte(secondPubKey.(string)), []byte(thirdPubKey.(string))}, Pointer2, Pointer3)
	return tx, err
}

func (self RpcServer) handleCreateRawSealLv1VoteDCBProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	var err error
	tx, err := self.buildRawSealLv1VoteDCBProposalTransaction(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
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
	tx, err := self.buildRawTransaction(params)
	arrayParams := common.InterfaceSlice(params)

	PointerToLv3Ballot := arrayParams[len(arrayParams)-7]
	Pointer3 := common.Hash{}
	copy(Pointer3[:], []byte(PointerToLv3Ballot.(string)))

	PointerToLv1Ballot := arrayParams[len(arrayParams)-6]
	Pointer1 := common.Hash{}
	copy(Pointer1[:], []byte(PointerToLv1Ballot.(string)))

	Seal1Data := arrayParams[len(arrayParams)-5]
	firstPubKey := arrayParams[len(arrayParams)-4]
	secondPubKey := arrayParams[len(arrayParams)-3]
	thirdPubKey := arrayParams[len(arrayParams)-2]
	thirdPrivateKey := arrayParams[len(arrayParams)-1]
	normalBallot := common.Decrypt(Seal1Data, thirdPrivateKey)
	tx.Metadata = metadata.NewNormalDCBBallotFromSealerMetadata(normalBallot.([]byte), [][]byte{firstPubKey.([]byte), secondPubKey.([]byte), thirdPubKey.([]byte)}, Pointer1, Pointer3)
	return tx, err
}

func (self RpcServer) handleCreateRawNormalVoteDCBProposalTransactionFromSealer(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	var err error
	tx, err := self.buildRawNormalVoteDCBProposalTransactionFromSealer(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
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
	tx, err := self.buildRawTransaction(params)
	arrayParams := common.InterfaceSlice(params)
	PointerToLv3Ballot := arrayParams[len(arrayParams)-5]
	Pointer := common.Hash{}
	copy(Pointer[:], []byte(PointerToLv3Ballot.(string)))
	firstPubKey := arrayParams[len(arrayParams)-4]
	secondPubKey := arrayParams[len(arrayParams)-3]
	thirdPubKey := arrayParams[len(arrayParams)-2]
	normalBallot := arrayParams[len(arrayParams)-1]
	tx.Metadata = metadata.NewNormalDCBBallotFromOwnerMetadata(normalBallot.([]byte), [][]byte{[]byte(firstPubKey.(string)), []byte(secondPubKey.(string)), []byte(thirdPubKey.(string))}, Pointer)
	return tx, err
}

func (self RpcServer) handleCreateRawNormalVoteDCBProposalTransactionFromOwner(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	var err error
	tx, err := self.buildRawNormalVoteDCBProposalTransactionFromOwner(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
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

func (self RpcServer) buildRawVoteDCBBoardTransaction(
	params interface{},
) (*transaction.TxCustomToken, error) {
	arrayParams := common.InterfaceSlice(params)
	tx, err := self.buildRawCustomTokenTransaction(params)
	candidatePaymentAddress := arrayParams[len(arrayParams)-1].(string)
	account, _ := wallet.Base58CheckDeserialize(candidatePaymentAddress)
	tx.Metadata = metadata.NewVoteDCBBoardMetadata(account.KeySet.PaymentAddress.Pk)
	return tx, err
}

func (self RpcServer) handleSendRawVoteBoardDCBTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	base58CheckDate := arrayParams[0].(string)
	rawTxBytes, _, err := base58.Base58Check{}.Decode(base58CheckDate)

	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	tx := transaction.TxCustomToken{}
	// Logger.log.Info(string(rawTxBytes))
	err = json.Unmarshal(rawTxBytes, &tx)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	hash, txDesc, err := self.config.TxMemPool.MaybeAcceptTransaction(&tx)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	Logger.log.Infof("there is hash of transaction: %s\n", hash.String())
	Logger.log.Infof("there is priority of transaction in pool: %d", txDesc.StartingPriority)

	// broadcast message
	txMsg, err := wire.MakeEmptyMessage(wire.CmdCustomToken)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	txMsg.(*wire.MessageTx).Transaction = &tx
	self.config.Server.PushMessageToAll(txMsg)

	return tx.Hash(), nil
}

func (self RpcServer) handleCreateRawVoteDCBBoardTransaction(
	params interface{},
	closeChan <-chan struct{},
) (interface{}, *RPCError) {
	var err error
	tx, err := self.buildRawVoteDCBBoardTransaction(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
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

func (self RpcServer) handleCreateAndSendVoteDCBBoardTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawVoteDCBBoardTransaction(params, closeChan)
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
	txId, err := self.handleSendRawVoteBoardDCBTransaction(newParam, closeChan)
	return txId, err
}

func (self RpcServer) buildRawSubmitDCBProposalTransaction(
	params interface{},
) (*transaction.Tx, *RPCError) {
	tx, err := self.buildRawTransaction(params)
	arrayParams := common.InterfaceSlice(params)
	NParams := len(arrayParams)
	DCBParams := *params2.NewDCBParamsFromRPC(arrayParams[NParams-4])
	executeDuration := arrayParams[NParams-3].(uint32)
	explanation := arrayParams[NParams-2].(string)
	paymentData := common.InterfaceSlice(arrayParams[NParams-1])
	address := privacy.PaymentAddress{
		Pk: paymentData[0].([]byte),
		Tk: paymentData[1].([]byte),
	}

	tx.Metadata = metadata.NewSubmitDCBProposalMetadata(DCBParams, executeDuration, explanation, &address)
	return tx, err
}

func (self RpcServer) handleCreateRawSubmitDCBProposalTransaction(
	params interface{},
	closeChan <-chan struct{},
) (interface{}, *RPCError) {
	var err error
	tx, err := self.buildRawSubmitDCBProposalTransaction(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
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

func (self RpcServer) handleSendRawSubmitDCBProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	var err error
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	base58CheckDate := arrayParams[0].(string)
	rawTxBytes, _, err := base58.Base58Check{}.Decode(base58CheckDate)

	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	tx := transaction.Tx{}
	// Logger.log.Info(string(rawTxBytes))
	err = json.Unmarshal(rawTxBytes, &tx)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	hash, txDesc, err := self.config.TxMemPool.MaybeAcceptTransaction(&tx)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	Logger.log.Infof("there is hash of transaction: %s\n", hash.String())
	Logger.log.Infof("there is priority of transaction in pool: %d", txDesc.StartingPriority)

	// broadcast message
	txMsg, err := wire.MakeEmptyMessage(wire.CmdTx)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	txMsg.(*wire.MessageTx).Transaction = &tx
	self.config.Server.PushMessageToAll(txMsg)

	return tx.Hash(), nil
}

func (self RpcServer) handleCreateAndSendSubmitDCBProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawSubmitDCBProposalTransaction(params, closeChan)
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
	txId, err := self.handleSendRawSubmitDCBProposalTransaction(newParam, closeChan)
	return txId, err
}
