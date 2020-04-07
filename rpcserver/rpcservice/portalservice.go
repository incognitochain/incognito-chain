package rpcservice

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"strconv"
)

type PortalService struct {
	BlockChain *blockchain.BlockChain
}

func (portal *PortalService) GetPortingRequestByByKey(txId string) (jsonresult.PortalPortingRequest, error) {
	portalStateDB := portal.BlockChain.BestState.Beacon.GetCopiedFeatureStateDB()
	portingRequestItem, err := statedb.GetPortalStateStatusMultiple(portalStateDB, statedb.PortalPortingRequestTxStatusPrefix(), []byte(txId))

	if err != nil {
		return jsonresult.PortalPortingRequest{}, err
	}

	var portingRequestStatus metadata.PortingRequestStatus
	err = json.Unmarshal(portingRequestItem, &portingRequestStatus)
	if err != nil {
		return jsonresult.PortalPortingRequest{}, err
	}

	result := jsonresult.PortalPortingRequest{
		PortingRequest: portingRequestStatus,
	}

	return result, nil
}

func (portal *PortalService) GetPortingRequestByByPortingId(portingId string) (jsonresult.PortalPortingRequest, error) {
	portalStateDB := portal.BlockChain.BestState.Beacon.GetCopiedFeatureStateDB()
	portingRequestItem, err := statedb.GetPortalStateStatusMultiple(portalStateDB, statedb.PortalPortingRequestStatusPrefix(), []byte(portingId))

	if err != nil {
		return jsonresult.PortalPortingRequest{}, err
	}

	var portingRequestStatus metadata.PortingRequestStatus
	err = json.Unmarshal(portingRequestItem, &portingRequestStatus)
	if err != nil {
		return jsonresult.PortalPortingRequest{}, err
	}

	result := jsonresult.PortalPortingRequest{
		PortingRequest: portingRequestStatus,
	}

	return result, nil
}

func (portal *PortalService) GetCustodianWithdrawByTxId(txId string) (jsonresult.PortalCustodianWithdrawRequest, error) {
	portalStateDB := portal.BlockChain.BestState.Beacon.GetCopiedFeatureStateDB()
	custodianWithdraw, err := statedb.GetPortalStateStatusMultiple(portalStateDB, statedb.PortalCustodianWithdrawStatusPrefix(), []byte(txId))

	if err != nil {
		return jsonresult.PortalCustodianWithdrawRequest{}, NewRPCError(GetCustodianWithdrawError, err)
	}

	if  custodianWithdraw == nil {
		return jsonresult.PortalCustodianWithdrawRequest{}, NewRPCError(GetCustodianWithdrawError, err)
	}

	var custodianWithdrawRequestStatus metadata.CustodianWithdrawRequestStatus
	err = json.Unmarshal(custodianWithdraw, &custodianWithdrawRequestStatus)
	if err != nil {
		return jsonresult.PortalCustodianWithdrawRequest{}, err
	}

	result := jsonresult.PortalCustodianWithdrawRequest{
		CustodianWithdrawRequest: custodianWithdrawRequestStatus,
	}

	return result, nil
}


func (portal *PortalService) GetFinalExchangeRates(beaconHeight uint64) (jsonresult.FinalExchangeRatesResult, error) {
	portalStateDB := portal.BlockChain.BestState.Beacon.GetCopiedFeatureStateDB()
	finalExchangeRates, err := statedb.GetFinalExchangeRatesByKey(portalStateDB, beaconHeight)

	if err != nil {
		return jsonresult.FinalExchangeRatesResult{}, err
	}

	item := make(map[string]jsonresult.FinalExchangeRatesDetailResult)

	for pTokenId, rates := range finalExchangeRates.Rates() {
		item[pTokenId] = jsonresult.FinalExchangeRatesDetailResult{
			Value: rates.Amount,
		}
	}

	result := jsonresult.FinalExchangeRatesResult{
		BeaconHeight: beaconHeight,
		Rates: item,
	}
	return result, nil
}

func (portal *PortalService) ConvertExchangeRates(tokenSymbol string, valuePToken uint64, beaconHeight uint64) (map[string]uint64, error) {

	result := make(map[string]uint64)
	finalExchangeRates, err := portal.getFinalExchangeRates(beaconHeight)
	if err != nil {
		return result, err
	}

	finalExchangeRatesObj := blockchain.NewConvertExchangeRatesObject(finalExchangeRates)
	exchange, err := finalExchangeRatesObj.ExchangePToken2PRVByTokenId(tokenSymbol, valuePToken)

	if err != nil {
		return result, err
	}
	result[tokenSymbol] = exchange

	return result, nil
}

func (portal *PortalService) GetPortingFees(tokenSymbol string, valuePToken uint64, beaconHeight uint64) (map[string]uint64, error) {
	result := make(map[string]uint64)
	finalExchangeRates, err := portal.getFinalExchangeRates(beaconHeight)
	if err != nil {
		return result, err
	}

	exchangePortingFees, err := blockchain.CalMinPortingFee(valuePToken, tokenSymbol, finalExchangeRates)

	if err != nil {
		return result, err
	}

	result[tokenSymbol] = exchangePortingFees

	return result, nil
}

