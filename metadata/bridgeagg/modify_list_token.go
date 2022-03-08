package bridgeagg

import (
	"encoding/base64"
	"encoding/json"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type ModifyListToken struct {
	NewListTokens map[common.Hash][]common.Hash `json:"NewListTokens"` // unifiedTokenID -> list tokenID
	metadataCommon.MetadataBase
}

func NewModifyListToken() *ModifyListToken {
	return &ModifyListToken{}
}

func NewModifyListTokenWithValue(
	newListTokens map[common.Hash][]common.Hash,
) *ModifyListToken {
	metadataBase := metadataCommon.MetadataBase{
		Type: metadataCommon.BridgeAggModifyListTokenMeta,
	}
	return &ModifyListToken{
		NewListTokens: newListTokens,
		MetadataBase:  metadataBase,
	}
}

func (request *ModifyListToken) ValidateTxWithBlockChain(
	tx metadataCommon.Transaction,
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (bool, error) {
	return true, nil
}

func (request *ModifyListToken) ValidateSanityData(
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	beaconHeight uint64,
	tx metadataCommon.Transaction,
) (bool, bool, error) {
	return true, true, nil
}

func (request *ModifyListToken) ValidateMetadataByItself() bool {
	return request.Type == metadataCommon.BridgeAggModifyListTokenMeta
}

func (request *ModifyListToken) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(&request)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (request *ModifyListToken) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(request)
}

func (request *ModifyListToken) BuildReqActions(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	content := metadataCommon.Action{
		Meta:    request,
		TxReqID: *(tx.Hash()),
	}
	contentBytes, err := json.Marshal(content)
	if err != nil {
		return [][]string{}, err
	}
	contentStr := base64.StdEncoding.EncodeToString(contentBytes)
	action := []string{strconv.Itoa(metadataCommon.BridgeAggModifyListTokenMeta), contentStr}
	return [][]string{action}, nil
}
