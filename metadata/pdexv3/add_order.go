package pdexv3

import (
	"encoding/json"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

// AddOrderRequest
type AddOrderRequest struct {
	TokenToSell         common.Hash                         `json:"TokenToSell"`
	PoolPairID          string                              `json:"PoolPairID"`
	SellAmount          uint64                              `json:"SellAmount"`
	MinAcceptableAmount uint64                              `json:"MinAcceptableAmount"`
	TradingFee          uint64                              `json:"TradingFee"`
	Receiver            map[common.Hash]privacy.OTAReceiver `json:"Receiver"`
	NftID               common.Hash                         `json:"NftID"`
	metadataCommon.MetadataBase
}

func NewAddOrderRequest(
	tokenToSell common.Hash,
	pairID string,
	sellAmount uint64,
	minAcceptableAmount uint64,
	tradingFee uint64,
	recv map[common.Hash]privacy.OTAReceiver,
	nftID common.Hash,
	metaType int,
) (*AddOrderRequest, error) {
	r := &AddOrderRequest{
		TokenToSell:         tokenToSell,
		PoolPairID:          pairID,
		SellAmount:          sellAmount,
		MinAcceptableAmount: minAcceptableAmount,
		TradingFee:          tradingFee,
		Receiver:            recv,
		NftID:               nftID,
		MetadataBase: metadataCommon.MetadataBase{
			Type: metaType,
		},
	}
	return r, nil
}

func (req AddOrderRequest) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

func (req AddOrderRequest) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {
	// OTAReceiver check
	for _, item := range req.Receiver {
		if !item.IsValid() {
			return false, false, metadataCommon.NewMetadataTxError(
				metadataCommon.PDEInvalidMetadataValueError, fmt.Errorf("Invalid OTAReceiver %v", item))
		}
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
	return true, true, nil
}

func (req AddOrderRequest) ValidateMetadataByItself() bool {
	return req.Type == metadataCommon.Pdexv3AddOrderRequestMeta
}

func (req AddOrderRequest) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(req)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (req *AddOrderRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(req)
}

func (req *AddOrderRequest) GetOTADeclarations() []metadataCommon.OTADeclaration {
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
