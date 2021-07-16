package metadata

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

// AddOrderRequest
type AddOrderRequest struct {
	TokenToSell         common.Hash         `json:"TokenToSell"`
	TokenToBuy          common.Hash         `json:"TokenToBuy"`
	PairID              string              `json:"PairID"`
	SellAmount          uint64              `json:"SellAmount"`
	MinAcceptableAmount uint64              `json:"MinAcceptableAmount"`
	TradingFee          uint64              `json:"TradingFee"`
	RefundReceiver      privacy.OTAReceiver `json:"RefundReceiver"`
	metadataCommon.MetadataBase
}

type AddOrderAction struct {
	Metadata    AddOrderRequest `json:"Metadata"`
	ShardID     byte            `json:"ShardID"`
	RequestTxID common.Hash     `json:"RequestTxID"`
}

type AcceptedAddOrder struct {
	TokenToBuy          common.Hash `json:"TokenToBuy"`
	PairID              string      `json:"PairID"`
	SellAmount          uint64      `json:"SellAmount"`
	MinAcceptableAmount uint64      `json:"MinAcceptableAmount"`
	ShardID             byte        `json:"ShardID"`
	RequestTxID         common.Hash `json:"RequestTxID"`
}

type RefundedAddOrder struct {
	Receiver    privacy.OTAReceiver `json:"Receiver"`
	TokenToSell common.Hash         `json:"TokenToSell"`
	Amount      uint64              `json:"Amount"`
	ShardID     byte                `json:"ShardID`
	RequestTxID common.Hash         `json:"RequestTxID"`
}

func NewAddOrderRequest(
	tokenToBuy common.Hash,
	tokenToSell common.Hash,
	pairID string,
	sellAmount uint64,
	minAcceptableAmount uint64,
	tradingFee uint64,
	refundRecv privacy.OTAReceiver,
	metaType int,
) (*AddOrderRequest, error) {
	pdeTradeRequest := &AddOrderRequest{
		TokenToSell:         tokenToSell,
		TokenToBuy:          tokenToBuy,
		PairID:              pairID,
		SellAmount:          sellAmount,
		MinAcceptableAmount: minAcceptableAmount,
		TradingFee:          tradingFee,
		RefundReceiver:      refundRecv,
		MetadataBase: metadataCommon.MetadataBase{
			Type: metaType,
		},
	}
	return pdeTradeRequest, nil
}

func (req AddOrderRequest) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

func (req AddOrderRequest) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {
	return true, true, nil
}

func (req AddOrderRequest) ValidateMetadataByItself() bool {
	return req.Type == metadataCommon.PDexV3AddOrderRequestMeta
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
	sellingToken := common.ConfidentialAssetID
	if req.TokenToSell == common.PRVCoinID {
		sellingToken = common.PRVCoinID
	}
	result = append(result, metadataCommon.OTADeclaration{PublicKey: req.RefundReceiver.PublicKey.ToBytes(), TokenID: sellingToken})
	return result
}
