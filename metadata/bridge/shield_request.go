package bridge

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type AcceptedShieldRequest struct {
	Receiver   string                      `json:"Receiver"`
	IncTokenID common.Hash                 `json:"IncTokenID"`
	TxReqID    common.Hash                 `json:"TxReqID"`
	IsReward   bool                        `json:"IsReward"`
	Data       []AcceptedShieldRequestData `json:"Data"`
}

type AcceptedShieldRequestData struct {
	IssuingAmount   uint64 `json:"IssuingAmount"`
	UniqTx          []byte `json:"UniqTx"`
	ExternalTokenID []byte `json:"ExternalTokenID"`
	NetworkID       uint   `json:"NetworkID"`
}

type ShieldRequestData struct {
	BlockHash []byte `json:"BlockHash"`
	TxIndex   uint   `json:"TxIndex"`
	Proof     []byte `json:"Proof"`
	NetworkID uint   `json:"NetworkID"`
}

type ShieldRequest struct {
	Data       []ShieldRequestData `json:"Data"`
	IncTokenID common.Hash         `json:"IncTokenID"`
	metadataCommon.MetadataBase
}

func NewShieldRequest() *ShieldRequest {
	return &ShieldRequest{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.ShieldUnifiedTokenRequestMeta,
		},
	}
}

func NewShieldRequestWithValue(
	data []ShieldRequestData, incTokenID common.Hash,
) *ShieldRequest {
	return &ShieldRequest{
		Data:       data,
		IncTokenID: incTokenID,
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.ShieldUnifiedTokenRequestMeta,
		},
	}
}

func (request *ShieldRequest) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

func (request *ShieldRequest) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {

	return true, true, nil
}

func (request *ShieldRequest) ValidateMetadataByItself() bool {
	return request.Type == metadataCommon.ShieldUnifiedTokenRequestMeta
}

func (request *ShieldRequest) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(&request)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (request *ShieldRequest) BuildReqActions(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	content, err := metadataCommon.NewActionWithValue(request, *tx.Hash(), nil).StringSlice(metadataCommon.ShieldUnifiedTokenRequestMeta)
	return [][]string{content}, err
}

func (request *ShieldRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(request)
}
