package pdexv3

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

// AddOrderResponse
type AddOrderResponse struct {
	Status      string      `json:"Status"`
	RequestTxID common.Hash `json:"RequestTxID"`
	metadataCommon.MetadataBase
}

type AcceptedAddOrder struct {
	TokenToBuy          common.Hash `json:"TokenToBuy"`
	PairID              string      `json:"PairID"`
	SellAmount          uint64      `json:"SellAmount"`
	MinAcceptableAmount uint64      `json:"MinAcceptableAmount"`
	ShardID             byte        `json:"ShardID"`
	RequestTxID         common.Hash `json:"RequestTxID"`
}

func (md AcceptedAddOrder) GetType() int {
	return metadataCommon.PDexV3AddOrderRequestMeta
}

func (md AcceptedAddOrder) GetStatus() string {
	return OrderAcceptedStatus
}

type RefundedAddOrder struct {
	Receiver    privacy.OTAReceiver `json:"Receiver"`
	TokenToSell common.Hash         `json:"TokenToSell"`
	Amount      uint64              `json:"Amount"`
	ShardID     byte                `json:"ShardID"`
	RequestTxID common.Hash         `json:"RequestTxID"`
}

func (md RefundedAddOrder) GetType() int {
	return metadataCommon.PDexV3AddOrderRequestMeta
}

func (md RefundedAddOrder) GetStatus() string {
	return OrderRefundedStatus
}

func (res AddOrderResponse) CheckTransactionFee(tx metadataCommon.Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (res AddOrderResponse) VerifyMinerCreatedTxBeforeGettingInBlock(mintData *metadataCommon.MintData, shardID byte, tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, ac *metadataCommon.AccumulatedValues, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever) (bool, error) {
	return true, nil
}

func (res AddOrderResponse) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

func (res AddOrderResponse) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {
	return true, true, nil
}

func (res AddOrderResponse) ValidateMetadataByItself() bool {
	return res.Type == metadataCommon.PDexV3AddOrderResponseMeta
}

func (res AddOrderResponse) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(res)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (res *AddOrderResponse) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(res)
}

