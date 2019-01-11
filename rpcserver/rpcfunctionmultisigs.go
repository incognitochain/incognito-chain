package rpcserver

import (
	"encoding/json"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/rpcserver/jsonresult"
)

func (self RpcServer) handleCreateRawTxWithMultiSigsReg(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	var err error
	arrayParams := common.InterfaceSlice(params)
	// Req param #4: multisigs registration info
	multiSigsReg := arrayParams[4].(map[string]interface{})
	paymentAddressMap := multiSigsReg["PaymentAddress"].(map[string]interface{})
	paymentAddress := privacy.PaymentAddress{
		Pk: []byte(paymentAddressMap["Pk"].(string)),
		Tk: []byte(paymentAddressMap["Tk"].(string)),
	}
	spendableMembers := multiSigsReg["SpendableMembers"].([]interface{})
	assertedSpendableMembers := [][]byte{}
	for _, pk := range spendableMembers {
		assertedSpendableMembers = append(assertedSpendableMembers, []byte(pk.(string)))
	}
	metaType := metadata.MultiSigsRegistrationMeta
	meta := metadata.NewMultiSigsRegistration(
		paymentAddress,
		assertedSpendableMembers,
		metaType,
	)

	normalTx, err := self.buildRawTransaction(params, meta)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	byteArrays, err := json.Marshal(normalTx)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            normalTx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (self RpcServer) handleCreateAndSendTxWithMultiSigsReg(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawTxWithMultiSigsReg(params, closeChan)
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

func (self RpcServer) handleCreateRawTxWithMultiSigsSpending(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	var err error
	arrayParams := common.InterfaceSlice(params)
	// Req param #4: multisigs spending info
	multiSigsSpending := arrayParams[4].(map[string]interface{})
	signs := multiSigsSpending["Signs"].(map[string]interface{})
	assertedSigns := map[string][]byte{}
	for k, s := range signs {
		assertedSigns[k] = []byte(s.(string))
	}
	metaType := metadata.MultiSigsSpendingMeta
	meta := metadata.NewMultiSigsSpending(
		assertedSigns,
		metaType,
	)

	normalTx, err := self.buildRawTransaction(params, meta)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	byteArrays, err := json.Marshal(normalTx)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            normalTx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (self RpcServer) handleCreateAndSendTxWithMultiSigsSpending(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawTxWithMultiSigsSpending(params, closeChan)
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
