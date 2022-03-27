package bridge

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type ShieldResponseData struct {
	ExternalTokenID []byte `json:"ExternalTokenID"`
	UniqTx          []byte `json:"UniqETHTx"`
	NetworkID       uint   `json:"NetworkID"`
}

type ShieldResponse struct {
	metadataCommon.MetadataBase
	TxReqID      common.Hash          `json:"TxReqID"`
	Data         []ShieldResponseData `json:"Data"`
	SharedRandom []byte               `json:"SharedRandom,omitempty"`
}

func NewShieldResponse() *ShieldResponse {
	return &ShieldResponse{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.ShieldUnifiedTokenResponseMeta,
		},
	}
}

func NewShieldResponseWithValue(
	data []ShieldResponseData, txReqID common.Hash, shardRandom []byte,
) *ShieldResponse {
	return &ShieldResponse{
		Data: data,
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.ShieldUnifiedTokenResponseMeta,
		},
		SharedRandom: shardRandom,
		TxReqID:      txReqID,
	}
}

func (response *ShieldResponse) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

func (response *ShieldResponse) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {

	return true, true, nil
}

func (response *ShieldResponse) ValidateMetadataByItself() bool {
	return response.Type == metadataCommon.ShieldUnifiedTokenResponseMeta
}

func (response *ShieldResponse) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(&response)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (response *ShieldResponse) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(response)
}

func (response ShieldResponse) VerifyMinerCreatedTxBeforeGettingInBlock(mintData *metadataCommon.MintData, shardID byte, tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, ac *metadataCommon.AccumulatedValues, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever) (bool, error) {
	idx := -1
	for i, inst := range mintData.Insts {
		if len(inst) < 4 { // this is not shieldRequest instruction
			continue
		}

		idx = i
		break
	}
	if idx == -1 { // not found the issuance request inst for this response
		return false, errors.New(fmt.Sprintf("no shield request inst found for ShieldResponse tx %s", tx.Hash().String()))
	}
	mintData.InstsUsed[idx] = 1
	return true, nil
}

func (response *ShieldResponse) SetSharedRandom(r []byte) {
	response.SharedRandom = r
}
