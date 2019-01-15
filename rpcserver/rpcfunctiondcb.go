package rpcserver

import (
	"encoding/json"
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
	constitution := self.config.BlockChain.BestState[0].BestBlock.Header.DCBConstitution
	dcbParam := constitution.DCBParams
	results := make(map[string]interface{})
	results["DCBParams"] = dcbParam
	results["ExecuteDuration"] = constitution.ExecuteDuration
	results["Explanation"] = constitution.Explanation
	return results, nil
}

// handleGetDCBConstitution - get dcb constitution
func (self RpcServer) handleGetDCBConstitution(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	constitution := self.config.BlockChain.BestState[0].BestBlock.Header.DCBConstitution
	return constitution, nil
}

// handleGetListDCBBoard - return list payment address of DCB board
func (self RpcServer) handleGetListDCBBoard(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return self.config.BlockChain.BestState[0].BestBlock.Header.DCBGovernor.BoardPaymentAddress, nil
}

func (self RpcServer) handleCreateRawTxWithIssuingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	// Req param #4: issuing request info
	issuingReq := arrayParams[4].(map[string]interface{})
	depositedAmount := uint64(issuingReq["DepositedAmount"].(float64))
	assetTypeBytes := []byte(issuingReq["AssetType"].(string))
	assetType := common.Hash{}
	copy(assetType[:], assetTypeBytes)
	metaType := metadata.IssuingRequestMeta
	receiverAddressMap := issuingReq["ReceiverAddress"].(map[string]interface{})
	receiverAddress := privacy.PaymentAddress{
		Pk: []byte(receiverAddressMap["Pk"].(string)),
		Tk: []byte(receiverAddressMap["Tk"].(string)),
	}

	meta := metadata.NewIssuingRequest(
		receiverAddress,
		depositedAmount,
		assetType,
		metaType,
	)

	normalTx, err := self.buildRawTransaction(params, meta)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}

	byteArrays, err1 := json.Marshal(normalTx)
	if err1 != nil {
		Logger.log.Error(err1)
		return nil, NewRPCError(ErrUnexpected, err1)
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
	metaType := metadata.ContractingRequestMeta
	meta := metadata.NewContractingRequest(metaType)
	normalTx, err := self.buildRawTransaction(params, meta)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}

	byteArrays, err1 := json.Marshal(normalTx)
	if err1 != nil {
		Logger.log.Error(err1)
		return nil, NewRPCError(ErrUnexpected, err1)
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

func (self RpcServer) buildRawVoteDCBBoardTransaction(
	params interface{},
) (*transaction.TxCustomToken, error) {
	arrayParams := common.InterfaceSlice(params)
	candidatePaymentAddress := arrayParams[len(arrayParams)-1].(string)
	account, _ := wallet.Base58CheckDeserialize(candidatePaymentAddress)
	metadata := metadata.NewVoteDCBBoardMetadata(account.KeySet.PaymentAddress.Pk)
	tx, err := self.buildRawCustomTokenTransaction(params, metadata)
	if err != nil {
		return nil, err
	}
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
	tx, err := self.buildRawVoteDCBBoardTransaction(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}

	byteArrays, err := json.MarshalIndent(tx, "", "\t")
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
	arrayParams := common.InterfaceSlice(params)
	NParams := len(arrayParams)

	newParams := arrayParams[NParams-1].(map[string]interface{})
	tmp, err := SenderKeyParamToMap(arrayParams[0])
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	newParams["PaymentAddress"] = tmp

	meta := metadata.NewSubmitDCBProposalMetadataFromJson(newParams)
	params = setBuildRawBurnSubmitProposalTransactionParams(params)
	tx, err1 := self.buildRawTransaction(params, meta)
	if err1 != nil {
		return nil, NewRPCError(ErrUnexpected, err1)
	}

	return tx, nil
}

func setBuildRawBurnSubmitProposalTransactionParams(params interface{}) interface{} {
	arrayParams := common.InterfaceSlice(params)
	x := make(map[string]interface{})
	x[common.BurningAddress] = float64(common.SubmitProposalFee)
	arrayParams[1] = x
	return arrayParams
}

func (self RpcServer) handleCreateRawSubmitDCBProposalTransaction(
	params interface{},
	closeChan <-chan struct{},
) (interface{}, *RPCError) {
	tx, err1 := self.buildRawSubmitDCBProposalTransaction(params)
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

func (self RpcServer) handleSendRawSubmitDCBProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
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

	hash, txDesc, err1 := self.config.TxMemPool.MaybeAcceptTransaction(&tx)
	if err1 != nil {
		return nil, NewRPCError(ErrUnexpected, err1)
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
