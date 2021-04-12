package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	portalcommonv4 "github.com/incognitochain/incognito-chain/portal/portalv4/common"
	"github.com/incognitochain/incognito-chain/wallet"
)

type PortalReplacementFeeRequest struct {
	MetadataBase
	TokenID string
	BatchID string
	Fee     uint
}

type PortalReplacementFeeRequestAction struct {
	Meta    PortalReplacementFeeRequest
	TxReqID common.Hash
	ShardID byte
}

type PortalReplacementFeeRequestContent struct {
	TokenID       string
	BatchID       string
	Fee           uint
	ExternalRawTx string
	UTXOs         []*statedb.UTXO
	TxReqID       common.Hash
	ShardID       byte
}

type PortalReplacementFeeRequestStatus struct {
	TokenID       string
	BatchID       string
	Fee           uint
	TxHash        string
	ExternalRawTx string
	BeaconHeight  uint64
	Status        int
}

func NewPortalReplacementFeeRequestStatus(tokenID, batchID string, fee uint, externalRawTx string, status int) *PortalReplacementFeeRequestStatus {
	return &PortalReplacementFeeRequestStatus{
		TokenID:       tokenID,
		BatchID:       batchID,
		Fee:           fee,
		ExternalRawTx: externalRawTx,
		Status:        status,
	}
}

func NewPortalReplacementFeeRequest(metaType int, tokenID, batchID string, fee uint) (*PortalReplacementFeeRequest, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}

	portalUnshieldReq := &PortalReplacementFeeRequest{
		TokenID: tokenID,
		BatchID: batchID,
		Fee:     fee,
	}

	portalUnshieldReq.MetadataBase = metadataBase

	return portalUnshieldReq, nil
}

func (repl PortalReplacementFeeRequest) ValidateTxWithBlockChain(
	txr Transaction,
	chainRetriever ChainRetriever,
	shardViewRetriever ShardViewRetriever,
	beaconViewRetriever BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	return true, nil
}

func (repl PortalReplacementFeeRequest) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	// validate IncAddressStr
	keyWallet, err := wallet.Base58CheckDeserialize(chainRetriever.GetPortalReplacementAddress(beaconHeight))
	if err != nil {
		return false, false, NewMetadataTxError(PortalV4FeeReplacementRequestMetaError, errors.New("Requester incognito address is invalid"))
	}
	incAddr := keyWallet.KeySet.PaymentAddress
	if len(incAddr.Pk) == 0 {
		return false, false, NewMetadataTxError(PortalV4FeeReplacementRequestMetaError, errors.New("Requester incognito address is invalid"))
	}

	if !bytes.Equal(tx.GetSigPubKey()[:], incAddr.Pk[:]) {
		return false, false, NewMetadataTxError(PortalV4FeeReplacementRequestMetaError, errors.New("Requester incognito address is not replace fee feeder"))
	}

	// check tx type
	if tx.GetType() != common.TxNormalType {
		return false, false, NewMetadataTxError(PortalV4FeeReplacementRequestMetaError, errors.New("tx replace transaction must be TxNormalType"))
	}

	// validate tokenID
	isPortalToken, err := chainRetriever.IsPortalToken(beaconHeight, repl.TokenID, common.PortalVersion4)
	if !isPortalToken || err != nil {
		return false, false, errors.New("TokenID is not supported currently on Portal v4")
	}

	// validate amount of pToken is divisible by the decimal difference between nano pToken and nano Token
	if repl.TokenID == portalcommonv4.PortalBTCIDStr && repl.Fee%10 != 0 {
		return false, false, errors.New("pBTC amount has to be divisible by 10")
	}

	if repl.BatchID == "" || repl.Fee < 1 {
		return false, false, errors.New("BatchID or Fee is invalid")
	}

	return true, true, nil
}

func (repl PortalReplacementFeeRequest) ValidateMetadataByItself() bool {
	return repl.Type == PortalV4FeeReplacementRequestMeta
}

func (repl PortalReplacementFeeRequest) Hash() *common.Hash {
	record := repl.MetadataBase.Hash().String()
	record += repl.TokenID
	record += repl.BatchID
	record += strconv.FormatUint(uint64(repl.Fee), 10)

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (repl *PortalReplacementFeeRequest) BuildReqActions(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	actionContent := PortalReplacementFeeRequestAction{
		Meta:    *repl,
		TxReqID: *tx.Hash(),
		ShardID: shardID,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(PortalV4FeeReplacementRequestMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (repl *PortalReplacementFeeRequest) CalculateSize() uint64 {
	return calculateSize(repl)
}
