package pdexv3

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

type AddLiquidity struct {
	poolPairID      string // only "" for the first contribution of pool
	pairHash        string
	receiverAddress string // receive nfct
	refundAddress   string // refund pToken
	tokenID         string
	tokenAmount     uint64
	amplifier       uint // only set for the first contribution
	metadataCommon.MetadataBase
}

func NewAddLiquidity() *AddLiquidity {
	return &AddLiquidity{}
}

func NewAddLiquidityWithValue(
	poolPairID, pairHash,
	receiverAddress, refundAddress,
	tokenID string, tokenAmount uint64, amplifier uint,
) *AddLiquidity {
	metadataBase := metadataCommon.MetadataBase{
		Type: metadataCommon.PDexV3AddLiquidityMeta,
	}
	return &AddLiquidity{
		poolPairID:      poolPairID,
		pairHash:        pairHash,
		receiverAddress: receiverAddress,
		refundAddress:   refundAddress,
		tokenID:         tokenID,
		tokenAmount:     tokenAmount,
		amplifier:       amplifier,
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
	if al.pairHash == "" {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Pair hash should not be empty"))
	}
	tokenID, err := common.Hash{}.NewHashFromStr(al.tokenID)
	if err != nil {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
	}
	if tokenID.IsZeroValue() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("TokenID should not be empty"))
	}
	receiverAddress := privacy.OTAReceiver{}
	err = receiverAddress.FromString(al.receiverAddress)
	if err != nil {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
	}
	if !receiverAddress.IsValid() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("ReceiverAddress is not valid"))
	}
	refundAddress := privacy.OTAReceiver{}
	err = refundAddress.FromString(al.refundAddress)
	if err != nil {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
	}
	if !refundAddress.IsValid() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("RefundAddress is not valid"))
	}
	//TODO: @tin add here amplifier can not smaller than 1.0
	if al.amplifier == 0 {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Amplifier is not valid"))
	}

	isBurned, burnCoin, burnedTokenID, err := tx.GetTxBurnData()
	if err != nil || !isBurned {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDENotBurningTxError, err)
	}
	if !bytes.Equal(burnedTokenID[:], tokenID[:]) {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Wrong request info's token id, it should be equal to tx's token id"))
	}
	if al.tokenAmount == 0 || al.tokenAmount != burnCoin.GetValue() {
		err := fmt.Errorf("Contributed amount is not valid expect %v but get %v", al.tokenAmount, burnCoin.GetValue())
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
	}
	if tx.GetType() == common.TxNormalType && tokenID.String() != common.PRVCoinID.String() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidTxTypeError, errors.New("With tx normal privacy, the tokenIDStr should be PRV, not custom token"))
	}
	if tx.GetType() == common.TxCustomTokenPrivacyType && tokenID.String() == common.PRVCoinID.String() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidTxTypeError, errors.New("With tx custome token privacy, the tokenIDStr should not be PRV, but custom token"))
	}
	return true, true, nil
}

func (al *AddLiquidity) ValidateMetadataByItself() bool {
	return al.Type == metadataCommon.PDexV3AddLiquidityMeta
}

func (al *AddLiquidity) Hash() *common.Hash {
	record := al.MetadataBase.Hash().String()
	record += al.poolPairID
	record += al.pairHash
	record += al.receiverAddress
	record += al.refundAddress
	record += al.tokenID
	record += strconv.FormatUint(uint64(al.amplifier), 10)
	record += strconv.FormatUint(al.tokenAmount, 10)
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (al *AddLiquidity) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(al)
}

func (al *AddLiquidity) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		PoolPairID      string `json:"PoolPairID"` // only "" for the first contribution of pool
		PairHash        string `json:"PairHash"`
		ReceiverAddress string `json:"ReceiverAddress"` // receive nfct
		RefundAddress   string `json:"RefundAddress"`   // refund pToken
		TokenID         string `json:"TokenID"`
		TokenAmount     uint64 `json:"TokenAmount"`
		Amplifier       uint   `json:"Amplifier"` // only set for the first contribution
		metadataCommon.MetadataBase
	}{
		PoolPairID:      al.poolPairID,
		PairHash:        al.pairHash,
		ReceiverAddress: al.receiverAddress,
		RefundAddress:   al.refundAddress,
		TokenID:         al.tokenID,
		TokenAmount:     al.tokenAmount,
		Amplifier:       al.amplifier,
		MetadataBase:    al.MetadataBase,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (al *AddLiquidity) UnmarshalJSON(data []byte) error {
	temp := struct {
		PoolPairID      string `json:"PoolPairID"` // only "" for the first contribution of pool
		PairHash        string `json:"PairHash"`
		ReceiverAddress string `json:"ReceiverAddress"` // receive nfct
		RefundAddress   string `json:"RefundAddress"`   // refund pToken
		TokenID         string `json:"TokenID"`
		TokenAmount     uint64 `json:"TokenAmount"`
		Amplifier       uint   `json:"Amplifier"` // only set for the first contribution
		metadataCommon.MetadataBase
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	al.poolPairID = temp.PoolPairID
	al.pairHash = temp.PairHash
	al.receiverAddress = temp.ReceiverAddress
	al.refundAddress = temp.RefundAddress
	al.tokenID = temp.TokenID
	al.tokenAmount = temp.TokenAmount
	al.amplifier = temp.Amplifier
	al.MetadataBase = temp.MetadataBase
	return nil
}

func (al *AddLiquidity) PoolPairID() string {
	return al.poolPairID
}

func (al *AddLiquidity) PairHash() string {
	return al.pairHash
}

func (al *AddLiquidity) ReceiverAddress() string {
	return al.receiverAddress
}

func (al *AddLiquidity) RefundAddress() string {
	return al.refundAddress
}

func (al *AddLiquidity) TokenID() string {
	return al.tokenID
}

func (al *AddLiquidity) TokenAmount() uint64 {
	return al.tokenAmount
}

func (al *AddLiquidity) Amplifier() uint {
	return al.amplifier
}
