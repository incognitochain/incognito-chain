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
	PoolPairID         string             `json:"PoolPairID"`
	FeeReceiverAddress FeeReceiverAddress `json:"FeeReceiverAddress"`
}

type WithdrawalProtocolFeeContent struct {
	PoolPairID string       `json:"PoolPairID"`
	TokenType  string       `json:"TokenType"`
	Receiver   ReceiverInfo `json:"Receiver"`
	TxReqID    common.Hash  `json:"TxReqID"`
	ShardID    byte         `json:"ShardID"`
}

func NewPdexv3WithdrawalProtocolFeeRequest(
	metaType int,
	pairID string,
	feeReceiverAddress FeeReceiverAddress,
) (*WithdrawalProtocolFeeRequest, error) {
	metadataBase := metadataCommon.NewMetadataBaseWithSignature(metaType)

	return &WithdrawalProtocolFeeRequest{
		MetadataBaseWithSignature: *metadataBase,
		PoolPairID:                pairID,
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
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.Pdexv3WithdrawProtocolFeeValidateSanityDataError, errors.New("Requester incognito address is invalid"))
	}
	incAddr := keyWallet.KeySet.PaymentAddress
	if len(incAddr.Pk) == 0 {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.Pdexv3WithdrawProtocolFeeValidateSanityDataError, errors.New("Requester incognito address is invalid"))
	}

	if ok, err := withdrawal.MetadataBaseWithSignature.VerifyMetadataSignature(incAddr.Pk, tx); err != nil || !ok {
		return false, false, errors.New("Sender is unauthorized")
	}

	// check tx type and version
	if tx.GetType() != common.TxNormalType {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.Pdexv3WithdrawProtocolFeeValidateSanityDataError, errors.New("Tx pDex v3 protocol fee withdrawal must be TxNormalType"))
	}

	if tx.GetVersion() != 2 {
		return false, false, metadataCommon.NewMetadataTxError(0, errors.New("Tx pDex v3 protocol fee withdrawal must be version 2"))
	}

	// Check OTA address string and tx random is valid
	shardID := byte(tx.GetValidationEnv().ShardID())
	_, err = isValidReceiverAddressStr(withdrawal.FeeReceiverAddress.Token0ReceiverAddress, shardID)
	if err != nil {
		return false, false, err
	}
	_, err = isValidReceiverAddressStr(withdrawal.FeeReceiverAddress.Token1ReceiverAddress, shardID)
	if err != nil {
		return false, false, err
	}
	_, err = isValidReceiverAddressStr(withdrawal.FeeReceiverAddress.PRVReceiverAddress, shardID)
	if err != nil {
		return false, false, err
	}
	_, err = isValidReceiverAddressStr(withdrawal.FeeReceiverAddress.PDEXReceiverAddress, shardID)
	if err != nil {
		return false, false, err
	}

	return true, true, nil
}

func (withdrawal WithdrawalProtocolFeeRequest) ValidateMetadataByItself() bool {
	return withdrawal.Type == metadataCommon.Pdexv3WithdrawProtocolFeeRequestMeta
}

func (withdrawal WithdrawalProtocolFeeRequest) Hash() *common.Hash {
	record := withdrawal.MetadataBaseWithSignature.Hash().String()
	record += withdrawal.PoolPairID
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
	record += withdrawal.PoolPairID
	record += withdrawal.FeeReceiverAddress.ToString()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (withdrawal *WithdrawalProtocolFeeRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(withdrawal)
}
