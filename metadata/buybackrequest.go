package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"reflect"
	"strconv"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/pkg/errors"
)

type BuyBackRequest struct {
	PaymentAddress privacy.PaymentAddress
	Amount         uint64

	TradeID []byte // To trade bond with DCB
	MetadataBase
}

func NewBuyBackRequest(
	paymentAddress privacy.PaymentAddress,
	amount uint64,
	metaType int,
) *BuyBackRequest {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &BuyBackRequest{
		PaymentAddress: paymentAddress,
		Amount:         amount,
		MetadataBase:   metadataBase,
	}
}

func (bbReq *BuyBackRequest) ValidateTxWithBlockChain(
	txr Transaction,
	bcr BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface,
) (bool, error) {
	if len(bbReq.TradeID) > 0 {
		// Validation for trading bonds
		_, buy, _, amount, _ := bcr.GetLatestTradeActivation(bbReq.TradeID)
		if amount < bbReq.Amount {
			return false, errors.Errorf("trade bond requested amount too high, %d > %d\n", bbReq.Amount, amount)
		}
		if buy {
			return false, errors.New("trade is for buying bonds, not selling")
		}
	}

	return true, nil
}

func (bbReq *BuyBackRequest) ValidateSanityData(
	bcr BlockchainRetriever,
	txr Transaction,
) (bool, bool, error) {
	// Note: the metadata was already verified with *transaction.TxCustomToken level so no need to verify with *transaction.Tx level again as *transaction.Tx is embedding property of *transaction.TxCustomToken
	if reflect.TypeOf(txr).String() == "*transaction.Tx" {
		return true, true, nil
	}

	if len(bbReq.PaymentAddress.Pk) == 0 {
		return false, false, errors.New("Wrong request info's payment address")
	}
	if len(bbReq.PaymentAddress.Tk) == 0 {
		return false, false, errors.New("Wrong request info's payment address")
	}
	if bbReq.Amount == 0 {
		return false, false, errors.New("Wrong request info's amount")
	}
	if !txr.IsCoinsBurning() {
		return false, false, errors.New("Must send bonds to burning address")
	}
	if txr.CalculateTxValue() < bbReq.Amount {
		return false, false, errors.New("Burning bond amount in Vouts should be equal metadata's amount")
	}
	if !bytes.Equal(txr.GetSigPubKey()[:], bbReq.PaymentAddress.Pk[:]) {
		return false, false, errors.New("PaymentAddress in metadata is not matched to sender address")
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
	record += bbReq.MetadataBase.Hash().String()
	record += string(bbReq.TradeID)
	hash := common.HashH([]byte(record))
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

func (bbReq *BuyBackRequest) CalculateSize() uint64 {
	return calculateSize(bbReq)
}
