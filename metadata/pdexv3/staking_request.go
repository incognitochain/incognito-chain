package pdexv3

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

type StakingRequest struct {
	metadataCommon.MetadataBase
	tokenID      string
	otaReceivers map[string]string
	nftID        string
	tokenAmount  uint64
}

func NewStakingRequest() *StakingRequest {
	return &StakingRequest{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.Pdexv3StakingRequestMeta,
		},
	}
}

func NewStakingRequestWithValue(
	tokenID, nftID string, tokenAmount uint64, otaReceivers map[string]string,
) *StakingRequest {
	return &StakingRequest{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.Pdexv3StakingRequestMeta,
		},
		tokenID:      tokenID,
		nftID:        nftID,
		tokenAmount:  tokenAmount,
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
	err := beaconViewRetriever.IsValidNftID(request.nftID)
	if err != nil {
		return false, err
	}
	for k := range request.otaReceivers {
		err := beaconViewRetriever.IsValidPdexv3StakingPool(k)
		if err != nil {
			return false, err
		}
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
	tokenID, err := common.Hash{}.NewHashFromStr(request.tokenID)
	if err != nil {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
	}
	if tokenID.IsZeroValue() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("TokenID should not be empty"))
	}
	nftID, err := common.Hash{}.NewHashFromStr(request.nftID)
	if err != nil {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
	}
	if nftID.IsZeroValue() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("NftID should not be empty"))
	}
	for _, v := range request.otaReceivers {
		otaReceive := privacy.OTAReceiver{}
		err = otaReceive.FromString(v)
		if err != nil {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
		}
		if !otaReceive.IsValid() {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("ReceiveAddress is not valid"))
		}
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
	return true, true, nil
}

func (request *StakingRequest) ValidateMetadataByItself() bool {
	return request.Type == metadataCommon.Pdexv3StakingRequestMeta
}

func (request *StakingRequest) Hash() *common.Hash {
	record := request.MetadataBase.Hash().String()
	data, _ := json.Marshal(request.otaReceivers)
	record += string(data)
	record += request.tokenID
	record += request.nftID
	record += strconv.FormatUint(request.tokenAmount, 10)
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (request *StakingRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(request)
}

func (request *StakingRequest) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		OtaReceivers map[string]string `json:"OtaReceivers"`
		TokenID      string            `json:"TokenID"`
		NftID        string            `json:"NftID"`
		TokenAmount  uint64            `json:"TokenAmount"`
		metadataCommon.MetadataBase
	}{
		OtaReceivers: request.otaReceivers,
		TokenID:      request.tokenID,
		NftID:        request.nftID,
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
		OtaReceivers map[string]string `json:"OtaReceivers"`
		TokenID      string            `json:"TokenID"`
		NftID        string            `json:"NftID"`
		TokenAmount  uint64            `json:"TokenAmount"`
		metadataCommon.MetadataBase
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	request.otaReceivers = temp.OtaReceivers
	request.tokenID = temp.TokenID
	request.nftID = temp.NftID
	request.tokenAmount = temp.TokenAmount
	request.MetadataBase = temp.MetadataBase
	return nil
}

func (request *StakingRequest) OtaReceivers() map[string]string {
	return request.otaReceivers
}

func (request *StakingRequest) TokenID() string {
	return request.tokenID
}

func (request *StakingRequest) TokenAmount() uint64 {
	return request.tokenAmount
}

func (request *StakingRequest) NftID() string {
	return request.nftID
}