func (portal *PortalService) getFinalExchangeRates(beaconHeight uint64) (*statedb.FinalExchangeRatesState, error) {
	portalStateDB := portal.BlockChain.BestState.Beacon.GetCopiedFeatureStateDB()
	finalExchangeRates, err := statedb.GetFinalExchangeRatesByKey(portalStateDB, beaconHeight)

	if err != nil {
		return statedb.NewFinalExchangeRatesState(), err
	}

	if err := blockchain.ValidationExchangeRates(finalExchangeRates) ; err != nil {
		return statedb.NewFinalExchangeRatesState(), err
	}

	if err := blockchain.ValidationExchangeRates(finalExchangeRates) ; err != nil {
		return statedb.NewFinalExchangeRatesState(), err
	}

	return finalExchangeRates, nil
}

func (portal *PortalService) CalculateAmountNeededCustodianDepositLiquidation(beaconHeight uint64, custodianAddress string, pTokenId string, isFreeCollateralSelected bool) (jsonresult.GetLiquidateAmountNeededCustodianDeposit, error) {
	portalStateDB := portal.BlockChain.BestState.Beacon.GetCopiedFeatureStateDB()
	custodian, err := statedb.GetOneCustodian(portalStateDB, beaconHeight, custodianAddress)

	if err != nil {
		return jsonresult.GetLiquidateAmountNeededCustodianDeposit{}, err
	}

	finalExchangeRates, err := portal.getFinalExchangeRates(beaconHeight)
	if err != nil {
		return jsonresult.GetLiquidateAmountNeededCustodianDeposit{}, err
	}

	amountNeeded, _, _, err := blockchain.CalAmountNeededDepositLiquidate(custodian, finalExchangeRates, pTokenId, isFreeCollateralSelected)

	result := jsonresult.GetLiquidateAmountNeededCustodianDeposit{
		IsFreeCollateralSelected: isFreeCollateralSelected,
		Amount: amountNeeded,
		TokenId: pTokenId,
		FreeCollateral: custodian.GetFreeCollateral(),
	}

	return result, nil
}

func (portal *PortalService) GetLiquidateTpExchangeRates(beaconHeight uint64 , custodianAddress string) (interface{}, error) {
	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	newTPKey := beaconHeightBytes
	newTPKey = append(newTPKey, []byte(custodianAddress)...)

	portalStateDB := portal.BlockChain.BestState.Beacon.GetCopiedFeatureStateDB()
	liquidateTpExchangeRates, err := statedb.GetPortalStateStatusMultiple(portalStateDB, statedb.PortalLiquidationTpExchangeRatesStatusPrefix(), newTPKey)

	if err != nil {
		return nil, err
	}

	var liquidateTopPercentileExchangeRatesStatus metadata.LiquidateTopPercentileExchangeRatesStatus
	err = json.Unmarshal(liquidateTpExchangeRates, &liquidateTopPercentileExchangeRatesStatus)
	if err != nil {
		return  nil, err
	}


	return liquidateTopPercentileExchangeRatesStatus, nil
}


func (portal *PortalService) GetLiquidateTpExchangeRatesByToken(beaconHeight uint64, custodianAddress string, tokenSymbol string) (jsonresult.GetLiquidateTpExchangeRates, error) {
	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	newTPKey := beaconHeightBytes
	newTPKey = append(newTPKey, []byte(custodianAddress)...)

	portalStateDB := portal.BlockChain.BestState.Beacon.GetCopiedFeatureStateDB()
	liquidateTpExchangeRates, err := statedb.GetPortalStateStatusMultiple(portalStateDB, statedb.PortalLiquidationTpExchangeRatesStatusPrefix(), newTPKey)

	if err != nil {
		return jsonresult.GetLiquidateTpExchangeRates{}, err
	}

	var liquidateTopPercentileExchangeRatesStatus metadata.LiquidateTopPercentileExchangeRatesStatus
	err = json.Unmarshal(liquidateTpExchangeRates, &liquidateTopPercentileExchangeRatesStatus)
	if err != nil {
		return jsonresult.GetLiquidateTpExchangeRates{}, err
	}

	topPercentile, ok := liquidateTopPercentileExchangeRatesStatus.Rates[tokenSymbol]

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

func (portal *PortalService) GetLiquidateExchangeRatesPool(
	beaconHeight uint64,
	tokenSymbol string,
) (jsonresult.GetLiquidateExchangeRates, error) {
	portalStateDB := portal.BlockChain.BestState.Beacon.GetCopiedFeatureStateDB()
	liquidateExchangeRates, err := statedb.GetLiquidateExchangeRatesPoolByKey(portalStateDB, beaconHeight)

	if err != nil {
		return jsonresult.GetLiquidateExchangeRates{}, err
	}

	liquidateExchangeRatesDetail, ok := liquidateExchangeRates.Rates()[tokenSymbol]

	if !ok {
		return jsonresult.GetLiquidateExchangeRates{}, nil
	}

	result := jsonresult.GetLiquidateExchangeRates{
		TokenId: tokenSymbol,
		Liquidation: liquidateExchangeRatesDetail,
	}
	return result, nil
}
