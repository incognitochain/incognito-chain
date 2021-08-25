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

type WithdrawLiquidityRequest struct {
	metadataCommon.MetadataBase
	poolPairID       string
	nftID            string
	otaReceiveNft    string
	otaReceiveToken0 string
	otaReceiveToken1 string
	shareAmount      uint64
}

func NewWithdrawLiquidityRequest() *WithdrawLiquidityRequest {
	return &WithdrawLiquidityRequest{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.Pdexv3WithdrawLiquidityRequestMeta,
		},
	}
}

func NewWithdrawLiquidityRequestWithValue(
	poolPairID, nftID, otaReceiveNft,
	otaReceiveToken0, otaReceiveToken1 string,
	shareAmount uint64,
) *WithdrawLiquidityRequest {
	return &WithdrawLiquidityRequest{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.Pdexv3WithdrawLiquidityRequestMeta,
		},
		poolPairID:       poolPairID,
		nftID:            nftID,
		otaReceiveNft:    otaReceiveNft,
		otaReceiveToken0: otaReceiveToken0,
		otaReceiveToken1: otaReceiveToken1,
		shareAmount:      shareAmount,
	}
}

func (request *WithdrawLiquidityRequest) ValidateTxWithBlockChain(
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
	err = beaconViewRetriever.IsValidPoolPairID(request.poolPairID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (request *WithdrawLiquidityRequest) ValidateSanityData(
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	beaconHeight uint64,
	tx metadataCommon.Transaction,
) (bool, bool, error) {
	if request.poolPairID == "" {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Pool pair id should not be empty"))
	}
	nftID, err := common.Hash{}.NewHashFromStr(request.nftID)
	if err != nil {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
	}
	if nftID.IsZeroValue() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("NftID should not be empty"))
	}
	otaReceiveNft := privacy.OTAReceiver{}
	err = otaReceiveNft.FromString(request.otaReceiveNft)
	if err != nil {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
	}
	if !otaReceiveNft.IsValid() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("otaReceiveNft is not valid"))
	}
	otaReceiveToken0 := privacy.OTAReceiver{}
	err = otaReceiveToken0.FromString(request.otaReceiveToken0)
	if err != nil {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
	}
	if !otaReceiveToken0.IsValid() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("otaReceiveNft is not valid"))
	}
	otaReceiveToken1 := privacy.OTAReceiver{}
	err = otaReceiveToken1.FromString(request.otaReceiveToken1)
	if err != nil {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
	}
	if !otaReceiveToken1.IsValid() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("otaReceiveNft is not valid"))
	}
	if request.shareAmount == 0 {
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

func (request *WithdrawLiquidityRequest) ValidateMetadataByItself() bool {
	return request.Type == metadataCommon.Pdexv3WithdrawLiquidityRequestMeta
}

func (request *WithdrawLiquidityRequest) Hash() *common.Hash {
	record := request.MetadataBase.Hash().String()
	record += request.poolPairID
	record += request.nftID
	record += request.otaReceiveNft
	record += request.otaReceiveToken0
	record += request.otaReceiveToken1
	record += strconv.FormatUint(uint64(request.shareAmount), 10)
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (request *WithdrawLiquidityRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(request)
}

func (request *WithdrawLiquidityRequest) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		PoolPairID       string `json:"PoolPairID"`
		NftID            string `json:"NftID"`
		OtaReceiveNft    string `json:"OtaReceiveNft"`
		OtaReceiveToken0 string `json:"OtaReceiveToken0"`
		OtaReceiveToken1 string `json:"OtaReceiveToken1"`
		ShareAmount      uint64 `json:"ShareAmount"`
		metadataCommon.MetadataBase
	}{
		PoolPairID:       request.poolPairID,
		NftID:            request.nftID,
		OtaReceiveNft:    request.otaReceiveNft,
		OtaReceiveToken0: request.otaReceiveToken0,
		OtaReceiveToken1: request.otaReceiveToken1,
		ShareAmount:      request.shareAmount,
		MetadataBase:     request.MetadataBase,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (request *WithdrawLiquidityRequest) UnmarshalJSON(data []byte) error {
	temp := struct {
		PoolPairID       string `json:"PoolPairID"`
		NftID            string `json:"NftID"`
		OtaReceiveNft    string `json:"OtaReceiveNft"`
		OtaReceiveToken0 string `json:"OtaReceiveToken0"`
		OtaReceiveToken1 string `json:"OtaReceiveToken1"`
		ShareAmount      uint64 `json:"ShareAmount"`
		metadataCommon.MetadataBase
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	request.poolPairID = temp.PoolPairID
	request.nftID = temp.NftID
	request.otaReceiveNft = temp.OtaReceiveNft
	request.otaReceiveToken0 = temp.OtaReceiveToken0
	request.otaReceiveToken1 = temp.OtaReceiveToken1
	request.shareAmount = temp.ShareAmount
	request.MetadataBase = temp.MetadataBase
	return nil
}

func (request *WithdrawLiquidityRequest) PoolPairID() string {
	return request.poolPairID
}

func (request *WithdrawLiquidityRequest) OtaReceiveNft() string {
	return request.otaReceiveNft
}

func (request *WithdrawLiquidityRequest) ShareAmount() uint64 {
	return request.shareAmount
}

func (request *WithdrawLiquidityRequest) NftID() string {
	return request.nftID
}

func (request *WithdrawLiquidityRequest) OtaReceiveToken0() string {
	return request.otaReceiveToken0
}

func (request *WithdrawLiquidityRequest) OtaReceiveToken1() string {
	return request.otaReceiveToken1
}
