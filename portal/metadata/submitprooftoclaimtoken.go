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

// PortalSubmitProof - portal submit proof
// metadata - user or custodian submit proof - create normal tx with this metadata
type PortalSubmitProof struct {
	basemeta.MetadataBase
	UniqueID        string
	TokenID         string // pTokenID in incognito chain
	IncogAddressStr string
	Amount          uint64
	Proof           string
	ActionType      uint // 0: porting 1: redeem
}

// PortalSubmitProofAction - shard validator creates instruction that contain this action content
type PortalSubmitProofAction struct {
	Meta    PortalSubmitProof
	TxReqID common.Hash
	ShardID byte
}

// PortalSubmitProofContent - Beacon builds a new instruction with this content after receiving a instruction from shard
// It will be appended to beaconBlock
// both accepted and rejected status
type PortalSubmitProofContent struct {
	UniqueID        string
	TokenID         string // pTokenID in incognito chain
	IncogAddressStr string
	Amount          uint64
	Proof           string
	TxReqID         common.Hash
	ShardID         byte
	ActionType      uint
}

// PortalSubmitProofStatus - Beacon tracks status of request ptokens into db
type PortalSubmitProofStatus struct {
	Status          byte
	UniqueID        string
	TokenID         string // pTokenID in incognito chain
	IncogAddressStr string
	Amount          uint64
	Proof           string
	TxReqID         common.Hash
	ActionType      uint
}

func NewPortalSubmitProof(
	metaType int,
	uniqueID string,
	tokenID string,
	incogAddressStr string,
	Amount uint64,
	Proof string,
	actionType uint) (*PortalSubmitProof, error) {
	metadataBase := basemeta.MetadataBase{
		Type: metaType,
	}
	requestPTokenMeta := &PortalSubmitProof{
		UniqueID:        uniqueID,
		TokenID:         tokenID,
		IncogAddressStr: incogAddressStr,
		Amount:          Amount,
		Proof:           Proof,
		ActionType:      actionType,
	}
	requestPTokenMeta.MetadataBase = metadataBase
	return requestPTokenMeta, nil
}

func (submitProof PortalSubmitProof) ValidateTxWithBlockChain(
	txr basemeta.Transaction,
	chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	return true, nil
}

func (submitProof PortalSubmitProof) ValidateSanityData(chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, beaconHeight uint64, txr basemeta.Transaction) (bool, bool, error) {
	// validate IncogAddressStr
	keyWallet, err := wallet.Base58CheckDeserialize(submitProof.IncogAddressStr)
	if err != nil {
		return false, false, basemeta.NewMetadataTxError(basemeta.PortalSubmitProofParamError, errors.New("Requester incognito address is invalid"))
	}
	incogAddr := keyWallet.KeySet.PaymentAddress
	if len(incogAddr.Pk) == 0 {
		return false, false, basemeta.NewMetadataTxError(basemeta.PortalSubmitProofParamError, errors.New("Requester incognito address is invalid"))
	}
	if !bytes.Equal(txr.GetSigPubKey()[:], incogAddr.Pk[:]) {
		return false, false, basemeta.NewMetadataTxError(basemeta.PortalSubmitProofParamError, errors.New("Requester incognito address is not signer"))
	}

	// check tx type
	if txr.GetType() != common.TxNormalType {
		return false, false, errors.New("tx custodian deposit must be TxNormalType")
	}

	// validate amount deposit
	if submitProof.Amount == 0 {
		return false, false, errors.New("porting amount should be larger than 0")
	}

	// validate tokenID and porting proof
	if !chainRetriever.IsPortalToken(beaconHeight, submitProof.TokenID) {
		return false, false, basemeta.NewMetadataTxError(basemeta.PortalSubmitProofParamError, errors.New("TokenID is not supported currently on Portal"))
	}
	if !chainRetriever.IsMultiSigSupported(beaconHeight, submitProof.TokenID) {
		return false, false, basemeta.NewMetadataTxError(basemeta.PortalSubmitProofParamError, errors.New("TokenID is not supported multisig currently on Portal"))
	}
	if chainRetriever.GetBCHeightBreakPointPortalV3() > beaconHeight {
		return false, false, basemeta.NewMetadataTxError(basemeta.PortalSubmitProofParamError, errors.New("Beacon height not reached to break point yet"))
	}

	return true, true, nil
}

func (submitProof PortalSubmitProof) ValidateMetadataByItself() bool {
	return submitProof.Type == basemeta.PortalSubmitProofToClaimToken
}

func (submitProof PortalSubmitProof) Hash() *common.Hash {
	record := submitProof.MetadataBase.Hash().String()
	record += submitProof.UniqueID
	record += submitProof.TokenID
	record += submitProof.IncogAddressStr
	record += strconv.FormatUint(submitProof.Amount, 10)
	record += submitProof.Proof
	record += strconv.FormatUint(uint64(submitProof.ActionType), 10)
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (submitProof *PortalSubmitProof) BuildReqActions(tx basemeta.Transaction, chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	actionContent := PortalSubmitProofAction{
		Meta:    *submitProof,
		TxReqID: *tx.Hash(),
		ShardID: shardID,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(basemeta.PortalSubmitProofToClaimToken), actionContentBase64Str}
	return [][]string{action}, nil
}

func (submitProof *PortalSubmitProof) CalculateSize() uint64 {
	return basemeta.CalculateSize(submitProof)
}
