package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/wallet"
)

type BuySellRequest struct {
	PaymentAddress privacy.PaymentAddress
	TokenID        common.Hash
	Amount         uint64
	BuyPrice       uint64 // in Constant unit

	TradeID []byte // To trade bond with DCB
	MetadataBase
}

func NewBuySellRequest(
	paymentAddress privacy.PaymentAddress,
	tokenID common.Hash,
	amount uint64,
	buyPrice uint64,
	metaType int,
) *BuySellRequest {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	result := &BuySellRequest{
		PaymentAddress: paymentAddress,
		Amount:         amount,
		BuyPrice:       buyPrice,
		MetadataBase:   metadataBase,
		TokenID:        tokenID,
	}
	return result
}

func (bsReq *BuySellRequest) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	// For DCB trading bods with GOV
	if len(bsReq.TradeID) > 0 {
		keyWalletDCBAccount, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
		dcbAddress := keyWalletDCBAccount.KeySet.PaymentAddress
		if !bytes.Equal(dcbAddress.Pk, bsReq.PaymentAddress.Pk) {
			return false, errors.New("buy bond request with TradeID must send assets to DCB's address")
		}
	}

	// no need to do other validation here since it'll be checked on beacon chain
	return true, nil
}

func (bsReq *BuySellRequest) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	if len(bsReq.PaymentAddress.Pk) == 0 {
		return false, false, errors.New("Wrong request info's payment address")
	}
	if len(bsReq.PaymentAddress.Tk) == 0 {
		return false, false, errors.New("Wrong request info's payment address")
	}
	if bsReq.BuyPrice == 0 {
		return false, false, errors.New("Wrong request info's buy price")
	}
	if bsReq.Amount == 0 {
		return false, false, errors.New("Wrong request info's amount")
	}
	if len(bsReq.TokenID) != common.HashSize {
		return false, false, errors.New("Wrong request info's asset type")
	}
	if txr.CalculateTxValue() < bsReq.BuyPrice*bsReq.Amount {
		return false, false, errors.New("Sending constant amount is not enough for buying bonds.")
	}
	if !txr.IsCoinsBurning() {
		return false, false, errors.New("Must send coin to burning address")
	}
	return true, true, nil
}

func (bsReq *BuySellRequest) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (bsReq *BuySellRequest) Hash() *common.Hash {
	record := bsReq.PaymentAddress.String()
	record += bsReq.TokenID.String()
	record += string(bsReq.Amount)
	record += string(bsReq.BuyPrice)
	record += bsReq.MetadataBase.Hash().String()
	record += string(bsReq.TradeID)

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (bsReq *BuySellRequest) BuildReqActions(tx Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	actionContent := map[string]interface{}{
		"txReqId": *(tx.Hash()),
		"meta":    *bsReq,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(BuyFromGOVRequestMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (bsReq *BuySellRequest) CalculateSize() uint64 {
	return calculateSize(bsReq)
}
