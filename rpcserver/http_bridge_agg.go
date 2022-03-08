package rpcserver

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	metadataBridgeAgg "github.com/incognitochain/incognito-chain/metadata/bridgeagg"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"github.com/incognitochain/incognito-chain/wallet"
)

func (httpServer *HttpServer) handleCreateAndSendTxBridgeAggModifyListTokens(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	var res interface{}
	data, err := httpServer.createBridgeAggModifyListTokensTransaction(params)
	if err != nil {
		return nil, err
	}
	base58CheckData := data.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	res, err = sendCreatedTransaction(httpServer, newParam, false, closeChan)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (httpServer *HttpServer) createBridgeAggModifyListTokensTransaction(
	params interface{},
) (*jsonresult.CreateTransactionResult, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	privateKey, ok := arrayParams[0].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("private key is invalid"))
	}
	privacyDetect, ok := arrayParams[3].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("privacy detection param need to be int"))
	}
	if int(privacyDetect) <= 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Tx has to be a privacy tx"))
	}
	if len(arrayParams) != 5 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid length of rpc expect %v but get %v", 4, len(arrayParams)))
	}
	keyWallet, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot deserialize private"))
	}
	if len(keyWallet.KeySet.PrivateKey) == 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid private key"))
	}

	// metadata object format to read from RPC parameters
	mdReader := &struct {
		NewList map[common.Hash][]common.Hash `json:"NewList"`
	}{}
	// parse params & metadata
	paramSelect, err := httpServer.pdexTxService.ReadParamsFrom(params, mdReader)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot deserialize parameters %v", err))
	}

	md := metadataBridgeAgg.NewModifyListTokenWithValue(mdReader.NewList)

	paramSelect.SetTokenID(common.PRVCoinID)
	paramSelect.SetMetadata(md)

	// get burning address
	bc := httpServer.pdexTxService.BlockChain
	bestState, err := bc.GetClonedBeaconBestState()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetClonedBeaconBestStateError, err)
	}
	temp := bc.GetBurningAddress(bestState.BeaconHeight)
	w, _ := wallet.Base58CheckDeserialize(temp)
	burnAddr := w.KeySet.PaymentAddress

	// burn selling amount for order, plus fee
	burnPayments := []*privacy.PaymentInfo{
		&privacy.PaymentInfo{
			PaymentAddress: burnAddr,
		},
	}
	paramSelect.PRV.PaymentInfos = burnPayments

	// create transaction
	tx, err := httpServer.pdexTxService.BuildTransaction(paramSelect, md)
	// error must be of type *RPCError for equality
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.CreateTxDataError, err)
	}

	marshaledTx, err := json.Marshal(tx)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.CreateTxDataError, err)
	}
	res := &jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(marshaledTx, 0x00),
	}
	return res, nil
}
