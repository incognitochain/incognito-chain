package bridgehub

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v1/schnorr"
	"strconv"
)

// ShieldingBTCRequest represents an BTC shielding request. Users create transactions with this metadata after
// sending public tokens to the corresponding validators. There are two ways to use this metadata,
// depending on which data has been enclosed with the depositing transaction:
//   - payment address: Receiver and Signature must be empty;
//   - using one-time depositing public key: Receiver must be an OTAReceiver, a signature is required.
type ShieldingBTCRequest struct {
	// Amount to shield
	Amount uint64 `json:"amount"`

	// BTCTxID btc transaction id send to associated wallet of validators
	BTCTxID common.Hash `json:"BTCTxID"`

	// IncTokenID is the Incognito tokenID of the shielding token.
	IncTokenID common.Hash

	// ExtChainID to distinguish between bridge hubs
	ExtChainID string `json:"ExtChainID"`

	// Signature is the signature for validating the authenticity of the request. This signature is different from a
	// MetadataBaseWithSignature type since it is signed with the tx privateKey.
	Signature []byte `json:"Signature"`

	// Receiver is the recipient of this shielding request. It is an OTAReceiver if OTDepositPubKey is not empty.
	Receiver string `json:"Receiver"`

	metadataCommon.MetadataBase
}

type ShieldingBTCReqAction struct {
	Meta    ShieldingBTCRequest `json:"meta"`
	TxReqID common.Hash         `json:"txReqId"`
}

type ShieldingBTCAcceptedInst struct {
	ShardID         byte        `json:"shardId"`
	IssuingAmount   uint64      `json:"issuingAmount"`
	Receiver        string      `json:"receiverAddrStr"`
	OTDepositKey    []byte      `json:"OTDepositKey,omitempty"`
	IncTokenID      common.Hash `json:"incTokenId"`
	TxReqID         common.Hash `json:"txReqId"`
	UniqTx          []byte      `json:"uniqBTCTx"`
	ExternalTokenID []byte      `json:"externalTokenId"`
}

func NewShieldingBTCRequest(
	amount uint64,
	btcTx common.Hash,
	incTokenID common.Hash,
	receiver string,
	signature []byte,
	extChainID string,
) (*ShieldingBTCRequest, error) {
	metadataBase := metadataCommon.MetadataBase{
		Type: metadataCommon.ShieldingBTCRequestMeta,
	}
	ShieldingBTCReq := &ShieldingBTCRequest{
		Amount:     amount,
		BTCTxID:    btcTx,
		IncTokenID: incTokenID,
		Receiver:   receiver,
		Signature:  signature,
		ExtChainID: extChainID,
	}
	ShieldingBTCReq.MetadataBase = metadataBase
	return ShieldingBTCReq, nil
}

func NewShieldingBTCRequestFromMap(
	data map[string]interface{},
	metaType int,
) (*ShieldingBTCRequest, error) {
	amount := data["amount"].(uint64)
	if amount == 0 {
		return nil, errors.New("BTCHub: not thing to shield")
	}
	btcTx, err := common.Hash{}.NewHashFromStr(data["btcTx"].(string))
	if err != nil {
		return nil, errors.New("BTCHub: invalid btc transaction id")
	}
	extBridgeId, ok := data["extBridgeId"].(string)
	if !ok {
		return nil, errors.New("BTCHub: invalid ext bridge id")
	}

	incTokenID, err := common.Hash{}.NewHashFromStr(data["IncTokenID"].(string))
	if err != nil {
		return nil, errors.New("BTCHub: Token incorrect")
	}

	var sig []byte
	tmpSig, ok := data["Signature"]
	if ok {
		sigStr, ok := tmpSig.(string)
		if ok {
			sig, _, err = base58.Base58Check{}.Decode(sigStr)
			if err != nil {
				return nil, errors.New("BTCHub: invalid base58-encoded signature")
			}
		}
	}

	tmpReceiver, ok := data["Receiver"]
	var receiver string
	if ok {
		receiver, _ = tmpReceiver.(string)
	}

	if _, ok := data["MetadataType"]; ok {
		tmpMdType, ok := data["MetadataType"].(float64)
		if ok {
			metaType = int(tmpMdType)
		}
	}

	req, _ := NewShieldingBTCRequest(
		amount,
		*btcTx,
		*incTokenID,
		receiver,
		sig,
		extBridgeId,
	)
	return req, nil
}

func ParseBTCIssuingInstContent(instContentStr string) (*ShieldingBTCReqAction, error) {
	contentBytes, err := base64.StdEncoding.DecodeString(instContentStr)
	if err != nil {
		return nil, metadataCommon.NewMetadataTxError(metadataCommon.IssuingBtcHubRequestDecodeInstructionError, err)
	}
	var issuingBTCHubReqAction ShieldingBTCReqAction
	err = json.Unmarshal(contentBytes, &issuingBTCHubReqAction)
	if err != nil {
		return nil, metadataCommon.NewMetadataTxError(metadataCommon.IssuingEvmRequestUnmarshalJsonError, err)
	}
	return &issuingBTCHubReqAction, nil
}

func (iReq ShieldingBTCRequest) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

func (iReq ShieldingBTCRequest) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {
	// check trigger feature or not
	if shardViewRetriever.GetTriggeredFeature()[metadataCommon.BridgeHubFeatureName] == 0 {
		return false, false, fmt.Errorf("Bridge Hub Feature has not been enabled yet %v", iReq.Type)
	}
	var err error
	// todo: update
	if iReq.IncTokenID.String() != "BTC_ID" {
		return false, false, fmt.Errorf("BTCHub: invalid token id")
	}

	// todo: add more validations

	if iReq.Receiver != "" {
		otaReceiver := new(privacy.OTAReceiver)
		err = otaReceiver.FromString(iReq.Receiver)
		if err != nil {
			return false, false, fmt.Errorf("BTCHub: invalid OTAReceiver")
		}
		if !otaReceiver.IsValid() {
			return false, false, fmt.Errorf("BTCHub: invalid OTAReceiver")
		}
	}

	if iReq.Signature != nil {
		schnorrSig := new(schnorr.SchnSignature)
		err = schnorrSig.SetBytes(iReq.Signature)
		if err != nil {
			return false, false, fmt.Errorf("BTCHub: invalid signature %v", iReq.Signature)
		}
	}

	return true, true, nil
}

func (iReq ShieldingBTCRequest) ValidateMetadataByItself() bool {
	if iReq.Type != metadataCommon.ShieldingBTCRequestMeta {
		return false
	}
	return true
}

func (iReq ShieldingBTCRequest) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(iReq)
	hash := common.HashH(rawBytes)
	return &hash
}

func (iReq *ShieldingBTCRequest) BuildReqActions(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	actionContent := map[string]interface{}{
		"meta":          *iReq,
		"RequestedTxID": tx.Hash(),
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(iReq.Type), actionContentBase64Str}
	return [][]string{action}, nil
}

func (iReq *ShieldingBTCRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(iReq)
}

func (iReq *ShieldingBTCRequest) GetOTADeclarations() []metadataCommon.OTADeclaration {
	var result []metadataCommon.OTADeclaration
	currentTokenID := common.ConfidentialAssetID
	if iReq.IncTokenID.String() == common.PRVIDStr {
		currentTokenID = common.PRVCoinID
	}
	otaReceiver := privacy.OTAReceiver{}
	otaReceiver.FromString(iReq.Receiver)
	result = append(result, metadataCommon.OTADeclaration{
		PublicKey: otaReceiver.PublicKey.ToBytes(), TokenID: currentTokenID,
	})
	return result
}
