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

// PortalRequestPTokens - portal user requests ptoken (after sending pubToken to custodians)
// metadata - user requests ptoken - create normal tx with this metadata
type PortalRequestPTokens struct {
	basemeta.MetadataBase
	UniquePortingID string
	TokenID         string // pTokenID in incognito chain
	IncogAddressStr string
	PortingAmount   uint64
	PortingProof    string
}

// PortalRequestPTokensAction - shard validator creates instruction that contain this action content
type PortalRequestPTokensAction struct {
	Meta    PortalRequestPTokens
	TxReqID common.Hash
	ShardID byte
}

// PortalRequestPTokensContent - Beacon builds a new instruction with this content after receiving a instruction from shard
// It will be appended to beaconBlock
// both accepted and rejected status
type PortalRequestPTokensContent struct {
	UniquePortingID string
	TokenID         string // pTokenID in incognito chain
	IncogAddressStr string
	PortingAmount   uint64
	PortingProof    string
	TxReqID         common.Hash
	ShardID         byte
}

// PortalRequestPTokensStatus - Beacon tracks status of request ptokens into db
type PortalRequestPTokensStatus struct {
	Status          byte
	UniquePortingID string
	TokenID         string // pTokenID in incognito chain
	IncogAddressStr string
	PortingAmount   uint64
	PortingProof    string
	TxReqID         common.Hash
}

func NewPortalRequestPTokens(
	metaType int,
	uniquePortingID string,
	tokenID string,
	incogAddressStr string,
	portingAmount uint64,
	portingProof string) (*PortalRequestPTokens, error) {
	metadataBase := basemeta.MetadataBase{
		Type: metaType,
	}
	requestPTokenMeta := &PortalRequestPTokens{
		UniquePortingID: uniquePortingID,
		TokenID:         tokenID,
		IncogAddressStr: incogAddressStr,
		PortingAmount:   portingAmount,
		PortingProof:    portingProof,
	}
	requestPTokenMeta.MetadataBase = metadataBase
	return requestPTokenMeta, nil
}

func (reqPToken PortalRequestPTokens) ValidateTxWithBlockChain(
	txr basemeta.Transaction,
	chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	return true, nil
}

func (reqPToken PortalRequestPTokens) ValidateSanityData(chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, beaconHeight uint64, txr basemeta.Transaction) (bool, bool, error) {
	// validate IncogAddressStr
	keyWallet, err := wallet.Base58CheckDeserialize(reqPToken.IncogAddressStr)
	if err != nil {
		return false, false, basemeta.NewMetadataTxError(basemeta.PortalRequestPTokenParamError, errors.New("Requester incognito address is invalid"))
	}
	incogAddr := keyWallet.KeySet.PaymentAddress
	if len(incogAddr.Pk) == 0 {
		return false, false, basemeta.NewMetadataTxError(basemeta.PortalRequestPTokenParamError, errors.New("Requester incognito address is invalid"))
	}
	if !bytes.Equal(txr.GetSigPubKey()[:], incogAddr.Pk[:]) {
		return false, false, basemeta.NewMetadataTxError(basemeta.PortalRequestPTokenParamError, errors.New("Requester incognito address is not signer"))
	}

	// check tx type
	if txr.GetType() != common.TxNormalType {
		return false, false, errors.New("tx custodian deposit must be TxNormalType")
	}

	// validate amount deposit
	if reqPToken.PortingAmount == 0 {
		return false, false, errors.New("porting amount should be larger than 0")
	}

	// validate tokenID and porting proof
	if !!chainRetriever.IsPortalToken(beaconHeight, reqPToken.TokenID) {
		return false, false, basemeta.NewMetadataTxError(basemeta.PortalRequestPTokenParamError, errors.New("TokenID is not supported currently on Portal"))
	}

	return true, true, nil
}

func (reqPToken PortalRequestPTokens) ValidateMetadataByItself() bool {
	return reqPToken.Type == basemeta.PortalUserRequestPTokenMeta
}

func (reqPToken PortalRequestPTokens) Hash() *common.Hash {
	record := reqPToken.MetadataBase.Hash().String()
	record += reqPToken.UniquePortingID
	record += reqPToken.TokenID
	record += reqPToken.IncogAddressStr
	record += strconv.FormatUint(reqPToken.PortingAmount, 10)
	record += reqPToken.PortingProof
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (reqPToken *PortalRequestPTokens) BuildReqActions(tx basemeta.Transaction, chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	actionContent := PortalRequestPTokensAction{
		Meta:    *reqPToken,
		TxReqID: *tx.Hash(),
		ShardID: shardID,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(basemeta.PortalUserRequestPTokenMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (reqPToken *PortalRequestPTokens) CalculateSize() uint64 {
	return basemeta.CalculateSize(reqPToken)
}
