package rpcservice

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)

type PortalService struct {
	BlockChain *blockchain.BlockChain
}

func (s *PortalService) GetPortingRequestByByTxID(txId string) (jsonresult.PortalPortingRequest, error) {
	portalStateDB := s.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
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

func (s *PortalService) GetPortingRequestByByPortingId(portingId string) (jsonresult.PortalPortingRequest, error) {
	portalStateDB := s.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
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

func (s *PortalService) GetCustodianWithdrawRequestStatusByTxId(txId string) (jsonresult.PortalCustodianWithdrawRequest, error) {
	portalStateDB := s.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
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

func (s *PortalService) GetFinalExchangeRates(stateDB *statedb.StateDB, beaconHeight uint64) (jsonresult.FinalExchangeRatesResult, error) {
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

func (s *PortalService) ConvertExchangeRates(
	finalExchangeRates *statedb.FinalExchangeRatesState, portalParams blockchain.PortalParams,
	amount uint64, tokenIDFrom string, tokenIDTo string) (uint64, error) {
	result := uint64(0)
	var err error
	exchangeTool := blockchain.NewPortalExchangeRateTool(finalExchangeRates, portalParams.SupportedCollateralTokens)
	if tokenIDTo != "" && tokenIDFrom != "" {
		result, err = exchangeTool.Convert(tokenIDFrom, tokenIDTo, amount)
	} else if tokenIDTo == "" {
		result, err = exchangeTool.ConvertToUSDT(tokenIDFrom, amount)
	} else if tokenIDFrom == "" {
		result, err = exchangeTool.ConvertFromUSDT(tokenIDTo, amount)
	}

	return result, err
}

func (s *PortalService) GetPortingFees(finalExchangeRates *statedb.FinalExchangeRatesState, tokenID string, valuePToken uint64, portalParam blockchain.PortalParams) (uint64, error) {
	return blockchain.CalMinPortingFee(valuePToken, tokenID, finalExchangeRates, portalParam)
}

func (s *PortalService) CalculateTopupAmountForCustodianState(
	stateDB *statedb.StateDB,
	custodianAddress string,
	portalTokenID string,
	collateralTokenID string,
	portalParam blockchain.PortalParams) (uint64, error) {
	custodian, err := statedb.GetCustodianByIncAddress(stateDB, custodianAddress)
	if err != nil {
		return 0, err
	}

	finalExchangeRates, err := statedb.GetFinalExchangeRatesState(stateDB)
	if err != nil {
		return 0, err
	}

	currentPortalState, err := blockchain.InitCurrentPortalStateFromDB(stateDB)
	if err != nil {
		return 0, err
	}

	topupAmount, err := blockchain.CalTopupAmountForCustodianState(currentPortalState, custodian, finalExchangeRates, portalTokenID, collateralTokenID, portalParam)

	return topupAmount, nil
}

func (s *PortalService) GetLiquidateExchangeRatesPool(
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

func (s *PortalService) GetCustodianWithdrawRequestStatusV3ByTxId(txId string) (metadata.CustodianWithdrawRequestStatusV3, error) {
	var res metadata.CustodianWithdrawRequestStatusV3

	portalStateDB := s.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
	statusBytes, err := statedb.GetPortalCustodianWithdrawCollateralStatusV3(portalStateDB, txId)
	if err != nil || statusBytes == nil || len(statusBytes) == 0 {
		return res, NewRPCError(GetCustodianWithdrawError, err)
	}
	err = json.Unmarshal(statusBytes, &res)
	return res, err
}

func (s *PortalService) GetWithdrawCollateralConfirm(txID common.Hash) (uint64, error) {
	stateDB := s.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
	return statedb.GetWithdrawCollateralConfirmProof(stateDB, txID)
}
