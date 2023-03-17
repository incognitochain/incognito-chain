package bridgehub

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

// whoever can send this type of tx
type RegisterBridgeRequest struct {
	ExtChainID       string   `json:"ExtChainID"`
	BridgePoolPubKey string   `json:"BridgePoolPubKey"` // TSS pubkey
	ValidatorPubKeys []string `json:"ValidatorPubKeys"` // pubkey to build TSS key
	VaultAddress     string   `json:"VaultAddress"`     // vault to receive external assets
	//Signature        string   `json:"Signature"`        // TSS sig
	metadataCommon.MetadataBase
}

type RegisterBridgeContentInst struct {
	ExtChainID       string   `json:"ExtChainID"`
	BridgePoolPubKey string   `json:"BridgePoolPubKey"` // TSS pubkey
	ValidatorPubKeys []string `json:"ValidatorPubKeys"` // pubkey to build TSS key
	VaultAddress     string   `json:"VaultAddress"`     // vault to receive external assets
	Signature        string   `json:"Signature"`        // TSS sig : TODO: 0xkraken: keep or remove?
	BridgeID         string   `json:"BridgeID,omitempty"`
	TxReqID          string   `json:"TxReqID"`
}

type RegisterBridgeMsg struct {
	ExtChainID       string   `json:"ExtChainID"`
	BridgePoolPubKey string   `json:"BridgePoolPubKey"`
	ValidatorPubKeys []string `json:"ValidatorPubKeys"` // pubkey to build TSS key
	VaultAddress     string   `json:"VaultAddress"`
}

func NewRegisterBridgeRequest(
	extChainID string,
	bridgePoolPubKey string, // TSS pubkey
	validatorPubKeys []string, // pubkey to build TSS key
	vaultAddress string, // vault to receive external assets
	//signature string, // TSS sig
) (*RegisterBridgeRequest, error) {
	metadataBase := metadataCommon.MetadataBase{
		Type: metadataCommon.BridgeHubRegisterBridgeMeta,
	}
	registerReq := &RegisterBridgeRequest{
		ExtChainID:       extChainID,
		BridgePoolPubKey: bridgePoolPubKey,
		ValidatorPubKeys: validatorPubKeys,
		VaultAddress:     vaultAddress,
		//Signature:        signature,
	}
	registerReq.MetadataBase = metadataBase
	return registerReq, nil
}

func (bReq RegisterBridgeRequest) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

func (bReq RegisterBridgeRequest) VerifySignature() (bool, error) {
	// validate TSS signature
	/*msg := RegisterBridgeMsg{*/
	/*ExtChainID:       bReq.ExtChainID,*/
	/*BridgePoolPubKey: bReq.BridgePoolPubKey,*/
	/*ValidatorPubKeys: bReq.ValidatorPubKeys,*/
	/*VaultAddress:     bReq.VaultAddress,*/
	/*}*/

	/*msgBytes, err := json.Marshal(msg)*/
	/*if err != nil {*/
	/*return false, fmt.Errorf("Marshal msg error: %v", err)*/
	/*}*/

	// TODO: review hash func on Bridge Validator Network
	//h := common.HashH(msgBytes)
	//return tss.VerifyTSSSig(bReq.BridgePoolPubKey, h.String(), bReq.Signature)
	return true, nil
}

func (bReq RegisterBridgeRequest) ValidateSanityData(
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	beaconHeight uint64,
	tx metadataCommon.Transaction,
) (bool, bool, error) {
	// check tx type
	if tx.GetType() != common.TxNormalType {
		return false, false, errors.New("Tx type must be n")
	}

	// check trigger feature or not
	// todo: 0xcryptolover disable to test
	//if shardViewRetriever.GetTriggeredFeature()[metadataCommon.BridgeHubFeatureName] == 0 {
	//	return false, false, fmt.Errorf("Bridge Hub Feature has not been enabled yet %v", bReq.Type)
	//}

	// vanity data
	// TODO 0xkraken: validate format address string
	if bReq.ExtChainID == "" {
		return false, false, errors.New("ExtChainID empty")
	}
	if bReq.BridgePoolPubKey == "" {
		return false, false, errors.New("BridgePoolPubKey empty")
	}
	if bReq.VaultAddress == "" {
		return false, false, errors.New("VaultAddress empty")
	}
	/*if bReq.Signature == "" {*/
	/*return false, false, errors.New("Signature empty")*/
	/*}*/
	if len(bReq.ValidatorPubKeys) == 0 {
		return false, false, errors.New("ValidatorPubKeys empty")
	}

	// validate sig
	isValidSig, err := bReq.VerifySignature()
	if err != nil || !isValidSig {
		return false, false, fmt.Errorf("Invalid Tss signature: %v", err)
	}

	return true, true, nil
}

func (bReq RegisterBridgeRequest) ValidateMetadataByItself() bool {
	return bReq.Type == metadataCommon.BridgeAggAddTokenMeta
}

func (bReq RegisterBridgeRequest) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(&bReq)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (bReq *RegisterBridgeRequest) BuildReqActions(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	actionContent := map[string]interface{}{
		"meta":          *bReq,
		"RequestedTxID": tx.Hash(),
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(bReq.Type), actionContentBase64Str}
	return [][]string{action}, nil
}

func (bReq *RegisterBridgeRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(bReq)
}
