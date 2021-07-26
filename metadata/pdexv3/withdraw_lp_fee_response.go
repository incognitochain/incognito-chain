package pdexv3

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type WithdrawalLPFeeResponse struct {
	metadataCommon.MetadataBase
	RequestStatus       string             `json:"RequestStatus"`
	ReqTxID             common.Hash        `json:"ReqTxID"`
	PairID              string             `json:"PairID"`
	NcftTokenID         string             `json:"NcftTokenID"`
	NcftReceiverAddress string             `json:"NcftReceiverAddress"`
	FeeReceiverAddress  FeeReceiverAddress `json:"FeeReceiverAddress"`
	FeeReceiverAmount   FeeReceiverAmount  `json:"FeeReceiverAmount"`
}

func NewPdexv3WithdrawalLPFeeResponse(
	metaType int,
	requestStatus string,
	reqTxID common.Hash,
	pairID string,
	ncftTokenID string,
	ncftReceiverAddress string,
	feeReceiverAddress FeeReceiverAddress,
	feeReceiverAmount FeeReceiverAmount,
) *WithdrawalLPFeeResponse {
	metadataBase := metadataCommon.NewMetadataBase(metaType)

	return &WithdrawalLPFeeResponse{
		MetadataBase:        *metadataBase,
		RequestStatus:       requestStatus,
		ReqTxID:             reqTxID,
		PairID:              pairID,
		NcftTokenID:         ncftTokenID,
		NcftReceiverAddress: ncftReceiverAddress,
		FeeReceiverAddress:  feeReceiverAddress,
		FeeReceiverAmount:   feeReceiverAmount,
	}
}

func (withdrawalResponse WithdrawalLPFeeResponse) CheckTransactionFee(
	tx metadataCommon.Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB,
) bool {
	// no need to have fee for this tx
	return true
}

func (withdrawalResponse WithdrawalLPFeeResponse) ValidateTxWithBlockChain(
	tx metadataCommon.Transaction,
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID)
	return false, nil
}

func (withdrawalResponse WithdrawalLPFeeResponse) ValidateSanityData(
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	beaconHeight uint64,
	tx metadataCommon.Transaction,
) (bool, bool, error) {
	return false, true, nil
}

func (withdrawalResponse WithdrawalLPFeeResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return withdrawalResponse.Type == metadataCommon.Pdexv3WithdrawLPFeeResponseMeta
}

func (withdrawalResponse WithdrawalLPFeeResponse) Hash() *common.Hash {
	record := withdrawalResponse.MetadataBase.Hash().String()
	record += withdrawalResponse.RequestStatus
	record += withdrawalResponse.ReqTxID.String()
	record += withdrawalResponse.PairID
	record += withdrawalResponse.NcftTokenID
	record += withdrawalResponse.NcftReceiverAddress
	record += withdrawalResponse.FeeReceiverAddress.ToString()
	record += withdrawalResponse.FeeReceiverAmount.ToString()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (withdrawalResponse *WithdrawalLPFeeResponse) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(withdrawalResponse)
}

func (withdrawalResponse WithdrawalLPFeeResponse) VerifyMinerCreatedTxBeforeGettingInBlock(
	mintData *metadataCommon.MintData,
	shardID byte, tx metadataCommon.Transaction,
	chainRetriever metadataCommon.ChainRetriever,
	ac *metadataCommon.AccumulatedValues,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
) (bool, error) {
	// TODO: verify mining tx with the request tx
	return true, nil
}
