package rpcserver

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/rpcserver/jsonresult"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/wallet"
	"github.com/ninjadotorg/constant/wire"
)

type metaConstructor func(map[string]interface{}) (metadata.Metadata, error)

var constructors = map[string]metaConstructor{
	CreateAndSendLoanRequest:  metadata.NewLoanRequest,
	CreateAndSendLoanResponse: metadata.NewLoanResponse,
	CreateAndSendLoanWithdraw: metadata.NewLoanWithdraw,
	CreateAndSendLoanPayment:  metadata.NewLoanPayment,
}

func (rpcServer RpcServer) handleGetLoanParams(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.config.BlockChain.BestState[0].BestBlock.Header.DCBConstitution.DCBParams.ListLoanParams, nil
}

func (rpcServer RpcServer) createRawLoanTx(params interface{}, closeChan <-chan struct{}, metaConstructor metaConstructor) (interface{}, *RPCError) {
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	loanDataRaw := arrayParams[len(arrayParams)-1].(map[string]interface{})
	loanMeta, errCons := metaConstructor(loanDataRaw)
	if errCons != nil {
		return nil, NewRPCError(ErrUnexpected, errCons)
	}
	tx, err := rpcServer.buildRawTransaction(params, loanMeta)
	fmt.Printf("sigPubKey after build: %v\n", tx.SigPubKey)
	if err != nil {
		return nil, err
	}
	byteArrays, errMarshal := json.Marshal(tx)
	if errMarshal != nil {
		// return hex for a new tx
		return nil, NewRPCError(ErrUnexpected, errMarshal)
	}
	fmt.Printf("Created raw loan tx: %+v\n", tx)
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (rpcServer RpcServer) sendRawLoanTx(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	base58CheckDate := arrayParams[0].(string)
	rawTxBytes, _, err := base58.Base58Check{}.Decode(base58CheckDate)

	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	tx := transaction.Tx{}
	err = json.Unmarshal(rawTxBytes, &tx)
	fmt.Printf("%+v\n", tx)
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
	txMsg, err := wire.MakeEmptyMessage(wire.CmdCLoanRequestToken)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	txMsg.(*wire.MessageTx).Transaction = &tx
	rpcServer.config.Server.PushMessageToAll(txMsg)

	result := jsonresult.CreateTransactionResult{
		TxID: tx.Hash().String(),
	}
	return result, nil
}

func (rpcServer RpcServer) createAndSendLoanTx(params interface{}, closeChan <-chan struct{}, createHandler, sendHandler commandHandler) (interface{}, *RPCError) {
	data, err := createHandler(rpcServer, params, closeChan)
	fmt.Printf("err create handler: %v\n", err)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	return sendHandler(rpcServer, newParam, closeChan)
}

func (rpcServer RpcServer) handleCreateRawLoanRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	constructor := constructors[CreateAndSendLoanRequest]
	return rpcServer.createRawLoanTx(params, closeChan, constructor)
}

func (rpcServer RpcServer) handleSendRawLoanRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.sendRawLoanTx(params, closeChan)
}

func (rpcServer RpcServer) handleCreateAndSendLoanRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.createAndSendLoanTx(
		params,
		closeChan,
		RpcServer.handleCreateRawLoanRequest,
		RpcServer.handleSendRawLoanRequest,
	)
}

func (rpcServer RpcServer) handleCreateRawLoanResponse(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	constructor := constructors[CreateAndSendLoanResponse]
	return rpcServer.createRawLoanTx(params, closeChan, constructor)
}

func (rpcServer RpcServer) handleSendRawLoanResponse(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.sendRawLoanTx(params, closeChan)
}

func (rpcServer RpcServer) handleCreateAndSendLoanResponse(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.createAndSendLoanTx(
		params,
		closeChan,
		RpcServer.handleCreateRawLoanResponse,
		RpcServer.handleSendRawLoanResponse,
	)
}

func (rpcServer RpcServer) handleCreateRawLoanWithdraw(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	constructor := constructors[CreateAndSendLoanWithdraw]
	return rpcServer.createRawLoanTx(params, closeChan, constructor)
}

func (rpcServer RpcServer) handleSendRawLoanWithdraw(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.sendRawLoanTx(params, closeChan)
}

func (rpcServer RpcServer) handleCreateAndSendLoanWithdraw(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.createAndSendLoanTx(
		params,
		closeChan,
		RpcServer.handleCreateRawLoanWithdraw,
		RpcServer.handleSendRawLoanWithdraw,
	)
}

func (rpcServer RpcServer) handleCreateRawLoanPayment(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	constructor := constructors[CreateAndSendLoanPayment]
	return rpcServer.createRawLoanTx(params, closeChan, constructor)
}

func (rpcServer RpcServer) handleSendRawLoanPayment(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.sendRawLoanTx(params, closeChan)
}

