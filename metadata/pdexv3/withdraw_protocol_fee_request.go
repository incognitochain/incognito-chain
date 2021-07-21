package pdexv3

import (
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/wallet"
)

type WithdrawalProtocolFeeRequest struct {
	metadataCommon.MetadataBaseWithSignature
	PairID string `json:"PairID"`
}

type WithdrawalProtocolFeeContent struct {
	PairID  string      `json:"PairID"`
	TxReqID common.Hash `json:"TxReqID"`
	ShardID byte        `json:"ShardID"`
}

type WithdrawalProtocolFeeStatus struct {
	Status int    `json:"Status"`
	PairID string `json:"PairID"`
}

func NewPDexV3WithdrawalProtocolFeeStatus(
	status int,
	pairID string,
) *WithdrawalProtocolFeeStatus {
	return &WithdrawalProtocolFeeStatus{
		PairID: pairID,
		Status: status,
	}
}

func NewPDexV3WithdrawalProtocolFeeRequest(
	metaType int,
	pairID string,
) (*WithdrawalProtocolFeeRequest, error) {
	metadataBase := metadataCommon.NewMetadataBaseWithSignature(metaType)
	withdrawal := &WithdrawalProtocolFeeRequest{}
	withdrawal.MetadataBaseWithSignature = *metadataBase
	withdrawal.PairID = pairID

	return withdrawal, nil
}

func (withdrawal WithdrawalProtocolFeeRequest) ValidateTxWithBlockChain(
	tx metadataCommon.Transaction,
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	return true, nil
}

func (withdrawal WithdrawalProtocolFeeRequest) ValidateSanityData(
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	beaconHeight uint64,
	tx metadataCommon.Transaction,
) (bool, bool, error) {
	// validate IncAddressStr
	keyWallet, err := wallet.Base58CheckDeserialize(config.Param().PDexParams.AdminAddress)
	if err != nil {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDexV3WitdrawProtocolFeeValidateSanityDataError, errors.New("Requester incognito address is invalid"))
	}
	incAddr := keyWallet.KeySet.PaymentAddress
	if len(incAddr.Pk) == 0 {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDexV3WitdrawProtocolFeeValidateSanityDataError, errors.New("Requester incognito address is invalid"))
	}

	if ok, err := withdrawal.MetadataBaseWithSignature.VerifyMetadataSignature(incAddr.Pk, tx); err != nil || !ok {
		return false, false, errors.New("Sender is unauthorized")
	}

	// check tx type and version
	if tx.GetType() != common.TxNormalType {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDexV3WitdrawProtocolFeeValidateSanityDataError, errors.New("Tx pDex v3 protocol fee withdrawal must be TxNormalType"))
	}

	if tx.GetVersion() != 2 {
		return false, false, metadataCommon.NewMetadataTxError(0, errors.New("Tx pDex v3 protocol fee withdrawal must be version 2"))
	}

	return true, true, nil
}

func (withdrawal WithdrawalProtocolFeeRequest) ValidateMetadataByItself() bool {
	return withdrawal.Type == metadataCommon.PDexV3WithdrawProtocolFeeRequestMeta
}

func (withdrawal WithdrawalProtocolFeeRequest) Hash() *common.Hash {
	record := withdrawal.MetadataBaseWithSignature.Hash().String()
	record += withdrawal.PairID

	if withdrawal.Sig != nil && len(withdrawal.Sig) != 0 {
		record += string(withdrawal.Sig)
	}

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (withdrawal WithdrawalProtocolFeeRequest) HashWithoutSig() *common.Hash {
	record := withdrawal.MetadataBaseWithSignature.Hash().String()
	record += withdrawal.PairID

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (withdrawal *WithdrawalProtocolFeeRequest) BuildReqActions(
	tx metadataCommon.Transaction,
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	shardID byte,
	shardHeight uint64,
) ([][]string, error) {
	return [][]string{}, nil
}

func (withdrawal *WithdrawalProtocolFeeRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(withdrawal)
}
