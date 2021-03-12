package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/wallet"
	"reflect"
	"strconv"
)

type PortalUnshieldRequest struct {
	MetadataBase
	IncAddressStr  string
	RemoteAddress  string
	TokenID        string
	UnshieldAmount uint64
}

type PortalUnshieldRequestAction struct {
	Meta    PortalUnshieldRequest
	TxReqID common.Hash
	ShardID byte
}

type PortalUnshieldRequestContent struct {
	IncAddressStr  string
	RemoteAddress  string
	TokenID        string
	UnshieldAmount uint64
	TxReqID        common.Hash
	ShardID        byte
}

type PortalUnshieldRequestStatus struct {
	IncAddressStr  string
	RemoteAddress  string
	TokenID        string
	UnshieldAmount uint64
	UnshieldID     string
	Status         int
}

func NewPortalUnshieldRequestStatus(incAddressStr, tokenID, remoteAddress string, burnAmount uint64, status int) *PortalUnshieldRequestStatus {
	return &PortalUnshieldRequestStatus{
		IncAddressStr:          incAddressStr,
		UnshieldAmount:         burnAmount,
		Status:                 status,
		TokenID:                tokenID,
		RemoteAddress:          remoteAddress,
	}
}

func NewPortalUnshieldRequest(metaType int, incAddressStr, tokenID, remoteAddress string, burnAmount uint64) (*PortalUnshieldRequest, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}

	portalUnshieldReq := &PortalUnshieldRequest{
		IncAddressStr:        incAddressStr,
		UnshieldAmount:       burnAmount,
		RemoteAddress:        remoteAddress,
		TokenID:              tokenID,
	}

	portalUnshieldReq.MetadataBase = metadataBase

	return portalUnshieldReq, nil
}

func (uReq PortalUnshieldRequest) ValidateTxWithBlockChain(
	txr Transaction,
	chainRetriever ChainRetriever,
	shardViewRetriever ShardViewRetriever,
	beaconViewRetriever BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	return true, nil
}

func (uReq PortalUnshieldRequest) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	// Note: the metadata was already verified with *transaction.TxCustomToken level so no need to verify with *transaction.Tx level again as *transaction.Tx is embedding property of *transaction.TxCustomToken
	if tx.GetType() == common.TxCustomTokenPrivacyType && reflect.TypeOf(tx).String() == "*transaction.Tx" {
		return true, true, nil
	}

	// validate RedeemerIncAddressStr
	keyWallet, err := wallet.Base58CheckDeserialize(uReq.IncAddressStr)
	if err != nil {
		return false, false, NewMetadataTxError(PortalV4UnshieldRequestValidateSanityDataError, errors.New("Requester incognito address is invalid"))
	}
	incAddr := keyWallet.KeySet.PaymentAddress
	if len(incAddr.Pk) == 0 {
		return false, false, NewMetadataTxError(PortalV4UnshieldRequestValidateSanityDataError, errors.New("Requester incognito address is invalid"))
	}
	if !bytes.Equal(tx.GetSigPubKey()[:], incAddr.Pk[:]) {
		return false, false, NewMetadataTxError(PortalV4UnshieldRequestValidateSanityDataError, errors.New("Requester incognito address is not signer"))
	}

	// check tx type
	if tx.GetType() != common.TxCustomTokenPrivacyType {
		return false, false, NewMetadataTxError(PortalV4UnshieldRequestValidateSanityDataError, errors.New("tx burn ptoken must be TxCustomTokenPrivacyType"))
	}

	if !tx.IsCoinsBurning(chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight) {
		return false, false, NewMetadataTxError(PortalV4UnshieldRequestValidateSanityDataError, errors.New("txprivacytoken in tx burn ptoken must be coin burning tx"))
	}

	// validate burning amount
	minUnshieldAmount := chainRetriever.GetPortalV4MinUnshieldAmount(uReq.TokenID, beaconHeight)
	if uReq.UnshieldAmount < minUnshieldAmount {
		return false, false, NewMetadataTxError(PortalV4UnshieldRequestValidateSanityDataError, fmt.Errorf("burning amount should be larger or equal to %v", minUnshieldAmount))
	}

	// validate value transfer of tx for redeem amount in ptoken
	if uReq.UnshieldAmount != tx.CalculateTxValue() {
		return false, false, NewMetadataTxError(PortalV4UnshieldRequestValidateSanityDataError, errors.New("burning amount should be equal to the tx value"))
	}

	// validate tokenID
	if uReq.TokenID != tx.GetTokenID().String() {
		return false, false, NewMetadataTxError(PortalV4UnshieldRequestValidateSanityDataError, errors.New("TokenID in metadata is not matched to tokenID in tx"))
	}
	// check tokenId is portal token or not
	if ok, err := chainRetriever.IsPortalToken(beaconHeight, uReq.TokenID, common.PortalVersion4); !ok || err != nil {
		return false, false, NewMetadataTxError(PortalV4UnshieldRequestValidateSanityDataError, errors.New("TokenID is not in portal tokens list v4"))
	}

	// validate RemoteAddress
	if len(uReq.RemoteAddress) == 0 {
		return false, false, NewMetadataTxError(PortalV4UnshieldRequestValidateSanityDataError, errors.New("Remote address is invalid"))
	}
	isValidRemoteAddress, err := chainRetriever.IsValidPortalRemoteAddress(uReq.TokenID, uReq.RemoteAddress, beaconHeight, common.PortalVersion4)
	if err != nil || !isValidRemoteAddress {
		return false, false, NewMetadataTxError(PortalV4UnshieldRequestValidateSanityDataError, fmt.Errorf("Remote address %v is not a valid address of tokenID %v - Error %v", uReq.RemoteAddress, uReq.TokenID, err))
	}

	return true, true, nil
}

func (uReq PortalUnshieldRequest) ValidateMetadataByItself() bool {
	return uReq.Type == PortalV4UnshieldingRequestMeta
}

func (uReq PortalUnshieldRequest) Hash() *common.Hash {
	record := uReq.MetadataBase.Hash().String()
	record += uReq.IncAddressStr
	record += uReq.RemoteAddress
	record += strconv.FormatUint(uReq.UnshieldAmount, 10)
	record += uReq.TokenID

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (uReq *PortalUnshieldRequest) BuildReqActions(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	actionContent := PortalUnshieldRequestAction{
		Meta:    *uReq,
		TxReqID: *tx.Hash(),
		ShardID: shardID,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(PortalV4UnshieldingRequestMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (uReq *PortalUnshieldRequest) CalculateSize() uint64 {
	return calculateSize(uReq)
}