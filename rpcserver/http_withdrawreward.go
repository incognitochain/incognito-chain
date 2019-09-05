package rpcserver

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

func (httpServer *HttpServer) handleCreateRawWithDrawTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	//VoteProposal - Step 2: Create Raw vote proposal transaction
	// params = setBuildRawBurnTransactionParams(params, FeeVote)
	arrayParams := common.InterfaceSlice(params)
	arrayParams[1] = nil
	param := map[string]interface{}{}
	keyWallet, err := wallet.Base58CheckDeserialize(arrayParams[0].(string))
	if err != nil {
		return []byte{}, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New(fmt.Sprintf("Wrong privatekey %+v", err)))
	}
	keyWallet.KeySet.InitFromPrivateKeyByte(keyWallet.KeySet.PrivateKey)
	param["PaymentAddress"] = keyWallet.Base58CheckSerialize(1)
	param["TokenID"] = arrayParams[4].(map[string]interface{})["TokenID"]
	arrayParams[4] = interface{}(param)
	return httpServer.createRawTxWithMetadata(
		arrayParams,
		closeChan,
		metadata.NewWithDrawRewardRequestFromRPC,
	)
}

func (httpServer *HttpServer) handleCreateAndSendWithDrawTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	//VoteProposal - Step 1: Client call rpc function to create vote proposal transaction
	return httpServer.createAndSendTxWithMetadata(
		params,
		closeChan,
		(*HttpServer).handleCreateRawWithDrawTransaction,
		(*HttpServer).handleSendRawTransaction,
	)
}

// handleGetRewardAmount - Get the reward amount of a payment address with all existed token
func (httpServer *HttpServer) handleGetRewardAmount(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	rewardAmountResult := make(map[string]uint64)
	rewardAmounts := make(map[common.Hash]uint64)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("key component invalid"))
	}
	paymentAddress := arrayParams[0]

	var keySet *incognitokey.KeySet

	if paymentAddress != "" {
		senderKey, err := wallet.Base58CheckDeserialize(paymentAddress.(string))
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
		}

		keySet = &senderKey.KeySet
	} else {
		keySet = httpServer.config.Server.GetUserKeySet()
	}

	if keySet == nil {
		return rewardAmountResult, nil
	}

	allCoinIDs, err := httpServer.config.BlockChain.GetAllCoinID()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	for _, coinID := range allCoinIDs {
		amount, err := (*httpServer.config.Database).GetCommitteeReward(keySet.PaymentAddress.Pk, coinID)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
		}
		if coinID == common.PRVCoinID {
			rewardAmountResult["PRV"] = amount
		} else {
			rewardAmounts[coinID] = amount
		}
	}

	cusPrivTok, crossPrivToken, err := httpServer.config.BlockChain.ListPrivacyCustomToken()

	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	for _, token := range cusPrivTok {
		if rewardAmounts[token.TxPrivacyTokenData.PropertyID] > 0 {
			rewardAmountResult[token.TxPrivacyTokenData.PropertyID.String()] = rewardAmounts[token.TxPrivacyTokenData.PropertyID]
		}
	}

	for _, token := range crossPrivToken {
		if rewardAmounts[token.TokenID] > 0 {
			rewardAmountResult[token.TokenID.String()] = rewardAmounts[token.TokenID]
		}
	}

	return rewardAmountResult, nil
}

// handleListRewardAmount - Get the reward amount of all committee with all existed token
func (httpServer *HttpServer) handleListRewardAmount(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	result := (*httpServer.config.Database).ListCommitteeReward()
	return result, nil
}
