package rpcserver

import (
	"fmt"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/wallet"
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
