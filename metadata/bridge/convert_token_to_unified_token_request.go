package bridge

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

type ConvertTokenToUnifiedTokenRequest struct {
	TokenID        common.Hash         `json:"TokenID"`
	UnifiedTokenID common.Hash         `json:"UnifiedTokenID"`
	NetworkID      uint                `json:"NetworkID"`
	Amount         uint64              `json:"Amount"`
	Receiver       privacy.OTAReceiver `json:"Receiver"`
	metadataCommon.MetadataBase
}

type RejectedConvertTokenToUnifiedToken struct {
	TokenID  common.Hash         `json:"TokenID"`
	Amount   uint64              `json:"Amount"`
	Receiver privacy.OTAReceiver `json:"Receiver"`
}

type AcceptedConvertTokenToUnifiedToken struct {
	ConvertTokenToUnifiedTokenRequest
	TxReqID    common.Hash `json:"TxReqID"`
	MintAmount uint64      `json:"MintAmount"`
}

func NewConvertTokenToUnifiedTokenRequest() *ConvertTokenToUnifiedTokenRequest {
	return &ConvertTokenToUnifiedTokenRequest{}
}

func NewConvertTokenToUnifiedTokenRequestWithValue(
	tokenID, unifiedTokenID common.Hash, networkID uint, amount uint64, receiver privacy.OTAReceiver,
) *ConvertTokenToUnifiedTokenRequest {
	metadataBase := metadataCommon.MetadataBase{
		Type: metadataCommon.BridgeAggConvertTokenToUnifiedTokenRequestMeta,
	}
	return &ConvertTokenToUnifiedTokenRequest{
		UnifiedTokenID: unifiedTokenID,
		TokenID:        tokenID,
		NetworkID:      networkID,
		Amount:         amount,
		Receiver:       receiver,
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
	if request.TokenID.IsZeroValue() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggConvertRequestValidateSanityDataError, errors.New("TokenID can not be empty"))
	}
	if request.UnifiedTokenID.IsZeroValue() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggConvertRequestValidateSanityDataError, errors.New("UnifiedTokenID can not be empty"))
	}
	if request.TokenID.String() == request.UnifiedTokenID.String() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggConvertRequestValidateSanityDataError, errors.New("TokenID and UnifiedTokenID cannot be the same"))
	}
	if !request.Receiver.IsValid() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggConvertRequestValidateSanityDataError, errors.New("receiver is not valid"))
	}
	if request.Receiver.GetShardID() != byte(tx.GetValidationEnv().ShardID()) {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggConvertRequestValidateSanityDataError, errors.New("otaReceiver shardID is different from txShardID"))
	}
	isBurned, burnCoin, burnedTokenID, err := tx.GetTxBurnData()
	if err != nil || !isBurned {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggConvertRequestValidateSanityDataError, err)
	}
	if !bytes.Equal(burnedTokenID[:], request.TokenID[:]) {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggConvertRequestValidateSanityDataError, errors.New("Wrong request info's token id, it should be equal to tx's token id"))
	}
	if request.Amount == 0 || request.Amount != burnCoin.GetValue() {
		err := fmt.Errorf("Amount is not valid metaAmount %v burntAmount %v", request.Amount, burnCoin.GetValue())
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggConvertRequestValidateSanityDataError, err)
	}
	if tx.GetType() != common.TxCustomTokenPrivacyType {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggConvertRequestValidateSanityDataError, errors.New("Convert tx need to be custom token type"))
	}
	if request.TokenID == common.PRVCoinID || request.UnifiedTokenID == common.PRVCoinID {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggConvertRequestValidateSanityDataError, errors.New("With tx custome token privacy, the tokenID should not be PRV, but custom token"))
	}
	if request.NetworkID == common.DefaultNetworkID {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggConvertRequestValidateSanityDataError, errors.New("NetworkID is invalid"))
	}

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

func (request *ConvertTokenToUnifiedTokenRequest) GetOTADeclarations() []metadataCommon.OTADeclaration {
	var result []metadataCommon.OTADeclaration
	result = append(result, metadataCommon.OTADeclaration{
		PublicKey: request.Receiver.PublicKey.ToBytes(), TokenID: common.ConfidentialAssetID,
	})
	return result
}

func (request *ConvertTokenToUnifiedTokenRequest) BuildReqActions(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	content, err := metadataCommon.NewActionWithValue(request, *tx.Hash(), nil).StringSlice(metadataCommon.BridgeAggConvertTokenToUnifiedTokenRequestMeta)
	return [][]string{content}, err
}
