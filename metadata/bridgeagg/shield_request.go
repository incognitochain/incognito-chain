package bridgeagg

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type ShieldData struct {
	BlockHash []byte `json:"BlockHash"`
	TxIndex   uint   `json:"TxIndex"`
	Proof     []byte `json:"Proof"`
	NetworkID uint   `json:"NetworkID"`
}

type ShieldRequest struct {
	ShieldDatas []ShieldData `json:"ShieldDatas"`
	IncTokenID  common.Hash  `json:"IncTokenID"`
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
	shieldDatas []ShieldData, incTokenID common.Hash,
) *ShieldRequest {
	return &ShieldRequest{
		ShieldDatas: shieldDatas,
		IncTokenID:  incTokenID,
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
	return true
}

func (request *ShieldRequest) Hash() *common.Hash {
	return nil
}

func (request *ShieldRequest) BuildReqActions(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	content, err := metadataCommon.NewActionWithValue(request, *tx.Hash(), nil).StringSlice(metadataCommon.ShieldUnifiedTokenRequestMeta)
	return [][]string{content}, err
}

func (request *ShieldRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(request)
}
