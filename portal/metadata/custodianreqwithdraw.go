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

type PortalCustodianWithdrawRequest struct {
	basemeta.MetadataBase
	PaymentAddress string
	Amount         uint64
}

type PortalCustodianWithdrawRequestAction struct {
	Meta    PortalCustodianWithdrawRequest
	TxReqID common.Hash
	ShardID byte
}

type PortalCustodianWithdrawRequestContent struct {
	PaymentAddress       string
	Amount               uint64
	RemainFreeCollateral uint64
	TxReqID              common.Hash
	ShardID              byte
}

type CustodianWithdrawRequestStatus struct {
	PaymentAddress                string
	Amount                        uint64
	Status                        int
	RemainCustodianFreeCollateral uint64
}

func NewCustodianWithdrawRequestStatus(paymentAddress string, amount uint64, status int, remainCustodianFreeCollateral uint64) *CustodianWithdrawRequestStatus {
	return &CustodianWithdrawRequestStatus{PaymentAddress: paymentAddress, Amount: amount, Status: status, RemainCustodianFreeCollateral: remainCustodianFreeCollateral}
}

func NewPortalCustodianWithdrawRequest(metaType int, paymentAddress string, amount uint64) (*PortalCustodianWithdrawRequest, error) {
	metadataBase := basemeta.MetadataBase{
		Type: metaType,
	}

	portalCustodianWithdrawReq := &PortalCustodianWithdrawRequest{
		PaymentAddress: paymentAddress,
		Amount:         amount,
	}

	portalCustodianWithdrawReq.MetadataBase = metadataBase

	return portalCustodianWithdrawReq, nil
}

func (withdrawReq PortalCustodianWithdrawRequest) ValidateTxWithBlockChain(
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

func (withdrawReq PortalCustodianWithdrawRequest) ValidateSanityData(chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, beaconHeight uint64, tx basemeta.Transaction) (bool, bool, error) {
	// validate Payment address
	if len(withdrawReq.PaymentAddress) <= 0 {
		return false, false, errors.New("PaymentAddress should be not empty")
	}
	keyWallet, err := wallet.Base58CheckDeserialize(withdrawReq.PaymentAddress)
	if err != nil {
		return false, false, errors.New("PaymentAddress incorrect")
	}
	incogAddr := keyWallet.KeySet.PaymentAddress
	if len(incogAddr.Pk) == 0 {
		return false, false, errors.New("wrong custodian incognito address")
	}
	if !bytes.Equal(tx.GetSigPubKey()[:], incogAddr.Pk[:]) {
		return false, false, errors.New("custodian incognito address is not signer tx")
	}

	// check tx type
	if tx.GetType() != common.TxNormalType {
		return false, false, errors.New("tx custodian deposit must be TxNormalType")
	}

	// check withdraw amount
	if withdrawReq.Amount <= 0 {
		return false, false, errors.New("Amount should be larger than 0")
	}

	return true, true, nil
}

func (withdrawReq PortalCustodianWithdrawRequest) ValidateMetadataByItself() bool {
	return withdrawReq.Type == basemeta.PortalCustodianWithdrawRequestMeta
}

func (withdrawReq PortalCustodianWithdrawRequest) Hash() *common.Hash {
	record := withdrawReq.MetadataBase.Hash().String()
	record += withdrawReq.PaymentAddress
	record += strconv.FormatUint(withdrawReq.Amount, 10)

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (withdrawReq *PortalCustodianWithdrawRequest) BuildReqActions(tx basemeta.Transaction, chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	actionContent := PortalCustodianWithdrawRequestAction{
		Meta:    *withdrawReq,
		TxReqID: *tx.Hash(),
		ShardID: shardID,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(basemeta.PortalCustodianWithdrawRequestMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (withdrawReq *PortalCustodianWithdrawRequest) CalculateSize() uint64 {
	return basemeta.CalculateSize(withdrawReq)
}
