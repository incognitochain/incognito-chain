package rpcserver

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

func (httpServer *HttpServer) handleCreateRawWithDrawTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	//VoteProposal - Step 2: Create Raw vote proposal transaction
	// params = setBuildRawBurnTransactionParams(params, FeeVote)
	arrayParams := common.InterfaceSlice(params)
	arrayParams[1] = nil
	param := map[string]interface{}{}
	keyWallet, err := wallet.Base58CheckDeserialize(arrayParams[0].(string))
	if err != nil {
		return []byte{}, NewRPCError(ErrRPCInvalidParams, errors.New(fmt.Sprintf("Wrong privatekey %+v", err)))
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

func (httpServer *HttpServer) handleCreateAndSendWithDrawTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	//VoteProposal - Step 1: Client call rpc function to create vote proposal transaction
	return httpServer.createAndSendTxWithMetadata(
		params,
		closeChan,
		(*HttpServer).handleCreateRawWithDrawTransaction,
		(*HttpServer).handleSendRawTransaction,
	)
}

// handleGetRewardAmount - Get the reward amount of a payment address with all existed token
func (httpServer *HttpServer) handleGetRewardAmount(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	rewardAmountResult := make(map[string]uint64)
	rewardAmounts := make(map[common.Hash]uint64)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("key component invalid"))
	}
	paymentAddress := arrayParams[0]

	var publicKey []byte

	if paymentAddress != "" {
		senderKey, err := wallet.Base58CheckDeserialize(paymentAddress.(string))
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
		}

		publicKey = senderKey.KeySet.PaymentAddress.Pk
	}

	if publicKey == nil {
		return rewardAmountResult, nil
	}

	allCoinIDs, err := httpServer.config.BlockChain.GetAllCoinID()
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	for _, coinID := range allCoinIDs {
		amount, err := (*httpServer.config.Database).GetCommitteeReward(publicKey, coinID)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
		}
		if coinID == common.PRVCoinID {
			rewardAmountResult["PRV"] = amount
		} else {
			rewardAmounts[coinID] = amount
		}
	}

	cusPrivTok, crossPrivToken, err := httpServer.config.BlockChain.ListPrivacyCustomToken()

	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	for _, token := range cusPrivTok {
		if rewardAmounts[token.TxTokenPrivacyData.PropertyID] > 0 {
			rewardAmountResult[token.TxTokenPrivacyData.PropertySymbol] = rewardAmounts[token.TxTokenPrivacyData.PropertyID]
		}
	}

	for _, token := range crossPrivToken {
		if rewardAmounts[token.TokenID] > 0 {
			rewardAmountResult[token.PropertySymbol] = rewardAmounts[token.TokenID]
		}
	}

	return rewardAmountResult, nil
}

// handleListRewardAmount - Get the reward amount of all committee with all existed token
func (httpServer *HttpServer) handleListRewardAmount(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	result := (*httpServer.config.Database).ListCommitteeReward()
	return result, nil
}
