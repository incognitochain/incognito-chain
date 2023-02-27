package metadata

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	btcrelaying "github.com/incognitochain/incognito-chain/relaying/btc"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

// @@NOTE: This tx is created when migration centralized bridge to portal v4
// PortalConvertVaultRequest
// metadata - create normal tx with this metadata
type PortalConvertVaultRequest struct {
	MetadataBase
	TokenID      string // pTokenID in incognito chain
	ConvertProof string
}

// PortalConvertVaultRequestAction - shard validator creates instruction that contain this action content
type PortalConvertVaultRequestAction struct {
	Meta    PortalConvertVaultRequest
	TxReqID common.Hash
	ShardID byte
}

// PortalConvertVaultRequestContent - Beacon builds a new instruction with this content after receiving a instruction from shard
// It will be appended to beaconBlock
// both accepted and rejected status
type PortalConvertVaultRequestContent struct {
	TokenID          string // pTokenID in incognito chain
	ConvertProofHash string
	ConvertingUTXO   []*statedb.UTXO
	ConvertingAmount uint64
	TxReqID          common.Hash
	ShardID          byte
}

// PortalConvertVaultRequestStatus - Beacon tracks status of request converting vault into db
type PortalConvertVaultRequestStatus struct {
	Status           byte
	TokenID          string // pTokenID in incognito chain
	ConvertProofHash string
	ConvertingUTXO   []*statedb.UTXO
	ConvertingAmount uint64
	TxReqID          common.Hash
	ErrorMsg         string
}

func NewPortalConvertVaultRequest(
	metaType int,
	tokenID string,
	convertingProof string) (*PortalConvertVaultRequest, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	convertRequestMeta := &PortalConvertVaultRequest{
		TokenID:      tokenID,
		ConvertProof: convertingProof,
	}
	convertRequestMeta.MetadataBase = metadataBase
	return convertRequestMeta, nil
}

func (convertVaultReq PortalConvertVaultRequest) ValidateTxWithBlockChain(
	txr Transaction,
	chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	return true, nil
}

func (convertVaultReq PortalConvertVaultRequest) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	// check proof is not empty
	if convertVaultReq.ConvertProof == "" {
		return false, false, NewMetadataTxError(metadataCommon.PortalV4ConvertVaultRequestMetaError,
			errors.New("Converting proof is empty"))
	}

	// check tx version and type
	if tx.GetVersion() != 2 {
		return false, false, NewMetadataTxError(metadataCommon.PortalV4ConvertVaultRequestMetaError,
			errors.New("Tx converting vault request must be version 2"))
	}
	if tx.GetType() != common.TxNormalType {
		return false, false, NewMetadataTxError(metadataCommon.PortalV4ConvertVaultRequestMetaError,
			errors.New("Tx converting vault request must be TxNormalType"))
	}

	// validate tokenID and shielding proof
	isPortalToken, err := chainRetriever.IsPortalToken(beaconHeight, convertVaultReq.TokenID, common.PortalVersion4)
	if !isPortalToken || err != nil {
		return false, false, NewMetadataTxError(metadataCommon.PortalV4ConvertVaultRequestMetaError,
			errors.New("TokenID is not supported currently on Portal v4"))
	}

	_, err = btcrelaying.ParseAndValidateSanityBTCProofFromB64EncodeStr(convertVaultReq.ConvertProof)
	if err != nil {
		return false, false, NewMetadataTxError(metadataCommon.PortalV4ConvertVaultRequestMetaError,
			fmt.Errorf("ConvertProof is invalid sanity %v", err))
	}

	return true, true, nil
}

func (convertVaultReq PortalConvertVaultRequest) ValidateMetadataByItself() bool {
	return convertVaultReq.Type == metadataCommon.PortalV4ConvertVaultRequestMeta
}

func (convertVaultReq PortalConvertVaultRequest) Hash() *common.Hash {
	record := convertVaultReq.MetadataBase.Hash().String()
	record += convertVaultReq.TokenID
	record += convertVaultReq.ConvertProof

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (convertVaultReq *PortalConvertVaultRequest) BuildReqActions(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	actionContent := PortalConvertVaultRequestAction{
		Meta:    *convertVaultReq,
		TxReqID: *tx.Hash(),
		ShardID: shardID,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(metadataCommon.PortalV4ConvertVaultRequestMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (convertVaultReq *PortalConvertVaultRequest) CalculateSize() uint64 {
	return calculateSize(convertVaultReq)
}

func (convertVaultReq *PortalConvertVaultRequest) ToCompactBytes() ([]byte, error) {
	return toCompactBytes(convertVaultReq)
}
