package rpcserver

import (
	"encoding/json"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/rpcserver/jsonresult"
	"github.com/pkg/errors"
)

func (self RpcServer) handleCreateAndSendTxWithCMBInitRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	normalTx, err := self.buildRawTransaction(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	// Req param #4: cmb init request
	paramsMap := arrayParams[4].(map[string]interface{})
	cmbInitRequest := metadata.NewCMBInitRequest(paramsMap)
	if cmbInitRequest == nil {
		return nil, NewRPCError(ErrUnexpected, errors.Errorf("Invalid CMBInitRequest data"))
	}
	normalTx.Metadata = cmbInitRequest
	byteArrays, marshalErr := json.Marshal(normalTx)
	if err != nil {
		Logger.log.Error(marshalErr)
		return nil, NewRPCError(ErrUnexpected, marshalErr)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            normalTx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}
