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
)

type UserMintNftRequest struct {
	metadataCommon.MetadataBase
	otaReceiver string
	amount      uint64
}

func NewUserMintNftRequest() *UserMintNftRequest {
	return &UserMintNftRequest{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.Pdexv3UserMintNftRequestMeta,
		},
	}
}

func NewUserMintNftRequestWithValue(otaReceiver string, amount uint64) *UserMintNftRequest {
	metadataBase := metadataCommon.MetadataBase{
		Type: metadataCommon.Pdexv3UserMintNftRequestMeta,
	}
	return &UserMintNftRequest{
		otaReceiver:  otaReceiver,
		amount:       amount,
		MetadataBase: metadataBase,
	}
}

func (request *UserMintNftRequest) ValidateTxWithBlockChain(
	tx metadataCommon.Transaction,
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (bool, error) {
	return beaconViewRetriever.IsValidPdexv3MintNftRequireAmount(request.amount)
}

func (request *UserMintNftRequest) ValidateSanityData(
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	beaconHeight uint64,
	tx metadataCommon.Transaction,
) (bool, bool, error) {
	if !chainRetriever.IsAfterPdexv3CheckPoint(beaconHeight) {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Feature pdexv3 has not been activated yet"))
	}
	otaReceiver := privacy.OTAReceiver{}
	err := otaReceiver.FromString(request.otaReceiver)
	if err != nil {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
	}
	if !otaReceiver.IsValid() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("ReceiveAddress is not valid"))
	}
	if request.amount == 0 {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("request.amount is 0"))
	}

	isBurned, burnCoin, burnedTokenID, err := tx.GetTxBurnData()
	if err != nil || !isBurned {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDENotBurningTxError, err)
	}
	if !bytes.Equal(burnedTokenID[:], common.PRVCoinID[:]) {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Wrong request info's token id, it should be equal to tx's token id"))
	}
	if burnCoin.GetValue() != request.amount {
		err := fmt.Errorf("Burnt amount is not valid expect %v but get %v", request.amount, burnCoin.GetValue())
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
	}
	if tx.GetType() != common.TxNormalType {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Tx type must be normal privacy type"))
	}
	if otaReceiver.GetShardID() != byte(tx.GetValidationEnv().ShardID()) {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("otaReceiver shardID is different from txShardID"))
	}
	return true, true, nil
}

func (request *UserMintNftRequest) ValidateMetadataByItself() bool {
	return request.Type == metadataCommon.Pdexv3UserMintNftRequestMeta
}

func (request *UserMintNftRequest) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(&request)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (request *UserMintNftRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(request)
}

func (request *UserMintNftRequest) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		OtaReceiver string `json:"OtaReceiver"`
		Amount      uint64 `json:"Amount"`
		metadataCommon.MetadataBase
	}{
		Amount:       request.amount,
		OtaReceiver:  request.otaReceiver,
		MetadataBase: request.MetadataBase,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (request *UserMintNftRequest) UnmarshalJSON(data []byte) error {
	temp := struct {
		OtaReceiver string `json:"OtaReceiver"`
		Amount      uint64 `json:"Amount"`
		metadataCommon.MetadataBase
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	request.amount = temp.Amount
	request.otaReceiver = temp.OtaReceiver
	request.MetadataBase = temp.MetadataBase
	return nil
}

func (request *UserMintNftRequest) OtaReceiver() string {
	return request.otaReceiver
}

func (request *UserMintNftRequest) Amount() uint64 {
	return request.amount
}

func (request *UserMintNftRequest) GetOTADeclarations() []metadataCommon.OTADeclaration {
	var result []metadataCommon.OTADeclaration
	otaReceiver := privacy.OTAReceiver{}
	otaReceiver.FromString(request.otaReceiver)
	result = append(result, metadataCommon.OTADeclaration{
		PublicKey: otaReceiver.PublicKey.ToBytes(), TokenID: common.PRVCoinID,
	})
	result = append(result, metadataCommon.OTADeclaration{
		PublicKey: otaReceiver.PublicKey.ToBytes(), TokenID: common.ConfidentialAssetID,
	})
	return result
}
