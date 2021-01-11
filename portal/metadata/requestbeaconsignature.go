package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/incognitochain/incognito-chain/basemeta"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/wallet"
	"strconv"
)

// PortalSignedTransaction - portal request signature from beacon
// metadata - user or custodian request signature of beacon to claim public token - create normal tx with this metadata
type PortalSignatureRequest struct {
	basemeta.MetadataBase
	SubmitProofTxID string
	TxFee           uint
	IncogAddressStr string
	RemoteAddress   string
	TokenID         string
}

// PortalSignatureAction - shard validator creates instruction that contain this action content
type PortalSignatureRequestAction struct {
	Meta    PortalSignatureRequest
	TxReqID common.Hash
	ShardID byte
}

// PortalSignatureContent - Beacon builds a new instruction with this content after receiving a instruction from shard
// It will be appended to beaconBlock
// both accepted and rejected status
type PortalSignatureRequestContent struct {
	SubmitProofTxID string
	SignedTx        string
	TxFee           uint
	IncogAddressStr string
	RemoteAddress   string
	TokenID         string
	TxReqID         common.Hash
	ShardID         byte
}

// PortalSignatureStatus - Beacon tracks status of request beacon signature into db
type PortalSignatureRequestStatus struct {
	Status          byte
	SubmitProofTxID string
	SignedTx        string
	TxFee           uint
	IncogAddressStr string
	RemoteAddress   string
	TokenID         string
	TxReqID         common.Hash
	ShardID         byte
}

func NewPortalSignatureRequest(
	metaType int,
	submitProofTxID string,
	txFee uint,
	incogAddressStr string,
	remoteAddress string,
	tokenID string) (*PortalSignatureRequest, error) {
	metadataBase := basemeta.MetadataBase{
		Type: metaType,
	}
	requestPTokenMeta := &PortalSignatureRequest{
		SubmitProofTxID: submitProofTxID,
		IncogAddressStr: incogAddressStr,
		TxFee:           txFee,
		RemoteAddress:   remoteAddress,
		TokenID:         tokenID,
	}
	requestPTokenMeta.MetadataBase = metadataBase
	return requestPTokenMeta, nil
}

func (sigRq PortalSignatureRequest) ValidateTxWithBlockChain(
	txr basemeta.Transaction,
	chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	return true, nil
}

func (sigRq PortalSignatureRequest) ValidateSanityData(chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, beaconHeight uint64, txr basemeta.Transaction) (bool, bool, error) {
	// validate IncogAddressStr
	keyWallet, err := wallet.Base58CheckDeserialize(sigRq.IncogAddressStr)
	if err != nil {
		return false, false, basemeta.NewMetadataTxError(basemeta.PortalSignatureRequestParamError, errors.New("Requester incognito address is invalid"))
	}
	incogAddr := keyWallet.KeySet.PaymentAddress
	if len(incogAddr.Pk) == 0 {
		return false, false, basemeta.NewMetadataTxError(basemeta.PortalSignatureRequestParamError, errors.New("Requester incognito address is invalid"))
	}
	if !bytes.Equal(txr.GetSigPubKey()[:], incogAddr.Pk[:]) {
		return false, false, basemeta.NewMetadataTxError(basemeta.PortalSignatureRequestParamError, errors.New("Requester incognito address is not signer"))
	}

	// check tx type
	if txr.GetType() != common.TxNormalType {
		return false, false, errors.New("tx custodian deposit must be TxNormalType")
	}

	if chainRetriever.GetBCHeightBreakPointPortalV3() > beaconHeight {
		return false, false, basemeta.NewMetadataTxError(basemeta.PortalSignatureRequestParamError, errors.New("Beacon height not reached to break point yet"))
	}

	// validate tokenID in request
	if !chainRetriever.IsPortalToken(beaconHeight, sigRq.TokenID) {
		return false, false, basemeta.NewMetadataTxError(basemeta.PortalSubmitProofParamError, errors.New("TokenID is not supported currently on Portal"))
	}
	if !chainRetriever.IsMultiSigSupported(beaconHeight, sigRq.TokenID) {
		return false, false, basemeta.NewMetadataTxError(basemeta.PortalSubmitProofParamError, errors.New("TokenID is not supported multisig currently on Portal"))
	}

	// validate remote addresses
	isValid, err := ValidatePortalRemoteAddresses(map[string]string{sigRq.TokenID: sigRq.RemoteAddress}, chainRetriever, beaconHeight)
	if !isValid || err != nil {
		return false, false, err
	}

	return true, true, nil
}

func (sigRq PortalSignatureRequest) ValidateMetadataByItself() bool {
	return sigRq.Type == basemeta.PortalRequestBeaconSignature
}

func (sigRq PortalSignatureRequest) Hash() *common.Hash {
	record := sigRq.MetadataBase.Hash().String()
	record += sigRq.SubmitProofTxID
	record += sigRq.IncogAddressStr
	record += sigRq.RemoteAddress
	record += sigRq.TokenID
	record += strconv.FormatUint(uint64(sigRq.TxFee), 10)
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (sigRq *PortalSignatureRequest) BuildReqActions(tx basemeta.Transaction, chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	actionContent := PortalSignatureRequestAction{
		Meta:    *sigRq,
		TxReqID: *tx.Hash(),
		ShardID: shardID,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(basemeta.PortalRequestBeaconSignature), actionContentBase64Str}
	return [][]string{action}, nil
}

func (sigRq *PortalSignatureRequest) CalculateSize() uint64 {
	return basemeta.CalculateSize(sigRq)
}
