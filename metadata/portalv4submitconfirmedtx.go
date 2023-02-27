package metadata

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	btcrelaying "github.com/incognitochain/incognito-chain/relaying/btc"
)

type PortalSubmitConfirmedTxRequest struct {
	MetadataBase
	TokenID       string // pTokenID in incognito chain
	UnshieldProof string
	BatchID       string
}

type PortalSubmitConfirmedTxAction struct {
	Meta    PortalSubmitConfirmedTxRequest
	TxReqID common.Hash
	ShardID byte
}

type PortalSubmitConfirmedTxContent struct {
	TokenID      string
	UTXOs        []*statedb.UTXO
	BatchID      string
	TxReqID      common.Hash
	ExternalTxID string
	ExternalFee  uint64
	ShardID      byte
}

type PortalSubmitConfirmedTxStatus struct {
	TokenID      string
	UTXOs        []*statedb.UTXO
	BatchID      string
	TxHash       string
	ExternalTxID string
	ExternalFee  uint64
	Status       int
}

func NewPortalSubmitConfirmedTxStatus(tokenID, batchID, externalTxID, txID string, UTXOs []*statedb.UTXO, status int, externalFee uint64) *PortalSubmitConfirmedTxStatus {
	return &PortalSubmitConfirmedTxStatus{
		TokenID:      tokenID,
		UTXOs:        UTXOs,
		BatchID:      batchID,
		TxHash:       txID,
		ExternalTxID: externalTxID,
		ExternalFee:  externalFee,
		Status:       status,
	}
}

func NewPortalSubmitConfirmedTxRequest(metaType int, unshieldProof, tokenID, batchID string) (*PortalSubmitConfirmedTxRequest, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}

	portalUnshieldReq := &PortalSubmitConfirmedTxRequest{
		TokenID:       tokenID,
		BatchID:       batchID,
		UnshieldProof: unshieldProof,
	}

	portalUnshieldReq.MetadataBase = metadataBase

	return portalUnshieldReq, nil
}

func (r PortalSubmitConfirmedTxRequest) ValidateTxWithBlockChain(
	txr Transaction,
	chainRetriever ChainRetriever,
	shardViewRetriever ShardViewRetriever,
	beaconViewRetriever BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	return true, nil
}

func (r PortalSubmitConfirmedTxRequest) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	// check tx type and version
	if tx.GetType() != common.TxNormalType {
		return false, false, NewMetadataTxError(metadataCommon.PortalV4SubmitConfirmedTxRequestMetaError, errors.New("tx replace transaction must be TxNormalType"))
	}

	if tx.GetVersion() != 2 {
		return false, false, NewMetadataTxError(metadataCommon.PortalV4SubmitConfirmedTxRequestMetaError,
			errors.New("Tx submit confirmed tx request must be version 2"))
	}

	// validate tokenID
	isPortalToken, err := chainRetriever.IsPortalToken(beaconHeight, r.TokenID, common.PortalVersion4)
	if !isPortalToken || err != nil {
		return false, false, errors.New("TokenID is not supported currently on Portal v4")
	}

	_, err = btcrelaying.ParseAndValidateSanityBTCProofFromB64EncodeStr(r.UnshieldProof)
	if r.BatchID == "" || err != nil {
		return false, false, errors.New("BatchID or UnshieldProof is invalid")
	}

	return true, true, nil
}

func (r PortalSubmitConfirmedTxRequest) ValidateMetadataByItself() bool {
	return r.Type == metadataCommon.PortalV4SubmitConfirmedTxMeta
}

func (r PortalSubmitConfirmedTxRequest) Hash() *common.Hash {
	record := r.MetadataBase.Hash().String()
	record += r.TokenID
	record += r.BatchID
	record += r.UnshieldProof

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (r *PortalSubmitConfirmedTxRequest) BuildReqActions(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	actionContent := PortalSubmitConfirmedTxAction{
		Meta:    *r,
		TxReqID: *tx.Hash(),
		ShardID: shardID,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(metadataCommon.PortalV4SubmitConfirmedTxMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (r *PortalSubmitConfirmedTxRequest) CalculateSize() uint64 {
	return calculateSize(r)
}

func (r *PortalSubmitConfirmedTxRequest) ToCompactBytes() ([]byte, error) {
	return metadataCommon.ToCompactBytes(r)
}
