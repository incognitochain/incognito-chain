package bridge

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
	"strconv"
)

type BurningPRVRequest struct {
	BurnerAddress     privacy.PaymentAddress // unused
	BurningAmount     uint64                 // must be equal to vout value
	TokenID           common.Hash
	TokenName         string // unused
	RemoteAddress     string
	RedepositReceiver *privacy.OTAReceiver `json:"RedepositReceiver,omitempty"`
	metadataCommon.MetadataBase
}

func NewBurningPRVRequest(
	burnerAddress privacy.PaymentAddress,
	burningAmount uint64,
	tokenID common.Hash,
	tokenName string,
	remoteAddress string,
	redepositReceiver privacy.OTAReceiver,
	metaType int,
) (*BurningPRVRequest, error) {
	metadataBase := metadataCommon.MetadataBase{
		Type: metaType,
	}
	burningReq := &BurningPRVRequest{
		BurnerAddress:     burnerAddress,
		BurningAmount:     burningAmount,
		TokenID:           tokenID,
		TokenName:         tokenName,
		RemoteAddress:     remoteAddress,
		RedepositReceiver: &redepositReceiver,
	}
	burningReq.MetadataBase = metadataBase
	return burningReq, nil
}

func (bReq BurningPRVRequest) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

func (bReq BurningPRVRequest) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {
	if shardViewRetriever.GetTriggeredFeature()["pdao"] == 0 {
		return false, false, fmt.Errorf("Feature not enabled yet")
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

	if _, err := hex.DecodeString(bReq.RemoteAddress); err != nil {
		return false, false, fmt.Errorf("invalid data %s, expect hex string", bReq.RemoteAddress)
	}

	if bReq.RedepositReceiver.GetShardID() != byte(tx.GetValidationEnv().ShardID()) {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, fmt.Errorf("otaReceiver shardID is different from txShardID"))
	}

	return true, true, nil
}

func (bReq BurningPRVRequest) ValidateMetadataByItself() bool {
	return bReq.Type == metadataCommon.BurningPRVRequestMeta
}

func (bReq BurningPRVRequest) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(bReq)
	hash := common.HashH(rawBytes)
	return &hash
}

func (bReq *BurningPRVRequest) BuildReqActions(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
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

func (bReq *BurningPRVRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(bReq)
}

func (bReq *BurningPRVRequest) GetOTADeclarations() []metadataCommon.OTADeclaration {
	var result []metadataCommon.OTADeclaration
	result = append(result, metadataCommon.OTADeclaration{
		PublicKey: bReq.RedepositReceiver.PublicKey.ToBytes(), TokenID: common.PRVCoinID,
	})
	return result
}
