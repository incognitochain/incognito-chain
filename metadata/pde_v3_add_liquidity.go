package metadata

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type PDEV3AddLiquidity struct {
	PoolPairID          string // only "" for the first contribution of pool
	PairHash            string
	OtaPublicKeyRefund  string // refund contributed token
	OtaTxRandomRefund   string
	OtaPublicKeyReceive string // receive nfct
	OtaTxRandomReceive  string
	TokenID             string
	TokenAmount         uint64
	Amplifier           uint // only set for the first contribution
	MetadataBase
}

func NewPDEV3AddLiquidity() *PDEV3AddLiquidity {
	return &PDEV3AddLiquidity{}
}

func NewPDEV3AddLiquidityWithValue(
	poolPairID, pairHash,
	otaPublicKeyRefund, otaTxRandomRefund,
	otaPublicKeyReceive, otaTxRandomReceive,
	tokenID string, tokenAmount uint64, amplifier uint,
) *PDEV3AddLiquidity {
	metadataBase := MetadataBase{
		Type: metadataCommon.PDexV3AddLiquidityMeta,
	}
	return &PDEV3AddLiquidity{
		PoolPairID:          poolPairID,
		PairHash:            pairHash,
		OtaPublicKeyRefund:  otaPublicKeyRefund,
		OtaTxRandomRefund:   otaTxRandomRefund,
		OtaPublicKeyReceive: otaPublicKeyReceive,
		OtaTxRandomReceive:  otaTxRandomReceive,
		TokenID:             tokenID,
		TokenAmount:         tokenAmount,
		Amplifier:           amplifier,
		MetadataBase:        metadataBase,
	}
}

func (pl *PDEV3AddLiquidity) ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	// NOTE: verify supported tokens pair as needed
	return true, nil
}

func (pl *PDEV3AddLiquidity) ValidateSanityData(
	chainRetriever ChainRetriever,
	shardViewRetriever ShardViewRetriever,
	beaconViewRetriever BeaconViewRetriever,
	beaconHeight uint64,
	tx Transaction,
) (bool, bool, error) {
	if pl.PairHash == "" {
		return false, false, NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Pair hash should not be empty"))
	}
	if pl.OtaPublicKeyRefund == "" {
		return false, false, NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Ota public key refund should not be empty"))
	}
	if pl.OtaPublicKeyReceive == "" {
		return false, false, NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Ota public key receive should not be empty"))
	}
	if pl.OtaTxRandomRefund == "" {
		return false, false, NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Ota tx random refund should not be empty"))
	}
	if pl.OtaTxRandomReceive == "" {
		return false, false, NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Ota tx random receive should not be empty"))
	}
	if pl.TokenID == "" {
		return false, false, NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("TokenID should not be empty"))
	}

	isBurned, burnCoin, burnedTokenID, err := tx.GetTxBurnData()
	if err != nil || !isBurned {
		return false, false, NewMetadataTxError(metadataCommon.PDENotBurningTxError, err)
	}

	/*fmt.Println("pl.TokenAmount:", pl.TokenAmount)*/
	/*fmt.Println("burnCoin.GetValue():", burnCoin.GetValue())*/

	if pl.TokenAmount == 0 || pl.TokenAmount != burnCoin.GetValue() {
		return false, false, NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Contributed amount is not valid"))
	}

	//TODO: @tin add here amplifier can not smaller than 1.0
	if pl.Amplifier == 0 {
		return false, false, NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Amplifier is not valid"))
	}

	tokenHash, err := common.Hash{}.NewHashFromStr(pl.TokenID)
	if err != nil {
		return false, false, NewMetadataTxError(metadataCommon.PDECouldNotGenerateHashFromStringError, errors.New("TokenIDStr incorrect"))
	}
	if !bytes.Equal(burnedTokenID[:], tokenHash[:]) {
		return false, false, NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Wrong request info's token id, it should be equal to tx's token id"))
	}

	if tx.GetType() == common.TxNormalType && pl.TokenID != common.PRVCoinID.String() {
		return false, false, NewMetadataTxError(metadataCommon.PDEInvalidTxTypeError, errors.New("With tx normal privacy, the tokenIDStr should be PRV, not custom token"))
	}

	if tx.GetType() == common.TxCustomTokenPrivacyType && pl.TokenID == common.PRVCoinID.String() {
		return false, false, NewMetadataTxError(metadataCommon.PDEInvalidTxTypeError, errors.New("With tx custome token privacy, the tokenIDStr should not be PRV, but custom token"))
	}

	_, err, ver := checkTraderAddress(pl.OtaPublicKeyRefund, pl.OtaTxRandomRefund)
	if err != nil {
		return false, false, NewMetadataTxError(metadataCommon.PDEInvalidTxTypeError, err)
	}
	if int8(ver) != tx.GetVersion() {
		err := fmt.Errorf("payment address version (%v) and tx version (%v) mismatch", ver, tx.GetVersion())
		return false, false, NewMetadataTxError(metadataCommon.PDEInvalidTxTypeError, err)
	}
	_, err, ver = checkTraderAddress(pl.OtaPublicKeyReceive, pl.OtaTxRandomReceive)
	if err != nil {
		return false, false, NewMetadataTxError(metadataCommon.PDEInvalidTxTypeError, err)
	}
	if int8(ver) != tx.GetVersion() {
		err := fmt.Errorf("payment address version (%v) and tx version (%v) mismatch", ver, tx.GetVersion())
		return false, false, NewMetadataTxError(metadataCommon.PDEInvalidTxTypeError, err)
	}

	return true, true, nil
}

func (pl *PDEV3AddLiquidity) ValidateMetadataByItself() bool {
	return pl.Type == metadataCommon.PDexV3AddLiquidityMeta
}

func (pl *PDEV3AddLiquidity) Hash() *common.Hash {
	record := pl.MetadataBase.Hash().String()
	record += pl.PoolPairID
	record += pl.PairHash
	record += pl.OtaPublicKeyRefund
	record += pl.OtaTxRandomRefund
	record += pl.OtaPublicKeyReceive
	record += pl.OtaTxRandomReceive
	record += pl.TokenID
	record += strconv.FormatUint(uint64(pl.Amplifier), 10)
	record += strconv.FormatUint(pl.TokenAmount, 10)
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (pl *PDEV3AddLiquidity) BuildReqActions(
	tx Transaction,
	chainRetriever ChainRetriever,
	shardViewRetriever ShardViewRetriever,
	beaconViewRetriever BeaconViewRetriever,
	shardID byte,
	shardHeight uint64,
) ([][]string, error) {
	return [][]string{}, nil
}

func (pl *PDEV3AddLiquidity) CalculateSize() uint64 {
	return calculateSize(pl)
}
