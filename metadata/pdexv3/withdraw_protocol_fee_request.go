package pdexv3

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/wallet"
)

type WithdrawalProtocolFeeRequest struct {
	metadataCommon.MetadataBaseWithSignature
	PoolPairID string `json:"PoolPairID"`
}

type WithdrawalProtocolFeeContent struct {
	PoolPairID string      `json:"PoolPairID"`
	Address    string      `json:"Address"`
	TokenID    common.Hash `json:"TokenID"`
	Amount     uint64      `json:"Amount"`
	IsLastInst bool        `json:"IsLastInst"`
	TxReqID    common.Hash `json:"TxReqID"`
	ShardID    byte        `json:"ShardID"`
}

type WithdrawalProtocolFeeStatus struct {
	Status int                    `json:"Status"`
	Amount map[common.Hash]uint64 `json:"Amount"`
}

func NewPdexv3WithdrawalProtocolFeeRequest(
	metaType int,
	pairID string,
) (*WithdrawalProtocolFeeRequest, error) {
	metadataBase := metadataCommon.NewMetadataBaseWithSignature(metaType)

	return &WithdrawalProtocolFeeRequest{
		MetadataBaseWithSignature: *metadataBase,
		PoolPairID:                pairID,
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
	if !chainRetriever.IsAfterPdexv3CheckPoint(beaconViewRetriever.GetHeight()) {
		return false, fmt.Errorf("Feature pdexv3 has not been activated yet")
	}
	pdexv3StateCached := chainRetriever.GetPdexv3Cached(beaconViewRetriever.BlockHash())
	err := beaconViewRetriever.IsValidPoolPairID(chainRetriever.GetBeaconChainDatabase(), pdexv3StateCached, withdrawal.PoolPairID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (withdrawal WithdrawalProtocolFeeRequest) ValidateSanityData(
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	beaconHeight uint64,
	tx metadataCommon.Transaction,
) (bool, bool, error) {
	if !chainRetriever.IsAfterPdexv3CheckPoint(beaconHeight) {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, fmt.Errorf("Feature pdexv3 has not been activated yet"))
	}

	// validate IncAddressStr
	keyWallet, err := wallet.Base58CheckDeserialize(config.Param().PDexParams.ProtocolFundAddress)
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

	return true, true, nil
}

func (withdrawal WithdrawalProtocolFeeRequest) ValidateMetadataByItself() bool {
	return withdrawal.Type == metadataCommon.Pdexv3WithdrawProtocolFeeRequestMeta
}

func (withdrawal WithdrawalProtocolFeeRequest) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(withdrawal)
	if withdrawal.Sig != nil && len(withdrawal.Sig) != 0 {
		rawBytes = append(rawBytes, withdrawal.Sig...)
	}

	// final hash
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (withdrawal WithdrawalProtocolFeeRequest) HashWithoutSig() *common.Hash {
	rawBytes, _ := json.Marshal(struct {
		Type       int    `json:"Type"`
		PoolPairID string `json:"PoolPairID"`
	}{
		Type:       metadataCommon.Pdexv3WithdrawProtocolFeeRequestMeta,
		PoolPairID: withdrawal.PoolPairID,
	})

	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (withdrawal *WithdrawalProtocolFeeRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(withdrawal)
}

func (withdrawal *WithdrawalProtocolFeeRequest) ToCompactBytes() ([]byte, error) {
	return metadataCommon.ToCompactBytes(withdrawal)
}
