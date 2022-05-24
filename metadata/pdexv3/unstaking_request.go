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
	stakingPoolID   string
	otaReceivers    map[string]string
	nftID           string
	unstakingAmount uint64
}

func NewUnstakingRequest() *UnstakingRequest {
	return &UnstakingRequest{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.Pdexv3UnstakingRequestMeta,
		},
	}
}

func NewUnstakingRequestWithValue(
	stakingPoolID, nftID string,
	otaReceivers map[string]string,
	unstakingAmount uint64,
) *UnstakingRequest {
	return &UnstakingRequest{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.Pdexv3UnstakingRequestMeta,
		},
		stakingPoolID:   stakingPoolID,
		nftID:           nftID,
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
	if !chainRetriever.IsAfterPdexv3CheckPoint(beaconViewRetriever.GetHeight()) {
		return false, fmt.Errorf("Feature pdexv3 has not been activated yet")
	}
	pdexv3StateCached := chainRetriever.GetPdexv3Cached(beaconViewRetriever.BlockHash())
	err := beaconViewRetriever.IsValidNftID(chainRetriever.GetBeaconChainDatabase(), pdexv3StateCached, request.nftID)
	if err != nil {
		return false, err
	}
	err = beaconViewRetriever.IsValidPdexv3UnstakingAmount(
		chainRetriever.GetBeaconChainDatabase(), pdexv3StateCached,
		request.stakingPoolID, request.nftID, request.unstakingAmount,
	)
	if err != nil {
		return false, err
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
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("NftID should not be empty"))
	}
	nftID, err := common.Hash{}.NewHashFromStr(request.nftID)
	if err != nil {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
	}
	if nftID.IsZeroValue() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("NftID should not be empty"))
	}
	if request.otaReceivers[request.nftID] == utils.EmptyString {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("NftID's ota receiver can not be empty"))
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
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("shareAmount can not be 0"))
	}

	isBurned, burnCoin, burnedTokenID, err := tx.GetTxBurnData()
	if err != nil || !isBurned {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDENotBurningTxError, err)
	}
	if !bytes.Equal(burnedTokenID[:], nftID[:]) {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Wrong request info's token id, it should be equal to tx's token id"))
	}
	if burnCoin.GetValue() != 1 {
		err := fmt.Errorf("Burnt amount is not valid expect %v but get %v", 1, burnCoin.GetValue())
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
	}
	if tx.GetType() != common.TxCustomTokenPrivacyType {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Tx type must be custom token privacy type"))
	}
	if nftID.String() == common.PRVCoinID.String() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidTxTypeError, errors.New("Invalid NftID"))
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
		StakingPoolID   string            `json:"StakingPoolID"`
		NftID           string            `json:"NftID"`
		OtaReceivers    map[string]string `json:"OtaReceivers"`
		UnstakingAmount uint64            `json:"UnstakingAmount"`
		metadataCommon.MetadataBase
	}{
		StakingPoolID:   request.stakingPoolID,
		NftID:           request.nftID,
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
		StakingPoolID   string            `json:"StakingPoolID"`
		NftID           string            `json:"NftID"`
		OtaReceivers    map[string]string `json:"OtaReceivers"`
		UnstakingAmount uint64            `json:"UnstakingAmount"`
		metadataCommon.MetadataBase
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	request.stakingPoolID = temp.StakingPoolID
	request.nftID = temp.NftID
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

func (request *UnstakingRequest) NftID() string {
	return request.nftID
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
