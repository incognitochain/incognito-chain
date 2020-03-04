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

type PortalCustodianWithDrawRequest struct {
	MetadataBase
	PaymentAddress string
	Amount uint64
}

type PortalCustodianWithDrawRequestAction struct {
	Meta PortalCustodianWithDrawRequest
	TxReqID common.Hash
	ShardID byte
}

type PortalCustodianWithDrawRequestContent struct {
	PaymentAddress string
	Amount uint64
	TxReqID common.Hash
}

func NewPortalCustodianWithDrawRequest(metaType int, paymentAddress string , amount uint64) (*PortalCustodianWithDrawRequest, error){
	metadataBase := MetadataBase{
		Type: metaType,
	}

	portalCustodianWithDrawReq := &PortalCustodianWithDrawRequest {
		PaymentAddress: paymentAddress,
		Amount: amount,
	}

	portalCustodianWithDrawReq.MetadataBase = metadataBase

	return portalCustodianWithDrawReq, nil
}

func (withDraw PortalCustodianWithDrawRequest) ValidateTxWithBlockChain(
	txr Transaction,
	bcr BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface,
) (bool, error) {
	// NOTE: verify supported tokens pair as needed
	return true, nil
}

func (withDraw PortalCustodianWithDrawRequest) ValidateSanityData(bcr BlockchainRetriever, txr Transaction, beaconHeight uint64) (bool, bool, error) {
	if txr.GetType() == common.TxCustomTokenPrivacyType && reflect.TypeOf(txr).String() == "*transaction.Tx" {
		return true, true, nil
	}

	if len(withDraw.PaymentAddress) <= 0 {
		return false, false, errors.New("Payment address should be not empty")
	}

	// validate Payment address
	keyWallet, err := wallet.Base58CheckDeserialize(withDraw.PaymentAddress)
	if err != nil {
		return false, false, NewMetadataTxError(IssuingRequestNewIssuingRequestFromMapEror, errors.New("ContributorAddressStr incorrect"))
	}

	incogAddr := keyWallet.KeySet.PaymentAddress
	if len(incogAddr.Pk) == 0 {
		return false, false, errors.New("wrong custodian incognito address")
	}
	if !bytes.Equal(txr.GetSigPubKey()[:], incogAddr.Pk[:]) {
		return false, false, errors.New("custodian incognito address is not signer tx")
	}

	// check tx type
	if txr.GetType() != common.TxNormalType {
		return false, false, errors.New("tx custodian deposit must be TxNormalType")
	}

	if withDraw.Amount <= 0 {
		return false, false, errors.New("Amount should be larger than 0")
	}

	return true, true, nil
}

func (withDraw PortalCustodianWithDrawRequest) ValidateMetadataByItself() bool {
	return withDraw.Type == PortalCustodianWithDrawRequestMeta
}

func (withDraw PortalCustodianWithDrawRequest) Hash() *common.Hash {
	record := withDraw.MetadataBase.Hash().String()
	record += withDraw.PaymentAddress
	record += strconv.FormatUint(withDraw.Amount, 10)

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (withDraw *PortalCustodianWithDrawRequest) BuildReqActions(tx Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	actionContent := PortalCustodianWithDrawRequestAction{
		Meta:    *withDraw,
		TxReqID: *tx.Hash(),
		ShardID: shardID,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(PortalCustodianWithDrawRequestMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (withDraw *PortalCustodianWithDrawRequest) CalculateSize() uint64 {
	return calculateSize(withDraw)
}