package bridgeagg

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

type ConvertTokenToUnifiedTokenRequest struct {
	tokenID        common.Hash
	unifiedTokenID common.Hash
	amount         uint64
	receivers      map[common.Hash]privacy.OTAReceiver
	metadataCommon.MetadataBase
}

func NewConvertTokenToUnifiedTokenRequest() *ConvertTokenToUnifiedTokenRequest {
	return &ConvertTokenToUnifiedTokenRequest{}
}

func NewConvertTokenToUnifiedTokenRequestWithValue(
	tokenID, unifiedTokenID common.Hash, amount uint64,
) *ConvertTokenToUnifiedTokenRequest {
	metadataBase := metadataCommon.MetadataBase{
		Type: metadataCommon.BridgeAggConvertTokenToUnifiedTokenRequestMeta,
	}
	return &ConvertTokenToUnifiedTokenRequest{
		unifiedTokenID: unifiedTokenID,
		tokenID:        tokenID,
		amount:         amount,
		MetadataBase:   metadataBase,
	}
}

func (request *ConvertTokenToUnifiedTokenRequest) ValidateTxWithBlockChain(
	tx metadataCommon.Transaction,
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (bool, error) {
	return true, nil
}

func (request *ConvertTokenToUnifiedTokenRequest) ValidateSanityData(
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	beaconHeight uint64,
	tx metadataCommon.Transaction,
) (bool, bool, error) {
	return true, true, nil
}

func (request *ConvertTokenToUnifiedTokenRequest) ValidateMetadataByItself() bool {
	return request.Type == metadataCommon.BridgeAggConvertTokenToUnifiedTokenRequestMeta
}

func (request *ConvertTokenToUnifiedTokenRequest) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(&request)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (request *ConvertTokenToUnifiedTokenRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(request)
}

func (request *ConvertTokenToUnifiedTokenRequest) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		TokenID string `json:"TokenID"`
		metadataCommon.MetadataBase
	}{
		TokenID:      request.tokenID,
		MetadataBase: request.MetadataBase,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (request *ConvertTokenToUnifiedTokenRequest) UnmarshalJSON(data []byte) error {
	temp := struct {
		TokenID common.Hash `json:"TokenID"`
		metadataCommon.MetadataBase
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	request.tokenID = temp.TokenID
	request.MetadataBase = temp.MetadataBase
	return nil
}

func (request *ConvertTokenToUnifiedTokenRequest) TokenID() common.Hash {
	return request.tokenID
}

func (request *ConvertTokenToUnifiedTokenRequest) UnifiedTokenID() common.Hash {
	return request.unifiedTokenID
}

func (request *ConvertTokenToUnifiedTokenRequest) Amount() uint64 {
	return request.amount
}

func (request *ConvertTokenToUnifiedTokenRequest) GetOTADeclarations() []metadataCommon.OTADeclaration {
	var result []metadataCommon.OTADeclaration
	for _, val := range request.receivers {
		result = append(result, metadataCommon.OTADeclaration{
			PublicKey: val.PublicKey.ToBytes(), TokenID: common.ConfidentialAssetID,
		})
	}
	return result
}
