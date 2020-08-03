package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"reflect"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

// whoever can send this type of tx
type BurningRequest struct {
	BurnerAddress privacy.PaymentAddress
	BurningAmount uint64 // must be equal to vout value
	TokenID       common.Hash
	TokenName     string
	RemoteAddress string
	MetadataBase
}

func NewBurningRequest(
	burnerAddress privacy.PaymentAddress,
	burningAmount uint64,
	tokenID common.Hash,
	tokenName string,
	remoteAddress string,
	metaType int,
) (*BurningRequest, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	burningReq := &BurningRequest{
		BurnerAddress: burnerAddress,
		BurningAmount: burningAmount,
		TokenID:       tokenID,
		TokenName:     tokenName,
		RemoteAddress: remoteAddress,
	}
	burningReq.MetadataBase = metadataBase
	return burningReq, nil
}

func (bReq BurningRequest) ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	bridgeTokenExisted, err := statedb.IsBridgeTokenExistedByType(beaconViewRetriever.GetBeaconFeatureStateDB(), bReq.TokenID, false)
	if err != nil {
		return false, err
	}
	if !bridgeTokenExisted {
		return false, errors.New("the burning token is not existed in bridge tokens")
	}
	return true, nil
}

func (bReq BurningRequest) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	// Note: the metadata was already verified with *transaction.TxCustomToken level so no need to verify with *transaction.Tx level again as *transaction.Tx is embedding property of *transaction.TxCustomToken
	if reflect.TypeOf(tx).String() == "*transaction.Tx" {
		return true, true, nil
	}

	if _, err := hex.DecodeString(bReq.RemoteAddress); err != nil {
		return false, false, err
	}
	if len(bReq.BurnerAddress.Pk) == 0 {
		return false, false, errors.New("Wrong request info's burner address")
	}
	if bReq.BurningAmount == 0 {
		return false, false, errors.New("Wrong request info's burned amount")
	}

	//if !tx.IsCoinsBurning(chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight) {
	//	return false, false, errors.New("Must send coin to burning address")
	//}
	//if bReq.BurningAmount != tx.CalculateTxValue() {
	//	return false, false, errors.New("BurningAmount incorrect")
	//}

	// check burning value
	isBurning, burningAmount := tx.CalculateBurningTxValue(chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight)
	if !isBurning {
		return false, false, errors.New("Must send coin to burning address")
	}
	if burningAmount == 0 || bReq.BurningAmount != burningAmount {
		return false, false, errors.New("BurningAmount incorrect")
	}

	if !bytes.Equal(tx.GetSigPubKey()[:], bReq.BurnerAddress.Pk[:]) {
		return false, false, errors.New("BurnerAddress incorrect")
	}
	if !bytes.Equal(tx.GetTokenID()[:], bReq.TokenID[:]) {
		return false, false, errors.New("Wrong request info's token id, it should be equal to tx's token id.")
	}
	return true, true, nil
}

func (bReq BurningRequest) ValidateMetadataByItself() bool {
	return bReq.Type == BurningRequestMeta || bReq.Type == BurningForDepositToSCRequestMeta
}

func (bReq BurningRequest) Hash() *common.Hash {
	record := bReq.MetadataBase.Hash().String()
	record += bReq.BurnerAddress.String()
	record += bReq.TokenID.String()
	record += strconv.FormatUint(bReq.BurningAmount, 10)
	record += bReq.TokenName
	record += bReq.RemoteAddress

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (bReq *BurningRequest) BuildReqActions(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte) ([][]string, error) {
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

func (bReq *BurningRequest) CalculateSize() uint64 {
	return calculateSize(bReq)
}
