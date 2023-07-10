package ink

import (
	"encoding/json"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

type InscribeRequest struct {
	Data     json.RawMessage     `json:"Data"`
	Receiver privacy.OTAReceiver `json:"Receiver"`
	metadataCommon.MetadataBase
}

type InscribeAcceptedAction struct {
	Receiver privacy.OTAReceiver `json:"Receiver"`
	TokenID  common.Hash         `json:"TokenID"`
}

func (acn *InscribeAcceptedAction) GetStatus() int {
	return 1
}

func (acn *InscribeAcceptedAction) GetType() int {
	return metadataCommon.InscribeRequestMeta
}

func (iReq InscribeRequest) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

func (iReq InscribeRequest) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {
	// if shardViewRetriever.GetBlockVersion() < metadataCommon.InscribeMinBlockVersion {
	// 	return false, false, fmt.Errorf("Inscribe feature not supported in block version %d, want %d", shardViewRetriever.GetBlockVersion(), metadataCommon.InscribeMinBlockVersion)
	// }

	// isBurned, burnCoin, burnedTokenID, err := tx.GetTxBurnData()
	// if err != nil || !isBurned {
	// 	return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggUnshieldValidateSanityDataError, fmt.Errorf("burn missing from tx %s - %v", tx.Hash(), err))
	// }
	// if *burnedTokenID != common.PRVCoinID {
	// 	return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggUnshieldValidateSanityDataError, fmt.Errorf("burned tokenID mismatch - %s", burnedTokenID.String()))
	// }
	// burnAmount := burnCoin.GetValue()
	// if burnAmount < 1000000000 {
	// 	return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggUnshieldValidateSanityDataError, fmt.Errorf("burn amount too small - %d", burnAmount))
	// }
	if !iReq.Receiver.IsValid() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, fmt.Errorf("invalid receiver"))
	}
	if iReq.Receiver.GetShardID() != byte(tx.GetValidationEnv().ShardID()) {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, fmt.Errorf("otaReceiver shardID is different from txShardID"))
	}
	return true, true, nil
}

func (iReq InscribeRequest) ValidateMetadataByItself() bool {
	return iReq.Type == metadataCommon.InscribeRequestMeta
}

func (iReq InscribeRequest) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(iReq)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (request *InscribeRequest) BuildReqActions(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	content, err := metadataCommon.NewActionWithValue(request, *tx.Hash(), nil).StringSlice(metadataCommon.InscribeRequestMeta)
	return [][]string{content}, err
}

func (iReq *InscribeRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(iReq)
}

func (iReq *InscribeRequest) GetOTADeclarations() []metadataCommon.OTADeclaration {
	var result []metadataCommon.OTADeclaration
	result = append(result, metadataCommon.OTADeclaration{
		PublicKey: iReq.Receiver.PublicKey.ToBytes(), TokenID: common.ConfidentialAssetID,
	})
	return result
}
