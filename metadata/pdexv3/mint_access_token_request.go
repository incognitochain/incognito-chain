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

type MintAccessTokenRequest struct {
	metadataCommon.MetadataBase
	otaReceiver *privacy.OTAReceiver
	amount      uint64
}

func NewMintAccessTokenRequest() *MintAccessTokenRequest {
	return &MintAccessTokenRequest{}
}

func NewMintAccessTokenRequestWithValue(
	amount uint64,
	otaReceiver *privacy.OTAReceiver,
) *MintAccessTokenRequest {
	return &MintAccessTokenRequest{
		amount:      amount,
		otaReceiver: otaReceiver,
	}
}

func (request *MintAccessTokenRequest) ValidateTxWithBlockChain(
	tx metadataCommon.Transaction,
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (bool, error) {
	if err := beaconViewRetriever.IsValidPdexv3MintAccessTokenAmount(request.amount); err != nil {
		return false, err
	}
	return true, nil
}

func (request *MintAccessTokenRequest) ValidateSanityData(
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	beaconHeight uint64,
	tx metadataCommon.Transaction,
) (bool, bool, error) {
	if !chainRetriever.IsAfterPdexv3CheckPoint(beaconHeight) {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Feature pdexv3 has not been activated yet"))
	}
	if !request.otaReceiver.IsValid() {
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
	if request.otaReceiver.GetShardID() != byte(tx.GetValidationEnv().ShardID()) {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("otaReceiver shardID is different from txShardID"))
	}
	return true, true, nil
}

func (request *MintAccessTokenRequest) ValidateMetadataByItself() bool {
	return request.Type == metadataCommon.Pdexv3MintAccessTokenRequestMeta
}

func (request *MintAccessTokenRequest) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(&request)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (request *MintAccessTokenRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(request)
}

func (request *MintAccessTokenRequest) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		OtaReceiver *privacy.OTAReceiver `json:"OtaReceiver"`
		Amount      uint64               `json:"Amount"`
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

func (request *MintAccessTokenRequest) UnmarshalJSON(data []byte) error {
	temp := struct {
		OtaReceiver *privacy.OTAReceiver `json:"OtaReceiver"`
		Amount      uint64               `json:"Amount"`
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

func (request *MintAccessTokenRequest) OtaReceiver() privacy.OTAReceiver {
	return *request.otaReceiver
}

func (request *MintAccessTokenRequest) Amount() uint64 {
	return request.amount
}

func (request *MintAccessTokenRequest) GetOTADeclarations() []metadataCommon.OTADeclaration {
	var result []metadataCommon.OTADeclaration
	result = append(result, metadataCommon.OTADeclaration{
		PublicKey: request.otaReceiver.PublicKey.ToBytes(), TokenID: common.PRVCoinID,
	})
	result = append(result, metadataCommon.OTADeclaration{
		PublicKey: request.otaReceiver.PublicKey.ToBytes(), TokenID: common.ConfidentialAssetID,
	})
	return result
}
