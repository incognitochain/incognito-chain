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
	NfctTokenID         string             `json:"NfctTokenID"`
	NfctReceiverAddress string             `json:"NfctReceiverAddress"`
	FeeReceiverAddress  FeeReceiverAddress `json:"FeeReceiverAddress"`
}

type WithdrawalLPFeeContent struct {
	PairID              string             `json:"PairID"`
	NfctTokenID         string             `json:"NfctTokenID"`
	NfctReceiverAddress string             `json:"NfctReceiverAddress"`
	FeeReceiverAddress  FeeReceiverAddress `json:"FeeReceiverAddress"`
	TxReqID             common.Hash        `json:"TxReqID"`
	ShardID             byte               `json:"ShardID"`
}

type WithdrawalLPFeeStatus struct {
	Status              int                `json:"Status"`
	PairID              string             `json:"PairID"`
	NfctTokenID         string             `json:"NfctTokenID"`
	NfctReceiverAddress string             `json:"NfctReceiverAddress"`
	FeeReceiverAddress  FeeReceiverAddress `json:"FeeReceiverAddress"`
}

func NewPdexv3WithdrawalLPFeeStatus(
	status int,
	pairID string,
	nfctTokenID string,
	nfctReceiverAddress string,
	feeReceiverAddress FeeReceiverAddress,
) *WithdrawalLPFeeStatus {
	return &WithdrawalLPFeeStatus{
		PairID:              pairID,
		NfctTokenID:         nfctTokenID,
		NfctReceiverAddress: nfctReceiverAddress,
		FeeReceiverAddress:  feeReceiverAddress,
		Status:              status,
	}
}

func NewPdexv3WithdrawalLPFeeRequest(
	metaType int,
	pairID string,
	nfctTokenID string,
	nfctReceiverAddress string,
	feeReceiverAddress FeeReceiverAddress,
) (*WithdrawalLPFeeRequest, error) {
	metadataBase := metadataCommon.NewMetadataBase(metaType)

	return &WithdrawalLPFeeRequest{
		MetadataBase:        *metadataBase,
		PairID:              pairID,
		NfctTokenID:         nfctTokenID,
		NfctReceiverAddress: nfctReceiverAddress,
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
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.Pdexv3WithdrawLPFeeValidateSanityDataError, errors.New("Tx pDex v3 LP fee withdrawal must be TxCustomTokenPrivacyType"))
	}

	if tx.GetVersion() != 2 {
		return false, false, metadataCommon.NewMetadataTxError(0, errors.New("Tx pDex v3 LP fee withdrawal must be version 2"))
	}

	// validate burn tx, tokenID & amount = 1
	isBurn, _, burnedCoin, burnedToken, err := tx.GetTxFullBurnData()
	if err != nil || !isBurn {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.Pdexv3WithdrawLPFeeValidateSanityDataError, fmt.Errorf("Tx is not a burn tx. Error %v", err))
	}
	burningAmt := burnedCoin.GetValue()
	burningTokenID := burnedToken.String()
	if burningAmt != 1 || burningTokenID != withdrawal.NfctTokenID {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.Pdexv3WithdrawLPFeeValidateSanityDataError, fmt.Errorf("Burning token ID or amount is wrong. Error %v", err))
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
	record += withdrawal.NfctTokenID
	record += withdrawal.NfctReceiverAddress
	record += withdrawal.FeeReceiverAddress.ToString()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (withdrawal *WithdrawalLPFeeRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(withdrawal)
}
