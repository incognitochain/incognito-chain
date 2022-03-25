package bridgeagg

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

type UnshieldRequestData struct {
	BurningAmount  uint64      `json:"BurningAmount"` // must be equal to vout value
	TokenID        common.Hash `json:"TokenID"`
	RemoteAddress  string      `json:"RemoteAddress"`
	IsDepositToSC  bool        `json:"IsDepositToSC"`
	NetworkID      uint        `json:"NetworkID"`
	ExpectedAmount uint64      `json:"ExpectedAmount"`
}

type UnshieldRequest struct {
	BurnerAddress privacy.PaymentAddress `json:"BurnerAddress"`
	Data          []UnshieldRequestData  `json:"Data"`
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
	data []UnshieldRequestData, burnerAddress privacy.PaymentAddress,
) *UnshieldRequest {
	return &UnshieldRequest{
		Data:          data,
		BurnerAddress: burnerAddress,
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
