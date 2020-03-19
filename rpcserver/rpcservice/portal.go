package rpcservice

import (
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/database/lvdb"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"strconv"
)

type Portal struct {
	BlockChain *blockchain.BlockChain
}

func (portal *Portal) GetPortingRequestByByKey(txHash string, db database.DatabaseInterface) (jsonresult.PortalPortingRequest, *RPCError) {

	portingRequestKey := lvdb.NewPortingRequestTxKey(txHash)
	portingRequestItem, err :=  blockchain.GetPortingRequestByKey(db, []byte(portingRequestKey))

	if err != nil {
		return jsonresult.PortalPortingRequest{}, NewRPCError(GetPortingRequestError, err)
	}

	if  portingRequestItem == nil {
		return jsonresult.PortalPortingRequest{}, NewRPCError(GetPortingRequestIsEmpty, err)
	}

	result := jsonresult.PortalPortingRequest{
		PortingRequest: *portingRequestItem,
	}

	return result, nil
}

func (portal *Portal) GetPortingRequestByByPortingId(portingId string, db database.DatabaseInterface) (jsonresult.PortalPortingRequest, *RPCError) {

	portingRequestKey := lvdb.NewPortingRequestKey(portingId)
	portingRequestItem, err :=  blockchain.GetPortingRequestByKey(db, []byte(portingRequestKey))

	if err != nil {
		return jsonresult.PortalPortingRequest{}, NewRPCError(GetPortingRequestError, err)
	}

	if  portingRequestItem == nil {
		return jsonresult.PortalPortingRequest{}, NewRPCError(GetPortingRequestIsEmpty, err)
	}

	result := jsonresult.PortalPortingRequest{
		PortingRequest: *portingRequestItem,
	}

	return result, nil
}

func (portal *Portal) GetCustodianWithdrawByTxId(txId string, db database.DatabaseInterface) (jsonresult.PortalCustodianWithdrawRequest, *RPCError) {
	key := lvdb.NewCustodianWithdrawRequest(txId)
	custodianWithdraw, err :=  blockchain.GetCustodianWithdrawRequestByKey(db, []byte(key))

	if err != nil {
		return jsonresult.PortalCustodianWithdrawRequest{}, NewRPCError(GetPortingRequestError, err)
	}

	if  custodianWithdraw == nil {
		return jsonresult.PortalCustodianWithdrawRequest{}, NewRPCError(GetPortingRequestIsEmpty, err)
	}

	result := jsonresult.PortalCustodianWithdrawRequest{
		CustodianWithdrawRequest: *custodianWithdraw,
	}

	return result, nil
}


func (portal *Portal) GetFinalExchangeRates(service *BlockService, db database.DatabaseInterface) (jsonresult.FinalExchangeRatesResult, *RPCError) {
	beaconBlock, err := service.GetBeaconBestBlock()
	if err != nil {
		return jsonresult.FinalExchangeRatesResult{}, NewRPCError(GetBeaconBestBlockError, err)
	}

	finalExchangeRatesKey := lvdb.NewFinalExchangeRatesKey(beaconBlock.GetHeight())
	finalExchangeRates, err := blockchain.GetFinalExchangeRatesByKey(
		db,
		[]byte(finalExchangeRatesKey),
	)

	if err != nil {
		return  jsonresult.FinalExchangeRatesResult{}, NewRPCError(GetExchangeRatesError, err)
	}

	item := make(map[string]jsonresult.FinalExchangeRatesDetailResult)

	for pTokenId, rates := range finalExchangeRates.Rates {
		item[pTokenId] = jsonresult.FinalExchangeRatesDetailResult{
			Value: rates.Amount,
		}
	}

	result := jsonresult.FinalExchangeRatesResult{
		BeaconHeight: beaconBlock.GetHeight(),
		Rates: item,
	}
	return result, nil
}

