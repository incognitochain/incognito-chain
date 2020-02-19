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