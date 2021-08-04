package pdexv3

import (
	"encoding/json"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

// TradeStatus containns the info tracked by feature statedb, which is then displayed in RPC status queries.
// For refunded trade, all fields except Status are ignored
type TradeStatus struct {
	Status     int         `json:"Status"`
	BuyAmount  uint64      `json:"BuyAmount"`
	TokenToBuy common.Hash `json:"TokenToBuy"`
}

// TradeResponse is the metadata inside response tx for trade
type TradeResponse struct {
	Status      int         `json:"Status"`
	RequestTxID common.Hash `json:"RequestTxID"`
	metadataCommon.MetadataBase
}

// AcceptedTrade is added as Content for produced beacon Instructions after handling a trade successfully
type AcceptedTrade struct {
	Receiver     privacy.OTAReceiver      `json:"Receiver"`
	Amount       uint64                   `json:"Amount"`
	TradePath    []string                 `json:"TradePath"`
	TokenToBuy   common.Hash              `json:"TokenToBuy"`
	PairChanges  [][2]*big.Int            `json:"PairChanges"`
	OrderChanges []map[string][2]*big.Int `json:"OrderChanges"`
}

func (md AcceptedTrade) GetType() int {
	return metadataCommon.Pdexv3TradeRequestMeta
}

func (md AcceptedTrade) GetStatus() int {
	return TradeAcceptedStatus
}

// RefundedTrade is added as Content for produced beacon instruction after failure to handle a trade
type RefundedTrade struct {
	Receiver    privacy.OTAReceiver `json:"Receiver"`
	TokenToSell common.Hash         `json:"TokenToSell"`
	Amount      uint64              `json:"Amount"`
}

func (md RefundedTrade) GetType() int {
	return metadataCommon.Pdexv3TradeRequestMeta
}

func (md RefundedTrade) GetStatus() int {
	return TradeRefundedStatus
}

func (res TradeResponse) CheckTransactionFee(tx metadataCommon.Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (res TradeResponse) VerifyMinerCreatedTxBeforeGettingInBlock(mintData *metadataCommon.MintData, shardID byte, tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, ac *metadataCommon.AccumulatedValues, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever) (bool, error) {
	return true, nil
}

func (res TradeResponse) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

func (res TradeResponse) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {
	return true, true, nil
}

func (res TradeResponse) ValidateMetadataByItself() bool {
	return res.Type == metadataCommon.Pdexv3TradeResponseMeta
}

func (res TradeResponse) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(res)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (res *TradeResponse) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(res)
}
