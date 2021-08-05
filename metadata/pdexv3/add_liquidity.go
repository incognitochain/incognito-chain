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
	"github.com/incognitochain/incognito-chain/utils"
)

type AddLiquidityRequest struct {
	poolPairID     string // only "" for the first contribution of pool
	pairHash       string
	receiveAddress string // receive nfct
	refundAddress  string // refund pToken
	tokenID        string
	nftID          string
	tokenAmount    uint64
	amplifier      uint // only set for the first contribution
	metadataCommon.MetadataBase
}

func NewAddLiquidity() *AddLiquidityRequest {
	return &AddLiquidityRequest{}
}

func NewAddLiquidityRequestWithValue(
	poolPairID, pairHash,
	receiveAddress, refundAddress,
	tokenID, nftID string, tokenAmount uint64, amplifier uint,
) *AddLiquidityRequest {
	metadataBase := metadataCommon.MetadataBase{
		Type: metadataCommon.Pdexv3AddLiquidityRequestMeta,
	}
	return &AddLiquidityRequest{
		poolPairID:     poolPairID,
		pairHash:       pairHash,
		receiveAddress: receiveAddress,
		refundAddress:  refundAddress,
		tokenID:        tokenID,
		nftID:          nftID,
		tokenAmount:    tokenAmount,
		amplifier:      amplifier,
		MetadataBase:   metadataBase,
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
	// NOTE: verify supported tokens pair as needed
	return true, nil
}

func (request *AddLiquidityRequest) ValidateSanityData(
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	beaconHeight uint64,
	tx metadataCommon.Transaction,
) (bool, bool, error) {
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
	if request.nftID != utils.EmptyString {
		nftID, err := common.Hash{}.NewHashFromStr(request.nftID)
		if err != nil {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
		}
		if nftID.IsZeroValue() {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("TokenID should not be empty"))
		}
	}
	receiveAddress := privacy.OTAReceiver{}
	err = receiveAddress.FromString(request.receiveAddress)
	if err != nil {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
	}
	if !receiveAddress.IsValid() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("ReceiveAddress is not valid"))
	}
	refundAddress := privacy.OTAReceiver{}
	err = refundAddress.FromString(request.refundAddress)
	if err != nil {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
	}
	if !refundAddress.IsValid() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("RefundAddress is not valid"))
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
	if tx.GetType() == common.TxNormalType && tokenID.String() != common.PRVCoinID.String() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidTxTypeError, errors.New("With tx normal privacy, the tokenIDStr should be PRV, not custom token"))
	}
	if tx.GetType() == common.TxCustomTokenPrivacyType && tokenID.String() == common.PRVCoinID.String() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidTxTypeError, errors.New("With tx custome token privacy, the tokenIDStr should not be PRV, but custom token"))
	}
	return true, true, nil
}

func (request *AddLiquidityRequest) ValidateMetadataByItself() bool {
	return request.Type == metadataCommon.Pdexv3AddLiquidityRequestMeta
}

