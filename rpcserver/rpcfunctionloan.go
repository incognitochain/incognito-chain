package rpcserver

import (
	"encoding/hex"
	"fmt"

	"github.com/big0t/constant-chain/cashec"
	"github.com/big0t/constant-chain/common"
	"github.com/big0t/constant-chain/metadata"
	"github.com/big0t/constant-chain/privacy"
	"github.com/big0t/constant-chain/rpcserver/jsonresult"
	"github.com/big0t/constant-chain/wallet"
)

func (rpcServer RpcServer) handleGetLoanParams(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.config.BlockChain.BestState.Beacon.StabilityInfo.DCBConstitution.DCBParams.ListLoanParams, nil
}

func (rpcServer RpcServer) handleCreateRawLoanRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	constructor := metaConstructors[CreateAndSendLoanRequest]
	return rpcServer.createRawTxWithMetadata(params, closeChan, constructor)
}

func (rpcServer RpcServer) handleSendRawLoanRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.sendRawTxWithMetadata(params, closeChan)
}

func (rpcServer RpcServer) handleCreateAndSendLoanRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.createAndSendTxWithMetadata(
		params,
		closeChan,
		RpcServer.handleCreateRawLoanRequest,
		RpcServer.handleSendRawLoanRequest,
	)
}

func (rpcServer RpcServer) handleCreateRawLoanResponse(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	constructor := metaConstructors[CreateAndSendLoanResponse]
	return rpcServer.createRawTxWithMetadata(params, closeChan, constructor)
}

func (rpcServer RpcServer) handleSendRawLoanResponse(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.sendRawTxWithMetadata(params, closeChan)
}

func (rpcServer RpcServer) handleCreateAndSendLoanResponse(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.createAndSendTxWithMetadata(
		params,
		closeChan,
		RpcServer.handleCreateRawLoanResponse,
		RpcServer.handleSendRawLoanResponse,
	)
}

func (rpcServer RpcServer) handleCreateRawLoanWithdraw(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	constructor := metaConstructors[CreateAndSendLoanWithdraw]
	return rpcServer.createRawTxWithMetadata(params, closeChan, constructor)
}

func (rpcServer RpcServer) handleSendRawLoanWithdraw(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.sendRawTxWithMetadata(params, closeChan)
}

func (rpcServer RpcServer) handleCreateAndSendLoanWithdraw(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.createAndSendTxWithMetadata(
		params,
		closeChan,
		RpcServer.handleCreateRawLoanWithdraw,
		RpcServer.handleSendRawLoanWithdraw,
	)
}

func (rpcServer RpcServer) handleCreateRawLoanPayment(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	constructor := metaConstructors[CreateAndSendLoanPayment]
	return rpcServer.createRawTxWithMetadata(params, closeChan, constructor)
}

func (rpcServer RpcServer) handleSendRawLoanPayment(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	fmt.Printf("[db] send raw loan payment tx\n")
	return rpcServer.sendRawTxWithMetadata(params, closeChan)
}

func (rpcServer RpcServer) handleCreateAndSendLoanPayment(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.createAndSendTxWithMetadata(
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
		txHashes := [][]byte{}
		_ = loanID
		// txHashes, err := (*rpcServer.config.Database).GetLoanTxs(loanID)
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
			if reqMeta, err := rpcServer.config.BlockChain.GetLoanRequestMeta(loanID); err == nil {
				lastByte := reqMeta.ReceiveAddress.Pk[len(reqMeta.ReceiveAddress.Pk)-1]
				shardID := common.GetShardIDFromLastByte(lastByte)
				height := rpcServer.config.BlockChain.GetChainHeight(shardID)
				loanPaymentInfo.Principle = priciple
				if height >= deadline { // Current term interest is not fully paid
					loanPaymentInfo.Interest = interest
				}
				loanPaymentInfo.Deadline = deadline + reqMeta.Params.Maturity
			} else {
				fmt.Printf("[db] calcLoanPay err: %+v\n", err)
			}
		} else {
			fmt.Printf("[db] calcLoanPay err: %+v\n", err)
		}
	}
	return loanPaymentInfo
}

// handleGetBankFund returns bank fund stored on Beacon chain
func (rpcServer RpcServer) handleGetBankFund(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	bankFund := rpcServer.config.BlockChain.BestState.Beacon.StabilityInfo.BankFund
	return bankFund, nil
}

// handleGetLoanRequestTxStatus checks if loan request is accepted by beacon
func (rpcServer RpcServer) handleGetLoanRequestTxStatus(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	result := map[string]string{}
	for _, p := range arrayParams {
		s := p.(string)
		result[s] = "unknown"
		if h, err := common.NewHashFromStr(s); err == nil {
			if _, _, _, tx, err := rpcServer.config.BlockChain.GetTransactionByHash(h); err == nil {
				reqMeta := tx.GetMetadata().(*metadata.LoanRequest)
				if _, err = rpcServer.config.BlockChain.GetLoanReq(reqMeta.LoanID); err == nil {
					result[s] = "mined"
				}
			}
		}
	}
	return result, nil
}
