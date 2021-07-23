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
	poolPairID     string // only "" for the first contribution of pool
	pairHash       string
	receiveAddress string // receive nfct
	refundAddress  string // refund pToken
	tokenID        string
	tokenAmount    uint64
	amplifier      uint // only set for the first contribution
	metadataCommon.MetadataBase
}

func NewAddLiquidity() *AddLiquidity {
	return &AddLiquidity{}
}

func NewAddLiquidityWithValue(
	poolPairID, pairHash,
	receiveAddress, refundAddress,
	tokenID string, tokenAmount uint64, amplifier uint,
) *AddLiquidity {
	metadataBase := metadataCommon.MetadataBase{
		Type: metadataCommon.Pdexv3AddLiquidityMeta,
	}
	return &AddLiquidity{
		poolPairID:     poolPairID,
		pairHash:       pairHash,
		receiveAddress: receiveAddress,
		refundAddress:  refundAddress,
		tokenID:        tokenID,
		tokenAmount:    tokenAmount,
		amplifier:      amplifier,
		MetadataBase:   metadataBase,
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
	receiveAddress := privacy.OTAReceiver{}
	err = receiveAddress.FromString(al.receiveAddress)
	if err != nil {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
	}
	if !receiveAddress.IsValid() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("ReceiveAddress is not valid"))
	}
	refundAddress := privacy.OTAReceiver{}
	err = refundAddress.FromString(al.refundAddress)
	if err != nil {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
	}
	if !refundAddress.IsValid() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("RefundAddress is not valid"))
	}
	if al.amplifier < DefaultAmplifier {
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
	return al.Type == metadataCommon.Pdexv3AddLiquidityMeta
}

func (al *AddLiquidity) Hash() *common.Hash {
	record := al.MetadataBase.Hash().String()
	record += al.poolPairID
	record += al.pairHash
	record += al.receiveAddress
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
		PoolPairID     string `json:"PoolPairID"` // only "" for the first contribution of pool
		PairHash       string `json:"PairHash"`
		ReceiveAddress string `json:"ReceiveAddress"` // receive nfct
		RefundAddress  string `json:"RefundAddress"`  // refund pToken
		TokenID        string `json:"TokenID"`
		TokenAmount    uint64 `json:"TokenAmount"`
		Amplifier      uint   `json:"Amplifier"` // only set for the first contribution
		metadataCommon.MetadataBase
	}{
		PoolPairID:     al.poolPairID,
		PairHash:       al.pairHash,
		ReceiveAddress: al.receiveAddress,
		RefundAddress:  al.refundAddress,
		TokenID:        al.tokenID,
		TokenAmount:    al.tokenAmount,
		Amplifier:      al.amplifier,
		MetadataBase:   al.MetadataBase,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (al *AddLiquidity) UnmarshalJSON(data []byte) error {
	temp := struct {
		PoolPairID     string `json:"PoolPairID"` // only "" for the first contribution of pool
		PairHash       string `json:"PairHash"`
		ReceiveAddress string `json:"ReceiveAddress"` // receive nfct
		RefundAddress  string `json:"RefundAddress"`  // refund pToken
		TokenID        string `json:"TokenID"`
		TokenAmount    uint64 `json:"TokenAmount"`
		Amplifier      uint   `json:"Amplifier"` // only set for the first contribution
		metadataCommon.MetadataBase
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	al.poolPairID = temp.PoolPairID
	al.pairHash = temp.PairHash
	al.receiveAddress = temp.ReceiveAddress
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

func (al *AddLiquidity) ReceiveAddress() string {
	return al.receiveAddress
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

func (al *AddLiquidity) FromStringSlice(source []string) error {
	if len(source) != 8 {
		return fmt.Errorf("Receive length %v but expect %v", len(source), 8)
	}
	if source[0] != strconv.Itoa(metadataCommon.Pdexv3AddLiquidityMeta) {
		return fmt.Errorf("Receive metaType %v but expect %v", source[0], metadataCommon.Pdexv3AddLiquidityMeta)
	}
	al.MetadataBase = metadataCommon.MetadataBase{Type: metadataCommon.Pdexv3AddLiquidityMeta}
	al.poolPairID = source[1]
	if source[2] == "" {
		return errors.New("Pair hash is invalid")
	}
	al.pairHash = source[2]
	receiveAddress := privacy.OTAReceiver{}
	err := receiveAddress.FromString(source[3])
	if err != nil {
		return err
	}
	if !receiveAddress.IsValid() {
		return errors.New("receive Address is invalid")
	}
	al.receiveAddress = source[3]
	refundAddress := privacy.OTAReceiver{}
	err = refundAddress.FromString(source[4])
	if err != nil {
		return err
	}
	if !refundAddress.IsValid() {
		return errors.New("refund Address is invalid")
	}
	al.refundAddress = source[4]
	tokenID, err := common.Hash{}.NewHashFromStr(source[5])
	if err != nil {
		return err
	}
	if tokenID.IsZeroValue() {
		return errors.New("TokenID is empty")
	}
	al.tokenID = source[5]
	tokenAmount, err := strconv.ParseUint(source[6], 10, 32)
	if err != nil {
		return err
	}
	al.tokenAmount = tokenAmount
	amplifier, err := strconv.ParseUint(source[7], 10, 32)
	if err != nil {
		return err
	}
	if amplifier < DefaultAmplifier {
		return fmt.Errorf("Amplifier can not be smaller than %v get %v", DefaultAmplifier, amplifier)
	}
	al.amplifier = uint(amplifier)
	return nil
}

func (al *AddLiquidity) StringSlice() []string {
	res := []string{strconv.Itoa(al.Type)}
	res = append(res, al.poolPairID)
	res = append(res, al.pairHash)
	res = append(res, al.receiveAddress)
	res = append(res, al.refundAddress)
	res = append(res, al.tokenID)
	tokenAmount := strconv.FormatUint(al.tokenAmount, 10)
	res = append(res, tokenAmount)
	amplifier := strconv.FormatUint(uint64(al.amplifier), 10)
	res = append(res, amplifier)
	return res
}
