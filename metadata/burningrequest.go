package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/pkg/errors"
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

func (bReq *BurningRequest) ValidateTxWithBlockChain(
	txr Transaction,
	bcr BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface,
) (bool, error) {
	bridgeTokenExisted, err := db.IsBridgeTokenExisted(bReq.TokenID)
	if err != nil {
		return false, err
	}
	if !bridgeTokenExisted {
		return false, errors.New("the burning token is not existed in bridge tokens")
	}
	return true, nil
}

func (bReq *BurningRequest) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {

	// Note: the metadata was already verified with *transaction.TxCustomToken level so no need to verify with *transaction.Tx level again as *transaction.Tx is embedding property of *transaction.TxCustomToken
	if reflect.TypeOf(txr).String() == "*transaction.Tx" {
		return true, true, nil
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
	if len(bReq.TokenID) != common.HashSize {
		return false, false, errors.New("Wrong request info's token id")
	}

	if !txr.IsCoinsBurning() {
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

func (bReq *BurningRequest) ValidateMetadataByItself() bool {
	if _, err := hex.DecodeString(bReq.RemoteAddress); err != nil {
		fmt.Printf("[db] err decode RemoteAddress: %v\n", err)
		return false
	}
	return bReq.Type == BurningRequestMeta
}

func (bReq *BurningRequest) Hash() *common.Hash {
	record := bReq.MetadataBase.Hash().String()
	record += bReq.BurnerAddress.String()
	record += bReq.TokenID.String()
	record += string(bReq.BurningAmount)
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
