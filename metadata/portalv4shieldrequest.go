package metadata

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	btcrelaying "github.com/incognitochain/incognito-chain/relaying/btc"
	"github.com/incognitochain/incognito-chain/wallet"
)

// PortalShieldingRequest - portal user requests ptoken (after sending pubToken to multisig wallet)
// metadata - portal user sends shielding request - create normal tx with this metadata
type PortalShieldingRequest struct {
	MetadataBase
	TokenID         string // pTokenID in incognito chain
	IncogAddressStr string
	ShieldingProof  string
}

// PortalShieldingRequestAction - shard validator creates instruction that contain this action content
type PortalShieldingRequestAction struct {
	Meta    PortalShieldingRequest
	TxReqID common.Hash
	ShardID byte
}

// PortalShieldingRequestContent - Beacon builds a new instruction with this content after receiving a instruction from shard
// It will be appended to beaconBlock
// both accepted and rejected status
type PortalShieldingRequestContent struct {
	TokenID         string // pTokenID in incognito chain
	IncogAddressStr string
	ProofHash       string
	ShieldingUTXO   []*statedb.UTXO
	MintingAmount   uint64
	TxReqID         common.Hash
	ExternalTxID    string
	ShardID         byte
}

// PortalRequestPTokensStatus - Beacon tracks status of request ptokens into db
type PortalShieldingRequestStatus struct {
	Status          byte
	Error           string
	TokenID         string // pTokenID in incognito chain
	IncogAddressStr string
	ProofHash       string
	ShieldingUTXO   []*statedb.UTXO
	MintingAmount   uint64
	TxReqID         common.Hash
	ExternalTxID    string
}

func NewPortalShieldingRequest(
	metaType int,
	tokenID string,
	incogAddressStr string,
	shieldingProof string) (*PortalShieldingRequest, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	shieldingRequestMeta := &PortalShieldingRequest{
		TokenID:         tokenID,
		IncogAddressStr: incogAddressStr,
		ShieldingProof:  shieldingProof,
	}
	shieldingRequestMeta.MetadataBase = metadataBase
	return shieldingRequestMeta, nil
}

func (shieldingReq PortalShieldingRequest) ValidateTxWithBlockChain(
	txr Transaction,
	chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	return true, nil
}

func (shieldingReq PortalShieldingRequest) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	// validate IncogAddressStr
	keyWallet, err := wallet.Base58CheckDeserialize(shieldingReq.IncogAddressStr)
	if err != nil {
		return false, false, NewMetadataTxError(metadataCommon.PortalV4ShieldRequestValidateSanityDataError, errors.New("Requester incognito address is invalid"))
	}
	incogAddr := keyWallet.KeySet.PaymentAddress
	if _, err := AssertPaymentAddressAndTxVersion(incogAddr, tx.GetVersion()); err != nil {
		return false, false, NewMetadataTxError(metadataCommon.PortalV4ShieldRequestValidateSanityDataError, errors.New("Requester incognito address is invalid"))
	}

	// check proof is not empty
	if shieldingReq.ShieldingProof == "" {
		return false, false, NewMetadataTxError(metadataCommon.PortalV4ShieldRequestValidateSanityDataError, errors.New("Shielding proof is empty"))
	}

	// check tx version and type
	if tx.GetVersion() != 2 {
		return false, false, NewMetadataTxError(metadataCommon.PortalV4ShieldRequestValidateSanityDataError, errors.New("Tx shielding request must be version 2"))
	}
	if tx.GetType() != common.TxNormalType {
		return false, false, NewMetadataTxError(metadataCommon.PortalV4ShieldRequestValidateSanityDataError, errors.New("Tx shielding request must be TxNormalType"))
	}

	// validate tokenID and shielding proof
	isPortalToken, err := chainRetriever.IsPortalToken(beaconHeight, shieldingReq.TokenID, common.PortalVersion4)
	if !isPortalToken || err != nil {
		return false, false, NewMetadataTxError(metadataCommon.PortalV4ShieldRequestValidateSanityDataError, errors.New("TokenID is not supported currently on Portal v4"))
	}

	_, err = btcrelaying.ParseAndValidateSanityBTCProofFromB64EncodeStr(shieldingReq.ShieldingProof)
	if err != nil {
		return false, false, NewMetadataTxError(metadataCommon.PortalV4ShieldRequestValidateSanityDataError,
			fmt.Errorf("ShieldingProof is invalid sanity %v", err))
	}

	return true, true, nil
}

func (shieldingReq PortalShieldingRequest) ValidateMetadataByItself() bool {
	return shieldingReq.Type == metadataCommon.PortalV4ShieldingRequestMeta
}

func (shieldingReq PortalShieldingRequest) Hash() *common.Hash {
	record := shieldingReq.MetadataBase.Hash().String()
	record += shieldingReq.TokenID
	record += shieldingReq.IncogAddressStr
	record += shieldingReq.ShieldingProof
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (shieldingReq *PortalShieldingRequest) BuildReqActions(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	actionContent := PortalShieldingRequestAction{
		Meta:    *shieldingReq,
		TxReqID: *tx.Hash(),
		ShardID: shardID,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(metadataCommon.PortalV4ShieldingRequestMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (shieldingReq *PortalShieldingRequest) CalculateSize() uint64 {
	return calculateSize(shieldingReq)
}

func (shieldingReq PortalShieldingRequest) ToCompactBytes() ([]byte, error) {
	return metadataCommon.ToCompactBytes(&shieldingReq)
}
