package metadata

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/wallet"
)

type PortalRedeemLiquidateExchangeRates struct {
	MetadataBase
	TokenID               string // pTokenID in incognito chain
	RedeemAmount          uint64
	RedeemerIncAddressStr string
}

type PortalRedeemLiquidateExchangeRatesAction struct {
	Meta    PortalRedeemLiquidateExchangeRates
	TxReqID common.Hash
	ShardID byte
}

type PortalRedeemLiquidateExchangeRatesContent struct {
	TokenID               string // pTokenID in incognito chain
	RedeemAmount          uint64
	RedeemerIncAddressStr string
	TxReqID               common.Hash
	ShardID               byte
	TotalPTokenReceived   uint64
}

type RedeemLiquidateExchangeRatesStatus struct {
	TxReqID             common.Hash
	TokenID             string
	RedeemerAddress     string
	RedeemAmount        uint64
	Status              byte
	TotalPTokenReceived uint64
}

func NewRedeemLiquidateExchangeRatesStatus(txReqID common.Hash, tokenID string, redeemerAddress string, redeemAmount uint64, status byte, totalPTokenReceived uint64) *RedeemLiquidateExchangeRatesStatus {
	return &RedeemLiquidateExchangeRatesStatus{TxReqID: txReqID, TokenID: tokenID, RedeemerAddress: redeemerAddress, RedeemAmount: redeemAmount, Status: status, TotalPTokenReceived: totalPTokenReceived}
}

func NewPortalRedeemLiquidateExchangeRates(
	metaType int,
	tokenID string,
	redeemAmount uint64,
	incAddressStr string,
) (*PortalRedeemLiquidateExchangeRates, error) {
	metadataBase := MetadataBase{Type: metaType}

	portalRedeemLiquidateExchangeRates := &PortalRedeemLiquidateExchangeRates{
		TokenID:               tokenID,
		RedeemAmount:          redeemAmount,
		RedeemerIncAddressStr: incAddressStr,
	}

	portalRedeemLiquidateExchangeRates.MetadataBase = metadataBase

	return portalRedeemLiquidateExchangeRates, nil
}

func (redeemReq PortalRedeemLiquidateExchangeRates) ValidateTxWithBlockChain(
	txr Transaction,
	chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	return true, nil
}

func (redeemReq PortalRedeemLiquidateExchangeRates) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, txr Transaction) (bool, bool, error) {
	// if txr.GetType() == common.TxCustomTokenPrivacyType && reflect.TypeOf(txr).String() == "*transaction.Tx" {
	// 	return true, true, nil
	// }
	// validate RedeemerIncAddressStr
	keyWallet, err := wallet.Base58CheckDeserialize(redeemReq.RedeemerIncAddressStr)
	if err != nil {
		return false, false, NewMetadataTxError(PortalRedeemLiquidateExchangeRatesParamError, errors.New("Address incognito redeem is invalid"))
	}
	if len(keyWallet.KeySet.PaymentAddress.Pk) == 0 {
		return false, false, NewMetadataTxError(PortalRedeemLiquidateExchangeRatesParamError, errors.New("Payment incognito address is invalid"))
	}
	// check burning tx
	isBurned, burnCoin, burnedTokenID, err := txr.GetTxBurnData()
	if err != nil || !isBurned {
		return false, false, errors.New("Error This is not Tx Burn")
	}
	// check tx type
	if txr.GetType() != common.TxCustomTokenPrivacyType {
		return false, false, errors.New("tx redeem request must be TxCustomTokenPrivacyType")
	}
	// validate redeem amount
	minAmount, err := chainRetriever.GetMinAmountPortalToken(redeemReq.TokenID, beaconHeight, common.PortalVersion3)
	if err != nil {
		return false, false, err
	}
	if redeemReq.RedeemAmount < minAmount {
		return false, false, fmt.Errorf("redeem amount should be larger or equal to %v", minAmount)
	}

	// validate value transfer of tx for redeem amount in ptoken
	if redeemReq.RedeemAmount != burnCoin.GetValue() {
		return false, false, errors.New("redeem amount should be equal to the tx value")
	}

	// validate tokenID
	if redeemReq.TokenID != burnedTokenID.String() {
		return false, false, NewMetadataTxError(PortalRedeemLiquidateExchangeRatesParamError, errors.New("TokenID in metadata is not matched to tokenID in tx"))
	}
	// check tokenId is portal token or not
	isPortalToken, err := chainRetriever.IsPortalToken(beaconHeight, redeemReq.TokenID, common.PortalVersion3)
	if !isPortalToken || err != nil {
		return false, false, errors.New("TokenID is not in portal tokens list")
	}

	// reject Redeem Request from Liquidation pool from BCHeightBreakPointPortalV3
	if beaconHeight >= config.Param().BCHeightBreakPointPortalV3 {
		return false, false, NewMetadataTxError(PortalRedeemLiquidateExchangeRatesParamError, fmt.Errorf("Should create redeem request from liquidation pool v3 after epoch %v", config.Param().BCHeightBreakPointPortalV3))
	}
	return true, true, nil
}

func (redeemReq PortalRedeemLiquidateExchangeRates) ValidateMetadataByItself() bool {
	return redeemReq.Type == PortalRedeemFromLiquidationPoolMeta
}

func (redeemReq PortalRedeemLiquidateExchangeRates) Hash() *common.Hash {
	record := redeemReq.MetadataBase.Hash().String()
	record += redeemReq.TokenID
	record += strconv.FormatUint(redeemReq.RedeemAmount, 10)
	record += redeemReq.RedeemerIncAddressStr
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (redeemReq *PortalRedeemLiquidateExchangeRates) BuildReqActions(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	actionContent := PortalRedeemLiquidateExchangeRatesAction{
		Meta:    *redeemReq,
		TxReqID: *tx.Hash(),
		ShardID: shardID,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(PortalRedeemFromLiquidationPoolMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (redeemReq *PortalRedeemLiquidateExchangeRates) CalculateSize() uint64 {
	return calculateSize(redeemReq)
}

func (redeemReq *PortalRedeemLiquidateExchangeRates) ToCompactBytes() ([]byte, error) {
	return toCompactBytes(redeemReq)
}