func (request *AddLiquidityRequest) Hash() *common.Hash {
	record := request.MetadataBase.Hash().String()
	record += request.poolPairID
	record += request.pairHash
	record += request.receiveAddress
	record += request.refundAddress
	record += request.tokenID
	record += request.nftID
	record += strconv.FormatUint(uint64(request.amplifier), 10)
	record += strconv.FormatUint(request.tokenAmount, 10)
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (request *AddLiquidityRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(request)
}

func (request *AddLiquidityRequest) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		PoolPairID     string `json:"PoolPairID"` // only "" for the first contribution of pool
		PairHash       string `json:"PairHash"`
		ReceiveAddress string `json:"ReceiveAddress"` // receive nfct
		RefundAddress  string `json:"RefundAddress"`  // refund pToken
		TokenID        string `json:"TokenID"`
		NftID          string `json:"NftID"`
		TokenAmount    uint64 `json:"TokenAmount"`
		Amplifier      uint   `json:"Amplifier"` // only set for the first contribution
		metadataCommon.MetadataBase
	}{
		PoolPairID:     request.poolPairID,
		PairHash:       request.pairHash,
		ReceiveAddress: request.receiveAddress,
		RefundAddress:  request.refundAddress,
		TokenID:        request.tokenID,
		NftID:          request.nftID,
		TokenAmount:    request.tokenAmount,
		Amplifier:      request.amplifier,
		MetadataBase:   request.MetadataBase,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (request *AddLiquidityRequest) UnmarshalJSON(data []byte) error {
	temp := struct {
		PoolPairID     string `json:"PoolPairID"` // only "" for the first contribution of pool
		PairHash       string `json:"PairHash"`
		ReceiveAddress string `json:"ReceiveAddress"` // receive nfct
		RefundAddress  string `json:"RefundAddress"`  // refund pToken
		TokenID        string `json:"TokenID"`
		NftID          string `json:"NftID"`
		TokenAmount    uint64 `json:"TokenAmount"`
		Amplifier      uint   `json:"Amplifier"` // only set for the first contribution
		metadataCommon.MetadataBase
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	request.poolPairID = temp.PoolPairID
	request.pairHash = temp.PairHash
	request.receiveAddress = temp.ReceiveAddress
	request.refundAddress = temp.RefundAddress
	request.tokenID = temp.TokenID
	request.nftID = temp.NftID
	request.tokenAmount = temp.TokenAmount
	request.amplifier = temp.Amplifier
	request.MetadataBase = temp.MetadataBase
	return nil
}

func (request *AddLiquidityRequest) PoolPairID() string {
	return request.poolPairID
}

func (request *AddLiquidityRequest) PairHash() string {
	return request.pairHash
}

func (request *AddLiquidityRequest) ReceiveAddress() string {
	return request.receiveAddress
}

func (request *AddLiquidityRequest) RefundAddress() string {
	return request.refundAddress
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

func (request *AddLiquidityRequest) NftID() string {
	return request.nftID
}

type AddLiquidityResponse struct {
	metadataCommon.MetadataBase
	status  string
	txReqID string
}

func NewAddLiquidityResponse() *AddLiquidityResponse {
	return &AddLiquidityResponse{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.Pdexv3AddLiquidityResponseMeta,
		},
	}
}

func NewAddLiquidityResponseWithValue(
	status, txReqID string,
) *AddLiquidityResponse {
	return &AddLiquidityResponse{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.Pdexv3AddLiquidityResponseMeta,
		},
		status:  status,
		txReqID: txReqID,
	}
}

func (response *AddLiquidityResponse) ValidateTxWithBlockChain(
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

func (response *AddLiquidityResponse) ValidateSanityData(
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	beaconHeight uint64,
	tx metadataCommon.Transaction,
) (bool, bool, error) {
	if response.status == "" {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("status can not be empty"))
	}
	txReqID, err := common.Hash{}.NewHashFromStr(response.txReqID)
	if err != nil {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
	}
	if txReqID.IsZeroValue() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("TxReqID should not be empty"))
	}
	return true, true, nil
}

func (response *AddLiquidityResponse) ValidateMetadataByItself() bool {
	return response.Type == metadataCommon.Pdexv3AddLiquidityResponseMeta
}

func (response *AddLiquidityResponse) Hash() *common.Hash {
	record := response.MetadataBase.Hash().String()
	record += response.status
	record += response.txReqID
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (response *AddLiquidityResponse) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(response)
}

func (response *AddLiquidityResponse) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Status  string `json:"Status"`
		TxReqID string `json:"TxReqID"`
		metadataCommon.MetadataBase
	}{
		Status:       response.status,
		TxReqID:      response.txReqID,
		MetadataBase: response.MetadataBase,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (response *AddLiquidityResponse) UnmarshalJSON(data []byte) error {
	temp := struct {
		Status  string `json:"Status"`
		TxReqID string `json:"TxReqID"`
		metadataCommon.MetadataBase
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	response.txReqID = temp.TxReqID
	response.status = temp.Status
	response.MetadataBase = temp.MetadataBase
	return nil
}

func (response *AddLiquidityResponse) TxReqID() string {
	return response.txReqID
}

func (response *AddLiquidityResponse) Status() string {
	return response.status
}
