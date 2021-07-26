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
	PairID             string             `json:"PairID"`
	FeeReceiverAddress FeeReceiverAddress `json:"FeeReceiverAddress"`
}

type WithdrawalProtocolFeeContent struct {
	PairID             string             `json:"PairID"`
	FeeReceiverAddress FeeReceiverAddress `json:"FeeReceiverAddress"`
	TxReqID            common.Hash        `json:"TxReqID"`
	ShardID            byte               `json:"ShardID"`
}

type WithdrawalProtocolFeeStatus struct {
	Status             int                `json:"Status"`
	PairID             string             `json:"PairID"`
	FeeReceiverAddress FeeReceiverAddress `json:"FeeReceiverAddress"`
}

func NewPdexv3WithdrawalProtocolFeeStatus(
	status int,
	pairID string,
	feeReceiverAddress FeeReceiverAddress,
) *WithdrawalProtocolFeeStatus {
	return &WithdrawalProtocolFeeStatus{
		PairID:             pairID,
		FeeReceiverAddress: feeReceiverAddress,
		Status:             status,
	}
}

func NewPdexv3WithdrawalProtocolFeeRequest(
	metaType int,
	pairID string,
	feeReceiverAddress FeeReceiverAddress,
) (*WithdrawalProtocolFeeRequest, error) {
	metadataBase := metadataCommon.NewMetadataBaseWithSignature(metaType)

	return &WithdrawalProtocolFeeRequest{
		MetadataBaseWithSignature: *metadataBase,
		PairID:                    pairID,
		FeeReceiverAddress:        feeReceiverAddress,
	}, nil
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
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.Pdexv3WitdrawProtocolFeeValidateSanityDataError, errors.New("Requester incognito address is invalid"))
	}
	incAddr := keyWallet.KeySet.PaymentAddress
	if len(incAddr.Pk) == 0 {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.Pdexv3WitdrawProtocolFeeValidateSanityDataError, errors.New("Requester incognito address is invalid"))
	}

	if ok, err := withdrawal.MetadataBaseWithSignature.VerifyMetadataSignature(incAddr.Pk, tx); err != nil || !ok {
		return false, false, errors.New("Sender is unauthorized")
	}

	// check tx type and version
	if tx.GetType() != common.TxNormalType {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.Pdexv3WitdrawProtocolFeeValidateSanityDataError, errors.New("Tx pDex v3 protocol fee withdrawal must be TxNormalType"))
	}

	if tx.GetVersion() != 2 {
		return false, false, metadataCommon.NewMetadataTxError(0, errors.New("Tx pDex v3 protocol fee withdrawal must be version 2"))
	}

	// TODO: Check OTA address string and tx random is valid

	return true, true, nil
}

func (withdrawal WithdrawalProtocolFeeRequest) ValidateMetadataByItself() bool {
	return withdrawal.Type == metadataCommon.Pdexv3WithdrawProtocolFeeRequestMeta
}

func (withdrawal WithdrawalProtocolFeeRequest) Hash() *common.Hash {
	record := withdrawal.MetadataBaseWithSignature.Hash().String()
	record += withdrawal.PairID
	record += withdrawal.FeeReceiverAddress.ToString()
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
	record += withdrawal.FeeReceiverAddress.ToString()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (withdrawal *WithdrawalProtocolFeeRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(withdrawal)
}
