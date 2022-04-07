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
	TokenID  common.Hash         `json:"TokenID"`
	Amount   uint64              `json:"Amount"`
	Receiver privacy.OTAReceiver `json:"Receiver"`
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
	TokenID  common.Hash           `json:"TokenID"`
	Data     []UnshieldRequestData `json:"Data"`
	Receiver privacy.OTAReceiver   `json:"Receiver"`
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
	tokenID common.Hash, data []UnshieldRequestData, receiver privacy.OTAReceiver,
) *UnshieldRequest {
	return &UnshieldRequest{
		TokenID:  tokenID,
		Data:     data,
		Receiver: receiver,
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.BurningUnifiedTokenRequestMeta,
		},
	}
}

func (request *UnshieldRequest) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	totalBurningAmount := uint64(0)
	for _, data := range request.Data {
		if _, err := hex.DecodeString(data.RemoteAddress); err != nil {
			return false, err
		}
		if data.BurningAmount == 0 {
			return false, fmt.Errorf("wrong request info's burned amount")
		}
		if data.NetworkID != common.BSCNetworkID && data.NetworkID != common.ETHNetworkID && data.NetworkID != common.PLGNetworkID && data.NetworkID != common.FTMNetworkID {
			return false, fmt.Errorf("Invalid networkID")
		}
		if data.BurningAmount < data.ExpectedAmount {
			return false, fmt.Errorf("burningAmount %v < expectedAmount %v", data.BurningAmount, data.ExpectedAmount)
		}
		ok, err := beaconViewRetriever.BridgeAggIsValidBurntAmount(data.BurningAmount, request.TokenID, data.NetworkID)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, fmt.Errorf("BurningAmount is not valid")
		}
		totalBurningAmount += data.BurningAmount
	}
	isBurned, burnCoin, burnedTokenID, err := tx.GetTxBurnData()
	if err != nil || !isBurned {
		return false, fmt.Errorf("it is not transaction burn. Error %v", err)
	}
	if !bytes.Equal(burnedTokenID[:], request.TokenID[:]) {
		return false, fmt.Errorf("wrong request info's token id and token burned")
	}
	burnAmount := burnCoin.GetValue()
	if burnAmount != totalBurningAmount || burnAmount == 0 {
		return false, fmt.Errorf("burn amount is incorrect %v", burnAmount)
	}
	return true, nil
}

func (request *UnshieldRequest) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {
	if len(request.Data) <= 0 || len(request.Data) >= config.Param().BridgeAggParam.MaxLenOfPath {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggUnshieldValidateSanityDataError, fmt.Errorf("Length of data %d need to be in [1..%d]", len(request.Data), config.Param().BridgeAggParam.MaxLenOfPath))
	}
	if !request.Receiver.IsValid() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggUnshieldValidateSanityDataError, fmt.Errorf("receiver is not valid"))
	}
	if request.Receiver.GetShardID() != byte(tx.GetValidationEnv().ShardID()) {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggUnshieldValidateSanityDataError, fmt.Errorf("otaReceiver shardID is different from txShardID"))
	}

	switch tx.GetType() {
	case common.TxNormalType:
		if request.TokenID != common.PRVCoinID {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidTxTypeError, fmt.Errorf("With tx normal privacy, the tokenIDStr should be PRV, not custom token"))
		}
	case common.TxCustomTokenPrivacyType:
		if request.TokenID == common.PRVCoinID {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidTxTypeError, fmt.Errorf("With tx custome token privacy, the tokenIDStr should not be PRV, but custom token"))
		}
	default:
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidTxTypeError, fmt.Errorf("Not recognize tx type"))
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
