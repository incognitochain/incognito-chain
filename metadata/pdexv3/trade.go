package pdexv3

import (
	"encoding/json"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

// TradeRequest
type TradeRequest struct {
	TradePath           []string                            `json:"TradePath"`
	TokenToSell         common.Hash                         `json:"TokenToSell"`
	SellAmount          uint64                              `json:"SellAmount"`
	MinAcceptableAmount uint64                              `json:"MinAcceptableAmount"`
	TradingFee          uint64                              `json:"TradingFee"`
	Receiver            map[common.Hash]privacy.OTAReceiver `json:"Receiver"`
	metadataCommon.MetadataBase
}

func NewTradeRequest(
	tradePath []string,
	tokenToSell common.Hash,
	sellAmount uint64,
	minAcceptableAmount uint64,
	tradingFee uint64,
	recv map[common.Hash]privacy.OTAReceiver,
	metaType int,
) (*TradeRequest, error) {
	pdeTradeRequest := &TradeRequest{
		TradePath:           tradePath,
		TokenToSell:         tokenToSell,
		SellAmount:          sellAmount,
		MinAcceptableAmount: minAcceptableAmount,
		TradingFee:          tradingFee,
		Receiver:            recv,
		MetadataBase: metadataCommon.MetadataBase{
			Type: metaType,
		},
	}
	return pdeTradeRequest, nil
}

func (req TradeRequest) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	if !chainRetriever.IsAfterPdexv3CheckPoint(beaconViewRetriever.GetHeight()) {
		return false, fmt.Errorf("Feature pdexv3 has not been activated yet")
	}
	pdexv3StateCached := chainRetriever.GetPdexv3Cached(beaconViewRetriever.BlockHash())
	for _, poolPairID := range req.TradePath {
		err := beaconViewRetriever.IsValidPoolPairID(chainRetriever.GetBeaconChainDatabase(), pdexv3StateCached, poolPairID)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func (req TradeRequest) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {
	if !chainRetriever.IsAfterPdexv3CheckPoint(beaconHeight) {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, fmt.Errorf("Feature pdexv3 has not been activated yet"))
	}

	// OTAReceiver check
	for _, item := range req.Receiver {
		if !item.IsValid() {
			return false, false, metadataCommon.NewMetadataTxError(
				metadataCommon.PDEInvalidMetadataValueError, fmt.Errorf("Invalid OTAReceiver %v", item))
		}
		if tx.GetSenderAddrLastByte() != item.GetShardID() {
			return false, false, metadataCommon.NewMetadataTxError(
				metadataCommon.PDEInvalidMetadataValueError,
				fmt.Errorf("Invalid shard %d for Receiver - must equal sender shard",
					item.GetShardID()))
		}
	}

	if req.SellAmount == 0 {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError,
			fmt.Errorf("SellAmount cannot be 0"))
	}

	// Burned coin check
	isBurn, burnedPRVCoin, burnedCoin, burnedTokenID, err := tx.GetTxFullBurnData()
	if err != nil || !isBurn {
		return false, false, metadataCommon.NewMetadataTxError(
			metadataCommon.PDEInvalidMetadataValueError,
			fmt.Errorf("Burned coins not found in trade request - %v", err))
	}
	if *burnedTokenID != req.TokenToSell {
		return false, false, metadataCommon.NewMetadataTxError(
			metadataCommon.PDEInvalidMetadataValueError,
			fmt.Errorf("Burned token ID mismatch - %v vs %v on metadata", *burnedTokenID, req.TokenToSell))
	}
	burnedTokenList := []common.Hash{*burnedTokenID}
	if burnedPRVCoin != nil {
		burnedTokenList = append(burnedTokenList, common.PRVCoinID)
	}
	for _, tokenID := range burnedTokenList {
		_, exists := req.Receiver[tokenID]
		if !exists {
			return false, false, metadataCommon.NewMetadataTxError(
				metadataCommon.PDEInvalidMetadataValueError,
				fmt.Errorf("Missing refund OTAReceiver for token %v", tokenID))
		}
	}

	// Type vs burned token id + amount check
	switch tx.GetType() {
	case common.TxNormalType:
		// PRV must be burned
		if req.TokenToSell != common.PRVCoinID || burnedPRVCoin == nil {
			return false, false, metadataCommon.NewMetadataTxError(
				metadataCommon.PDEInvalidMetadataValueError,
				fmt.Errorf("Burned token invalid - must be PRV"))
		}
		// range check before adding
		if req.SellAmount > burnedPRVCoin.GetValue() {
			return false, false, metadataCommon.NewMetadataTxError(
				metadataCommon.PDEInvalidMetadataValueError,
				fmt.Errorf("Sell amount invalid - must not exceed %d PRV burned", burnedPRVCoin.GetValue()))
		}
		if req.SellAmount+req.TradingFee != burnedPRVCoin.GetValue() {
			return false, false, metadataCommon.NewMetadataTxError(
				metadataCommon.PDEInvalidMetadataValueError,
				fmt.Errorf("Sell amount invalid - must equal burned amount %d PRV after fee",
					burnedPRVCoin.GetValue()))
		}
	case common.TxCustomTokenPrivacyType:
		if req.TokenToSell != *burnedTokenID || burnedCoin == nil {
			return false, false, metadataCommon.NewMetadataTxError(
				metadataCommon.PDEInvalidMetadataValueError,
				fmt.Errorf("Burned token invalid - must be %v", req.TokenToSell))
		}
		if burnedPRVCoin == nil {
			// pay fee with the same token
			// range check before adding
			if req.SellAmount > burnedCoin.GetValue() {
				return false, false, metadataCommon.NewMetadataTxError(
					metadataCommon.PDEInvalidMetadataValueError,
					fmt.Errorf("Sell amount invalid - must not exceed %d burned", burnedCoin.GetValue()))
			}
			if req.SellAmount+req.TradingFee != burnedCoin.GetValue() {
				return false, false, metadataCommon.NewMetadataTxError(
					metadataCommon.PDEInvalidMetadataValueError,
					fmt.Errorf("Sell amount invalid - must equal burned amount %d after fee",
						burnedCoin.GetValue()))
			}
		} else {
			// pay fee using PRV
			if req.TradingFee != burnedPRVCoin.GetValue() {
				return false, false, metadataCommon.NewMetadataTxError(
					metadataCommon.PDEInvalidMetadataValueError,
					fmt.Errorf("Trading fee in PRV invalid - must equal burned amount %d",
						burnedPRVCoin.GetValue()))
			}
			if req.SellAmount != burnedCoin.GetValue() {
				return false, false, metadataCommon.NewMetadataTxError(
					metadataCommon.PDEInvalidMetadataValueError,
					fmt.Errorf("Sell amount in - must equal burned amount %d of token %v",
						burnedCoin.GetValue(), *burnedTokenID))
			}
		}
	default:
		return false, false, fmt.Errorf("Invalid transaction type %v for trade request", tx.GetType())
	}

	if len(req.TradePath) == 0 {
		return false, false, fmt.Errorf("Invalid trade request - path empty")
	}

	// trade path length check
	if len(req.TradePath) > MaxTradePathLength {
		return false, false, fmt.Errorf("Trade path length of %d exceeds maximum", len(req.TradePath))
	}
	return true, true, nil
}

func (req TradeRequest) ValidateMetadataByItself() bool {
	return req.Type == metadataCommon.Pdexv3TradeRequestMeta
}

func (req TradeRequest) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(req)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (req *TradeRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(req)
}

func (req TradeRequest) GetOTADeclarations() []metadataCommon.OTADeclaration {
	var result []metadataCommon.OTADeclaration
	for currentTokenID, val := range req.Receiver {
		if currentTokenID != common.PRVCoinID {
			currentTokenID = common.ConfidentialAssetID
		}
		result = append(result, metadataCommon.OTADeclaration{
			PublicKey: val.PublicKey.ToBytes(), TokenID: currentTokenID,
		})
	}
	return result
}
