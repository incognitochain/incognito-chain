package rpcserver

import (
	"encoding/hex"
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
func (rpcServer RpcServer) handleGetDCBParams(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	// constitution := rpcServer.config.BlockChain.BestState[0].BestBlock.Header.DCBConstitution
	// dcbParam := constitution.DCBParams
	// results := make(map[string]interface{})
	// results["DCBParams"] = dcbParam
	// results["ExecuteDuration"] = constitution.ExecuteDuration
	// results["Explanation"] = constitution.Explanation
	// return results, nil
	return nil, nil
}

// handleGetDCBConstitution - get dcb constitution
func (rpcServer RpcServer) handleGetDCBConstitution(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	// constitution := rpcServer.config.BlockChain.BestState[0].BestBlock.Header.DCBConstitution
	// return constitution, nil
	return nil, nil
}

// handleGetListDCBBoard - return list payment address of DCB board
func (rpcServer RpcServer) handleGetListDCBBoard(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	// res := ListPaymentAddressToListString(rpcServer.config.BlockChain.BestState[0].BestBlock.Header.DCBGovernor.BoardPaymentAddress)
	// return res, nil
	return nil, nil
}

func (rpcServer RpcServer) handleAppendListDCBBoard(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	// arrayParams := common.InterfaceSlice(params)
	// senderKey := arrayParams[0].(string)
	// paymentAddress, _ := rpcServer.GetPaymentAddressFromSenderKeyParams(senderKey)
	// rpcServer.config.BlockChain.BestState[0].BestBlock.Header.DCBGovernor.BoardPaymentAddress = append(rpcServer.config.BlockChain.BestState[0].BestBlock.Header.DCBGovernor.BoardPaymentAddress, *paymentAddress)
	// res := ListPaymentAddressToListString(rpcServer.config.BlockChain.BestState[0].BestBlock.Header.DCBGovernor.BoardPaymentAddress)
	// return res, nil
	return nil, nil
}

func ListPaymentAddressToListString(addresses []privacy.PaymentAddress) []string {
	res := make([]string, 0)
	for _, i := range addresses {
		pk := hex.EncodeToString(i.Pk)
		res = append(res, pk)
	}
	return res
}

func (rpcServer RpcServer) handleCreateRawTxWithIssuingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
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

	normalTx, err := rpcServer.buildRawTransaction(params, meta)
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

func (rpcServer RpcServer) handleCreateAndSendTxWithIssuingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := rpcServer.handleCreateRawTxWithIssuingRequest(params, closeChan)
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

func (rpcServer RpcServer) handleCreateRawTxWithContractingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	metaType := metadata.ContractingRequestMeta
	meta := metadata.NewContractingRequest(metaType)
	normalTx, err := rpcServer.buildRawTransaction(params, meta)
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

func (rpcServer RpcServer) handleCreateAndSendTxWithContractingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := rpcServer.handleCreateRawTxWithContractingRequest(params, closeChan)
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

func getAmountVote(receiversPaymentAddressParam map[string]interface{}) int64 {
	sumAmount := int64(0)
	for paymentAddressStr, amount := range receiversPaymentAddressParam {
		if paymentAddressStr == common.BurningAddress {
			sumAmount += int64(amount.(float64))
		}
	}
	return sumAmount
}

func (rpcServer RpcServer) buildRawVoteDCBBoardTransaction(
	params interface{},
) (*transaction.TxCustomToken, error) {
	arrayParams := common.InterfaceSlice(params)
	candidatePaymentAddress := arrayParams[len(arrayParams)-1].(string)
	account, _ := wallet.Base58CheckDeserialize(candidatePaymentAddress)

	receiversPaymentAddressParam := make(map[string]interface{})
	if arrayParams[1] != nil {
		receiversPaymentAddressParam = arrayParams[1].(map[string]interface{})
	}
	amount := getAmountVote(receiversPaymentAddressParam)

	metadata := metadata.NewVoteDCBBoardMetadata(account.KeySet.PaymentAddress, amount)
	tx, err := rpcServer.buildRawCustomTokenTransaction(params, metadata)

	if err != nil {
		return nil, err
	}
	return tx, err
}

func (rpcServer RpcServer) handleSendRawVoteBoardDCBTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
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

	hash, txDesc, err := rpcServer.config.TxMemPool.MaybeAcceptTransaction(&tx)
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
	rpcServer.config.Server.PushMessageToAll(txMsg)

	return tx.Hash(), nil
}

func (rpcServer RpcServer) handleCreateRawVoteDCBBoardTransaction(
	params interface{},
	closeChan <-chan struct{},
) (interface{}, *RPCError) {
	tx, err := rpcServer.buildRawVoteDCBBoardTransaction(params)
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

func (rpcServer RpcServer) handleCreateAndSendVoteDCBBoardTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := rpcServer.handleCreateRawVoteDCBBoardTransaction(params, closeChan)
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
	txId, err := rpcServer.handleSendRawVoteBoardDCBTransaction(newParam, closeChan)
	return txId, err
}

func (rpcServer RpcServer) buildRawSubmitDCBProposalTransaction(
	params interface{},
) (*transaction.Tx, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	NParams := len(arrayParams)

	newParams := arrayParams[NParams-1].(map[string]interface{})
	tmp, err := rpcServer.GetPaymentAddressFromPrivateKeyParams(arrayParams[0].(string))
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	newParams["PaymentAddress"] = tmp

	meta := metadata.NewSubmitDCBProposalMetadataFromJson(newParams)
	params = setBuildRawBurnTransactionParams(params, common.FeeSubmitProposal)
	tx, err1 := rpcServer.buildRawTransaction(params, meta)
	if err1 != nil {
		return nil, NewRPCError(ErrUnexpected, err1)
	}

	return tx, nil
}

func setBuildRawBurnTransactionParams(params interface{}, fee float64) interface{} {
	arrayParams := common.InterfaceSlice(params)
	x := make(map[string]interface{})
	x[common.BurningAddress] = fee
	arrayParams[1] = x
	return arrayParams
}

func (rpcServer RpcServer) handleCreateRawSubmitDCBProposalTransaction(
	params interface{},
	closeChan <-chan struct{},
) (interface{}, *RPCError) {
	tx, err1 := rpcServer.buildRawSubmitDCBProposalTransaction(params)
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

func (rpcServer RpcServer) handleSendRawSubmitDCBProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
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

	hash, txDesc, err1 := rpcServer.config.TxMemPool.MaybeAcceptTransaction(&tx)
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
	rpcServer.config.Server.PushMessageToAll(txMsg)

	return tx.Hash(), nil
}

func (rpcServer RpcServer) handleCreateAndSendSubmitDCBProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := rpcServer.handleCreateRawSubmitDCBProposalTransaction(params, closeChan)
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
	txId, err := rpcServer.handleSendRawSubmitDCBProposalTransaction(newParam, closeChan)
	return txId, err
}
