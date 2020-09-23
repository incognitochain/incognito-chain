package rpcservice

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)

type PortalService struct {
	BlockChain *blockchain.BlockChain
}

func (portal *PortalService) GetPortingRequestByByTxID(txId string) (jsonresult.PortalPortingRequest, error) {
	portalStateDB := portal.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
	portingRequestItem, err := statedb.GetPortalPortingRequestByTxIDStatus(portalStateDB, txId)

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
	portalStateDB := portal.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
	portingRequestItem, err := statedb.GetPortalPortingRequestStatus(portalStateDB, portingId)

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

func (portal *PortalService) GetCustodianWithdrawRequestStatusByTxId(txId string) (jsonresult.PortalCustodianWithdrawRequest, error) {
	portalStateDB := portal.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
	custodianWithdraw, err := statedb.GetPortalCustodianWithdrawCollateralStatus(portalStateDB, txId)

	if err != nil {
		return jsonresult.PortalCustodianWithdrawRequest{}, NewRPCError(GetCustodianWithdrawError, err)
	}

	if custodianWithdraw == nil {
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

func (portal *PortalService) GetFinalExchangeRates(stateDB *statedb.StateDB, beaconHeight uint64) (jsonresult.FinalExchangeRatesResult, error) {
	finalExchangeRates, err := statedb.GetFinalExchangeRatesState(stateDB)

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
		Rates:        item,
	}
	return result, nil
}

func (portal *PortalService) ConvertExchangeRates(stateDB *statedb.StateDB, tokenID string, valuePToken uint64) (map[string]uint64, error) {
	result := make(map[string]uint64)
	finalExchangeRates, err := statedb.GetFinalExchangeRatesState(stateDB)
	if err != nil {
		return result, err
	}

	finalExchangeRatesObj := blockchain.NewConvertExchangeRatesObject(finalExchangeRates)
	exchange, err := finalExchangeRatesObj.ExchangePToken2PRVByTokenId(tokenID, valuePToken)

	if err != nil {
		return result, err
	}
	result[tokenID] = exchange

	return result, nil
}

func (portal *PortalService) GetPortingFees(stateDB *statedb.StateDB, tokenID string, valuePToken uint64, portalParam blockchain.PortalParams) (map[string]uint64, error) {
	result := make(map[string]uint64)
	finalExchangeRates, err := statedb.GetFinalExchangeRatesState(stateDB)
	if err != nil {
		return result, err
	}

	exchangePortingFees, err := blockchain.CalMinPortingFee(valuePToken, tokenID, finalExchangeRates, portalParam.MinPercentPortingFee)

	if err != nil {
		return result, err
	}

	result[tokenID] = exchangePortingFees

	return result, nil
}

func (portal *PortalService) CalculateAmountNeededCustodianDepositLiquidation(
	stateDB *statedb.StateDB,
	custodianAddress string,
	pTokenId string,
	portalParam blockchain.PortalParams) (jsonresult.GetLiquidateAmountNeededCustodianDeposit, error) {
	custodian, err := statedb.GetOneCustodian(stateDB, custodianAddress)
	if err != nil {
		return jsonresult.GetLiquidateAmountNeededCustodianDeposit{}, err
	}

	finalExchangeRates, err := statedb.GetFinalExchangeRatesState(stateDB)
	if err != nil {
		return jsonresult.GetLiquidateAmountNeededCustodianDeposit{}, err
	}

	currentPortalState, err := blockchain.InitCurrentPortalStateFromDB(stateDB)
	if err != nil {
		return jsonresult.GetLiquidateAmountNeededCustodianDeposit{}, err
	}

	amountNeeded, err := blockchain.CalAmountNeededDepositLiquidate(currentPortalState, custodian, finalExchangeRates, pTokenId, portalParam)

	result := jsonresult.GetLiquidateAmountNeededCustodianDeposit{
		Amount:  amountNeeded,
		TokenId: pTokenId,
	}

	return result, nil
}

func (portal *PortalService) GetLiquidateExchangeRatesPool(
	stateDB *statedb.StateDB,
	tokenID string,
) (jsonresult.GetLiquidateExchangeRates, error) {
	liquidateExchangeRates, err := statedb.GetLiquidateExchangeRatesPoolByKey(stateDB)
	if err != nil {
		Logger.log.Errorf("Error when getting liquidation pool %v", err)
		return jsonresult.GetLiquidateExchangeRates{}, nil
	}

	liquidateExchangeRatesDetail, ok := liquidateExchangeRates.Rates()[tokenID]

	if !ok {
		return jsonresult.GetLiquidateExchangeRates{}, nil
	}

	result := jsonresult.GetLiquidateExchangeRates{
		TokenId:     tokenID,
		Liquidation: liquidateExchangeRatesDetail,
	}
	return result, nil
}

func (portal *PortalService) GetCustodianWithdrawRequestStatusV3ByTxId(txId string) (metadata.CustodianWithdrawRequestStatusV3, error) {
	var res metadata.CustodianWithdrawRequestStatusV3

	portalStateDB := portal.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
	statusBytes, err := statedb.GetPortalCustodianWithdrawCollateralStatusV3(portalStateDB, txId)
	if err != nil || statusBytes == nil || len(statusBytes) == 0 {
		return res, NewRPCError(GetCustodianWithdrawError, err)
	}
	err = json.Unmarshal(statusBytes, &res)
	return res, err
}