func (portal *Portal) ConvertExchangeRates(tokenSymbol string, valuePToken uint64, service *BlockService, db database.DatabaseInterface) (map[string]uint64, *RPCError) {

	result := make(map[string]uint64)
	beaconBlock, err := service.GetBeaconBestBlock()
	if err != nil {
		return result, NewRPCError(GetBeaconBestBlockError, err)
	}

	finalExchangeRatesKey := lvdb.NewFinalExchangeRatesKey(beaconBlock.GetHeight())
	finalExchangeRates, err := blockchain.GetFinalExchangeRatesByKey(db, []byte(finalExchangeRatesKey))

	if err != nil {
		return result, NewRPCError(GetExchangeRatesError, err)
	}

	if err := blockchain.ValidationExchangeRates(finalExchangeRates) ; err != nil {
		return result, NewRPCError(GetExchangeRatesError, err)
	}

	exchange, err := finalExchangeRates.ExchangePToken2PRVByTokenId(tokenSymbol, valuePToken)

	if err != nil {
		return result, NewRPCError(GetExchangeRatesError, err)
	}
	result[tokenSymbol] = exchange

	return result, nil
}

func (portal *Portal) GetPortingFees(tokenSymbol string, valuePToken uint64, service *BlockService, db database.DatabaseInterface) (map[string]uint64, *RPCError) {

	result := make(map[string]uint64)
	beaconBlock, err := service.GetBeaconBestBlock()
	if err != nil {
		return result, NewRPCError(GetBeaconBestBlockError, err)
	}

	finalExchangeRatesKey := lvdb.NewFinalExchangeRatesKey(beaconBlock.GetHeight())
	finalExchangeRates, err := blockchain.GetFinalExchangeRatesByKey(db, []byte(finalExchangeRatesKey))

	if err != nil {
		return result, NewRPCError(GetExchangeRatesError, err)
	}

	if err := blockchain.ValidationExchangeRates(finalExchangeRates) ; err != nil {
		return result, NewRPCError(GetExchangeRatesError, err)
	}

	exchangePortingFees, err := blockchain.CalMinPortingFee(valuePToken, tokenSymbol, finalExchangeRates)

	if err != nil {
		return result, NewRPCError(GetExchangeRatesError, err)
	}

	result[tokenSymbol] = exchangePortingFees

	return result, nil
}

func (portal *Portal) GetLiquidateTpExchangeRates(beaconHeight uint64, custodianAddress string, tokenSymbol string, service *BlockService, db database.DatabaseInterface) (jsonresult.GetLiquidateTpExchangeRates, error) {
	liquidateTpExchangeRatesKey := lvdb.NewPortalLiquidateTPExchangeRatesKey(beaconHeight, custodianAddress)
	liquidateTpExchangeRates, err := blockchain.GetLiquidateTPExchangeRatesByKey(db, []byte(liquidateTpExchangeRatesKey))

	if err != nil {
		return jsonresult.GetLiquidateTpExchangeRates{}, err
	}

	topPercentile, ok := liquidateTpExchangeRates.Rates[tokenSymbol]

	if !ok {
		return jsonresult.GetLiquidateTpExchangeRates{}, nil
	}

	tp := "TP" + strconv.Itoa(topPercentile.TPKey)
	result := jsonresult.GetLiquidateTpExchangeRates{
		TokenId: tokenSymbol,
		TopPercentile: tp,
		Data: topPercentile,
	}

	return result, nil
}

func (portal *Portal) GetLiquidateExchangeRates(
	beaconHeight uint64,
	tokenSymbol string,
	db database.DatabaseInterface,
) (jsonresult.GetLiquidateExchangeRates, error) {
	liquidateExchangeRatesKey := lvdb.NewPortalLiquidateExchangeRatesKey(beaconHeight)
	liquidateExchangeRates, err := blockchain.GetLiquidateExchangeRatesByKey(db, []byte(liquidateExchangeRatesKey))

	if err != nil {
		return jsonresult.GetLiquidateExchangeRates{}, err
	}

	liquidateExchangeRatesDetail, ok := liquidateExchangeRates.Rates[tokenSymbol]

	if !ok {
		return jsonresult.GetLiquidateExchangeRates{}, nil
	}

	result := jsonresult.GetLiquidateExchangeRates{
		TokenId: tokenSymbol,
		Liquidation: liquidateExchangeRatesDetail,
	}
	return result, nil
}
