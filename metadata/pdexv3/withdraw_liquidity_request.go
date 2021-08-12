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
	index                 string
	token0Amount          uint64
	token1Amount          uint64
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
	poolPairID, nftID, otaReceiveNft, index string,
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
		index:                 index,
		token0Amount:          token0Amount,
		token1Amount:          token1Amount,
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
	txIndex, err := common.Hash{}.NewHashFromStr(request.index)
	if err != nil {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
	}
	if txIndex.IsZeroValue() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("NftID should not be empty"))
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
	if tx.GetType() == common.TxCustomTokenPrivacyType || nftID.String() == common.PRVCoinID.String() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidTxTypeError, errors.New("With tx custome token privacy, the tokenIDStr should not be PRV, but custom token"))
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
	record += request.index
	record += strconv.FormatUint(uint64(request.token0Amount), 10)
	record += strconv.FormatUint(uint64(request.token1Amount), 10)
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
		Index                 string            `json:"Index"`
		Token0Amount          uint64            `json:"Token0Amount"`
		Token1Amount          uint64            `json:"Token1Amount"`
		OtaReceiveTradingFees map[string]string `json:"OtaReceiveTradingFees"`
		metadataCommon.MetadataBase
	}{
		PoolPairID:            request.poolPairID,
		NftID:                 request.nftID,
		Index:                 request.index,
		Token0Amount:          request.token0Amount,
		Token1Amount:          request.token1Amount,
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
		Index                 string            `json:"Index"`
		Token0Amount          uint64            `json:"Token0Amount"`
		Token1Amount          uint64            `json:"Token1Amount"`
		OtaReceiveTradingFees map[string]string `json:"OtaReceiveTradingFees"`
		metadataCommon.MetadataBase
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	request.poolPairID = temp.PoolPairID
	request.nftID = temp.NftID
	request.index = temp.Index
	request.token0Amount = temp.Token0Amount
	request.token1Amount = temp.Token1Amount
	if temp.OtaReceiveTradingFees != nil {
		request.otaReceiveTradingFees = temp.OtaReceiveTradingFees
	}
	request.MetadataBase = temp.MetadataBase
	return nil
}

func (request *WithdrawLiquidityRequest) PoolPairID() string {
	return request.poolPairID
}

func (request *WithdrawLiquidityRequest) Index() string {
	return request.index
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
