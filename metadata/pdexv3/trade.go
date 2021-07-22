package pdexv3

import (
	"encoding/json"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

// TradeRequest
type TradeRequest struct {
	TradePath           []common.Hash       `json:"TradePath"`
	SellAmount          uint64              `json:"SellAmount"`
	MinAcceptableAmount uint64              `json:"MinAcceptableAmount"`
	TradingFee          uint64              `json:"TradingFee"`
	Receiver            privacy.OTAReceiver `json:"Receiver"`
	RefundReceiver      privacy.OTAReceiver `json:"RefundReceiver"`
	metadataCommon.MetadataBase
}

type TradeAction struct {
	Metadata    TradeRequest `json:"Metadata"`
	ShardID     byte         `json:"ShardID"`
	RequestTxID common.Hash  `json:"RequestTxID"`
}

type AcceptedTrade struct {
	Receiver      privacy.OTAReceiver `json:"Receiver"`
	ReceiveAmount uint64              `json:"ReceiveAmount"`
	TokenToBuy    common.Hash         `json:"TokenToBuy"`
	PairID        string              `json:"PairID"`
	Token0Change  big.Int             `json:"Token0Change"`
	Token1Change  big.Int             `json:"Token1Change"`
	ShardID       byte                `json:"ShardID"`
	RequestTxID   common.Hash         `json:"RequestTxID"`
}

type RefundedTrade struct {
	Receiver    privacy.OTAReceiver `json:"Receiver"`
	TokenToSell common.Hash         `json:"TokenToSell"`
	Amount      uint64              `json:"Amount"`
	ShardID     byte                `json:"ShardID"`
	RequestTxID common.Hash         `json:"RequestTxID"`
}

func NewTradeRequest(
	tradePath []common.Hash,
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

func (req *TradeRequest) GetOTADeclarations() []metadataCommon.OTADeclaration {
	var result []metadataCommon.OTADeclaration
	sellingToken := common.ConfidentialAssetID
	if len(req.TradePath) < 1 {
		// this would be an invalid request
		return nil
	}
	if req.TradePath[0] == common.PRVCoinID {
		sellingToken = common.PRVCoinID
	}
	result = append(result, metadataCommon.OTADeclaration{PublicKey: req.Receiver.PublicKey.ToBytes(), TokenID: sellingToken})
	result = append(result, metadataCommon.OTADeclaration{PublicKey: req.RefundReceiver.PublicKey.ToBytes(), TokenID: common.PRVCoinID})
	return result
}
