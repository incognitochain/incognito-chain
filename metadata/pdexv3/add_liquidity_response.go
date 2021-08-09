package pdexv3

import (
	"encoding/json"
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

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

func (response *AddLiquidityResponse) CheckTransactionFee(tx metadataCommon.Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
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

type RefundAddLiquidity struct {
	Contribution *statedb.Pdexv3ContributionState `json:"Contribution"`
}

type MatchAddLiquidity struct {
	Contribution  statedb.Pdexv3ContributionState `json:"Contribution"`
	NewPoolPairID string                          `json:"NewPoolPairID"`
	NftID         common.Hash                     `json:"NftID"`
}

type MatchAndReturnAddLiquidity struct {
	ShareAmount              uint64                           `json:"ShareAmount"`
	Contribution             *statedb.Pdexv3ContributionState `json:"Contribution"`
	ReturnAmount             uint64                           `json:"ReturnAmount"`
	ExistedTokenActualAmount uint64                           `json:"ExistedTokenActualAmount"`
	ExistedTokenReturnAmount uint64                           `json:"ExistedTokenReturnAmount"`
	ExistedTokenID           common.Hash                      `json:"ExistedTokenID"`
	NftID                    common.Hash                      `json:"NftID"`
}

func (response *AddLiquidityResponse) VerifyMinerCreatedTxBeforeGettingInBlock(
	mintData *metadataCommon.MintData,
	shardID byte,
	tx metadataCommon.Transaction,
	chainRetriever metadataCommon.ChainRetriever,
	ac *metadataCommon.AccumulatedValues,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
) (bool, error) {
	return true, nil
}
