package pdexv3

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
)

type WithdrawalProtocolFeeRequest struct {
	metadataCommon.MetadataBaseWithSignature
	PoolPairID string                              `json:"PoolPairID"`
	Receivers  map[common.Hash]privacy.OTAReceiver `json:"Receivers"`
}

type WithdrawalProtocolFeeContent struct {
	PoolPairID string                       `json:"PoolPairID"`
	TokenID    common.Hash                  `json:"TokenID"`
	Receivers  map[common.Hash]ReceiverInfo `json:"Receivers"`
	TxReqID    common.Hash                  `json:"TxReqID"`
	ShardID    byte                         `json:"ShardID"`
}

type WithdrawalProtocolFeeStatus struct {
	Status    int                          `json:"Status"`
	Receivers map[common.Hash]ReceiverInfo `json:"Receivers"`
}

func NewPdexv3WithdrawalProtocolFeeRequest(
	metaType int,
	pairID string,
	receivers map[common.Hash]privacy.OTAReceiver,
) (*WithdrawalProtocolFeeRequest, error) {
	metadataBase := metadataCommon.NewMetadataBaseWithSignature(metaType)

	return &WithdrawalProtocolFeeRequest{
		MetadataBaseWithSignature: *metadataBase,
		PoolPairID:                pairID,
		Receivers:                 receivers,
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

	if len(withdrawal.Receivers) > MaxPoolPairWithdrawalReceiver {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.Pdexv3WithdrawProtocolFeeValidateSanityDataError, fmt.Errorf("Too many receivers"))
	}

	// Check OTA address string and tx random is valid
	shardID := byte(tx.GetValidationEnv().ShardID())
	for _, receiver := range withdrawal.Receivers {
		_, err = isValidOTAReceiver(receiver, shardID)
		if err != nil {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
		}
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
	rawBytes, _ := json.Marshal(withdrawal)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (withdrawal *WithdrawalProtocolFeeRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(withdrawal)
}

func (withdrawal *WithdrawalProtocolFeeRequest) GetOTADeclarations() []metadataCommon.OTADeclaration {
	result := []metadataCommon.OTADeclaration{}
	for currentTokenID, val := range withdrawal.Receivers {
		if currentTokenID != common.PRVCoinID {
			currentTokenID = common.ConfidentialAssetID
		}
		result = append(result, metadataCommon.OTADeclaration{
			PublicKey: val.PublicKey.ToBytes(), TokenID: currentTokenID,
		})
	}
	return result
}
