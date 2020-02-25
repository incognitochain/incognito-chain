package rpcservice

import (
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/database/lvdb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/pkg/errors"
)

type PortalExchangeRatesService struct {
	BlockChain *blockchain.BlockChain
}

func (portalExchangeRatesService *PortalExchangeRatesService) GetExchangeRates(service *BlockService, db database.DatabaseInterface) (jsonresult.FinalExchangeRatesResult, *RPCError) {
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

	result := jsonresult.FinalExchangeRatesResult{Rates:item}
	return result, nil
}

func (portalExchangeRatesService *PortalExchangeRatesService) ConvertExchangeRates(tokenSymbol string, valuePToken uint64, service *BlockService) (jsonresult.ExchangeRatesResult, *RPCError) {

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

	if tokenSymbol == metadata.PortalTokenSymbolBTC {
		btcExchange := finalExchangeRates.ExchangeBTC2PRV(valuePToken)
		item[metadata.PortalTokenSymbolBTC] = btcExchange
	} else if tokenSymbol == metadata.PortalTokenSymbolBNB {
		bnbExchange := finalExchangeRates.ExchangeBNB2PRV(valuePToken)
		item[metadata.PortalTokenSymbolBNB] = bnbExchange
	} else if tokenSymbol == metadata.PortalTokenSymbolPRV {
		return jsonresult.ExchangeRatesResult{}, NewRPCError(GetExchangeRatesIsEmpty, errors.New("PRV Token not support"))
	}

	result := jsonresult.ExchangeRatesResult{Rates:item}
	return result, nil
}