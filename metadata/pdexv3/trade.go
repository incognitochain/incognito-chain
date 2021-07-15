package metadata

import (
	"encoding/base64"
	"encoding/json"
	"math/big"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

// TradeRequest
type TradeRequest struct {
	TokenToSell         common.Hash         `json:"TokenToSell"`
	TokenToBuy          common.Hash         `json:"TokenToBuy"`
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
	ShardID     byte                `json:"ShardID`
	RequestTxID common.Hash         `json:"RequestTxID"`
}

func NewTradeRequest(
	tokenToBuy common.Hash,
	tokenToSell common.Hash,
	sellAmount uint64,
	minAcceptableAmount uint64,
	tradingFee uint64,
	recv, refundRecv privacy.OTAReceiver,
	metaType int,
) (*TradeRequest, error) {
	pdeTradeRequest := &TradeRequest{
		TokenToSell:         tokenToSell,
		TokenToBuy:          tokenToBuy,
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

func (req *TradeRequest) BuildReqActions(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	action := TradeAction{
		Metadata:    *req,
		ShardID:     shardID,
		RequestTxID: *tx.Hash(),
	}
	b, err := json.Marshal(action)
	if err != nil {
		return [][]string{}, err
	}
	actionEncoded := base64.StdEncoding.EncodeToString(b)
	return [][]string{
		[]string{
			strconv.Itoa(metadataCommon.PDexV3TradeRequestMeta),
			actionEncoded,
		},
	}, nil
}

func (req *TradeRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(req)
}

func (req *TradeRequest) GetOTADeclarations() []metadataCommon.OTADeclaration {
	var result []metadataCommon.OTADeclaration
	sellingToken := common.ConfidentialAssetID
	if req.TokenToSell == common.PRVCoinID {
		sellingToken = common.PRVCoinID
	}
	result = append(result, metadataCommon.OTADeclaration{PublicKey: req.Receiver.PublicKey.ToBytes(), TokenID: sellingToken})
	result = append(result, metadataCommon.OTADeclaration{PublicKey: req.RefundReceiver.PublicKey.ToBytes(), TokenID: common.PRVCoinID})
	return result
}
