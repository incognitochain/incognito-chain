package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/basemeta"
	"github.com/incognitochain/incognito-chain/wallet"
	"reflect"
	"strconv"
)

type PortalRedeemFromLiquidationPoolV3 struct {
	basemeta.MetadataBase
	TokenID               string // portalTokenID in incognito chain
	RedeemAmount          uint64
	RedeemerIncAddressStr string
	RedeemerExtAddressStr string
}

type PortalRedeemFromLiquidationPoolActionV3 struct {
	Meta    PortalRedeemFromLiquidationPoolV3
	TxReqID common.Hash
	ShardID byte
}

type PortalRedeemFromLiquidationPoolContentV3 struct {
	TokenID               string // portalTokenID in incognito chain
	RedeemAmount          uint64
	RedeemerIncAddressStr string
	RedeemerExtAddressStr string
	TxReqID               common.Hash
	ShardID               byte
	MintedPRVCollateral   uint64
	UnlockedTokenCollaterals map[string]uint64
}

type PortalRedeemFromLiquidationPoolStatusV3 struct {
	TokenID               string // portalTokenID in incognito chain
	RedeemAmount          uint64
	RedeemerIncAddressStr string
	RedeemerExtAddressStr string
	TxReqID               common.Hash
	MintedPRVCollateral   uint64
	UnlockedTokenCollaterals map[string]uint64
	Status byte
}

func NewPortalRedeemFromLiquidationPoolV3(
	metaType int,
	tokenID string,
	redeemAmount uint64,
	incAddressStr string,
	extAddressStr string,
) (*PortalRedeemFromLiquidationPoolV3, error) {
	portalRedeemLiquidateExchangeRates := &PortalRedeemFromLiquidationPoolV3{
		MetadataBase:          basemeta.MetadataBase{Type: metaType},
		TokenID:               tokenID,
		RedeemAmount:          redeemAmount,
		RedeemerIncAddressStr: incAddressStr,
		RedeemerExtAddressStr: extAddressStr,
	}

	return portalRedeemLiquidateExchangeRates, nil
}

func (redeemReq PortalRedeemFromLiquidationPoolV3) ValidateTxWithBlockChain(
	txr basemeta.Transaction,
	chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	return true, nil
}

func (redeemReq PortalRedeemFromLiquidationPoolV3) ValidateSanityData(chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, beaconHeight uint64, txr basemeta.Transaction) (bool, bool, error) {
	if txr.GetType() == common.TxCustomTokenPrivacyType && reflect.TypeOf(txr).String() == "*transaction.Tx" {
		return true, true, nil
	}
	// validate RedeemerIncAddressStr
	keyWallet, err := wallet.Base58CheckDeserialize(redeemReq.RedeemerIncAddressStr)
	if err != nil {
		return false, false, basemeta.NewMetadataTxError(basemeta.PortalRedeemLiquidateExchangeRatesParamError, errors.New("Address incognito redeem is invalid"))
	}

	incAddr := keyWallet.KeySet.PaymentAddress
	if len(incAddr.Pk) == 0 {
		return false, false, basemeta.NewMetadataTxError(basemeta.PortalRedeemLiquidateExchangeRatesParamError, errors.New("Payment incognito address is invalid"))
	}
	if !bytes.Equal(txr.GetSigPubKey()[:], incAddr.Pk[:]) {
		return false, false, basemeta.NewMetadataTxError(basemeta.PortalRedeemLiquidateExchangeRatesParamError, errors.New("Address incognito redeem is not signer"))
	}

	// check tx type
	if txr.GetType() != common.TxCustomTokenPrivacyType {
		return false, false, errors.New("tx redeem request must be TxCustomTokenPrivacyType")
	}

	if !txr.IsCoinsBurning(chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight) {
		return false, false, errors.New("txprivacytoken in tx redeem request must be coin burning tx")
	}

	// validate redeem amount
	minAmount := common.MinAmountPortalPToken[redeemReq.TokenID]
	if redeemReq.RedeemAmount < minAmount {
		return false, false, fmt.Errorf("redeem amount should be larger or equal to %v", minAmount)
	}

	// validate value transfer of tx for redeem amount in ptoken
	if redeemReq.RedeemAmount != txr.CalculateTxValue() {
		return false, false, errors.New("redeem amount should be equal to the tx value")
	}

	// validate tokenID
	if redeemReq.TokenID != txr.GetTokenID().String() {
		return false, false, basemeta.NewMetadataTxError(basemeta.PortalRedeemLiquidateExchangeRatesParamError, errors.New("TokenID in metadata is not matched to tokenID in tx"))
	}
	// check tokenId is portal token or not
	if !IsPortalToken(redeemReq.TokenID) {
		return false, false, basemeta.NewMetadataTxError(basemeta.PortalRedeemLiquidateExchangeRatesParamError, errors.New("TokenID is not in portal tokens list"))
	}

	// checkout ext address
	if common.Has0xPrefix(redeemReq.RedeemerExtAddressStr) {
		return false, false, errors.New("Redeem from liquidation v3: RedeemerExtAddressStr shouldn't have 0x prefix")
	}
	if isValid, err := ValidatePortalExternalAddress(common.ETHChainName, common.Remove0xPrefix(common.EthAddrStr), redeemReq.RedeemerExtAddressStr); !isValid || err != nil {
		return false, false, errors.New("Redeem from liquidation v3: RedeemerExtAddressStr is invalid")
	}
	return true, true, nil
}

func (redeemReq PortalRedeemFromLiquidationPoolV3) ValidateMetadataByItself() bool {
	return redeemReq.Type == basemeta.PortalRedeemFromLiquidationPoolMetaV3
}

func (redeemReq PortalRedeemFromLiquidationPoolV3) Hash() *common.Hash {
	record := redeemReq.MetadataBase.Hash().String()
	record += redeemReq.TokenID
	record += strconv.FormatUint(redeemReq.RedeemAmount, 10)
	record += redeemReq.RedeemerIncAddressStr
	record += redeemReq.RedeemerExtAddressStr
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (redeemReq *PortalRedeemFromLiquidationPoolV3) BuildReqActions(tx basemeta.Transaction, chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	actionContent := PortalRedeemFromLiquidationPoolActionV3{
		Meta:    *redeemReq,
		TxReqID: *tx.Hash(),
		ShardID: shardID,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(basemeta.PortalRedeemFromLiquidationPoolMetaV3), actionContentBase64Str}
	return [][]string{action}, nil
}

func (redeemReq *PortalRedeemFromLiquidationPoolV3) CalculateSize() uint64 {
	return basemeta.CalculateSize(redeemReq)
}
