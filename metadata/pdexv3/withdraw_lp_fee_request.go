package pdexv3

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

type WithdrawalLPFeeRequest struct {
	metadataCommon.MetadataBase
	PoolPairID string                              `json:"PoolPairID"`
	NftID      common.Hash                         `json:"NftID"`
	Receivers  map[common.Hash]privacy.OTAReceiver `json:"Receivers"`
}

type WithdrawalLPFeeContent struct {
	PoolPairID string       `json:"PoolPairID"`
	NftID      common.Hash  `json:"NftID"`
	TokenID    common.Hash  `json:"TokenID"`
	Receiver   ReceiverInfo `json:"Receiver"`
	IsLastInst bool         `json:"IsLastInst"`
	TxReqID    common.Hash  `json:"TxReqID"`
	ShardID    byte         `json:"ShardID"`
}

type WithdrawalLPFeeStatus struct {
	Status    int                          `json:"Status"`
	Receivers map[common.Hash]ReceiverInfo `json:"Receivers"`
}

func NewPdexv3WithdrawalLPFeeRequest(
	metaType int,
	pairID string,
	nftID common.Hash,
	receivers map[common.Hash]privacy.OTAReceiver,
) (*WithdrawalLPFeeRequest, error) {
	metadataBase := metadataCommon.NewMetadataBase(metaType)

	return &WithdrawalLPFeeRequest{
		MetadataBase: *metadataBase,
		PoolPairID:   pairID,
		NftID:        nftID,
		Receivers:    receivers,
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
	if !chainRetriever.IsAfterPdexv3CheckPoint(beaconViewRetriever.GetHeight()) {
		return false, fmt.Errorf("Feature pdexv3 has not been activated yet")
	}
	pdexv3StateCached := chainRetriever.GetPdexv3Cached(beaconViewRetriever.BlockHash())
	err := beaconViewRetriever.IsValidNftID(chainRetriever.GetBeaconChainDatabase(), pdexv3StateCached, withdrawal.NftID.String())
	if err != nil {
		return false, err
	}
	err = beaconViewRetriever.IsValidPoolPairID(chainRetriever.GetBeaconChainDatabase(), pdexv3StateCached, withdrawal.PoolPairID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (withdrawal WithdrawalLPFeeRequest) ValidateSanityData(
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	beaconHeight uint64,
	tx metadataCommon.Transaction,
) (bool, bool, error) {
	if !chainRetriever.IsAfterPdexv3CheckPoint(beaconHeight) {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, fmt.Errorf("Feature pdexv3 has not been activated yet"))
	}

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

	if len(withdrawal.Receivers) > MaxPoolPairWithdrawalReceiver {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.Pdexv3WithdrawLPFeeValidateSanityDataError, fmt.Errorf("Too many receivers"))
	}

	// Check OTA address string and tx random is valid
	shardID := byte(tx.GetValidationEnv().ShardID())
	for _, receiver := range withdrawal.Receivers {
		_, err = isValidOTAReceiver(receiver, shardID)
		if err != nil {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
		}
	}

	_, isExisted := withdrawal.Receivers[withdrawal.NftID]
	if !isExisted {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.Pdexv3WithdrawLPFeeValidateSanityDataError, fmt.Errorf("Nft Receiver is not existed"))
	}

	return true, true, nil
}

func (withdrawal WithdrawalLPFeeRequest) ValidateMetadataByItself() bool {
	return withdrawal.Type == metadataCommon.Pdexv3WithdrawLPFeeRequestMeta
}

func (withdrawal WithdrawalLPFeeRequest) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(withdrawal)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (withdrawal *WithdrawalLPFeeRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(withdrawal)
}

func (withdrawal *WithdrawalLPFeeRequest) GetOTADeclarations() []metadataCommon.OTADeclaration {
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
