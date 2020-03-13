package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/wallet"
	"reflect"
	"strconv"
)

type PortalRedeemLiquidateExchangeRates struct {
	MetadataBase
	TokenID               string // pTokenID in incognito chain
	RedeemAmount          uint64
	RedeemerIncAddressStr string
	RemoteAddress         string // btc/bnb/etc address
	RedeemFee             uint64 // redeem fee in PRV, 0.01% redeemAmount in PRV
}

type PortalRedeemLiquidateExchangeRatesAction struct {
	Meta    PortalRedeemLiquidateExchangeRates
	TxReqID common.Hash
	ShardID byte
}

type PortalRedeemLiquidateExchangeRatesContent struct {
	TokenID                 string // pTokenID in incognito chain
	RedeemAmount            uint64
	RedeemerIncAddressStr   string
	RemoteAddress           string // btc/bnb/etc address
	RedeemFee               uint64 // redeem fee in PRV, 0.01% redeemAmount in PRV
	TxReqID                 common.Hash
	ShardID                 byte
}

func NewPortalRedeemLiquidateExchangeRates(
	metaType int,
	tokenID string,
	redeemAmount uint64,
	incAddressStr string,
	remoteAddr string,
	redeemFee uint64,
) (*PortalRedeemLiquidateExchangeRates, error) {
	metadataBase := MetadataBase{Type: metaType}

	portalRedeemLiquidateExchangeRates := &PortalRedeemLiquidateExchangeRates {
		TokenID:               tokenID,
		RedeemAmount:          redeemAmount,
		RedeemerIncAddressStr: incAddressStr,
		RemoteAddress:         remoteAddr,
		RedeemFee:             redeemFee,
	}

	portalRedeemLiquidateExchangeRates.MetadataBase = metadataBase

	return portalRedeemLiquidateExchangeRates, nil
}

func (redeemReq PortalRedeemLiquidateExchangeRates) ValidateTxWithBlockChain(
	txr Transaction,
	bcr BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface,
) (bool, error) {
	return true, nil
}

func (redeemReq PortalRedeemLiquidateExchangeRates) ValidateSanityData(bcr BlockchainRetriever, txr Transaction, beaconHeight uint64) (bool, bool, error) {
	// Note: the metadata was already verified with *transaction.TxCustomToken level so no need to verify with *transaction.Tx level again as *transaction.Tx is embedding property of *transaction.TxCustomToken
	if txr.GetType() == common.TxCustomTokenPrivacyType && reflect.TypeOf(txr).String() == "*transaction.Tx" {
		if !txr.IsCoinsBurning(bcr, beaconHeight) {
			return false, false, errors.New("txnormal in tx redeem request must be coin burning tx")
		}
		// validate value transfer of tx for redeem fee in prv
		if redeemReq.RedeemFee != txr.CalculateTxValue() {
			return false, false, errors.New("redeem fee amount should be equal to the tx value")
		}
		return true, true, nil
	}

	// validate RedeemerIncAddressStr
	keyWallet, err := wallet.Base58CheckDeserialize(redeemReq.RedeemerIncAddressStr)
	if err != nil {
		return false, false, NewMetadataTxError(PortalRedeemLiquidateExchangeRatesParamError, errors.New("Address incognito redeem is invalid"))
	}


	incAddr := keyWallet.KeySet.PaymentAddress
	if len(incAddr.Pk) == 0 {
		return false, false, NewMetadataTxError(PortalRedeemLiquidateExchangeRatesParamError, errors.New("Payment incognito address is invalid"))
	}
	if !bytes.Equal(txr.GetSigPubKey()[:], incAddr.Pk[:]) {
		return false, false, NewMetadataTxError(PortalRedeemLiquidateExchangeRatesParamError, errors.New("Address incognito redeem is not signer"))
	}

	// check tx type
	if txr.GetType() != common.TxCustomTokenPrivacyType {
		return false, false, errors.New("tx redeem request must be TxCustomTokenPrivacyType")
	}

	if !txr.IsCoinsBurning(bcr, beaconHeight) {
		return false, false, errors.New("txprivacytoken in tx redeem request must be coin burning tx")
	}

	// validate redeem amount
	if redeemReq.RedeemAmount <= 0 {
		return false, false, errors.New("redeem amount should be larger than 0")
	}

	// validate redeem fee
	if redeemReq.RedeemFee <= 0 {
		return false, false, errors.New("redeem fee should be larger than 0")
	}

	// validate value transfer of tx for redeem amount in ptoken
	if redeemReq.RedeemAmount != txr.CalculateTxValue() {
		return false, false, errors.New("redeem amount should be equal to the tx value")
	}

	// validate tokenID
	if redeemReq.TokenID != txr.GetTokenID().String() {
		return false, false, NewMetadataTxError(PortalRedeemLiquidateExchangeRatesParamError, errors.New("TokenID in metadata is not matched to tokenID in tx"))
	}
	// check tokenId is portal token or not
	if !IsPortalToken(redeemReq.TokenID) {
		return false, false, NewMetadataTxError(PortalRedeemLiquidateExchangeRatesParamError, errors.New("TokenID is not in portal tokens list"))
	}

	//validate RemoteAddress
	// todo:
	if len(redeemReq.RemoteAddress) == 0 {
		return false, false, NewMetadataTxError(PortalRedeemLiquidateExchangeRatesParamError, errors.New("Remote address is invalid"))
	}

	return true, true, nil
}

func (redeemReq PortalRedeemLiquidateExchangeRates) ValidateMetadataByItself() bool {
	return redeemReq.Type == PortalRedeemLiquidateExchangeRatesMeta
}

func (redeemReq PortalRedeemLiquidateExchangeRates) Hash() *common.Hash {
	record := redeemReq.MetadataBase.Hash().String()
	record += redeemReq.TokenID
	record += strconv.FormatUint(redeemReq.RedeemAmount, 10)
	record += strconv.FormatUint(redeemReq.RedeemFee, 10)
	record += redeemReq.RedeemerIncAddressStr
	record += redeemReq.RemoteAddress
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (redeemReq *PortalRedeemLiquidateExchangeRates) BuildReqActions(tx Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
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
	action := []string{strconv.Itoa(PortalRedeemLiquidateExchangeRatesMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (redeemReq *PortalRedeemLiquidateExchangeRates) CalculateSize() uint64 {
	return calculateSize(redeemReq)
}