func (rpcServer RpcServer) handleCreateAndSendLoanPayment(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.createAndSendLoanTx(
		params,
		closeChan,
		RpcServer.handleCreateRawLoanPayment,
		RpcServer.handleSendRawLoanPayment,
	)
}

// Input: LoanIDs
// Output: for each loan:
//  - approved (bool)
//  - approvers (list pubkeys)
func (rpcServer RpcServer) handleGetLoanResponseApproved(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	result := jsonresult.ListLoanResponseApproved{
		Approvers: make(map[string][]string),
		Approved:  make(map[string]bool),
	}
	for _, param := range arrayParams {
		strLoanID := param.(string)
		approvers := rpcServer.getResponseAddresses(strLoanID, metadata.Accept)
		approveReq := rpcServer.config.BlockChain.GetDCBParams().MinLoanResponseRequire
		approved := len(approvers) >= int(approveReq)
		result.Approvers[strLoanID] = approvers
		result.Approved[strLoanID] = approved
	}
	return result, nil
}

// Input: LoanIDs
// Output: for each loan:
//  - rejected (bool)
//  - rejectors (list pubkeys)
func (rpcServer RpcServer) handleGetLoanResponseRejected(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	result := jsonresult.ListLoanResponseRejected{
		Rejectors: make(map[string][]string),
		Rejected:  make(map[string]bool),
	}
	for _, param := range arrayParams {
		strLoanID := param.(string)
		rejectors := rpcServer.getResponseAddresses(strLoanID, metadata.Reject)
		approveReq := rpcServer.config.BlockChain.GetDCBParams().MinLoanResponseRequire
		rejectReq := common.NumberOfDCBGovernors - approveReq
		rejected := len(rejectors) > int(rejectReq)
		result.Rejectors[strLoanID] = rejectors
		result.Rejected[strLoanID] = rejected
	}
	return result, nil
}

func (rpcServer RpcServer) getResponseAddresses(strLoanID string, respType metadata.ValidLoanResponse) []string {
	addresses := []string{}
	loanID, err := hex.DecodeString(strLoanID)
	if err == nil {
		txHashes, err := (*rpcServer.config.Database).GetLoanTxs(loanID)
		fmt.Printf("GetLoanTxs found: %x\n", txHashes)
		if err == nil {
			respData := metadata.GetLoanResponses(txHashes, rpcServer.config.BlockChain)
			for _, resp := range respData {
				if resp.Response == respType {
					address := getPaymentAddressStrFromPubKey(resp.PublicKey)
					addresses = append(addresses, address)
				}
			}
		}
	}
	return addresses
}

func getPaymentAddressStrFromPubKey(pubkey []byte) string {
	key := &wallet.KeyWallet{
		KeySet: cashec.KeySet{
			PaymentAddress: privacy.PaymentAddress{
				Pk: pubkey,
				Tk: make([]byte, 33),
			},
		},
	}
	return key.Base58CheckSerialize(wallet.PaymentAddressType)
}

// Input: LoanIDs
// Output: for each loan:
//  - rejected (bool)
//  - rejectors (list pubkeys)
func (rpcServer RpcServer) handleGetLoanPaymentInfo(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	result := jsonresult.ListLoanPaymentInfo{
		Info: make(map[string]jsonresult.LoanPaymentInfo),
	}
	for _, param := range arrayParams {
		strLoanID := param.(string)
		result.Info[strLoanID] = rpcServer.calcLoanPaymentInfo(strLoanID)
	}
	return result, nil
}

func (rpcServer RpcServer) calcLoanPaymentInfo(strLoanID string) jsonresult.LoanPaymentInfo {
	loanPaymentInfo := jsonresult.LoanPaymentInfo{}
	if loanID, err := hex.DecodeString(strLoanID); err == nil {
		if priciple, interest, deadline, err := (*rpcServer.config.Database).GetLoanPayment(loanID); err == nil {
			if txReqHash, err := (*rpcServer.config.Database).GetLoanRequestTx(loanID); err == nil {
				hash, _ := (&common.Hash{}).NewHash(txReqHash)
				if _, _, _, txReq, err := (*rpcServer.config.BlockChain).GetTransactionByHash(hash); err == nil {
					reqMeta, _ := txReq.GetMetadata().(*metadata.LoanRequest)
					chainID, _ := common.GetTxSenderChain(txReq.GetSenderAddrLastByte())
					height := rpcServer.config.BlockChain.GetChainHeight(chainID)
					loanPaymentInfo.Principle = priciple
					if height >= deadline { // Current term interest is not fully paid
						loanPaymentInfo.Interest = interest
					}
					loanPaymentInfo.Deadline = deadline + reqMeta.Params.Maturity
				}
			}
		}
	}
	return loanPaymentInfo
}
