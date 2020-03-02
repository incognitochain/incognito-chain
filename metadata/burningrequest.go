package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
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

func (bReq BurningRequest) ValidateTxWithBlockChain(
	txr Transaction,
	bcr BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface,
) (bool, error) {
	bridgeTokenExisted, err := db.IsBridgeTokenExistedByType(bReq.TokenID, false)
	if err != nil {
		return false, err
	}
	if !bridgeTokenExisted {
		return false, errors.New("the burning token is not existed in bridge tokens")
	}
	return true, nil
}

func (bReq BurningRequest) ValidateSanityData(bcr BlockchainRetriever, txr Transaction, beaconHeight uint64) (bool, bool, error) {
	// Note: the metadata was already verified with *transaction.TxCustomToken level so no need to verify with *transaction.Tx level again as *transaction.Tx is embedding property of *transaction.TxCustomToken
	if reflect.TypeOf(txr).String() == "*transaction.Tx" {
		return true, true, nil
	}

	if _, err := hex.DecodeString(bReq.RemoteAddress); err != nil {
		return false, false, err
	}
	if bReq.Type != BurningRequestMeta {
		return false, false, errors.New("Wrong request info's meta type")
	}
	if len(bReq.BurnerAddress.Pk) == 0 {
		return false, false, errors.New("Wrong request info's burner address")
	}
	if bReq.BurningAmount == 0 {
		return false, false, errors.New("Wrong request info's burned amount")
	}
	if !txr.IsCoinsBurning(bcr, beaconHeight) {
		return false, false, errors.New("Must send coin to burning address")
	}
	if bReq.BurningAmount != txr.CalculateTxValue() {
		return false, false, errors.New("BurningAmount incorrect")
	}
	if !bytes.Equal(txr.GetSigPubKey()[:], bReq.BurnerAddress.Pk[:]) {
		return false, false, errors.New("BurnerAddress incorrect")
	}
	return true, true, nil
}

func (bReq BurningRequest) ValidateMetadataByItself() bool {
	return bReq.Type == BurningRequestMeta
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

func (bReq *BurningRequest) BuildReqActions(tx Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	actionContent := map[string]interface{}{
		"meta":          *bReq,
		"RequestedTxID": tx.Hash(),
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(BurningRequestMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (bReq *BurningRequest) CalculateSize() uint64 {
	return calculateSize(bReq)
}
