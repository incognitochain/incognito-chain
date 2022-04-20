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

type StakingRequest struct {
	metadataCommon.MetadataBase
	tokenID      string
	otaReceiver  string
	otaReceivers map[common.Hash]privacy.OTAReceiver // receive tokens
	AccessOption
	tokenAmount uint64
}

func NewStakingRequest() *StakingRequest {
	return &StakingRequest{
		AccessOption: *NewAccessOption(),
		otaReceivers: make(map[common.Hash]privacy.OTAReceiver),
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.Pdexv3StakingRequestMeta,
		},
	}
}

func NewStakingRequestWithValue(
	tokenID, otaReceiver string, tokenAmount uint64, accessOption AccessOption,
	otaReceivers map[common.Hash]privacy.OTAReceiver,
) *StakingRequest {
	return &StakingRequest{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.Pdexv3StakingRequestMeta,
		},
		tokenID:      tokenID,
		AccessOption: accessOption,
		tokenAmount:  tokenAmount,
		otaReceiver:  otaReceiver,
		otaReceivers: otaReceivers,
	}
}

func (request *StakingRequest) ValidateTxWithBlockChain(
	tx metadataCommon.Transaction,
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (bool, error) {
	err := request.AccessOption.IsValid(tx, request.otaReceivers, beaconViewRetriever, transactionStateDB, false, false, "")
	if err != nil {
		return false, err
	}
	tokenHash, err := common.Hash{}.NewHashFromStr(request.tokenID)
	if err != nil {
		return false, err
	}
	err = request.AccessOption.ValidateOtaReceivers(tx, request.otaReceiver, request.otaReceivers, *tokenHash, false)
	if err != nil {
		return false, err
	}
	ok, err := beaconViewRetriever.IsValidPdexv3StakingPool(request.tokenID)
	if err != nil || !ok {
		if err == nil {
			err = fmt.Errorf("StakingPoolID %s is not valid", request.tokenID)
		}
		return false, err
	}
	if !request.AccessOption.UseNft() && request.AccessOption.AccessID != nil {
		return beaconViewRetriever.IsValidPdexv3Staker(request.tokenID, request.AccessID.String())
	}
	return true, nil
}

func (request *StakingRequest) ValidateSanityData(
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	beaconHeight uint64,
	tx metadataCommon.Transaction,
) (bool, bool, error) {
	if !chainRetriever.IsAfterPdexv3CheckPoint(beaconHeight) {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Feature pdexv3 has not been activated yet"))
	}
	tokenID, err := common.Hash{}.NewHashFromStr(request.tokenID)
	if err != nil {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
	}
	if tokenID.IsZeroValue() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("TokenID should not be empty"))
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
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidTxTypeError, errors.New("cannot staking pdex access token"))
		}
		if tokenID.String() == common.PRVCoinID.String() {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidTxTypeError, errors.New("With tx custome token privacy, the tokenIDStr should not be PRV, but custom token"))
		}
	default:
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidTxTypeError, errors.New("Not recognize tx type"))
	}
	return true, true, nil
}

func (request *StakingRequest) ValidateMetadataByItself() bool {
	return request.Type == metadataCommon.Pdexv3StakingRequestMeta
}

func (request *StakingRequest) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(&request)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (request *StakingRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(request)
}

func (request *StakingRequest) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		OtaReceiver  string                              `json:"OtaReceiver,omitempty"`
		OtaReceivers map[common.Hash]privacy.OTAReceiver `json:"OtaReceivers,omitempty"`
		TokenID      string                              `json:"TokenID"`
		AccessOption
		TokenAmount uint64 `json:"TokenAmount"`
		metadataCommon.MetadataBase
	}{
		OtaReceivers: request.otaReceivers,
		OtaReceiver:  request.otaReceiver,
		TokenID:      request.tokenID,
		AccessOption: request.AccessOption,
		TokenAmount:  request.tokenAmount,
		MetadataBase: request.MetadataBase,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (request *StakingRequest) UnmarshalJSON(data []byte) error {
	temp := struct {
		OtaReceiver  string                              `json:"OtaReceiver,omitempty"`
		OtaReceivers map[common.Hash]privacy.OTAReceiver `json:"OtaReceivers,omitempty"`
		TokenID      string                              `json:"TokenID"`
		AccessOption
		TokenAmount uint64 `json:"TokenAmount"`
		metadataCommon.MetadataBase
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	request.otaReceivers = temp.OtaReceivers
	request.otaReceiver = temp.OtaReceiver
	request.tokenID = temp.TokenID
	request.AccessOption = temp.AccessOption
	request.tokenAmount = temp.TokenAmount
	request.MetadataBase = temp.MetadataBase
	return nil
}

//OtaReceivers read only function
func (request *StakingRequest) OtaReceivers() map[common.Hash]privacy.OTAReceiver {
	return request.otaReceivers
}

func (request *StakingRequest) OtaReceiver() string {
	return request.otaReceiver
}

func (request *StakingRequest) TokenID() string {
	return request.tokenID
}

func (request *StakingRequest) TokenAmount() uint64 {
	return request.tokenAmount
}

func (request *StakingRequest) GetOTADeclarations() []metadataCommon.OTADeclaration {
	var result []metadataCommon.OTADeclaration
	if request.otaReceiver != utils.EmptyString {
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
	return result
}
