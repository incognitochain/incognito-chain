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
	"math/big"
	"strconv"
)

type PortalCustodianWithdrawRequestV3 struct {
	basemeta.MetadataBase
	CustodianIncAddress      string
	CustodianExternalAddress string // there is no 0x prefix
	ExternalTokenID          string // collateral token ID, there is no 0x prefix
	Amount                   uint64
}

type PortalCustodianWithdrawRequestActionV3 struct {
	Meta    PortalCustodianWithdrawRequestV3
	TxReqID common.Hash
	ShardID byte
}

type PortalCustodianWithdrawRequestContentV3 struct {
	CustodianIncAddress      string
	CustodianExternalAddress string
	ExternalTokenID          string // collateral token ID
	Amount                   *big.Int
	TxReqID                  common.Hash
	ShardID                  byte
}

type CustodianWithdrawRequestStatusV3 struct {
	CustodianIncAddress      string
	CustodianExternalAddress string
	ExternalTokenID          string // collateral token ID
	Amount                   *big.Int
	TxReqID                  common.Hash
	Status                   int
}

func NewCustodianWithdrawRequestStatusV3(
	cusIncAddress string,
	cusExternalAddress string,
	externalTokenID string,
	amount *big.Int,
	txReqID common.Hash,
	status int) *CustodianWithdrawRequestStatusV3 {
	return &CustodianWithdrawRequestStatusV3{
		CustodianIncAddress:      cusIncAddress,
		CustodianExternalAddress: cusExternalAddress,
		ExternalTokenID:          externalTokenID,
		Amount:                   amount,
		TxReqID:                  txReqID,
		Status:                   status,
	}
}

func NewPortalCustodianWithdrawRequestV3(
	metaType int,
	cusIncAddress string,
	cusExternalAddress string,
	externalTokenID string,
	amount uint64, ) (*PortalCustodianWithdrawRequestV3, error) {
	portalCustodianWithdrawReq := &PortalCustodianWithdrawRequestV3{
		MetadataBase: basemeta.MetadataBase{
			Type: metaType,
		},
		CustodianIncAddress:      cusIncAddress,
		CustodianExternalAddress: cusExternalAddress,
		ExternalTokenID:          externalTokenID,
		Amount:                   amount,
	}
	return portalCustodianWithdrawReq, nil
}

func (req PortalCustodianWithdrawRequestV3) ValidateTxWithBlockChain(
	txr basemeta.Transaction,
	chainRetriever basemeta.ChainRetriever,
	shardViewRetriever basemeta.ShardViewRetriever,
	beaconViewRetriever basemeta.BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	// NOTE: verify supported tokens pair as needed
	return true, nil
}

func (req PortalCustodianWithdrawRequestV3) ValidateSanityData(chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, beaconHeight uint64, tx basemeta.Transaction) (bool, bool, error) {
	// check tx type
	if tx.GetType() != common.TxNormalType {
		return false, false, errors.New("custodian request withdraw v3: tx type must be TxNormalType")
	}

	// validate custodian payment address
	if len(req.CustodianIncAddress) <= 0 {
		return false, false, errors.New("custodian request withdraw v3: Payment address should be not empty")
	}
	keyWallet, err := wallet.Base58CheckDeserialize(req.CustodianIncAddress)
	if err != nil {
		return false, false, errors.New("custodian request withdraw v3: CustodianIncAddress is incorrect")
	}
	incogAddr := keyWallet.KeySet.PaymentAddress
	if len(incogAddr.Pk) == 0 {
		return false, false, errors.New("custodian request withdraw v3: Custodian public key is incorrect")
	}
	if !bytes.Equal(tx.GetSigPubKey()[:], incogAddr.Pk[:]) {
		return false, false, errors.New("custodian request withdraw v3: custodian incognito address is not tx's sender address")
	}

	// validate amount
	if req.Amount <= 0 {
		return false, false, errors.New("custodian request withdraw v3: amount should be larger than 0")
	}

	// validate ExternalTokenID
	if common.Has0xPrefix(req.ExternalTokenID) {
		return false, false, errors.New("custodian request withdraw v3: external tokenID shouldn't have 0x prefix")
	}
	// check externalTokenID should be one of supported collateral tokenIDs
	if !IsSupportedTokenCollateralV3(chainRetriever, beaconHeight, req.ExternalTokenID) {
		return false, false, errors.New("custodian request withdraw v3: ExternalTokenID is not portal collateral")
	}

	// validate remote address
	if common.Has0xPrefix(req.CustodianExternalAddress) {
			return false, false, errors.New("custodian request withdraw v3: external tokenID shouldn't have 0x prefix")
	}
	if isValid, err := ValidatePortalExternalAddress(common.ETHChainName, req.ExternalTokenID, req.CustodianExternalAddress); !isValid || err != nil {
		return false, false, errors.New("custodian request withdraw v3: custodian external address is invalid")
	}

	return true, true, nil
}

func (req PortalCustodianWithdrawRequestV3) ValidateMetadataByItself() bool {
	return req.Type == basemeta.PortalCustodianWithdrawRequestMetaV3
}

func (req PortalCustodianWithdrawRequestV3) Hash() *common.Hash {
	record := req.MetadataBase.Hash().String()
	record += req.CustodianIncAddress
	record += req.CustodianExternalAddress
	record += req.ExternalTokenID
	record += strconv.FormatUint(req.Amount, 10)

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (req *PortalCustodianWithdrawRequestV3) BuildReqActions(tx basemeta.Transaction, chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	actionContent := PortalCustodianWithdrawRequestActionV3{
		Meta:    *req,
		TxReqID: *tx.Hash(),
		ShardID: shardID,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(basemeta.PortalCustodianWithdrawRequestMetaV3), actionContentBase64Str}
	return [][]string{action}, nil
}

func (req *PortalCustodianWithdrawRequestV3) CalculateSize() uint64 {
	return basemeta.CalculateSize(req)
}
