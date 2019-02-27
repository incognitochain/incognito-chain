package metadata

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
)

type BuyBackRequest struct {
	PaymentAddress privacy.PaymentAddress
	Amount         uint64
	TokenID        common.Hash
	MetadataBase
}

func NewBuyBackRequest(
	paymentAddress privacy.PaymentAddress,
	amount uint64,
	tokenID common.Hash,
	metaType int,
) *BuyBackRequest {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &BuyBackRequest{
		PaymentAddress: paymentAddress,
		Amount:         amount,
		TokenID:        tokenID,
		MetadataBase:   metadataBase,
	}
}

func (bbReq *BuyBackRequest) ValidateTxWithBlockChain(
	txr Transaction,
	bcr BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface,
) (bool, error) {

	// TODO: need to check vin's amt and burning address in vout
	return true, nil
}

func (bbReq *BuyBackRequest) ValidateSanityData(
	bcr BlockchainRetriever,
	txr Transaction,
) (bool, bool, error) {
	if len(bbReq.PaymentAddress.Pk) == 0 {
		return false, false, errors.New("Wrong request info's payment address")
	}
	if len(bbReq.PaymentAddress.Tk) == 0 {
		return false, false, errors.New("Wrong request info's payment address")
	}
	if bbReq.Amount == 0 {
		return false, false, errors.New("Wrong request info's amount")
	}
	if len(bbReq.TokenID) != common.HashSize {
		return false, false, errors.New("Wrong request info's token id")
	}
	return true, true, nil
}

func (bbReq *BuyBackRequest) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (bbReq *BuyBackRequest) Hash() *common.Hash {
	record := bbReq.PaymentAddress.String()
	record += string(bbReq.Amount)
	record += bbReq.TokenID.String()
	record += bbReq.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (bbReq *BuyBackRequest) BuildReqActions(tx Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	prevMeta, err := tx.GetMetadataFromVinsTx(bcr)
	if err != nil {
		return [][]string{}, err
	}

	actionContent := map[string]interface{}{
		"txReqId":        *(tx.Hash()),
		"buyBackReqMeta": bbReq,
		"prevMeta":       prevMeta,
	}

	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(BuyBackRequestMeta), actionContentBase64Str}
	return [][]string{action}, nil
}
