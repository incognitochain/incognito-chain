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
	PoolPairID         string             `json:"PoolPairID"`
	NftID              common.Hash        `json:"NftID"`
	NftReceiverAddress string             `json:"NftReceiverAddress"`
	FeeReceiverAddress FeeReceiverAddress `json:"FeeReceiverAddress"`
}

type WithdrawalLPFeeContent struct {
	PoolPairID string       `json:"PoolPairID"`
	NftID      common.Hash  `json:"NftID"`
	TokenType  string       `json:"TokenType"`
	Receiver   ReceiverInfo `json:"Receiver"`
	TxReqID    common.Hash  `json:"TxReqID"`
	ShardID    byte         `json:"ShardID"`
}

func NewPdexv3WithdrawalLPFeeRequest(
	metaType int,
	pairID string,
	nftID common.Hash,
	nftReceiverAddress string,
	feeReceiverAddress FeeReceiverAddress,
) (*WithdrawalLPFeeRequest, error) {
	metadataBase := metadataCommon.NewMetadataBase(metaType)

	return &WithdrawalLPFeeRequest{
		MetadataBase:       *metadataBase,
		PoolPairID:         pairID,
		NftID:              nftID,
		NftReceiverAddress: nftReceiverAddress,
		FeeReceiverAddress: feeReceiverAddress,
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
	burningTokenID := burnedToken
	if burningAmt != 1 || *burningTokenID != withdrawal.NftID {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.Pdexv3WithdrawLPFeeValidateSanityDataError, fmt.Errorf("Burning token ID or amount is wrong. Error %v", err))
	}

	// Check OTA address string and tx random is valid
	shardID := byte(tx.GetValidationEnv().ShardID())
	_, err = isValidReceiverAddressStr(withdrawal.NftReceiverAddress, shardID)
	if err != nil {
		return false, false, err
	}
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

func (withdrawal WithdrawalLPFeeRequest) ValidateMetadataByItself() bool {
	return withdrawal.Type == metadataCommon.Pdexv3WithdrawLPFeeRequestMeta
}

func (withdrawal WithdrawalLPFeeRequest) Hash() *common.Hash {
	record := withdrawal.MetadataBase.Hash().String()
	record += withdrawal.PoolPairID
	record += withdrawal.NftID.String()
	record += withdrawal.NftReceiverAddress
	record += withdrawal.FeeReceiverAddress.ToString()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (withdrawal *WithdrawalLPFeeRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(withdrawal)
}
