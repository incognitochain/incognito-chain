package metadata

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/incognitochain/incognito-chain/privacy/coin"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type PortalUnshieldRequest struct {
	MetadataBase
	OTAPubKeyStr   string // OTA
	TxRandomStr    string
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
	OTAPubKeyStr   string
	TxRandomStr    string
	RemoteAddress  string
	TokenID        string
	UnshieldAmount uint64
	TxReqID        common.Hash
	ShardID        byte
}

type PortalUnshieldRequestStatus struct {
	OTAPubKeyStr   string
	TxRandomStr    string
	RemoteAddress  string
	TokenID        string
	UnshieldAmount uint64
	UnshieldID     string
	ExternalTxID   string
	ExternalFee    uint64
	Status         int
}

func NewPortalUnshieldRequestStatus(otaPubKeyStr, txRandomStr, tokenID, remoteAddress, unshieldID, externalTxID string, burnAmount, externalFee uint64, status int) *PortalUnshieldRequestStatus {
	return &PortalUnshieldRequestStatus{
		OTAPubKeyStr:   otaPubKeyStr,
		TxRandomStr:    txRandomStr,
		RemoteAddress:  remoteAddress,
		TokenID:        tokenID,
		UnshieldAmount: burnAmount,
		UnshieldID:     unshieldID,
		ExternalTxID:   externalTxID,
		ExternalFee:    externalFee,
		Status:         status,
	}
}

func NewPortalUnshieldRequest(metaType int, otaPubKeyStr, txRandomStr string, tokenID, remoteAddress string, burnAmount uint64) (*PortalUnshieldRequest, error) {
	portalUnshieldReq := &PortalUnshieldRequest{
		OTAPubKeyStr:   otaPubKeyStr,
		TxRandomStr:    txRandomStr,
		UnshieldAmount: burnAmount,
		RemoteAddress:  remoteAddress,
		TokenID:        tokenID,
	}

	portalUnshieldReq.MetadataBase = MetadataBase{
		Type: metaType,
	}

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

	// check tx type
	if tx.GetType() != common.TxCustomTokenPrivacyType {
		return false, false, NewMetadataTxError(metadataCommon.PortalV4UnshieldRequestValidateSanityDataError, errors.New("tx burn ptoken must be TxCustomTokenPrivacyType"))
	}

	// check ota address string and tx random is valid
	_, err, ver := metadataCommon.CheckIncognitoAddress(uReq.OTAPubKeyStr, uReq.TxRandomStr)
	if err != nil {
		return false, false, NewMetadataTxError(metadataCommon.PortalV4UnshieldRequestValidateSanityDataError,
			fmt.Errorf("payment address string or txrandom is not corrrect format %v", err))
	}

	// check tx version
	if int8(ver) != tx.GetVersion() {
		return false, false, NewMetadataTxError(metadataCommon.PortalV4UnshieldRequestValidateSanityDataError,
			fmt.Errorf("payment address version (%v) and tx version (%v) mismatch", ver, tx.GetVersion()))
	}
	if tx.GetVersion() != 2 {
		return false, false, NewMetadataTxError(metadataCommon.PortalV4UnshieldRequestValidateSanityDataError,
			errors.New("Tx unshielding request must be version 2"))
	}

	// check tx burn or not
	isBurn, _, burnedCoin, burnedToken, err := tx.GetTxFullBurnData()
	if err != nil || !isBurn {
		return false, false, NewMetadataTxError(metadataCommon.PortalV4UnshieldRequestValidateSanityDataError, fmt.Errorf("this is not burn tx. Error %v", err))
	}
	burningAmt := burnedCoin.GetValue()
	burningTokenID := burnedToken.String()

	// validate tokenID
	if uReq.TokenID != burningTokenID {
		return false, false, NewMetadataTxError(metadataCommon.PortalV4UnshieldRequestValidateSanityDataError, errors.New("TokenID in metadata is not matched to burning tokenID in tx"))
	}
	// check tokenId is portal token or not
	if ok, err := chainRetriever.IsPortalToken(beaconHeight, uReq.TokenID, common.PortalVersion4); !ok || err != nil {
		return false, false, NewMetadataTxError(metadataCommon.PortalV4UnshieldRequestValidateSanityDataError, errors.New("TokenID is not in portal tokens list v4"))
	}

	// validate burning amount
	// check unshielding amount is equal to burning amount
	if uReq.UnshieldAmount != burningAmt {
		return false, false, NewMetadataTxError(metadataCommon.PortalV4UnshieldRequestValidateSanityDataError, errors.New("burning amount %v should be equal to unshielding amount"))
	}

	// check unshielding amount is not less then minimum unshielding amount
	minUnshieldAmount := chainRetriever.GetPortalV4MinUnshieldAmount(uReq.TokenID, beaconHeight)
	if uReq.UnshieldAmount < minUnshieldAmount {
		return false, false, NewMetadataTxError(metadataCommon.PortalV4UnshieldRequestValidateSanityDataError, fmt.Errorf("unshielding amount should be larger or equal to %v", minUnshieldAmount))
	}

	// validate RemoteAddress
	isValidRemoteAddress, err := chainRetriever.IsValidPortalRemoteAddress(uReq.TokenID, uReq.RemoteAddress, beaconHeight, common.PortalVersion4)
	if err != nil || !isValidRemoteAddress {
		return false, false, NewMetadataTxError(metadataCommon.PortalV4UnshieldRequestValidateSanityDataError, fmt.Errorf("Remote address %v is not a valid address of tokenID %v - Error %v", uReq.RemoteAddress, uReq.TokenID, err))
	}

	return true, true, nil
}

func (uReq PortalUnshieldRequest) ValidateMetadataByItself() bool {
	return uReq.Type == metadataCommon.PortalV4UnshieldingRequestMeta
}

func (uReq PortalUnshieldRequest) Hash() *common.Hash {
	record := uReq.MetadataBase.Hash().String()
	record += uReq.OTAPubKeyStr
	record += uReq.TxRandomStr
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
	action := []string{strconv.Itoa(metadataCommon.PortalV4UnshieldingRequestMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (uReq *PortalUnshieldRequest) CalculateSize() uint64 {
	return calculateSize(uReq)
}

func (uReq *PortalUnshieldRequest) ToCompactBytes() ([]byte, error) {
	return toCompactBytes(uReq)
}

func (uReq *PortalUnshieldRequest) GetOTADeclarations() []OTADeclaration {
	result := []OTADeclaration{}
	pk, _, err := coin.ParseOTAInfoFromString(uReq.OTAPubKeyStr, uReq.TxRandomStr)
	tokenID := common.ConfidentialAssetID
	if err == nil {
		result = append(result, OTADeclaration{PublicKey: pk.ToBytes(), TokenID: tokenID})
	}
	return result
}
