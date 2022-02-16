package metadata

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v1/schnorr"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	btcrelaying "github.com/incognitochain/incognito-chain/relaying/btc"
	"github.com/incognitochain/incognito-chain/wallet"
)

// PortalV4ShieldingRequest represents a shielding request of Portal V4. Users create transactions with this metadata after
// sending public tokens to multi-sig wallets. There are two ways to use this metadata, depending on how the corresponding
// multi-sig wallet (a.k.a. depositing address) is generated:
// 	- using payment address: Receiver must be a payment address, OTDepositPubKey, Signature must be empty and the corresponding
//	deposit address must be built with Receiver as the chain-code;
//	- using one-time depositing public key: Receiver must be an OTAReceiver, OTDepositPubKey must not be empty,
// 	a signature is required and the corresponding deposit address must be built with OTDepositPubKey as the chain-code.
type PortalV4ShieldingRequest struct {
	MetadataBase

	// TokenID is the Incognito tokenID of the shielding token.
	TokenID string // pTokenID in incognito chain

	// OTDepositPubKey is the base58-encoded public key for this shielding request, used to validate the authenticity of the request.
	// This field is only used with one-time depositing addresses.
	// If set to empty, Receiver must be a payment address. Otherwise, Receiver must be an OTAReceiver.
	OTDepositPubKey string `json:"OTDepositPubKey,omitempty"`

	// Signature is the signature for validating the authenticity of the request. This signature is different from a
	// MetadataBaseWithSignature type since it is signed with the tx privateKey.
	Signature []byte `json:"Signature,omitempty"`

	// Receiver is the recipient of this shielding request.
	// Receiver is
	//	- an Incognito payment address if OTDepositPubKey is empty;
	//	- an OTAReceiver if OTDepositPubKey is not empty.
	Receiver string `json:"IncogAddressStr"` // the json-tag is required for backward-compatibility.

	// ShieldingProof is the generated proof for this shielding request.
	ShieldingProof string
}

// PortalShieldingRequestAction - shard validator creates instruction that contain this action content
type PortalShieldingRequestAction struct {
	Meta    PortalV4ShieldingRequest
	TxReqID common.Hash
	ShardID byte
}

// PortalShieldingRequestContent represents a beacon instruction (either accepted or rejected) for a PortalV4ShieldingRequest.
type PortalShieldingRequestContent struct {
	// TokenID is the Incognito ID of the shielding token.
	TokenID string

	// OTDepositPubKey is the same as in the shielding request.
	OTDepositPubKey string `json:"OTDepositPubKey,omitempty"`

	// Receiver is the same as in the shielding request.
	Receiver string `json:"IncogAddressStr"` // the json-tag is required for backward-compatibility.

	// ProofHash is the hash of the shielding proof.
	ProofHash string

	// ShieldingUTXO is the list of public UTXOs sent to the corresponding shielding address.
	ShieldingUTXO []*statedb.UTXO

	// MintingAmount is the shielding amount.
	MintingAmount uint64

	// TxReqID is the request for this instruction.
	TxReqID common.Hash

	// ExternalTxID is the ID of the corresponding public transaction.
	ExternalTxID string

	// ShardID is the shard where this instruction resides in.
	ShardID byte
}

// PortalShieldingRequestStatus is used for beacon to track the status of a shielding request.
type PortalShieldingRequestStatus struct {
	Status          byte
	Error           string
	TokenID         string
	OTDepositPubKey string `json:"OTDepositPubKey,omitempty"`
	Receiver        string `json:"IncogAddressStr"` // the json-tag is required for backward-compatibility.
	ProofHash       string
	ShieldingUTXO   []*statedb.UTXO
	MintingAmount   uint64
	TxReqID         common.Hash
	ExternalTxID    string
}

// NewPortalShieldingRequest creates a new PortalV4ShieldingRequest based on given data.
// If depositPubKey is not nil or empty, it will create a request with a signature.
func NewPortalShieldingRequest(
	metaType int,
	tokenID string,
	receiver string,
	shieldingProof string,
	depositPubKey string,
	signature []byte) (*PortalV4ShieldingRequest, error) {
	shieldingRequestMeta := &PortalV4ShieldingRequest{
		TokenID:        tokenID,
		Receiver:       receiver,
		ShieldingProof: shieldingProof,
		MetadataBase:   MetadataBase{Type: metaType},
	}

	if len(depositPubKey) != 0 {
		shieldingRequestMeta.Signature = signature
		shieldingRequestMeta.OTDepositPubKey = depositPubKey
	}

	return shieldingRequestMeta, nil
}

func (req PortalV4ShieldingRequest) ValidateTxWithBlockChain(
	txr Transaction,
	chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	return true, nil
}

