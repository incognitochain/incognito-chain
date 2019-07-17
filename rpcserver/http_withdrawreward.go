package rpcserver

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/incognitokey"

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
	keyWallet.KeySet.ImportFromPrivateKeyByte(keyWallet.KeySet.PrivateKey)
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
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("key component invalid"))
	}
	senderKeyParam := arrayParams[0]

	var keySet *incognitokey.KeySet

	if senderKeyParam != "" {
		senderKey, err := wallet.Base58CheckDeserialize(senderKeyParam.(string))
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
		}
		err = senderKey.KeySet.ImportFromPrivateKey(&senderKey.KeySet.PrivateKey)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
		}
		keySet = &senderKey.KeySet
	} else {
		keySet = httpServer.config.Server.GetUserKeySet()
	}

	allCoinIDs, err := httpServer.config.BlockChain.GetAllCoinID()
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	rewardAmounts := make(map[common.Hash]uint64)
	for _, coinID := range allCoinIDs {
		amount, err := (*httpServer.config.Database).GetCommitteeReward(keySet.PaymentAddress.Pk, coinID)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
		}
		rewardAmounts[coinID] = amount
	}

	return rewardAmounts, nil
}

// handleListRewardAmount - Get the reward amount of all committee with all existed token
func (httpServer *HttpServer) handleListRewardAmount(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	result := (*httpServer.config.Database).ListCommitteeReward()
	return result, nil
}
