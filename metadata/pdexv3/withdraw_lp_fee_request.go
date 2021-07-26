package pdexv3

import (
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type WithdrawalLPFeeRequest struct {
	metadataCommon.MetadataBase
	PairID              string             `json:"PairID"`
	NcftTokenID         string             `json:"NcftTokenID"`
	NcftReceiverAddress string             `json:"NcftReceiverAddress"`
	FeeReceiverAddress  FeeReceiverAddress `json:"FeeReceiverAddress"`
}

type WithdrawalLPFeeContent struct {
	PairID              string             `json:"PairID"`
	NcftTokenID         string             `json:"NcftTokenID"`
	NcftReceiverAddress string             `json:"NcftReceiverAddress"`
	FeeReceiverAddress  FeeReceiverAddress `json:"FeeReceiverAddress"`
	TxReqID             common.Hash        `json:"TxReqID"`
	ShardID             byte               `json:"ShardID"`
}

type WithdrawalLPFeeStatus struct {
	Status              int                `json:"Status"`
	PairID              string             `json:"PairID"`
	NcftTokenID         string             `json:"NcftTokenID"`
	NcftReceiverAddress string             `json:"NcftReceiverAddress"`
	FeeReceiverAddress  FeeReceiverAddress `json:"FeeReceiverAddress"`
}

func NewPdexv3WithdrawalLPFeeStatus(
	status int,
	pairID string,
	ncftTokenID string,
	ncftReceiverAddress string,
	feeReceiverAddress FeeReceiverAddress,
) *WithdrawalLPFeeStatus {
	return &WithdrawalLPFeeStatus{
		PairID:              pairID,
		NcftTokenID:         ncftTokenID,
		NcftReceiverAddress: ncftReceiverAddress,
		FeeReceiverAddress:  feeReceiverAddress,
		Status:              status,
	}
}

func NewPdexv3WithdrawalLPFeeRequest(
	metaType int,
	pairID string,
	ncftTokenID string,
	ncftReceiverAddress string,
	feeReceiverAddress FeeReceiverAddress,
) (*WithdrawalLPFeeRequest, error) {
	metadataBase := metadataCommon.NewMetadataBase(metaType)

	return &WithdrawalLPFeeRequest{
		MetadataBase:        *metadataBase,
		PairID:              pairID,
		NcftTokenID:         ncftTokenID,
		NcftReceiverAddress: ncftReceiverAddress,
		FeeReceiverAddress:  feeReceiverAddress,
	}, nil
}

func (withdrawal WithdrawalLPFeeRequest) ValidateTxWithBlockChain(
	tx metadataCommon.Transaction,
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	return true, nil
}

func (withdrawal WithdrawalLPFeeRequest) ValidateSanityData(
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	beaconHeight uint64,
	tx metadataCommon.Transaction,
) (bool, bool, error) {
	// check tx type and version
	if tx.GetType() != common.TxCustomTokenPrivacyType {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.Pdexv3WitdrawLPFeeValidateSanityDataError, errors.New("Tx pDex v3 LP fee withdrawal must be TxCustomTokenPrivacyType"))
	}

	if tx.GetVersion() != 2 {
		return false, false, metadataCommon.NewMetadataTxError(0, errors.New("Tx pDex v3 LP fee withdrawal must be version 2"))
	}

	// validate burn tx, tokenID & amount = 1
	isBurn, _, burnedCoin, burnedToken, err := tx.GetTxFullBurnData()
	if err != nil || !isBurn {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.Pdexv3WitdrawLPFeeValidateSanityDataError, fmt.Errorf("Tx is not a burn tx. Error %v", err))
	}
	burningAmt := burnedCoin.GetValue()
	burningTokenID := burnedToken.String()
	if burningAmt != 1 || burningTokenID != withdrawal.NcftTokenID {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.Pdexv3WitdrawLPFeeValidateSanityDataError, fmt.Errorf("Burning token ID or amount is wrong. Error %v", err))
	}

	// TODO: Check OTA address string and tx random is valid

	return true, true, nil
}

func (withdrawal WithdrawalLPFeeRequest) ValidateMetadataByItself() bool {
	return withdrawal.Type == metadataCommon.Pdexv3WithdrawLPFeeRequestMeta
}

func (withdrawal WithdrawalLPFeeRequest) Hash() *common.Hash {
	record := withdrawal.MetadataBase.Hash().String()
	record += withdrawal.PairID
	record += withdrawal.NcftTokenID
	record += withdrawal.NcftReceiverAddress
	record += withdrawal.FeeReceiverAddress.ToString()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (withdrawal *WithdrawalLPFeeRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(withdrawal)
}
