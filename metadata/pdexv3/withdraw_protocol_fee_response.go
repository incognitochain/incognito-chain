package pdexv3

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type WithdrawalProtocolFeeResponse struct {
	metadataCommon.MetadataBase
	RequestStatus      string             `json:"RequestStatus"`
	ReqTxID            common.Hash        `json:"ReqTxID"`
	PairID             string             `json:"PairID"`
	FeeReceiverAddress FeeReceiverAddress `json:"FeeReceiverAddress"`
	FeeReceiverAmount  FeeReceiverAmount  `json:"FeeReceiverAmount"`
}

func NewPdexv3WithdrawalProtocolFeeResponse(
	metaType int,
	requestStatus string,
	reqTxID common.Hash,
	pairID string,
	feeReceiverAddress FeeReceiverAddress,
	feeReceiverAmount FeeReceiverAmount,
) *WithdrawalProtocolFeeResponse {
	metadataBase := metadataCommon.NewMetadataBase(metaType)

	return &WithdrawalProtocolFeeResponse{
		MetadataBase:       *metadataBase,
		RequestStatus:      requestStatus,
		ReqTxID:            reqTxID,
		PairID:             pairID,
		FeeReceiverAddress: feeReceiverAddress,
		FeeReceiverAmount:  feeReceiverAmount,
	}
}

func (withdrawalResponse WithdrawalProtocolFeeResponse) CheckTransactionFee(
	tx metadataCommon.Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB,
) bool {
	// no need to have fee for this tx
	return true
}

func (withdrawalResponse WithdrawalProtocolFeeResponse) ValidateTxWithBlockChain(
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

func (withdrawalResponse WithdrawalProtocolFeeResponse) ValidateSanityData(
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	beaconHeight uint64,
	tx metadataCommon.Transaction,
) (bool, bool, error) {
	return false, true, nil
}

func (withdrawalResponse WithdrawalProtocolFeeResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return withdrawalResponse.Type == metadataCommon.Pdexv3WithdrawProtocolFeeResponseMeta
}

func (withdrawalResponse WithdrawalProtocolFeeResponse) Hash() *common.Hash {
	record := withdrawalResponse.MetadataBase.Hash().String()
	record += withdrawalResponse.RequestStatus
	record += withdrawalResponse.ReqTxID.String()
	record += withdrawalResponse.PairID
	record += withdrawalResponse.FeeReceiverAddress.ToString()
	record += withdrawalResponse.FeeReceiverAmount.ToString()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (withdrawalResponse *WithdrawalProtocolFeeResponse) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(withdrawalResponse)
}

func (withdrawalResponse WithdrawalProtocolFeeResponse) VerifyMinerCreatedTxBeforeGettingInBlock(
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
