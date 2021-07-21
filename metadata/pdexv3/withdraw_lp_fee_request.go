package pdexv3

import (
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/wallet"
)

type WithdrawalLPFeeRequest struct {
	metadataCommon.MetadataBase
	PairID      string `json:"PairID"`
	NcftTokenID string `json:"NcftTokenID"`
}

type WithdrawalLPFeeContent struct {
	PairID      string      `json:"PairID"`
	NcftTokenID string      `json:"NcftTokenID"`
	TxReqID     common.Hash `json:"TxReqID"`
	ShardID     byte        `json:"ShardID"`
}

type WithdrawalLPFeeStatus struct {
	Status      int    `json:"Status"`
	PairID      string `json:"PairID"`
	NcftTokenID string `json:"NcftTokenID"`
}

func NewPDexV3WithdrawalLPFeeStatus(
	status int,
	pairID string,
	ncftTokenID string,
) *WithdrawalLPFeeStatus {
	return &WithdrawalLPFeeStatus{
		PairID:      pairID,
		NcftTokenID: ncftTokenID,
		Status:      status,
	}
}

func NewPDexV3WithdrawalLPFeeRequest(
	metaType int,
	pairID string,
	ncftTokenID string,
) (*WithdrawalLPFeeRequest, error) {
	metadataBase := metadataCommon.NewMetadataBase(metaType)
	withdrawal := &WithdrawalLPFeeRequest{}
	withdrawal.MetadataBase = *metadataBase
	withdrawal.PairID = pairID
	withdrawal.NcftTokenID = ncftTokenID

	return withdrawal, nil
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
	// validate IncAddressStr
	keyWallet, err := wallet.Base58CheckDeserialize(config.Param().PDexParams.AdminAddress)
	if err != nil {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDexV3WitdrawLPFeeValidateSanityDataError, errors.New("Requester incognito address is invalid"))
	}
	incAddr := keyWallet.KeySet.PaymentAddress
	if len(incAddr.Pk) == 0 {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDexV3WitdrawLPFeeValidateSanityDataError, errors.New("Requester incognito address is invalid"))
	}

	// check tx type and version
	if tx.GetType() != common.TxNormalType {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDexV3WitdrawLPFeeValidateSanityDataError, errors.New("Tx pDex v3 LP fee withdrawal must be TxNormalType"))
	}

	if tx.GetVersion() != 2 {
		return false, false, metadataCommon.NewMetadataTxError(0, errors.New("Tx pDex v3 LP fee withdrawal must be version 2"))
	}

	return true, true, nil
}

func (withdrawal WithdrawalLPFeeRequest) ValidateMetadataByItself() bool {
	return withdrawal.Type == metadataCommon.PDexV3WithdrawLPFeeRequestMeta
}

func (withdrawal WithdrawalLPFeeRequest) Hash() *common.Hash {
	record := withdrawal.MetadataBase.Hash().String()
	record += withdrawal.PairID
	record += withdrawal.NcftTokenID

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (withdrawal *WithdrawalLPFeeRequest) BuildReqActions(
	tx metadataCommon.Transaction,
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	shardID byte,
	shardHeight uint64,
) ([][]string, error) {
	return [][]string{}, nil
}

func (withdrawal *WithdrawalLPFeeRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(withdrawal)
}
