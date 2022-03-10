package metadata

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/wallet"
	"strconv"
)

type PortalExchangeRates struct {
	MetadataBaseWithSignature
	SenderAddress string
	Rates         []*ExchangeRateInfo //amount * 10^6 (USDT)
}

type PortalExchangeRatesAction struct {
	Meta     PortalExchangeRates
	TxReqID  common.Hash
	LockTime int64
	ShardID  byte
}

type ExchangeRateInfo struct {
	PTokenID string
	Rate     uint64
}

type ExchangeRatesRequestStatus struct {
	//Status        byte 		// dont need this field
	SenderAddress string
	Rates         []*ExchangeRateInfo
}

func NewExchangeRatesRequestStatus(senderAddress string, rates []*ExchangeRateInfo) *ExchangeRatesRequestStatus {
	return &ExchangeRatesRequestStatus{SenderAddress: senderAddress, Rates: rates}
}

func NewPortalExchangeRates(metaType int, senderAddress string, currency []*ExchangeRateInfo) (*PortalExchangeRates, error) {
	metadataBase := NewMetadataBaseWithSignature(metaType)

	portalExchangeRates := &PortalExchangeRates{
		SenderAddress: senderAddress,
		Rates:         currency,
	}

	portalExchangeRates.MetadataBaseWithSignature = *metadataBase

	return portalExchangeRates, nil
}

type PortalExchangeRatesContent struct {
	SenderAddress string
	Rates         []*ExchangeRateInfo
	TxReqID       common.Hash
	LockTime      int64
}

func (portalExchangeRates PortalExchangeRates) ValidateTxWithBlockChain(
	txr Transaction,
	chainRetriever ChainRetriever,
	shardViewRetriever ShardViewRetriever,
	beaconViewRetriever BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	// NOTE: verify supported tokens pair as needed
	return true, nil
}

func (portalExchangeRates PortalExchangeRates) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, txr Transaction) (bool, bool, error) {
	feederAddress := chainRetriever.GetPortalFeederAddress(beaconHeight)
	isEqual, err := wallet.ComparePaymentAddresses(portalExchangeRates.SenderAddress, feederAddress)
	if err != nil {
		return false, false, fmt.Errorf("cannot compare payment address %v and %v: %v", portalExchangeRates.SenderAddress, feederAddress, err)
	}
	if !isEqual {
		return false, false, fmt.Errorf("sender address and feeder address mismatch (%v != %v)\n", portalExchangeRates.SenderAddress, feederAddress)
	}

	keyWallet, err := wallet.Base58CheckDeserialize(portalExchangeRates.SenderAddress)
	if err != nil {
		return false, false, errors.New("SenderAddress incorrect")
	}

	senderAddr := keyWallet.KeySet.PaymentAddress
	if len(senderAddr.Pk) == 0 {
		return false, false, errors.New("Sender address invalid, sender address must be incognito address")
	}
	if ok, err := portalExchangeRates.MetadataBaseWithSignature.VerifyMetadataSignature(senderAddr.Pk, txr); err != nil || !ok {
		return false, false, errors.New("Sender is unauthorized")
	}

	if txr.GetType() != common.TxNormalType {
		return false, false, errors.New("Tx exchange rates must be TxNormalType")
	}

	for _, value := range portalExchangeRates.Rates {
		if !chainRetriever.IsPortalExchangeRateToken(beaconHeight, value.PTokenID) {
			return false, false, errors.New("Public token is not supported currently")
		}

		if value.Rate <= 0 {
			return false, false, errors.New("Exchange rates should be larger than 0")
		}
	}

	return true, true, nil
}

func (portalExchangeRates PortalExchangeRates) ValidateMetadataByItself() bool {
	return portalExchangeRates.Type == PortalExchangeRatesMeta
}

func (portalExchangeRates PortalExchangeRates) Hash() *common.Hash {
	record := portalExchangeRates.MetadataBaseWithSignature.Hash().String()
	record += portalExchangeRates.SenderAddress
	for _, rateInfo := range portalExchangeRates.Rates {
		record += rateInfo.PTokenID
		record += strconv.FormatUint(rateInfo.Rate, 10)
	}
	if portalExchangeRates.Sig != nil && len(portalExchangeRates.Sig) != 0 {
		record += string(portalExchangeRates.Sig)
	}
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (portalExchangeRates PortalExchangeRates) HashWithoutSig() *common.Hash {
	record := portalExchangeRates.MetadataBaseWithSignature.Hash().String()
	record += portalExchangeRates.SenderAddress
	for _, rateInfo := range portalExchangeRates.Rates {
		record += rateInfo.PTokenID
		record += strconv.FormatUint(rateInfo.Rate, 10)
	}
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (portalExchangeRates *PortalExchangeRates) BuildReqActions(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	actionContent := PortalExchangeRatesAction{
		Meta:     *portalExchangeRates,
		TxReqID:  *tx.Hash(),
		LockTime: tx.GetLockTime(),
		ShardID:  shardID,
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

func (portalExchangeRates *PortalExchangeRates) ToCompactBytes() ([]byte, error) {
	return toCompactBytes(portalExchangeRates)
}
