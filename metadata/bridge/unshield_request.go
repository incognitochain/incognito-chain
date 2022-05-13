package bridge

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

type RejectedUnshieldRequest struct {
	UnifiedTokenID common.Hash         `json:"UnifiedTokenID"`
	Amount         uint64              `json:"Amount"`
	Receiver       privacy.OTAReceiver `json:"Receiver"`
}

type AcceptedInstUnshieldRequest struct {
	UnifiedTokenID common.Hash                   `json:"UnifiedTokenID"`
	IsDepositToSC  bool                          `json:"IsDepositToSC"`
	TxReqID        common.Hash                   `json:"TxReqID"`
	Data           []AcceptedUnshieldRequestData `json:"Data"`
	IsWaiting      bool                          `json:"IsWaiting"`
}

type AcceptedUnshieldRequestData struct {
	BurningAmount  uint64      `json:"BurningAmount"`
	ReceivedAmount uint64      `json:"ReceivedAmount"`
	IncTokenID     common.Hash `json:"IncTokenID"`
}

type UnshieldRequestData struct {
	IncTokenID        common.Hash `json:"IncTokenID"`
	BurningAmount     uint64      `json:"BurningAmount"`
	MinExpectedAmount uint64      `json:"MinExpectedAmount"`
	RemoteAddress     string      `json:"RemoteAddress"`
}

type UnshieldRequest struct {
	UnifiedTokenID common.Hash           `json:"UnifiedTokenID"`
	Data           []UnshieldRequestData `json:"Data"`
	Receiver       privacy.OTAReceiver   `json:"Receiver"`
	IsDepositToSC  bool                  `json:"IsDepositToSC"`
	metadataCommon.MetadataBase
}

func NewUnshieldRequest() *UnshieldRequest {
	return &UnshieldRequest{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.BurningUnifiedTokenRequestMeta,
		},
	}
}

func NewUnshieldRequestWithValue(
	unifiedTokenID common.Hash, data []UnshieldRequestData, receiver privacy.OTAReceiver, isDepositToSC bool,
) *UnshieldRequest {
	return &UnshieldRequest{
		UnifiedTokenID: unifiedTokenID,
		Data:           data,
		Receiver:       receiver,
		IsDepositToSC:  isDepositToSC,
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.BurningUnifiedTokenRequestMeta,
		},
	}
}

func (request *UnshieldRequest) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

func (request *UnshieldRequest) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {
	usedTokenIDs := make(map[common.Hash]bool)
	if request.UnifiedTokenID.IsZeroValue() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggConvertRequestValidateSanityDataError, fmt.Errorf("UnifiedTokenID can not be empty"))
	}
	if len(request.Data) <= 0 || len(request.Data) > config.Param().BridgeAggParam.MaxLenOfPath {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggUnshieldValidateSanityDataError, fmt.Errorf("Length of data %d need to be in [1..%d]", len(request.Data), config.Param().BridgeAggParam.MaxLenOfPath))
	}
	if !request.Receiver.IsValid() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggUnshieldValidateSanityDataError, fmt.Errorf("receiver is not valid"))
	}
	if request.Receiver.GetShardID() != byte(tx.GetValidationEnv().ShardID()) {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggUnshieldValidateSanityDataError, fmt.Errorf("otaReceiver shardID is different from txShardID"))
	}
	usedTokenIDs[request.UnifiedTokenID] = true
	totalBurningAmount := uint64(0)
	for _, data := range request.Data {
		if _, err := hex.DecodeString(data.RemoteAddress); err != nil {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggUnshieldValidateSanityDataError, err)
		}
		if data.BurningAmount == 0 {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggUnshieldValidateSanityDataError, fmt.Errorf("wrong request info's burned amount"))
		}
		if data.IncTokenID.IsZeroValue() {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggUnshieldValidateSanityDataError, fmt.Errorf("IncTokenID cannot be empty"))
		}
		if usedTokenIDs[data.IncTokenID] {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggUnshieldValidateSanityDataError, fmt.Errorf("Duplicate tokenID %s", data.IncTokenID.String()))
		}
		usedTokenIDs[data.IncTokenID] = true
		if data.BurningAmount < data.MinExpectedAmount {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggUnshieldValidateSanityDataError, fmt.Errorf("burningAmount %v < expectedAmount %v", data.BurningAmount, data.MinExpectedAmount))
		}
		totalBurningAmount += data.BurningAmount
		if totalBurningAmount < data.BurningAmount {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggUnshieldValidateSanityDataError, fmt.Errorf("Out of range uint64"))
		}
	}
	isBurned, burnCoin, burnedTokenID, err := tx.GetTxBurnData()
	if err != nil || !isBurned {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggUnshieldValidateSanityDataError, fmt.Errorf("it is not transaction burn. Error %v", err))
	}
	if !bytes.Equal(burnedTokenID[:], request.UnifiedTokenID[:]) {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggUnshieldValidateSanityDataError, fmt.Errorf("wrong request info's token id and token burned"))
	}
	burnAmount := burnCoin.GetValue()
	if burnAmount != totalBurningAmount || burnAmount == 0 {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggUnshieldValidateSanityDataError, fmt.Errorf("burn amount is incorrect %v", burnAmount))
	}

	if tx.GetType() != common.TxCustomTokenPrivacyType {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggUnshieldValidateSanityDataError, fmt.Errorf("tx is not custom token privacy type"))
	}
	for k := range usedTokenIDs {
		if k == common.PRVCoinID || k == common.PDEXCoinID {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggUnshieldValidateSanityDataError, fmt.Errorf("tokenID must not be special token"))
		}
	}

	return true, true, nil
}

func (request *UnshieldRequest) ValidateMetadataByItself() bool {
	return request.Type == metadataCommon.BurningUnifiedTokenRequestMeta
}

func (request *UnshieldRequest) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(&request)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (request *UnshieldRequest) BuildReqActions(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	content, err := metadataCommon.NewActionWithValue(request, *tx.Hash(), nil).StringSlice(metadataCommon.BurningUnifiedTokenRequestMeta)
	return [][]string{content}, err
}

func (request *UnshieldRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(request)
}

func (request *UnshieldRequest) GetOTADeclarations() []metadataCommon.OTADeclaration {
	var result []metadataCommon.OTADeclaration
	result = append(result, metadataCommon.OTADeclaration{
		PublicKey: request.Receiver.PublicKey.ToBytes(), TokenID: common.ConfidentialAssetID,
	})
	return result
}
