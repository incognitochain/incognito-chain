package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/wallet"
	"strconv"
	"reflect"
	"errors"
)

type ExchangeRate struct {
	Amount uint64 //amount * 10^6 (USDT)
}

type PortalExchangeRates struct {
	MetadataBase
	SenderAddress string
	Rates map[string]ExchangeRate //Key: Crypto Symbol
}

type PortalExchangeRatesAction struct {
	Meta PortalExchangeRates
	TxReqID common.Hash
	LockTime int64
	ShardID byte
}

func NewPortalExchangeRates(metaType int, senderAddress string, currency map[string]ExchangeRate) (*PortalExchangeRates, error) {
	metadataBase := MetadataBase{Type: metaType}

	portalExchangeRates := &PortalExchangeRates {
		SenderAddress: senderAddress,
		Rates: currency,
	}

	portalExchangeRates.MetadataBase = metadataBase

	return portalExchangeRates, nil
}

type PortalExchangeRatesContent struct {
	SenderAddress string
	Rates map[string]ExchangeRate
	TxReqID common.Hash
	LockTime int64
}

func (portalExchangeRates PortalExchangeRates) ValidateTxWithBlockChain(
	txr Transaction,
	bcr BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface,
) (bool, error) {
	// NOTE: verify supported tokens pair as needed
	return true, nil
}

func (portalExchangeRates PortalExchangeRates) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	if txr.GetType() == common.TxCustomTokenPrivacyType && reflect.TypeOf(txr).String() == "*transaction.Tx" {
		return true, true, nil
	}

	keyWallet, err := wallet.Base58CheckDeserialize(portalExchangeRates.SenderAddress)
	if err != nil {
		return false, false, NewMetadataTxError(IssuingRequestNewIssuingRequestFromMapEror, errors.New("SenderAddress incorrect"))
	}

	senderAddr := keyWallet.KeySet.PaymentAddress
	if len(senderAddr.Pk) == 0 {
		return false, false, errors.New("Sender address invalid, sender address must be incognito address")
	}

	if !bytes.Equal(txr.GetSigPubKey()[:], senderAddr.Pk[:]) {
		return false, false, errors.New("Sender address is not signer tx")
	}

	if txr.GetType() != common.TxNormalType {
		return false, false, errors.New("tx exchange rates must be TxNormalType")
	}

	// check burning tx
	if !txr.IsCoinsBurning(bcr) {
		return false, false, errors.New("must send coin to burning address")
	}

	for pTokenId, exchangeRate := range portalExchangeRates.Rates {
		isSupported, err := common.SliceExists(PortalSupportedExchangeRatesSymbols, pTokenId)
		if err != nil || !isSupported {
			return false, false, errors.New("Public token is not supported currently")
		}

		if exchangeRate.Amount == 0 {
			return false, false, errors.New("porting fee should be larger than 0")
		}
	}

	return true, true, nil
}

func (portalExchangeRates PortalExchangeRates) ValidateMetadataByItself() bool {
	return portalExchangeRates.Type == PortalExchangeRatesMeta
}

func (portalExchangeRates PortalExchangeRates) Hash() *common.Hash {
	record := portalExchangeRates.MetadataBase.Hash().String()
	record += portalExchangeRates.SenderAddress

	for pTokenId, exchange := range portalExchangeRates.Rates {
		record += pTokenId
		record += strconv.FormatUint(exchange.Amount, 10)
	}

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (portalExchangeRates *PortalExchangeRates) BuildReqActions(tx Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	actionContent := PortalExchangeRatesAction{
		Meta:    *portalExchangeRates,
		TxReqID: *tx.Hash(),
		LockTime: tx.GetLockTime(),
		ShardID: shardID,
	}

	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(PortalExchangeRatesMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (portalExchangeRates *PortalExchangeRates) CalculateSize() uint64 {
	return calculateSize(portalExchangeRates)
}