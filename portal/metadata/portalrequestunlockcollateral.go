package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/basemeta"
	"github.com/incognitochain/incognito-chain/wallet"
	"strconv"
)

// PortalRequestUnlockCollateral - portal custodian requests unlock collateral (after returning pubToken to user)
// metadata - custodian requests unlock collateral - create normal tx with this metadata
type PortalRequestUnlockCollateral struct {
	basemeta.MetadataBase
	UniqueRedeemID      string
	TokenID             string // pTokenID in incognito chain
	CustodianAddressStr string
	RedeemAmount        uint64
	RedeemProof         string
}

// PortalRequestUnlockCollateralAction - shard validator creates instruction that contain this action content
type PortalRequestUnlockCollateralAction struct {
	Meta    PortalRequestUnlockCollateral
	TxReqID common.Hash
	ShardID byte
}

// PortalRequestUnlockCollateralContent - Beacon builds a new instruction with this content after receiving a instruction from shard
// It will be appended to beaconBlock
// both accepted and rejected status
type PortalRequestUnlockCollateralContent struct {
	UniqueRedeemID      string
	TokenID             string // pTokenID in incognito chain
	CustodianAddressStr string
	RedeemAmount        uint64
	UnlockAmount        uint64 // prv
	RedeemProof         string
	TxReqID             common.Hash
	ShardID             byte
}

// PortalRequestUnlockCollateralStatus - Beacon tracks status of request unlock collateral amount into db
type PortalRequestUnlockCollateralStatus struct {
	Status              byte
	UniqueRedeemID      string
	TokenID             string // pTokenID in incognito chain
	CustodianAddressStr string
	RedeemAmount        uint64
	UnlockAmount        uint64 // prv
	RedeemProof         string
	TxReqID             common.Hash
}

func NewPortalRequestUnlockCollateral(
	metaType int,
	uniqueRedeemID string,
	tokenID string,
	incogAddressStr string,
	redeemAmount uint64,
	redeemProof string) (*PortalRequestUnlockCollateral, error) {
	metadataBase := basemeta.MetadataBase{
		Type: metaType,
	}
	requestPTokenMeta := &PortalRequestUnlockCollateral{
		UniqueRedeemID:      uniqueRedeemID,
		TokenID:             tokenID,
		CustodianAddressStr: incogAddressStr,
		RedeemAmount:        redeemAmount,
		RedeemProof:         redeemProof,
	}
	requestPTokenMeta.MetadataBase = metadataBase
	return requestPTokenMeta, nil
}

func (meta PortalRequestUnlockCollateral) ValidateTxWithBlockChain(
	txr basemeta.Transaction,
	chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	return true, nil
}

func (meta PortalRequestUnlockCollateral) ValidateSanityData(chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, beaconHeight uint64, txr basemeta.Transaction) (bool, bool, error) {
	// validate CustodianAddressStr
	keyWallet, err := wallet.Base58CheckDeserialize(meta.CustodianAddressStr)
	if err != nil {
		return false, false, basemeta.NewMetadataTxError(basemeta.PortalRequestPTokenParamError, errors.New("Custodian incognito address is invalid"))
	}
	incogAddr := keyWallet.KeySet.PaymentAddress
	if len(incogAddr.Pk) == 0 {
		return false, false, basemeta.NewMetadataTxError(basemeta.PortalRequestPTokenParamError, errors.New("Custodian incognito address is invalid"))
	}
	if !bytes.Equal(txr.GetSigPubKey()[:], incogAddr.Pk[:]) {
		return false, false, basemeta.NewMetadataTxError(basemeta.PortalRequestPTokenParamError, errors.New("Custodian incognito address is not signer"))
	}

	// check tx type
	if txr.GetType() != common.TxNormalType {
		return false, false, errors.New("tx custodian deposit must be TxNormalType")
	}

	// validate amount redeem
	if meta.RedeemAmount == 0 {
		return false, false, errors.New("redeem amount should be larger than 0")
	}

	// validate tokenID
	if !chainRetriever.IsPortalToken(beaconHeight, meta.TokenID) {
		return false, false, errors.New("TokenID is not a portal token")
	}

	return true, true, nil
}

func (meta PortalRequestUnlockCollateral) ValidateMetadataByItself() bool {
	return meta.Type == basemeta.PortalRequestUnlockCollateralMeta || meta.Type == basemeta.PortalRequestUnlockCollateralMetaV3
}

func (meta PortalRequestUnlockCollateral) Hash() *common.Hash {
	record := meta.MetadataBase.Hash().String()
	record += meta.UniqueRedeemID
	record += meta.TokenID
	record += meta.CustodianAddressStr
	record += strconv.FormatUint(meta.RedeemAmount, 10)
	record += meta.RedeemProof
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (meta *PortalRequestUnlockCollateral) BuildReqActions(tx basemeta.Transaction, chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	actionContent := PortalRequestUnlockCollateralAction{
		Meta:    *meta,
		TxReqID: *tx.Hash(),
		ShardID: shardID,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(meta.Type), actionContentBase64Str}
	return [][]string{action}, nil
}

func (meta *PortalRequestUnlockCollateral) CalculateSize() uint64 {
	return basemeta.CalculateSize(meta)
}
