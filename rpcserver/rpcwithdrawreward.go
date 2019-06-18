package rpcserver

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

func (rpcServer RpcServer) handleCreateRawWithDrawTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
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
	arrayParams[4] = interface{}(param)
	return rpcServer.createRawTxWithMetadata(
		arrayParams,
		closeChan,
		metadata.NewWithDrawRewardRequestFromRPC,
	)
}

func (rpcServer RpcServer) handleCreateAndSendWithDrawTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	//VoteProposal - Step 1: Client call rpc function to create vote proposal transaction
	return rpcServer.createAndSendTxWithMetadata(
		params,
		closeChan,
		RpcServer.handleCreateRawWithDrawTransaction,
		RpcServer.handleSendRawTransaction,
	)
}

// Get the reward amount of a private key
func (rpcServer RpcServer) handleGetRewardAmount(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("key component invalid"))
	}
	senderKeyParam := arrayParams[0]
	senderKey, err := wallet.Base58CheckDeserialize(senderKeyParam.(string))
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	senderKey.KeySet.ImportFromPrivateKey(&senderKey.KeySet.PrivateKey)

	allCoinIDs, err := rpcServer.config.BlockChain.GetAllCoinID()
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	rewardAmounts := make(map[common.Hash]uint64)
	for _, coinID := range allCoinIDs {
		amount, err := (*rpcServer.config.Database).GetCommitteeReward(senderKey.KeySet.PaymentAddress.Pk, coinID)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
		}
		rewardAmounts[coinID] = amount
	}

	return rewardAmounts, nil
}
