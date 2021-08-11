package pdexv3

import (
	"encoding/json"

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
	return true, nil
}

func (req TradeRequest) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {
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
