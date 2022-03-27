package bridge

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type RejectedUnshieldRequest struct {
	TokenID       common.Hash `json:"TokenID"`
	BurningAmount uint64      `json:"BurningAmount"`
}

type AcceptedUnshieldRequest struct {
	TokenID common.Hash                   `json:"TokenID"`
	TxReqID common.Hash                   `json:"TxReqID"`
	Data    []AcceptedUnshieldRequestData `json:"data"`
}

type AcceptedUnshieldRequestData struct {
	Amount        uint64 `json:"BurningAmount"`
	NetworkID     uint   `json:"NetworkID,omitempty"`
	Fee           uint64 `json:"Fee"`
	IsDepositToSC bool   `json:"IsDepositToSC"`
}

type UnshieldRequestData struct {
	BurningAmount  uint64 `json:"BurningAmount"`
	RemoteAddress  string `json:"RemoteAddress"`
	IsDepositToSC  bool   `json:"IsDepositToSC"`
	NetworkID      uint   `json:"NetworkID"`
	ExpectedAmount uint64 `json:"ExpectedAmount"`
}

type UnshieldRequest struct {
	TokenID common.Hash           `json:"TokenID"`
	Data    []UnshieldRequestData `json:"Data"`
	metadataCommon.MetadataBase
}

func NewUnshieldRequest() *UnshieldRequest {
	return &UnshieldRequest{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.UnshieldUnifiedTokenRequestMeta,
		},
	}
}

func NewUnshieldRequestWithValue(
	data []UnshieldRequestData,
) *UnshieldRequest {
	return &UnshieldRequest{
		Data: data,
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.UnshieldUnifiedTokenRequestMeta,
		},
	}
}

func (request *UnshieldRequest) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

func (request *UnshieldRequest) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {

	return true, true, nil
}

func (request *UnshieldRequest) ValidateMetadataByItself() bool {
	return request.Type == metadataCommon.UnshieldUnifiedTokenRequestMeta
}

func (request *UnshieldRequest) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(&request)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (request *UnshieldRequest) BuildReqActions(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	content, err := metadataCommon.NewActionWithValue(request, *tx.Hash(), nil).StringSlice(metadataCommon.UnshieldUnifiedTokenRequestMeta)
	return [][]string{content}, err
}

func (request *UnshieldRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(request)
}
