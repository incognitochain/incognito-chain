package bridgehub

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"strconv"
)

type BridgeHubUnshieldRequest struct {
	BurningAmount uint64 // must be equal to vout value
	TokenID       common.Hash
	TokenName     string // unused
	RemoteAddress string
	metadataCommon.MetadataBase
}

func NewBridgeHubUnshieldRequest(
	burningAmount uint64,
	tokenID common.Hash,
	tokenName string,
	remoteAddress string,
	metaType int,
) (*BridgeHubUnshieldRequest, error) {
	metadataBase := metadataCommon.MetadataBase{
		Type: metaType,
	}
	burningReq := &BridgeHubUnshieldRequest{
		BurningAmount: burningAmount,
		TokenID:       tokenID,
		TokenName:     tokenName,
		RemoteAddress: remoteAddress,
	}
	burningReq.MetadataBase = metadataBase
	return burningReq, nil
}

func (bReq BridgeHubUnshieldRequest) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

func (bReq BridgeHubUnshieldRequest) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {
	// check trigger feature or not
	if shardViewRetriever.GetTriggeredFeature()[metadataCommon.BridgeHubFeatureName] == 0 {
		return false, false, fmt.Errorf("Bridge Hub Feature has not been enabled yet %v", bReq.Type)
	}

	if bReq.BurningAmount == 0 {
		return false, false, fmt.Errorf("wrong request info's burned amount")
	}

	isBurned, burnCoin, burnedTokenID, err := tx.GetTxBurnData()
	if err != nil || !isBurned {
		return false, false, fmt.Errorf("it is not transaction burn. Error %v", err)
	}

	if !bytes.Equal(burnedTokenID[:], bReq.TokenID[:]) {
		return false, false, fmt.Errorf("wrong request info's token id and token burned")
	}

	burnAmount := burnCoin.GetValue()
	if burnAmount == 0 || burnAmount != bReq.BurningAmount {
		return false, false, fmt.Errorf("burn amount is incorrect %v", burnAmount)
	}

	// validate RemoteAddress for btc only
	isValidRemoteAddress, err := chainRetriever.IsValidPortalRemoteAddress(bReq.TokenID.String(), bReq.RemoteAddress, beaconHeight, common.PortalVersion4)
	if err != nil || !isValidRemoteAddress {
		return false, false, fmt.Errorf("invalid bitcoin address %v", bReq.RemoteAddress)
	}

	return true, true, nil
}

func (bReq BridgeHubUnshieldRequest) ValidateMetadataByItself() bool {
	return bReq.Type == metadataCommon.BridgeHubUnshieldRequest
}

func (bReq BridgeHubUnshieldRequest) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(bReq)
	hash := common.HashH(rawBytes)
	return &hash
}

func (bReq *BridgeHubUnshieldRequest) BuildReqActions(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
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

func (bReq *BridgeHubUnshieldRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(bReq)
}
