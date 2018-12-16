package rpcserver

import (
	"encoding/hex"
	"encoding/json"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	privacy "github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/rpcserver/jsonresult"
)

func (self RpcServer) handleGetBondTypes(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	tempRes1 := jsonresult.GetBondTypeResult{
		BondID:         []byte("12345abc"),
		StartSellingAt: 123,
		Maturity:       300,
		BuyBackPrice:   100000,
	}
	tempRes2 := jsonresult.GetBondTypeResult{
		BondID:         []byte("12345xyz"),
		StartSellingAt: 95,
		Maturity:       200,
		BuyBackPrice:   200000,
	}
	return []jsonresult.GetBondTypeResult{tempRes1, tempRes2}, nil
}

func (self RpcServer) handleGetGOVParams(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	govParam := self.config.BlockChain.BestState[0].BestBlock.Header.GOVConstitution.GOVParams
	return govParam, nil
}

func (self RpcServer) handleGetGOVConstitution(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	constitution := self.config.BlockChain.BestState[0].BestBlock.Header.GOVConstitution
	return constitution, nil
}

func (self RpcServer) handleGetListGOVBoard(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return self.config.BlockChain.BestState[0].BestBlock.Header.GOVGovernor.GOVBoardPubKeys, nil
}

func (self RpcServer) handleCreateRawTxWithBuyBackRequest(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	arrayParams := common.InterfaceSlice(params)
	normalTx, err := self.buildRawTransaction(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	// Req param #4: issuing request info
	buyBackReq := arrayParams[4].(map[string]interface{})
	voutIndex := int(buyBackReq["voutIndex"].(float64))
	buyBackFromTxIDBytes := []byte(buyBackReq["buyBackFromTxID"].(string))
	buyBackFromTxID := common.Hash{}
	copy(buyBackFromTxID[:], buyBackFromTxIDBytes)
	metaType := metadata.BuyBackRequestMeta
	normalTx.Metadata = metadata.NewBuyBackRequest(
		buyBackFromTxID,
		voutIndex,
		metaType,
	)
	byteArrays, err := json.Marshal(normalTx)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	hexData := hex.EncodeToString(byteArrays)
	result := jsonresult.CreateTransactionResult{
		HexData: hexData,
	}
	return result, nil
}

func (self RpcServer) handleCreateAndSendTxWithBuyBackRequest(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	data, err := self.handleCreateRawTxWithBuyBackRequest(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	hexStrOfTx := tx.HexData
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, hexStrOfTx)
	sendResult, err := self.handleSendRawTransaction(newParam, closeChan)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID: sendResult.(jsonresult.CreateTransactionResult).TxID,
	}
	return result, nil
}

func (self RpcServer) handleCreateRawTxWithBuySellRequest(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	arrayParams := common.InterfaceSlice(params)
	normalTx, err := self.buildRawTransaction(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	// Req param #4: buy/sell request info
	buySellReq := arrayParams[4].(map[string]interface{})

	paymentAddressMap := buySellReq["paymentAddress"].(map[string]interface{})
	paymentAddress := privacy.PaymentAddress{
		Pk: []byte(paymentAddressMap["pk"].(string)),
		Tk: []byte(paymentAddressMap["tk"].(string)),
	}
	assetTypeBytes := []byte(buySellReq["assetType"].(string))
	assetType := common.Hash{}
	copy(assetType[:], assetTypeBytes)
	amount := uint64(buySellReq["amount"].(float64))
	buyPrice := uint64(buySellReq["buyPrice"].(float64))
	metaType := metadata.BuyFromGOVRequestMeta
	normalTx.Metadata = metadata.NewBuySellRequest(
		paymentAddress,
		assetType,
		amount,
		buyPrice,
		metaType,
	)
	byteArrays, err := json.Marshal(normalTx)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	hexData := hex.EncodeToString(byteArrays)
	result := jsonresult.CreateTransactionResult{
		HexData: hexData,
	}
	return result, nil
}

func (self RpcServer) handleCreateAndSendTxWithBuySellRequest(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	data, err := self.handleCreateRawTxWithBuySellRequest(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	hexStrOfTx := tx.HexData
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, hexStrOfTx)
	sendResult, err := self.handleSendRawTransaction(newParam, closeChan)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID: sendResult.(jsonresult.CreateTransactionResult).TxID,
	}
	return result, nil
}
