package rpcserver

import (
	"encoding/json"
	"errors"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/common/base58"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/rpcserver/jsonresult"
	"github.com/constant-money/constant-chain/wallet"
)

func (rpcServer RpcServer) handleCreateRawTxWithMultiSigsReg(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	var err error
	arrayParams := common.InterfaceSlice(params)
	// Req param #4: multisigs registration info
	multiSigsReg := arrayParams[4].(map[string]interface{})
	registeringPaymentAddrStr := multiSigsReg["RegisteringPaymentAddressStr"].(string)
	registeringPaymentAddrKey, err := wallet.Base58CheckDeserialize(registeringPaymentAddrStr)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	registeringPaymentAddr := registeringPaymentAddrKey.KeySet.PaymentAddress

	spendableMembersAddrStrs := multiSigsReg["SpendableMembersAddrStrs"].([]interface{})
	spendableMembers := make([][]byte, len(spendableMembersAddrStrs))
	for i, addrStr := range spendableMembersAddrStrs {
		assertedAddrStr, ok := addrStr.(string)
		if !ok {
			return nil, NewRPCError(ErrUnexpected, errors.New("wrong type on payment address string"))
		}
		paymentAddrKey, err := wallet.Base58CheckDeserialize(assertedAddrStr)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
		}
		paymentAddr := paymentAddrKey.KeySet.PaymentAddress
		spendableMembers[i] = paymentAddr.Pk
	}

	metaType := metadata.MultiSigsRegistrationMeta
	meta := metadata.NewMultiSigsRegistration(
		registeringPaymentAddr,
		spendableMembers,
		metaType,
	)

	normalTx, err := rpcServer.buildRawTransaction(params, meta)
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

func (rpcServer RpcServer) handleCreateAndSendTxWithMultiSigsReg(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := rpcServer.handleCreateRawTxWithMultiSigsReg(params, closeChan)
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
	sendResult, err := rpcServer.handleSendRawTransaction(newParam, closeChan)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID: sendResult.(jsonresult.CreateTransactionResult).TxID,
	}
	return result, nil
}

func (rpcServer RpcServer) handleCreateRawTxWithMultiSigsSpending(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	var err error
	arrayParams := common.InterfaceSlice(params)
	// Req param #4: multisigs spending info
	multiSigsSpending, ok := arrayParams[4].(map[string]interface{})
	if !ok {
		return nil, NewRPCError(ErrUnexpected, errors.New("could not parse requesting multiSigsSpending metadata"))
	}

	signs := multiSigsSpending["Signs"].(map[string]interface{})
	assertedSigns := map[string][]byte{}
	for paymentAddrStr, sign := range signs {
		signerPaymentAddrKey, err := wallet.Base58CheckDeserialize(paymentAddrStr)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
		}
		signerPaymentAddr := signerPaymentAddrKey.KeySet.PaymentAddress
		assertedSigns[string(signerPaymentAddr.Pk)] = []byte(sign.(string))
	}
	metaType := metadata.MultiSigsSpendingMeta
	meta := metadata.NewMultiSigsSpending(
		assertedSigns,
		metaType,
	)

	normalTx, err := rpcServer.buildRawTransaction(params, meta)
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

func (rpcServer RpcServer) handleCreateAndSendTxWithMultiSigsSpending(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := rpcServer.handleCreateRawTxWithMultiSigsSpending(params, closeChan)
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
	sendResult, err := rpcServer.handleSendRawTransaction(newParam, closeChan)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID: sendResult.(jsonresult.CreateTransactionResult).TxID,
	}
	return result, nil
}
