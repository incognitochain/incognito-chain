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

func (s *ShieldData) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

func (s *ShieldData) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {

	return true, true, nil
}

func (s *ShieldData) ValidateMetadataByItself() bool {
	return true
}

func (s *ShieldData) Hash() *common.Hash {
	record := iReq.BlockHash.String()
	record += string(iReq.TxIndex)
	proofStrs := iReq.ProofStrs
	for _, proofStr := range proofStrs {
		record += proofStr
	}
	record += iReq.MetadataBase.Hash().String()
	record += iReq.IncTokenID.String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (s *ShieldData) BuildReqActions(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {

}

func (request *ShieldRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(request)
}
