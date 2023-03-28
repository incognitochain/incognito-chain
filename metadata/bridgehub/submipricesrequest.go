package bridgehub

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/wallet"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

// whoever can send this type of tx
type SubmitPriceRequest struct {
	TokenIds []string `json:"TokenIds"` // list of token ids
	Prices   []uint64 `json:"Prices"`   // price list
	metadataCommon.MetadataBase
}

type SubmitPricesContentInst struct {
	TokenIds []string `json:"TokenIds"` // list of token ids
	Prices   []uint64 `json:"Prices"`   // price list
	TxReqID  string   `json:"TxReqID"`
}

func NewSubmitPriceRequest(tokenIds []string, prices []uint64) (*SubmitPriceRequest, error) {
	metadataBase := metadataCommon.MetadataBase{
		Type: metadataCommon.BridgeHubSubmitPrices,
	}
	submitPriceReq := &SubmitPriceRequest{
		TokenIds: tokenIds,
		Prices:   prices,
	}
	submitPriceReq.MetadataBase = metadataBase
	return submitPriceReq, nil
}

func (bReq SubmitPriceRequest) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

func (bReq SubmitPriceRequest) ValidateSanityData(
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

	keyWallet, err := wallet.Base58CheckDeserialize(config.Param().BridgeAggParam.AdminAddress)
	if err != nil {
		return false, false, fmt.Errorf("[Bridge Hub] Can not parse protocol fund address: %v", err)
	}

	incAddr := keyWallet.KeySet.PaymentAddress
	if len(incAddr.Pk) == 0 {
		return false, false, fmt.Errorf("[Bridge Hub] Requester incognito address is invalid")
	}

	// check sender must be bridgehub oracle
	if ok, err := bReq.VerifyMetadataSignature(incAddr.Pk, tx); !ok || err != nil {
		return false, false, fmt.Errorf("[Bridge Hub] CheckAuthorizedSender fail")
	}

	return true, true, nil
}

func (bReq SubmitPriceRequest) ValidateMetadataByItself() bool {
	return bReq.Type == metadataCommon.BridgeHubSubmitPrices
}

func (bReq SubmitPriceRequest) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(&bReq)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (bReq *SubmitPriceRequest) BuildReqActions(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
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

func (bReq *SubmitPriceRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(bReq)
}
