package pdexv3

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

type AddLiquidity struct {
	PoolPairID      string              `json:"PoolPairID"` // only "" for the first contribution of pool
	PairHash        string              `json:"PairHash"`
	ReceiverAddress privacy.OTAReceiver `json:"ReceiverAddress"` // receive nfct
	RefundAddress   privacy.OTAReceiver `json:"RefundAddress"`   // refund pToken
	TokenID         common.Hash         `json:"TokenID"`
	TokenAmount     uint64              `json:"TokenAmount"`
	Amplifier       uint                `json:"Amplifier"` // only set for the first contribution
	metadataCommon.MetadataBase
}

func NewAddLiquidity() *AddLiquidity {
	return &AddLiquidity{}
}

func NewAddLiquidityWithValue(
	poolPairID, pairHash string,
	receiverAddress, refundAddress privacy.OTAReceiver,
	tokenID common.Hash, tokenAmount uint64, amplifier uint,
) *AddLiquidity {
	metadataBase := metadataCommon.MetadataBase{
		Type: metadataCommon.PDexV3AddLiquidityMeta,
	}
	return &AddLiquidity{
		PoolPairID:      poolPairID,
		PairHash:        pairHash,
		ReceiverAddress: receiverAddress,
		RefundAddress:   refundAddress,
		TokenID:         tokenID,
		TokenAmount:     tokenAmount,
		Amplifier:       amplifier,
		MetadataBase:    metadataBase,
	}
}

func (al *AddLiquidity) ValidateTxWithBlockChain(
	tx metadataCommon.Transaction,
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (bool, error) {
	// NOTE: verify supported tokens pair as needed
	return true, nil
}

func (al *AddLiquidity) ValidateSanityData(
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	beaconHeight uint64,
	tx metadataCommon.Transaction,
) (bool, bool, error) {
	if al.PairHash == "" {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Pair hash should not be empty"))
	}
	if al.TokenID.IsZeroValue() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("TokenID should not be empty"))
	}
	if !al.ReceiverAddress.IsValid() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("ReceiverAddress is not valid"))
	}
	if !al.RefundAddress.IsValid() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("RefundAddress is not valid"))
	}
	//TODO: @tin add here amplifier can not smaller than 1.0
	if al.Amplifier == 0 {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Amplifier is not valid"))
	}

	isBurned, burnCoin, burnedTokenID, err := tx.GetTxBurnData()
	if err != nil || !isBurned {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDENotBurningTxError, err)
	}

	if !bytes.Equal(burnedTokenID[:], al.TokenID[:]) {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Wrong request info's token id, it should be equal to tx's token id"))
	}
	if al.TokenAmount == 0 || al.TokenAmount != burnCoin.GetValue() {
		err := fmt.Errorf("Contributed amount is not valid expect %v but get %v", al.TokenAmount, burnCoin.GetValue())
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
	}
	if tx.GetType() == common.TxNormalType && al.TokenID.String() != common.PRVCoinID.String() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidTxTypeError, errors.New("With tx normal privacy, the tokenIDStr should be PRV, not custom token"))
	}
	if tx.GetType() == common.TxCustomTokenPrivacyType && al.TokenID.String() == common.PRVCoinID.String() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidTxTypeError, errors.New("With tx custome token privacy, the tokenIDStr should not be PRV, but custom token"))
	}

	return true, true, nil
}

func (al *AddLiquidity) ValidateMetadataByItself() bool {
	return al.Type == metadataCommon.PDexV3AddLiquidityMeta
}

func (al *AddLiquidity) Hash() *common.Hash {
	record := al.MetadataBase.Hash().String()
	record += al.PoolPairID
	record += al.PairHash
	str, _ := al.ReceiverAddress.String()
	record += str
	str, _ = al.RefundAddress.String()
	record += str
	record += al.TokenID.String()
	record += strconv.FormatUint(uint64(al.Amplifier), 10)
	record += strconv.FormatUint(al.TokenAmount, 10)
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (al *AddLiquidity) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(al)
}
