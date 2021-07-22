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
	TradePath           []string            `json:"TradePath"`
	TokenToSell         common.Hash         `json:"TokenToSell"`
	SellAmount          uint64              `json:"SellAmount"`
	MinAcceptableAmount uint64              `json:"MinAcceptableAmount"`
	TradingFee          uint64              `json:"TradingFee"`
	Receiver            privacy.OTAReceiver `json:"Receiver"`
	RefundReceiver      privacy.OTAReceiver `json:"RefundReceiver"`
	metadataCommon.MetadataBase
}

func NewTradeRequest(
	tradePath []string,
	sellAmount uint64,
	minAcceptableAmount uint64,
	tradingFee uint64,
	recv, refundRecv privacy.OTAReceiver,
	metaType int,
) (*TradeRequest, error) {
	pdeTradeRequest := &TradeRequest{
		TradePath:           tradePath,
		SellAmount:          sellAmount,
		MinAcceptableAmount: minAcceptableAmount,
		TradingFee:          tradingFee,
		Receiver:            recv,
		RefundReceiver:      refundRecv,
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
	return req.Type == metadataCommon.PDexV3TradeRequestMeta
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
	sellingToken := common.ConfidentialAssetID
	if req.TokenToSell == common.PRVCoinID {
		sellingToken = common.PRVCoinID
	}
	result = append(result, metadataCommon.OTADeclaration{PublicKey: req.Receiver.PublicKey.ToBytes(), TokenID: sellingToken})
	result = append(result, metadataCommon.OTADeclaration{PublicKey: req.RefundReceiver.PublicKey.ToBytes(), TokenID: common.PRVCoinID})
	return result
}
