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
	poolPairID            string
	nftID                 string
	otaReceiveNft         string
	token0Amount          uint64
	otaReceiveToken0      string
	token1Amount          uint64
	otaReceiveToken1      string
	otaReceiveTradingFees map[string]string
}

func NewWithdrawLiquidityRequest() *WithdrawLiquidityRequest {
	return &WithdrawLiquidityRequest{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.Pdexv3WithdrawLiquidityRequestMeta,
		},
		otaReceiveTradingFees: make(map[string]string),
	}
}

func NewWithdrawLiquidityRequestWithValue(
	poolPairID, nftID, otaReceiveNft,
	otaReceiveToken0, otaReceiveToken1 string,
	token0Amount, token1Amount uint64,
	otaReceiveTradingFees map[string]string,
) *WithdrawLiquidityRequest {
	return &WithdrawLiquidityRequest{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.Pdexv3WithdrawLiquidityRequestMeta,
		},
		poolPairID:            poolPairID,
		nftID:                 nftID,
		otaReceiveNft:         otaReceiveNft,
		token0Amount:          token0Amount,
		otaReceiveToken0:      otaReceiveToken0,
		token1Amount:          token1Amount,
		otaReceiveToken1:      otaReceiveToken1,
		otaReceiveTradingFees: otaReceiveTradingFees,
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
	// NOTE: verify supported tokens pair as needed
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
	if request.token0Amount == 0 {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("token0Amount can not be 0"))
	}
	if request.token1Amount == 0 {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("token1Amount can not be 0"))
	}
	if len(request.otaReceiveTradingFees) != 0 {
		for k, v := range request.otaReceiveTradingFees {
			tokenID, err := common.Hash{}.NewHashFromStr(k)
			if err != nil {
				return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
			}
			if tokenID.IsZeroValue() {
				return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("NftID should not be empty"))
			}
			otaReceive := privacy.OTAReceiver{}
			err = otaReceive.FromString(v)
			if err != nil {
				return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
			}
			if !otaReceive.IsValid() {
				return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("otaReceiveNft is not valid"))
			}
		}
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
	} else {
		if nftID.String() == common.PRVCoinID.String() {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidTxTypeError, errors.New("With tx custome token privacy, the tokenIDStr should not be PRV, but custom token"))
		}
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
	record += strconv.FormatUint(uint64(request.token0Amount), 10)
	record += strconv.FormatUint(uint64(request.token1Amount), 10)
	record += request.otaReceiveToken0
	record += request.otaReceiveToken1
	data, _ := json.Marshal(request.otaReceiveTradingFees)
	record += string(data)
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (request *WithdrawLiquidityRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(request)
}

func (request *WithdrawLiquidityRequest) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		PoolPairID            string            `json:"PoolPairID"`
		NftID                 string            `json:"NftID"`
		OtaReceiveNft         string            `json:"OtaReceiveNft"`
		Token0Amount          uint64            `json:"Token0Amount"`
		OtaReceiveToken0      string            `json:"OtaReceiveToken0"`
		Token1Amount          uint64            `json:"Token1Amount"`
		OtaReceiveToken1      string            `json:"OtaReceiveToken1"`
		OtaReceiveTradingFees map[string]string `json:"OtaReceiveTradingFees"`
		metadataCommon.MetadataBase
	}{
		PoolPairID:            request.poolPairID,
		NftID:                 request.nftID,
		OtaReceiveNft:         request.otaReceiveNft,
		Token0Amount:          request.token0Amount,
		OtaReceiveToken0:      request.otaReceiveToken0,
		Token1Amount:          request.token1Amount,
		OtaReceiveToken1:      request.otaReceiveToken1,
		OtaReceiveTradingFees: request.otaReceiveTradingFees,
		MetadataBase:          request.MetadataBase,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (request *WithdrawLiquidityRequest) UnmarshalJSON(data []byte) error {
	temp := struct {
		PoolPairID            string            `json:"PoolPairID"`
		NftID                 string            `json:"NftID"`
		OtaReceiveNft         string            `json:"OtaReceiveNft"`
		Token0Amount          uint64            `json:"Token0Amount"`
		OtaReceiveToken0      string            `json:"OtaReceiveToken0"`
		Token1Amount          uint64            `json:"Token1Amount"`
		OtaReceiveToken1      string            `json:"OtaReceiveToken1"`
		OtaReceiveTradingFees map[string]string `json:"OtaReceiveTradingFees"`
		metadataCommon.MetadataBase
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	request.poolPairID = temp.PoolPairID
	request.nftID = temp.NftID
	request.token0Amount = temp.Token0Amount
	request.token1Amount = temp.Token1Amount
	if temp.OtaReceiveTradingFees != nil {
		request.otaReceiveTradingFees = temp.OtaReceiveTradingFees
	}
	request.otaReceiveNft = temp.OtaReceiveNft
	request.otaReceiveToken0 = temp.OtaReceiveToken0
	request.otaReceiveToken1 = temp.OtaReceiveToken1
	request.MetadataBase = temp.MetadataBase
	return nil
}

func (request *WithdrawLiquidityRequest) PoolPairID() string {
	return request.poolPairID
}

func (request *WithdrawLiquidityRequest) OtaReceiveNft() string {
	return request.otaReceiveNft
}

func (request *WithdrawLiquidityRequest) Token0Amount() uint64 {
	return request.token0Amount
}

func (request *WithdrawLiquidityRequest) Token1Amount() uint64 {
	return request.token1Amount
}

func (request *WithdrawLiquidityRequest) OtaReceiveTradingFees() map[string]string {
	return request.otaReceiveTradingFees
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
