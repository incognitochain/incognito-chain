package pdexv3

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/utils"
)

type AddLiquidityRequest struct {
	poolPairID  string // only "" for the first contribution of pool
	pairHash    string
	otaReceiver string // refund pToken
	tokenID     string
	tokenAmount uint64
	amplifier   uint // only set for the first contribution
	AccessOption
	metadataCommon.MetadataBase
}

func NewAddLiquidity() *AddLiquidityRequest {
	return &AddLiquidityRequest{
		AccessOption: *NewAccessOption(),
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.Pdexv3AddLiquidityRequestMeta,
		},
	}
}

func NewAddLiquidityRequestWithValue(
	poolPairID, pairHash, otaReceiver, tokenID string,
	tokenAmount uint64, amplifier uint,
	accessOption *AccessOption,
) *AddLiquidityRequest {
	metadataBase := metadataCommon.MetadataBase{
		Type: metadataCommon.Pdexv3AddLiquidityRequestMeta,
	}
	return &AddLiquidityRequest{
		poolPairID:   poolPairID,
		pairHash:     pairHash,
		otaReceiver:  otaReceiver,
		tokenID:      tokenID,
		AccessOption: *accessOption,
		tokenAmount:  tokenAmount,
		amplifier:    amplifier,
		MetadataBase: metadataBase,
	}
}

func (request *AddLiquidityRequest) ValidateTxWithBlockChain(
	tx metadataCommon.Transaction,
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (bool, error) {
	err := request.AccessOption.isValid(tx, beaconViewRetriever, transactionStateDB, false)
	if err != nil {
		return false, err
	}
	if request.poolPairID != utils.EmptyString {
		err := beaconViewRetriever.IsValidPoolPairID(request.poolPairID)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func (request *AddLiquidityRequest) ValidateSanityData(
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	beaconHeight uint64,
	tx metadataCommon.Transaction,
) (bool, bool, error) {
	if !chainRetriever.IsAfterPdexv3CheckPoint(beaconHeight) {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Feature pdexv3 has not been activated yet"))
	}
	if request.pairHash == "" {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Pair hash should not be empty"))
	}
	tokenID, err := common.Hash{}.NewHashFromStr(request.tokenID)
	if err != nil {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
	}
	if tokenID.IsZeroValue() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("TokenID should not be empty"))
	}
	otaReceiver := privacy.OTAReceiver{}
	err = otaReceiver.FromString(request.otaReceiver)
	if err != nil {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
	}
	if !otaReceiver.IsValid() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("RefundAddress is not valid"))
	}
	if request.amplifier < BaseAmplifier {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Amplifier is not valid"))
	}
	isBurned, burnCoin, burnedTokenID, err := tx.GetTxBurnData()
	if err != nil || !isBurned {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDENotBurningTxError, err)
	}
	if !bytes.Equal(burnedTokenID[:], tokenID[:]) {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Wrong request info's token id, it should be equal to tx's token id"))
	}
	if request.tokenAmount == 0 || request.tokenAmount != burnCoin.GetValue() {
		err := fmt.Errorf("Contributed amount is not valid expect %v but get %v", request.tokenAmount, burnCoin.GetValue())
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
	}
	switch tx.GetType() {
	case common.TxNormalType:
		if tokenID.String() != common.PRVCoinID.String() {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidTxTypeError, errors.New("With tx normal privacy, the tokenIDStr should be PRV, not custom token"))
		}
	case common.TxCustomTokenPrivacyType:
		if tokenID.String() == common.PRVCoinID.String() {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidTxTypeError, errors.New("With tx custome token privacy, the tokenIDStr should not be PRV, but custom token"))
		}
	default:
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidTxTypeError, errors.New("Not recognize tx type"))
	}
	if otaReceiver.GetShardID() != byte(tx.GetValidationEnv().ShardID()) {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("otaReceiver shardID is different from txShardID"))
	}
	return true, true, nil
}

func (request *AddLiquidityRequest) ValidateMetadataByItself() bool {
	return request.Type == metadataCommon.Pdexv3AddLiquidityRequestMeta
}

func (request *AddLiquidityRequest) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(&request)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (request *AddLiquidityRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(request)
}

func (request *AddLiquidityRequest) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		PoolPairID  string `json:"PoolPairID"` // only "" for the first contribution of pool
		PairHash    string `json:"PairHash"`
		OtaReceiver string `json:"OtaReceiver"` // receive pToken
		TokenID     string `json:"TokenID"`
		NftID       string `json:"NftID,omitempty"` // will be removed when next ota is ready
		TokenAmount uint64 `json:"TokenAmount"`
		Amplifier   uint   `json:"Amplifier"` // only set for the first contribution
		AccessOption
		metadataCommon.MetadataBase
	}{
		PoolPairID:   request.poolPairID,
		PairHash:     request.pairHash,
		OtaReceiver:  request.otaReceiver,
		TokenID:      request.tokenID,
		TokenAmount:  request.tokenAmount,
		Amplifier:    request.amplifier,
		AccessOption: request.AccessOption,
		MetadataBase: request.MetadataBase,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (request *AddLiquidityRequest) UnmarshalJSON(data []byte) error {
	temp := struct {
		PoolPairID  string `json:"PoolPairID"` // only "" for the first contribution of pool
		PairHash    string `json:"PairHash"`
		OtaReceiver string `json:"OtaReceiver"` // receive pToken
		TokenID     string `json:"TokenID"`
		NftID       string `json:"NftID,omitempty"` // will be disabled when next ota is ready
		TokenAmount uint64 `json:"TokenAmount"`
		Amplifier   uint   `json:"Amplifier"` // only set for the first contribution
		AccessOption
		metadataCommon.MetadataBase
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	request.poolPairID = temp.PoolPairID
	request.pairHash = temp.PairHash
	request.otaReceiver = temp.OtaReceiver
	request.tokenID = temp.TokenID
	request.tokenAmount = temp.TokenAmount
	request.amplifier = temp.Amplifier
	request.MetadataBase = temp.MetadataBase
	request.AccessOption = temp.AccessOption
	return nil
}

func (request *AddLiquidityRequest) PoolPairID() string {
	return request.poolPairID
}

func (request *AddLiquidityRequest) PairHash() string {
	return request.pairHash
}

func (request *AddLiquidityRequest) OtaReceiver() string {
	return request.otaReceiver
}

func (request *AddLiquidityRequest) TokenID() string {
	return request.tokenID
}

func (request *AddLiquidityRequest) TokenAmount() uint64 {
	return request.tokenAmount
}

func (request *AddLiquidityRequest) Amplifier() uint {
	return request.amplifier
}

func (request *AddLiquidityRequest) GetOTADeclarations() []metadataCommon.OTADeclaration {
	var result []metadataCommon.OTADeclaration
	currentTokenID := common.ConfidentialAssetID
	if request.TokenID() == common.PRVIDStr {
		currentTokenID = common.PRVCoinID
	}
	otaReceiver := privacy.OTAReceiver{}
	otaReceiver.FromString(request.otaReceiver)
	result = append(result, metadataCommon.OTADeclaration{
		PublicKey: otaReceiver.PublicKey.ToBytes(), TokenID: currentTokenID,
	})
	return result
}
