package rpcservice

import (
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/database/lvdb"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)

type PortalExchangeRatesService struct {
	BlockChain *blockchain.BlockChain
}

func (portalExchangeRatesService *PortalExchangeRatesService) GetExchangeRates(senderAddress string, service *BlockService) (jsonresult.FinalExchangeRatesResult, *RPCError) {

	_, _, err := GetKeySetFromPaymentAddressParam(senderAddress)
	if err != nil {
		return jsonresult.FinalExchangeRatesResult{}, NewRPCError(InvalidSenderPrivateKeyError, err)
	}

	beaconBlock, err := service.GetBeaconBestBlock()
	if err != nil {
		return jsonresult.FinalExchangeRatesResult{}, NewRPCError(GetBeaconBestBlockError, err)
	}

	finalExchangeRatesKey := lvdb.NewFinalExchangeRatesKey(beaconBlock.GetHeight())
	finalExchangeRates, err := blockchain.GetFinalExchangeRatesByKey(portalExchangeRatesService.BlockChain.GetDatabase(), []byte(finalExchangeRatesKey))

	if err != nil {
		return  jsonresult.FinalExchangeRatesResult{}, NewRPCError(GetExchangeRatesError, err)
	}

	item := make(map[string]jsonresult.FinalExchangeRatesDetailResult)

	for pTokenId, rates := range finalExchangeRates.Rates {
		item[pTokenId] = jsonresult.FinalExchangeRatesDetailResult{
			Value: rates.Amount,
		}
	}

	result := jsonresult.FinalExchangeRatesResult{Rates:item}
	return result, nil
}

func (portalExchangeRatesService *PortalExchangeRatesService) ConvertExchangeRates(senderAddress string, valuePToken uint64, service *BlockService) (jsonresult.ExchangeRatesResult, *RPCError) {

	_, _, err := GetKeySetFromPaymentAddressParam(senderAddress)
	if err != nil {
		return jsonresult.ExchangeRatesResult{}, NewRPCError(InvalidSenderPrivateKeyError, err)
	}

	beaconBlock, err := service.GetBeaconBestBlock()
	if err != nil {
		return jsonresult.ExchangeRatesResult{}, NewRPCError(GetBeaconBestBlockError, err)
	}

	finalExchangeRatesKey := lvdb.NewFinalExchangeRatesKey(beaconBlock.GetHeight())
	finalExchangeRates, err := blockchain.GetFinalExchangeRatesByKey(portalExchangeRatesService.BlockChain.GetDatabase(), []byte(finalExchangeRatesKey))

	if err != nil {
		return jsonresult.ExchangeRatesResult{}, NewRPCError(GetExchangeRatesError, err)
	}

	if finalExchangeRates == nil {
		return jsonresult.ExchangeRatesResult{}, NewRPCError(GetExchangeRatesIsEmpty, err)
	}

	item := make(map[string]uint64)
	btcExchange := finalExchangeRates.ExchangeBTC2PRV(valuePToken)
	item["BTC"] = btcExchange
	bnbExchange := finalExchangeRates.ExchangeBNB2PRV(valuePToken)
	item["BNB"] = bnbExchange

	result := jsonresult.ExchangeRatesResult{Rates:item}
	return result, nil
}