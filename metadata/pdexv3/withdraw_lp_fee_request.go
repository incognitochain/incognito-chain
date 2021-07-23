package pdexv3

import (
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type WithdrawalLPFeeRequest struct {
	metadataCommon.MetadataBase
	PairID                string `json:"PairID"`
	NcftTokenID           string `json:"NcftTokenID"`
	Token0ReceiverAddress string `json:"Token0ReceiverAddress"`
	Token1ReceiverAddress string `json:"Token1ReceiverAddress"`
	PRVReceiverAddress    string `json:"PRVReceiverAddress"`
	PDEXReceiverAddress   string `json:"PDEXReceiverAddress"`
}

type WithdrawalLPFeeContent struct {
	PairID                string      `json:"PairID"`
	NcftTokenID           string      `json:"NcftTokenID"`
	Token0ReceiverAddress string      `json:"Token0ReceiverAddress"`
	Token1ReceiverAddress string      `json:"Token1ReceiverAddress"`
	PRVReceiverAddress    string      `json:"PRVReceiverAddress"`
	PDEXReceiverAddress   string      `json:"PDEXReceiverAddress"`
	TxReqID               common.Hash `json:"TxReqID"`
	ShardID               byte        `json:"ShardID"`
}

type WithdrawalLPFeeStatus struct {
	Status                int    `json:"Status"`
	PairID                string `json:"PairID"`
	NcftTokenID           string `json:"NcftTokenID"`
	Token0ReceiverAddress string `json:"Token0ReceiverAddress"`
	Token1ReceiverAddress string `json:"Token1ReceiverAddress"`
	PRVReceiverAddress    string `json:"PRVReceiverAddress"`
	PDEXReceiverAddress   string `json:"PDEXReceiverAddress"`
}

func NewPdexv3WithdrawalLPFeeStatus(
	status int,
	pairID string,
	ncftTokenID string,
	token0ReceiverAddress string,
	token1ReceiverAddress string,
	prvReceiverAddress string,
	pdexReceiverAddress string,
) *WithdrawalLPFeeStatus {
	return &WithdrawalLPFeeStatus{
		PairID:                pairID,
		NcftTokenID:           ncftTokenID,
		Token0ReceiverAddress: token0ReceiverAddress,
		Token1ReceiverAddress: token1ReceiverAddress,
		PRVReceiverAddress:    prvReceiverAddress,
		PDEXReceiverAddress:   pdexReceiverAddress,
		Status:                status,
	}
}

func NewPdexv3WithdrawalLPFeeRequest(
	metaType int,
	pairID string,
	ncftTokenID string,
	token0ReceiverAddress string,
	token1ReceiverAddress string,
	prvReceiverAddress string,
	pdexReceiverAddress string,
) (*WithdrawalLPFeeRequest, error) {
	metadataBase := metadataCommon.NewMetadataBase(metaType)

	return &WithdrawalLPFeeRequest{
		MetadataBase:          *metadataBase,
		PairID:                pairID,
		NcftTokenID:           ncftTokenID,
		Token0ReceiverAddress: token0ReceiverAddress,
		Token1ReceiverAddress: token1ReceiverAddress,
		PRVReceiverAddress:    prvReceiverAddress,
		PDEXReceiverAddress:   pdexReceiverAddress,
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

	// TODO: Check OTA address string and tx random is valid

	// TODO: validate receiver & amount = 1

	// TODO: validate transfer token & Ncft Token ID

	return true, true, nil
}

func (withdrawal WithdrawalLPFeeRequest) ValidateMetadataByItself() bool {
	return withdrawal.Type == metadataCommon.Pdexv3WithdrawLPFeeRequestMeta
}

func (withdrawal WithdrawalLPFeeRequest) Hash() *common.Hash {
	record := withdrawal.MetadataBase.Hash().String()
	record += withdrawal.PairID
	record += withdrawal.NcftTokenID
	record += withdrawal.Token0ReceiverAddress
	record += withdrawal.Token1ReceiverAddress
	record += withdrawal.PRVReceiverAddress
	record += withdrawal.PDEXReceiverAddress

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (withdrawal *WithdrawalLPFeeRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(withdrawal)
}
