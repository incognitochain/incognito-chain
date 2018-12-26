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
	chainID := byte(arrayParams[2].(float64))
	pubKey := wallet.GetPubKeyFromPaymentAddress(paymentAddress)
	startedBlock := self.config.BlockChain.BestState[chainID].BestBlock.Header.Height
	db := *self.config.Database
	result := jsonresult.GetAmountVoteTokenResult{}
	var err error
	if boardType == "dcb" {
		result.DCBVoteTokenAmount, err = db.GetDCBVoteTokenAmount(uint32(startedBlock), pubKey)
		if err != nil {
			result.DCBVoteTokenAmount = 0
		}
	} else if boardType == "gov" {
		result.GOVVoteTokenAmount, err = db.GetGOVVoteTokenAmount(uint32(startedBlock), pubKey)
		if err != nil {
			result.GOVVoteTokenAmount = 0
		}
	} else if boardType == "" {
		result.DCBVoteTokenAmount, err = db.GetDCBVoteTokenAmount(uint32(startedBlock), pubKey)
		if err != nil {
			result.DCBVoteTokenAmount = 0
		}
		result.GOVVoteTokenAmount, err = db.GetGOVVoteTokenAmount(uint32(startedBlock), pubKey)
		if err != nil {
			result.GOVVoteTokenAmount = 0
		}
	}
	return result, nil
}
