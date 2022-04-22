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

type UnstakingRequest struct {
	metadataCommon.MetadataBase
	stakingPoolID string
	otaReceivers  map[string]string
	AccessOption
	unstakingAmount uint64
}

func NewUnstakingRequest() *UnstakingRequest {
	return &UnstakingRequest{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.Pdexv3UnstakingRequestMeta,
		},
		otaReceivers: make(map[string]string),
	}
}

func NewUnstakingRequestWithValue(
	stakingPoolID string,
	otaReceivers map[string]string,
	unstakingAmount uint64,
	accessOption AccessOption,
) *UnstakingRequest {
	return &UnstakingRequest{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.Pdexv3UnstakingRequestMeta,
		},
		stakingPoolID:   stakingPoolID,
		AccessOption:    accessOption,
		otaReceivers:    otaReceivers,
		unstakingAmount: unstakingAmount,
	}
}

func (request *UnstakingRequest) ValidateTxWithBlockChain(
	tx metadataCommon.Transaction,
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (bool, error) {
	err := request.AccessOption.IsValid(tx, request.getParsedOtaReceivers(), beaconViewRetriever, transactionStateDB, true, false, "")
	if err != nil {
		return false, err
	}
	isBurned, burnCoin, burnedTokenID, err := tx.GetTxBurnData()
	if err != nil || !isBurned {
		return false, metadataCommon.NewMetadataTxError(metadataCommon.PDENotBurningTxError, err)
	}
	accessID := utils.EmptyString
	expectBurntTokenID := common.Hash{}
	if request.AccessOption.UseNft() {
		if request.otaReceivers[request.AccessOption.NftID.String()] == utils.EmptyString {
			return false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("NftID's ota receiver cannot be empty"))
		}
		expectBurntTokenID = *request.AccessOption.NftID
		accessID = request.AccessOption.NftID.String()
		if *request.AccessOption.NftID == common.PRVCoinID {
			return false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidTxTypeError, errors.New("NftID cannot be prv"))
		}
	} else {
		expectBurntTokenID = common.PdexAccessCoinID
		accessID = request.AccessOption.AccessID.String()
	}
	if !bytes.Equal(burnedTokenID[:], expectBurntTokenID[:]) {
		return false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Wrong request info's token id, it should be equal to tx's token id"))
	}
	if burnCoin.GetValue() != 1 {
		err := fmt.Errorf("Burnt amount is not valid expect %v but get %v", 1, burnCoin.GetValue())
		return false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
	}
	if tx.GetType() != common.TxCustomTokenPrivacyType {
		return false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Tx type must be custom token privacy type"))
	}
	ok, err := beaconViewRetriever.IsValidPdexv3UnstakingAmount(request.stakingPoolID, accessID, request.unstakingAmount)
	if err != nil || !ok {
		if !ok {
			err = fmt.Errorf("UnstakingAmount is invalid")
		}
		return false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
	}
	if !request.UseNft() {
		return beaconViewRetriever.IsValidAccessOTAWithPdexState(
			*metadataCommon.NewPdexv3ExtendAccessIDWithValue(
				request.stakingPoolID,
				*request.AccessID,
				request.BurntOTA.ToBytesS(),
				metadataCommon.Pdexv3UnstakingRequestMeta,
				utils.EmptyString,
			),
		)
	}
	return true, nil
}

func (request *UnstakingRequest) ValidateSanityData(
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	beaconHeight uint64,
	tx metadataCommon.Transaction,
) (bool, bool, error) {
	if !chainRetriever.IsAfterPdexv3CheckPoint(beaconHeight) {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Feature pdexv3 has not been activated yet"))
	}
	stakingPoolID, err := common.Hash{}.NewHashFromStr(request.stakingPoolID)
	if err != nil {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
	}
	if stakingPoolID.IsZeroValue() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("StakingPoolID should not be empty"))
	}
	for tokenID, otaReceiverStr := range request.otaReceivers {
		_, err := common.Hash{}.NewHashFromStr(tokenID)
		if err != nil {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
		}
		otaReceiver := privacy.OTAReceiver{}
		err = otaReceiver.FromString(otaReceiverStr)
		if err != nil {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
		}
		if !otaReceiver.IsValid() {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("otaReceiveNft is not valid"))
		}
		if otaReceiver.GetShardID() != byte(tx.GetValidationEnv().ShardID()) {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("otaReceiver shardID is different from txShardID"))
		}
	}
	if request.unstakingAmount == 0 {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("unstakingAmount cannot be 0"))
	}

	return true, true, nil
}

func (request *UnstakingRequest) ValidateMetadataByItself() bool {
	return request.Type == metadataCommon.Pdexv3UnstakingRequestMeta
}

func (request *UnstakingRequest) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(&request)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (request *UnstakingRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(request)
}

func (request *UnstakingRequest) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		StakingPoolID string `json:"StakingPoolID"`
		AccessOption
		OtaReceivers    map[string]string `json:"OtaReceivers"`
		UnstakingAmount uint64            `json:"UnstakingAmount"`
		metadataCommon.MetadataBase
	}{
		StakingPoolID:   request.stakingPoolID,
		AccessOption:    request.AccessOption,
		OtaReceivers:    request.otaReceivers,
		UnstakingAmount: request.unstakingAmount,
		MetadataBase:    request.MetadataBase,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (request *UnstakingRequest) UnmarshalJSON(data []byte) error {
	temp := struct {
		StakingPoolID string `json:"StakingPoolID"`
		AccessOption
		OtaReceivers    map[string]string `json:"OtaReceivers"`
		UnstakingAmount uint64            `json:"UnstakingAmount"`
		metadataCommon.MetadataBase
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	request.stakingPoolID = temp.StakingPoolID
	request.AccessOption = temp.AccessOption
	request.otaReceivers = temp.OtaReceivers
	request.unstakingAmount = temp.UnstakingAmount
	request.MetadataBase = temp.MetadataBase
	return nil
}

func (request *UnstakingRequest) StakingPoolID() string {
	return request.stakingPoolID
}

func (request *UnstakingRequest) OtaReceivers() map[string]string {
	return request.otaReceivers
}

func (request *UnstakingRequest) UnstakingAmount() uint64 {
	return request.unstakingAmount
}

func (request *UnstakingRequest) GetOTADeclarations() []metadataCommon.OTADeclaration {
	var result []metadataCommon.OTADeclaration
	for tokenID, val := range request.otaReceivers {
		tokenHash := common.PRVCoinID
		if tokenID != common.PRVIDStr {
			tokenHash = common.ConfidentialAssetID
		}
		otaReceiver := privacy.OTAReceiver{}
		otaReceiver.FromString(val)
		result = append(result, metadataCommon.OTADeclaration{
			PublicKey: otaReceiver.PublicKey.ToBytes(), TokenID: tokenHash,
		})
	}
	return result
}

// requires passing sanity check
func (request *UnstakingRequest) getParsedOtaReceivers() map[common.Hash]privacy.OTAReceiver {
	result := make(map[common.Hash]privacy.OTAReceiver)
	for k, v := range request.otaReceivers {
		tokenID, _ := common.Hash{}.NewHashFromStr(k)
		recv := &privacy.OTAReceiver{}
		recv.FromString(v)
		result[*tokenID] = *recv
	}
	return result
}
