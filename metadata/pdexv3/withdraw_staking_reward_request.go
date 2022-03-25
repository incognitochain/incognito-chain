package pdexv3

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/utils"
)

type WithdrawalStakingRewardRequest struct {
	metadataCommon.MetadataBase
	StakingPoolID string `json:"StakingPoolID"`
	AccessOption
	Receivers map[common.Hash]privacy.OTAReceiver `json:"Receivers"`
}

type WithdrawalStakingRewardContent struct {
	StakingPoolID string `json:"StakingPoolID"`
	AccessOption
	TokenID    common.Hash  `json:"TokenID"`
	Receiver   ReceiverInfo `json:"Receiver"`
	IsLastInst bool         `json:"IsLastInst"`
	TxReqID    common.Hash  `json:"TxReqID"`
	ShardID    byte         `json:"ShardID"`
	AccessOTA  []byte       `json:"AccessOTA,omitempty"`
}

type WithdrawalStakingRewardStatus struct {
	Status    int                          `json:"Status"`
	Receivers map[common.Hash]ReceiverInfo `json:"Receivers"`
}

func NewPdexv3WithdrawalStakingRewardRequest(
	metaType int,
	stakingToken string,
	accessOption AccessOption,
	receivers map[common.Hash]privacy.OTAReceiver,
) (*WithdrawalStakingRewardRequest, error) {
	metadataBase := metadataCommon.NewMetadataBase(metaType)

	return &WithdrawalStakingRewardRequest{
		MetadataBase:  *metadataBase,
		StakingPoolID: stakingToken,
		AccessOption:  accessOption,
		Receivers:     receivers,
	}, nil
}

func (withdrawal WithdrawalStakingRewardRequest) ValidateTxWithBlockChain(
	tx metadataCommon.Transaction,
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	err := withdrawal.AccessOption.IsValid(tx, withdrawal.Receivers, beaconViewRetriever, db, true, false, "")
	if err != nil {
		return false, err
	}
	// validate burn tx, tokenID & amount = 1
	isBurn, _, burnedCoin, burnedToken, err := tx.GetTxFullBurnData()
	if err != nil || !isBurn {
		return false, metadataCommon.NewMetadataTxError(metadataCommon.Pdexv3WithdrawStakingRewardValidateSanityDataError, fmt.Errorf("Tx is not a burn tx. Error %v", err))
	}
	burningAmt := burnedCoin.GetValue()
	expectBurntTokenID := common.Hash{}
	if withdrawal.AccessOption.UseNft() {
		expectBurntTokenID = *withdrawal.NftID
		_, isExisted := withdrawal.Receivers[*withdrawal.NftID]
		if !isExisted {
			return false, metadataCommon.NewMetadataTxError(metadataCommon.Pdexv3WithdrawLPFeeValidateSanityDataError, fmt.Errorf("Nft Receiver is not existed"))
		}
	} else {
		expectBurntTokenID = common.PdexAccessCoinID
	}
	burningTokenID := burnedToken
	if burningAmt != 1 || *burningTokenID != expectBurntTokenID {
		return false, metadataCommon.NewMetadataTxError(metadataCommon.Pdexv3WithdrawStakingRewardValidateSanityDataError, fmt.Errorf("Burning token ID or amount is wrong. Error %v", err))
	}

	if len(withdrawal.Receivers) > MaxStakingRewardWithdrawalReceiver {
		return false, metadataCommon.NewMetadataTxError(metadataCommon.Pdexv3WithdrawStakingRewardValidateSanityDataError, fmt.Errorf("Too many receivers"))
	}

	// Check OTA address string and tx random is valid
	for _, receiver := range withdrawal.Receivers {
		_, err = isValidOTAReceiver(receiver, shardID)
		if err != nil {
			return false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
		}
	}
	ok, err := beaconViewRetriever.IsValidPdexv3StakingPool(withdrawal.StakingPoolID)
	if err != nil || !ok {
		if !ok {
			err = fmt.Errorf("Cannot find stakingPoolID %s", withdrawal.StakingPoolID)
		}
		return false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
	}
	if !withdrawal.UseNft() {
		return beaconViewRetriever.IsValidAccessOTAWithPdexState(
			*metadataCommon.NewPdexv3ExtendAccessIDWithValue(
				withdrawal.StakingPoolID,
				*withdrawal.AccessID,
				withdrawal.BurntOTA.ToBytesS(),
				metadataCommon.Pdexv3WithdrawStakingRewardRequestMeta,
				utils.EmptyString,
			),
		)
	}
	return true, nil
}

func (withdrawal WithdrawalStakingRewardRequest) ValidateSanityData(
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
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.Pdexv3WithdrawStakingRewardValidateSanityDataError, errors.New("Tx pDex v3 staking reward withdrawal must be TxCustomTokenPrivacyType"))
	}

	if tx.GetVersion() != 2 {
		return false, false, metadataCommon.NewMetadataTxError(0, errors.New("Tx pDex v3 staking reward withdrawal must be version 2"))
	}
	return true, true, nil
}

func (withdrawal WithdrawalStakingRewardRequest) ValidateMetadataByItself() bool {
	return withdrawal.Type == metadataCommon.Pdexv3WithdrawStakingRewardRequestMeta
}

func (withdrawal WithdrawalStakingRewardRequest) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(withdrawal)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (withdrawal *WithdrawalStakingRewardRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(withdrawal)
}

func (withdrawal *WithdrawalStakingRewardRequest) GetOTADeclarations() []metadataCommon.OTADeclaration {
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
