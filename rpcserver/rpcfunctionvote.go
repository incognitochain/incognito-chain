package rpcserver

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/rpcserver/jsonresult"
	"github.com/ninjadotorg/constant/wallet"
)

func (self RpcServer) handleGetAmountVoteToken(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	boardType := arrayParams[0].(string)
	paymentAddress := arrayParams[1].(string)
	chainID := arrayParams[2].(byte)
	pubKey := wallet.GetPubKeyFromPaymentAddress(paymentAddress)
	startedBlock := self.config.BlockChain.BestState[chainID].BestBlock.Header.Height
	db := *self.config.Database
	result := jsonresult.GetAmountVoteTokenResult{}
	if boardType == "dcb" {
		result.DCBVoteTokenAmount = db.GetDCBVoteTokenAmount(uint32(startedBlock), pubKey)
	} else if boardType == "gov" {
		result.GOVVoteTokenAmount = db.GetGOVVoteTokenAmount(uint32(startedBlock), pubKey)
	} else if boardType == "" {
		result.DCBVoteTokenAmount = db.GetDCBVoteTokenAmount(uint32(startedBlock), pubKey)
		result.GOVVoteTokenAmount = db.GetGOVVoteTokenAmount(uint32(startedBlock), pubKey)
	}
	return result, nil
}
