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

func (self RpcServer) handleGetLoanParams(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return self.config.BlockChain.BestState[0].BestBlock.Header.DCBConstitution.DCBParams.LoanParams, nil
}

func (self RpcServer) createRawLoanTx(params interface{}, closeChan <-chan struct{}, metaConstructor metaConstructor) (interface{}, *RPCError) {
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	loanDataRaw := arrayParams[len(arrayParams)-1].(map[string]interface{})
	loanMeta, errCons := metaConstructor(loanDataRaw)
	if errCons != nil {
		return nil, NewRPCError(ErrUnexpected, errCons)
	}
	tx, err := self.buildRawTransaction(params, loanMeta)
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

func (self RpcServer) sendRawLoanTx(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
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

	hash, txDesc, err := self.config.TxMemPool.MaybeAcceptTransaction(&tx)
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
	self.config.Server.PushMessageToAll(txMsg)

	result := jsonresult.CreateTransactionResult{
		TxID: tx.Hash().String(),
	}
	return result, nil
}

func (self RpcServer) createAndSendLoanTx(params interface{}, closeChan <-chan struct{}, createHandler, sendHandler commandHandler) (interface{}, *RPCError) {
	data, err := createHandler(self, params, closeChan)
	fmt.Printf("err create handler: %v\n", err)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	return sendHandler(self, newParam, closeChan)
}

func (self RpcServer) handleCreateRawLoanRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	constructor := constructors[CreateAndSendLoanRequest]
	return self.createRawLoanTx(params, closeChan, constructor)
}

func (self RpcServer) handleSendRawLoanRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return self.sendRawLoanTx(params, closeChan)
}

func (self RpcServer) handleCreateAndSendLoanRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return self.createAndSendLoanTx(
		params,
		closeChan,
		RpcServer.handleCreateRawLoanRequest,
		RpcServer.handleSendRawLoanRequest,
	)
}

func (self RpcServer) handleCreateRawLoanResponse(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	constructor := constructors[CreateAndSendLoanResponse]
	return self.createRawLoanTx(params, closeChan, constructor)
}

func (self RpcServer) handleSendRawLoanResponse(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return self.sendRawLoanTx(params, closeChan)
}

func (self RpcServer) handleCreateAndSendLoanResponse(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return self.createAndSendLoanTx(
		params,
		closeChan,
		RpcServer.handleCreateRawLoanResponse,
		RpcServer.handleSendRawLoanResponse,
	)
}

func (self RpcServer) handleCreateRawLoanWithdraw(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	constructor := constructors[CreateAndSendLoanWithdraw]
	return self.createRawLoanTx(params, closeChan, constructor)
}

func (self RpcServer) handleSendRawLoanWithdraw(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return self.sendRawLoanTx(params, closeChan)
}

func (self RpcServer) handleCreateAndSendLoanWithdraw(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return self.createAndSendLoanTx(
		params,
		closeChan,
		RpcServer.handleCreateRawLoanWithdraw,
		RpcServer.handleSendRawLoanWithdraw,
	)
}

func (self RpcServer) handleCreateRawLoanPayment(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	constructor := constructors[CreateAndSendLoanPayment]
	return self.createRawLoanTx(params, closeChan, constructor)
}

func (self RpcServer) handleSendRawLoanPayment(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return self.sendRawLoanTx(params, closeChan)
}

func (self RpcServer) handleCreateAndSendLoanPayment(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return self.createAndSendLoanTx(
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
func (self RpcServer) handleGetLoanResponseApproved(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	result := jsonresult.ListLoanResponseApproved{
		Approvers: make(map[string][]string),
		Approved:  make(map[string]bool),
	}
	for _, param := range arrayParams {
		strLoanID := param.(string)
		approvers := self.getResponseAddresses(strLoanID, metadata.Accept)
		approveReq := self.config.BlockChain.GetDCBParams().MinLoanResponseRequire
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
func (self RpcServer) handleGetLoanResponseRejected(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	result := jsonresult.ListLoanResponseRejected{
		Rejectors: make(map[string][]string),
		Rejected:  make(map[string]bool),
	}
	for _, param := range arrayParams {
		strLoanID := param.(string)
		rejectors := self.getResponseAddresses(strLoanID, metadata.Reject)
		approveReq := self.config.BlockChain.GetDCBParams().MinLoanResponseRequire
		rejectReq := common.NumberOfDCBGovernors - approveReq
		rejected := len(rejectors) > int(rejectReq)
		result.Rejectors[strLoanID] = rejectors
		result.Rejected[strLoanID] = rejected
	}
	return result, nil
}

func (self RpcServer) getResponseAddresses(strLoanID string, respType metadata.ValidLoanResponse) []string {
	addresses := []string{}
	loanID, err := hex.DecodeString(strLoanID)
	if err == nil {
		txHashes, err := (*self.config.Database).GetLoanTxs(loanID)
		fmt.Printf("GetLoanTxs found: %x\n", txHashes)
		if err == nil {
			respData := metadata.GetLoanResponses(txHashes, self.config.BlockChain)
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
	key := &wallet.Key{
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
func (self RpcServer) handleGetLoanPaymentInfo(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	result := jsonresult.ListLoanPaymentInfo{
		Info: make(map[string]jsonresult.LoanPaymentInfo),
	}
	for _, param := range arrayParams {
		strLoanID := param.(string)
		loanID, err := hex.DecodeString(strLoanID)
		loanPaymentInfo := jsonresult.LoanPaymentInfo{}
		if err == nil {
			priciple, interest, deadline, err := (*self.config.Database).GetLoanPayment(loanID)
			if err == nil {
				reqMeta, err := (*self.config.BlockChain).GetLoanRequestMeta(loanID)
				if err == nil {
					loanPaymentInfo.Principle = priciple
					loanPaymentInfo.Interest = interest
					loanPaymentInfo.Deadline = deadline + reqMeta.Params.Maturity
				}
			}
		}
		result.Info[strLoanID] = loanPaymentInfo
	}
	return result, nil
}