// ValidateSanityData proceeds with the following steps.
//	1. The Receiver of the request must be valid.
//		1.1. If Receiver is a payment address
//			- The payment address must be of the same version of the transaction.
//			- OTDepositPubKey and Signature must be empty.
//		1.2. If Receiver is an OTAReceiver
//			- OTDepositPubKey, Signature must not be empty.
//			- The Signature must be valid against OTDepositPubKey.
//	2. The transaction must be of version 2, with type `n`.
//	3. The shielding token must be a supported portal token.
//	4. The ShieldingProof must be in a good form: not empty, sanity-checked.
func (req PortalV4ShieldingRequest) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	// Step 1
	keyWallet, err := wallet.Base58CheckDeserialize(req.Receiver)
	if err == nil { // 1.1
		incAddr := keyWallet.KeySet.PaymentAddress
		if _, err = AssertPaymentAddressAndTxVersion(incAddr, tx.GetVersion()); err != nil {
			return false, false, NewMetadataTxError(metadataCommon.PortalV4ShieldRequestValidateSanityDataError, fmt.Errorf("invalid Incognito address"))
		}
		if req.OTDepositPubKey != "" || req.Signature != nil {
			return false, false, NewMetadataTxError(metadataCommon.PortalV4ShieldRequestValidateSanityDataError, fmt.Errorf("Signature and OTDepositPubKey must be empty"))
		}
	} else {
		otaReceiver := new(privacy.OTAReceiver)
		err = otaReceiver.FromString(req.Receiver)
		if err != nil {
			return false, false, NewMetadataTxError(metadataCommon.PortalV4ShieldRequestValidateSanityDataError, fmt.Errorf("invalid OTAReceiver"))
		}
		if !otaReceiver.IsValid() {
			return false, false, NewMetadataTxError(metadataCommon.PortalV4ShieldRequestValidateSanityDataError, fmt.Errorf("invalid OTAReceiver"))
		}
		otaReceiverBytes, _ := otaReceiver.Bytes()

		// 1.2
		depositPubKeyBytes, _, err := base58.Base58Check{}.Decode(req.OTDepositPubKey)
		if err != nil {
			return false, false, NewMetadataTxError(metadataCommon.PortalV4ShieldRequestValidateSanityDataError, fmt.Errorf("cannot decode OTDepositPubKey %v", req.OTDepositPubKey))
		}
		depositPubKey, err := new(operation.Point).FromBytesS(depositPubKeyBytes)
		if err != nil {
			return false, false, NewMetadataTxError(metadataCommon.PortalV4ShieldRequestValidateSanityDataError, fmt.Errorf("invalid OTDepositPubKey %v", req.OTDepositPubKey))
		}
		schnorrKey := new(privacy.SchnorrPublicKey)
		schnorrKey.Set(depositPubKey)
		schnorrSig := new(schnorr.SchnSignature)
		err = schnorrSig.SetBytes(req.Signature)
		if err != nil {
			return false, false, NewMetadataTxError(metadataCommon.PortalV4ShieldRequestValidateSanityDataError, fmt.Errorf("invalid signature %v", req.Signature))
		}

		if isValid := schnorrKey.Verify(schnorrSig, common.HashB(otaReceiverBytes)); !isValid {
			return false, false, NewMetadataTxError(metadataCommon.PortalV4ShieldRequestValidateSanityDataError, fmt.Errorf("invalid signature"))
		}
	}

	// Step 2
	if tx.GetVersion() != 2 {
		return false, false, NewMetadataTxError(metadataCommon.PortalV4ShieldRequestValidateSanityDataError, fmt.Errorf("tx must be of version 2"))
	}
	if tx.GetType() != common.TxNormalType {
		return false, false, NewMetadataTxError(metadataCommon.PortalV4ShieldRequestValidateSanityDataError, fmt.Errorf("tx must be of type `n`"))
	}

	// Step 3
	isPortalToken, err := chainRetriever.IsPortalToken(beaconHeight, req.TokenID, common.PortalVersion4)
	if !isPortalToken || err != nil {
		return false, false, NewMetadataTxError(metadataCommon.PortalV4ShieldRequestValidateSanityDataError, fmt.Errorf("tokenID not supported"))
	}

	// Step 4.
	if req.ShieldingProof == "" {
		return false, false, NewMetadataTxError(metadataCommon.PortalV4ShieldRequestValidateSanityDataError, fmt.Errorf("shieldingProof is empty"))
	}
	_, err = btcrelaying.ParseAndValidateSanityBTCProofFromB64EncodeStr(req.ShieldingProof)
	if err != nil {
		return false, false, NewMetadataTxError(metadataCommon.PortalV4ShieldRequestValidateSanityDataError,
			fmt.Errorf("shieldingProof is invalid %v", err))
	}

	return true, true, nil
}

func (req PortalV4ShieldingRequest) ValidateMetadataByItself() bool {
	return req.Type == metadataCommon.PortalV4ShieldingRequestMeta
}

func (req PortalV4ShieldingRequest) Hash() *common.Hash {
	var record string
	if req.OTDepositPubKey != "" {
		jsb, _ := json.Marshal(req)
		hash := common.HashH(jsb)
		return &hash
	}

	// old shielding request
	record = req.MetadataBase.Hash().String()
	record += req.TokenID
	record += req.Receiver
	record += req.ShieldingProof
	hash := common.HashH([]byte(record))

	return &hash
}

func (req *PortalV4ShieldingRequest) BuildReqActions(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	actionContent := PortalShieldingRequestAction{
		Meta:    *req,
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

func (req *PortalV4ShieldingRequest) CalculateSize() uint64 {
	return calculateSize(req)
}

func (req *PortalV4ShieldingRequest) GetOTADeclarations() []metadataCommon.OTADeclaration {
	var result []metadataCommon.OTADeclaration

	if req.OTDepositPubKey != "" {
		otaReceiver := privacy.OTAReceiver{}
		_ = otaReceiver.FromString(req.Receiver)
		result = append(result, metadataCommon.OTADeclaration{
			PublicKey: otaReceiver.PublicKey.ToBytes(), TokenID: common.ConfidentialAssetID,
		})
	}

	return result
}
