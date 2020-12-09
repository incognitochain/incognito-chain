package metadata

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/basemeta"
	"strconv"
)

// PortalRedeemRequestV3 - portal user redeem requests to get public token by burning ptoken
// metadata - redeem request - create normal tx with this metadata
type PortalLiquidateCustodian struct {
	basemeta.MetadataBase
	UniqueRedeemID           string
	TokenID                  string // pTokenID in incognito chain
	RedeemPubTokenAmount     uint64
	MintedCollateralAmount   uint64 // minted PRV amount for sending back to users
	RedeemerIncAddressStr    string
	CustodianIncAddressStr   string
	LiquidatedByExchangeRate bool
}

// PortalLiquidateCustodianContent - Beacon builds a new instruction with this content after detecting custodians run away
// It will be appended to beaconBlock
type PortalLiquidateCustodianContent struct {
	basemeta.MetadataBase
	UniqueRedeemID                 string
	TokenID                        string // pTokenID in incognito chain
	RedeemPubTokenAmount           uint64
	LiquidatedCollateralAmount     uint64 // minted PRV amount for sending back to users
	RemainUnlockAmountForCustodian uint64
	RedeemerIncAddressStr          string
	CustodianIncAddressStr         string
	LiquidatedByExchangeRate       bool
	ShardID                        byte

	RemainUnlockAmountsForCustodian map[string]uint64
	LiquidatedCollateralAmounts     map[string]uint64 // minted PRV amount for sending back to users
}

// PortalLiquidateCustodianStatus - Beacon tracks status of custodian liquidation into db
type PortalLiquidateCustodianStatus struct {
	Status                         byte
	UniqueRedeemID                 string
	TokenID                        string // pTokenID in incognito chain
	RedeemPubTokenAmount           uint64
	LiquidatedCollateralAmount     uint64 // minted PRV amount for sending back to users
	RemainUnlockAmountForCustodian uint64
	RedeemerIncAddressStr          string
	CustodianIncAddressStr         string
	LiquidatedByExchangeRate       bool
	ShardID                        byte
	LiquidatedBeaconHeight         uint64

	RemainUnlockAmountsForCustodian map[string]uint64
	LiquidatedCollateralAmounts     map[string]uint64 // minted PRV amount for sending back to users
}

func NewPortalLiquidateCustodian(
	metaType int,
	uniqueRedeemID string,
	tokenID string,
	redeemAmount uint64,
	mintedCollateralAmount uint64,
	redeemerIncAddressStr string,
	custodianIncAddressStr string,
	liquidatedByExchangeRate bool) (*PortalLiquidateCustodian, error) {
	metadataBase := basemeta.MetadataBase{
		Type: metaType,
	}
	liquidCustodianMeta := &PortalLiquidateCustodian{
		UniqueRedeemID:           uniqueRedeemID,
		TokenID:                  tokenID,
		RedeemPubTokenAmount:     redeemAmount,
		MintedCollateralAmount:   mintedCollateralAmount,
		RedeemerIncAddressStr:    redeemerIncAddressStr,
		CustodianIncAddressStr:   custodianIncAddressStr,
		LiquidatedByExchangeRate: liquidatedByExchangeRate,
	}
	liquidCustodianMeta.MetadataBase = metadataBase
	return liquidCustodianMeta, nil
}

func (liqCustodian PortalLiquidateCustodian) ValidateTxWithBlockChain(
	txr basemeta.Transaction,
	chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	return true, nil
}

func (liqCustodian PortalLiquidateCustodian) ValidateSanityData(chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, beaconHeight uint64, txr basemeta.Transaction) (bool, bool, error) {
	return true, true, nil
}

func (liqCustodian PortalLiquidateCustodian) ValidateMetadataByItself() bool {
	return liqCustodian.Type == basemeta.PortalLiquidateCustodianMeta || liqCustodian.Type == basemeta.PortalLiquidateCustodianMetaV3
}

func (liqCustodian PortalLiquidateCustodian) Hash() *common.Hash {
	record := liqCustodian.MetadataBase.Hash().String()
	record += liqCustodian.UniqueRedeemID
	record += liqCustodian.TokenID
	record += strconv.FormatUint(liqCustodian.RedeemPubTokenAmount, 10)
	record += strconv.FormatUint(liqCustodian.MintedCollateralAmount, 10)
	record += liqCustodian.RedeemerIncAddressStr
	record += liqCustodian.CustodianIncAddressStr
	record += strconv.FormatBool(liqCustodian.LiquidatedByExchangeRate)
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (liqCustodian *PortalLiquidateCustodian) CalculateSize() uint64 {
	return basemeta.CalculateSize(liqCustodian)
}
