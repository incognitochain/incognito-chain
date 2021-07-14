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

type PDEContributionV3 struct {
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

func NewPDEContributionV3() *PDEContributionV3 {
	return &PDEContributionV3{}
}

func NewPDEContributionV3WithValue(
	poolPairID, pairHash,
	otaPublicKeyRefund, otaTxRandomRefund,
	otaPublicKeyReceive, otaTxRandomReceive,
	tokenID string, tokenAmount uint64, amplifier uint,
) *PDEContributionV3 {
	metadataBase := MetadataBase{
		Type: metadataCommon.PDexV3AddLiquidityMeta,
	}
	return &PDEContributionV3{
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

func (pc *PDEContributionV3) ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	// NOTE: verify supported tokens pair as needed
	return true, nil
}

func (pc *PDEContributionV3) ValidateSanityData(
	chainRetriever ChainRetriever,
	shardViewRetriever ShardViewRetriever,
	beaconViewRetriever BeaconViewRetriever,
	beaconHeight uint64,
	tx Transaction,
) (bool, bool, error) {
	if pc.PairHash == "" {
		return false, false, NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Pair hash should not be empty"))
	}
	if pc.OtaPublicKeyRefund == "" {
		return false, false, NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Ota public key refund should not be empty"))
	}
	if pc.OtaPublicKeyReceive == "" {
		return false, false, NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Ota public key receive should not be empty"))
	}
	if pc.OtaTxRandomRefund == "" {
		return false, false, NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Ota tx random refund should not be empty"))
	}
	if pc.OtaTxRandomReceive == "" {
		return false, false, NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Ota tx random receive should not be empty"))
	}
	if pc.TokenID == "" {
		return false, false, NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("TokenID should not be empty"))
	}

	isBurned, burnCoin, burnedTokenID, err := tx.GetTxBurnData()
	if err != nil || !isBurned {
		return false, false, NewMetadataTxError(metadataCommon.PDENotBurningTxError, err)
	}

	if pc.TokenAmount == 0 || pc.TokenAmount != burnCoin.GetValue() {
		return false, false, NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Contributed amount is not valid"))
	}

	//TODO: @tin add here amplifier can not smaller than 1.0
	if pc.Amplifier == 0 {
		return false, false, NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Amplifier is not valid"))
	}

	tokenHash, err := common.Hash{}.NewHashFromStr(pc.TokenID)
	if err != nil {
		return false, false, NewMetadataTxError(metadataCommon.PDECouldNotGenerateHashFromStringError, errors.New("TokenIDStr incorrect"))
	}
	if !bytes.Equal(burnedTokenID[:], tokenHash[:]) {
		return false, false, NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Wrong request info's token id, it should be equal to tx's token id"))
	}

	if tx.GetType() == common.TxNormalType && pc.TokenID != common.PRVCoinID.String() {
		return false, false, NewMetadataTxError(metadataCommon.PDEInvalidTxTypeError, errors.New("With tx normal privacy, the tokenIDStr should be PRV, not custom token"))
	}

	if tx.GetType() == common.TxCustomTokenPrivacyType && pc.TokenID == common.PRVCoinID.String() {
		return false, false, NewMetadataTxError(metadataCommon.PDEInvalidTxTypeError, errors.New("With tx custome token privacy, the tokenIDStr should not be PRV, but custom token"))
	}

	_, err, ver := checkTraderAddress(pc.OtaPublicKeyRefund, pc.OtaTxRandomRefund)
	if err != nil {
		return false, false, NewMetadataTxError(metadataCommon.PDEInvalidTxTypeError, err)
	}
	if int8(ver) != tx.GetVersion() {
		err := fmt.Errorf("payment address version (%v) and tx version (%v) mismatch", ver, tx.GetVersion())
		return false, false, NewMetadataTxError(metadataCommon.PDEInvalidTxTypeError, err)
	}
	_, err, ver = checkTraderAddress(pc.OtaPublicKeyReceive, pc.OtaTxRandomReceive)
	if err != nil {
		return false, false, NewMetadataTxError(metadataCommon.PDEInvalidTxTypeError, err)
	}
	if int8(ver) != tx.GetVersion() {
		err := fmt.Errorf("payment address version (%v) and tx version (%v) mismatch", ver, tx.GetVersion())
		return false, false, NewMetadataTxError(metadataCommon.PDEInvalidTxTypeError, err)
	}

	return true, true, nil
}

func (pc *PDEContributionV3) ValidateMetadataByItself() bool {
	return pc.Type == metadataCommon.PDexV3AddLiquidityMeta
}

func (pc *PDEContributionV3) Hash() *common.Hash {
	record := pc.MetadataBase.Hash().String()
	record += pc.PoolPairID
	record += pc.PairHash
	record += pc.OtaPublicKeyRefund
	record += pc.OtaTxRandomRefund
	record += pc.OtaPublicKeyReceive
	record += pc.OtaTxRandomReceive
	record += pc.TokenID
	record += strconv.FormatUint(uint64(pc.Amplifier), 10)
	record += strconv.FormatUint(pc.TokenAmount, 10)
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (pc *PDEContributionV3) BuildReqActions(
	tx Transaction,
	chainRetriever ChainRetriever,
	shardViewRetriever ShardViewRetriever,
	beaconViewRetriever BeaconViewRetriever,
	shardID byte,
	shardHeight uint64,
) ([][]string, error) {
	return [][]string{}, nil
}

func (pc *PDEContributionV3) CalculateSize() uint64 {
	return calculateSize(pc)
}
