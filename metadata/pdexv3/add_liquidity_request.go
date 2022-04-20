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
	poolPairID   string // only "" for the first contribution of pool
	pairHash     string
	otaReceiver  string                              // receive refunded token or accessToken
	otaReceivers map[common.Hash]privacy.OTAReceiver // receive tokens
	tokenID      string
	AccessOption
	tokenAmount uint64
	amplifier   uint // only set for the first contribution
	metadataCommon.MetadataBase
}

func NewAddLiquidity() *AddLiquidityRequest {
	return &AddLiquidityRequest{
		AccessOption: *NewAccessOption(),
		otaReceivers: make(map[common.Hash]privacy.OTAReceiver),
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.Pdexv3AddLiquidityRequestMeta,
		},
	}
}

func NewAddLiquidityRequestWithValue(
	poolPairID, pairHash, otaReceiver, tokenID string,
	tokenAmount uint64, amplifier uint,
	accessOption *AccessOption,
	otaReceivers map[common.Hash]privacy.OTAReceiver,
) *AddLiquidityRequest {
	metadataBase := metadataCommon.MetadataBase{
		Type: metadataCommon.Pdexv3AddLiquidityRequestMeta,
	}
	return &AddLiquidityRequest{
		poolPairID:   poolPairID,
		pairHash:     pairHash,
		otaReceiver:  otaReceiver,
		tokenID:      tokenID,
		tokenAmount:  tokenAmount,
		amplifier:    amplifier,
		MetadataBase: metadataBase,
		AccessOption: *accessOption,
		otaReceivers: otaReceivers,
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
	isNewAccessOTALpRequest := request.otaReceiver != utils.EmptyString && len(request.otaReceivers) != 0
	err := request.AccessOption.IsValid(tx, request.otaReceivers, beaconViewRetriever, transactionStateDB, false, isNewAccessOTALpRequest, request.otaReceiver)
	if err != nil {
		return false, err
	}
	tokenHash, err := common.Hash{}.NewHashFromStr(request.tokenID)
	if err != nil {
		return false, err
	}
	err = request.AccessOption.ValidateOtaReceivers(tx, request.otaReceiver, request.otaReceivers, *tokenHash, isNewAccessOTALpRequest)
	if err != nil {
		return false, err
	}
	if request.poolPairID != utils.EmptyString {
		ok, err := beaconViewRetriever.IsValidPdexv3PoolPairID(request.poolPairID)
		if err != nil || !ok {
			if err == nil {
				err = fmt.Errorf("poolPairID %s is not valid", request.poolPairID)
			}
			return ok, err
		}
	}
	if !request.AccessOption.UseNft() && request.AccessOption.AccessID != nil {
		return beaconViewRetriever.IsValidPdexv3LP(request.poolPairID, request.AccessID.String())
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
		if *tokenID == common.PdexAccessCoinID {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidTxTypeError, errors.New("cannot contribute pdex access token"))
		}
		if tokenID.String() == common.PRVCoinID.String() {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidTxTypeError, errors.New("With tx custome token privacy, the tokenIDStr should not be PRV, but custom token"))
		}
	default:
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidTxTypeError, errors.New("Not recognize tx type"))
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
		PoolPairID   string                              `json:"PoolPairID"` // only "" for the first contribution of pool
		PairHash     string                              `json:"PairHash"`
		OtaReceiver  string                              `json:"OtaReceiver,omitempty"` // receive pToken
		OtaReceivers map[common.Hash]privacy.OTAReceiver `json:"OtaReceivers,omitempty"`
		TokenID      string                              `json:"TokenID"`
		AccessOption
		TokenAmount uint64 `json:"TokenAmount"`
		Amplifier   uint   `json:"Amplifier"` // only set for the first contribution
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
		OtaReceivers: request.otaReceivers,
	})

	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (request *AddLiquidityRequest) UnmarshalJSON(data []byte) error {
	temp := struct {
		PoolPairID   string                              `json:"PoolPairID"` // only "" for the first contribution of pool
		PairHash     string                              `json:"PairHash"`
		OtaReceiver  string                              `json:"OtaReceiver,omitempty"` // receive pToken
		OtaReceivers map[common.Hash]privacy.OTAReceiver `json:"OtaReceivers,omitempty"`
		TokenID      string                              `json:"TokenID"`
		AccessOption
		TokenAmount uint64 `json:"TokenAmount"`
		Amplifier   uint   `json:"Amplifier"` // only set for the first contribution
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
	request.otaReceivers = temp.OtaReceivers
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

//OtaReceivers read only function
func (request *AddLiquidityRequest) OtaReceivers() map[common.Hash]privacy.OTAReceiver {
	return request.otaReceivers
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
	if request.otaReceivers != nil {
		for tokenID, val := range request.otaReceivers {
			if tokenID != common.PRVCoinID {
				tokenID = common.ConfidentialAssetID
			}
			result = append(result, metadataCommon.OTADeclaration{
				PublicKey: val.PublicKey.ToBytes(), TokenID: tokenID,
			})
		}
	}
	// request.otaReceiver now will store receiver for nftID
	// and define which receiver will receive accessOTA (no need to declare OTA in this case)
	// receiver has been declared in block code above
	if request.otaReceiver != utils.EmptyString && len(request.otaReceivers) == 0 {
		currentTokenID := common.ConfidentialAssetID
		if request.TokenID() == common.PRVIDStr {
			currentTokenID = common.PRVCoinID
		}
		otaReceiver := privacy.OTAReceiver{}
		otaReceiver.FromString(request.otaReceiver)
		result = append(result, metadataCommon.OTADeclaration{
			PublicKey: otaReceiver.PublicKey.ToBytes(), TokenID: currentTokenID,
		})
	}
	return result
}
