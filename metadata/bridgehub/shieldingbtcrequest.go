package bridgehub

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
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
	ExtChainID int `json:"ExtChainID"`

	// BridgePoolPubKey to verify signature and update state
	BridgePoolPubKey string `json:"BridgePoolPubKey"`

	// Signature is the signature for validating the authenticity of the request. This signature is different from a
	// MetadataBaseWithSignature type since it is signed with the tx privateKey.
	Signature string `json:"Signature"`

	// Receiver is the recipient of this shielding request. It is an OTAReceiver if OTDepositPubKey is not empty.
	Receiver privacy.OTAReceiver `json:"Receiver"`

	metadataCommon.MetadataBase
}

type ShieldingBTCReqAction struct {
	Meta    ShieldingBTCRequest `json:"meta"`
	TxReqID common.Hash         `json:"txReqId"`
}

type ShieldingBTCAcceptedInst struct {
	ShardID          byte                `json:"shardId"`
	IssuingAmount    uint64              `json:"issuingAmount"`
	Receiver         privacy.OTAReceiver `json:"receiverAddrStr"`
	IncTokenID       common.Hash         `json:"incTokenId"`
	TxReqID          common.Hash         `json:"txReqId"`
	UniqTx           []byte              `json:"uniqTx"`
	BridgePoolPubKey string              `json:"BridgePoolPubKey"`
	ExtChainID       int                 `json:"ExtChainID"`
}

func NewShieldingBTCRequest(
	amount uint64,
	btcTx common.Hash,
	incTokenID common.Hash,
	receiver privacy.OTAReceiver,
	signature string,
	extChainID int,
	bridgePoolPubKey string,
) (*ShieldingBTCRequest, error) {
	metadataBase := metadataCommon.MetadataBase{
		Type: metadataCommon.ShieldingBTCRequestMeta,
	}
	ShieldingBTCReq := &ShieldingBTCRequest{
		Amount:           amount,
		BTCTxID:          btcTx,
		IncTokenID:       incTokenID,
		Receiver:         receiver,
		Signature:        signature,
		ExtChainID:       extChainID,
		BridgePoolPubKey: bridgePoolPubKey,
	}
	ShieldingBTCReq.MetadataBase = metadataBase
	return ShieldingBTCReq, nil
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
	// todo: 0xcryptolover disable to test
	//if shardViewRetriever.GetTriggeredFeature()[metadataCommon.BridgeHubFeatureName] == 0 {
	//	return false, false, fmt.Errorf("Bridge Hub Feature has not been enabled yet %v", iReq.Type)
	//}
	var err error
	// todo: update btc id
	if iReq.IncTokenID.String() != "0000000000000000000000000000000000000000000000000000000000000010" {
		return false, false, fmt.Errorf("BTCHub: invalid token id")
	}

	if iReq.Amount == 0 {
		return false, false, fmt.Errorf("BTCHub: invalid shielding amount")
	}

	// todo: add more validations
	if iReq.Receiver.GetShardID() != byte(tx.GetValidationEnv().ShardID()) {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, fmt.Errorf("otaReceiver shardID is different from txShardID"))
	}

	// todo: 0xCryptoLover check field format
	sigInBytes, err := base64.StdEncoding.DecodeString(iReq.Signature)
	if err != nil {
		return false, false, fmt.Errorf("BTCHub: can not decode signature %v with error %v", iReq.Signature, err.Error())
	}

	schnorrSig := new(schnorr.SchnSignature)
	err = schnorrSig.SetBytes(sigInBytes)
	if err != nil {
		return false, false, fmt.Errorf("BTCHub: invalid signature %v", iReq.Signature)
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
		"meta":    *iReq,
		"TxReqID": tx.Hash(),
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
	result = append(result, metadataCommon.OTADeclaration{
		PublicKey: iReq.Receiver.PublicKey.ToBytes(), TokenID: common.ConfidentialAssetID,
	})
	return result
}
